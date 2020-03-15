package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// activateCred activates a credential.
func activateCred() {
	ensureAllPassed(fActivateSet, credInFlagName, secretInFlagName,
		handleFlagName, protectorFlagName)

	// Read the credential blob and encrypted secret.
	cred, err := ioutil.ReadFile(*fActivateCredIn)
	if err != nil {
		log.Fatalf("failed to read credential blob: %v", err)
	}

	secret, err := ioutil.ReadFile(*fActivateSecretIn)
	if err != nil {
		log.Fatalf("failed to read encrypted secret: %v", err)
	}

	// Activate the credential.
	t := getTPM(*fActivateTPM)
	defer t.Close()

	cred, err = tpm2.ActivateCredential(t, tpmutil.Handle(fActivateHandle),
		tpmutil.Handle(fActivateProtector), *fActivatePassword,
		*fActivateProtectorPassword, cred, secret)
	if err != nil {
		log.Fatalf("failed to activate credential: %v", err)
	}

	// Output the credential.
	os.Stdout.Write(cred)
}
