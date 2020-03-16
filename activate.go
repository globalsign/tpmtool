package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// activateCred activates a credential.
func activateCred() error {
	err := ensureAllPassed(fActivateSet, credInFlagName, secretInFlagName,
		handleFlagName, protectorFlagName)
	if err != nil {
		return err
	}

	// Read the credential blob and encrypted secret.
	cred, err := ioutil.ReadFile(*fActivateCredIn)
	if err != nil {
		return fmt.Errorf("failed to read credential blob: %v", err)
	}

	secret, err := ioutil.ReadFile(*fActivateSecretIn)
	if err != nil {
		return fmt.Errorf("failed to read encrypted secret: %v", err)
	}

	// Activate the credential.
	t, err := getTPM(*fActivateTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	cred, err = tpm2.ActivateCredential(t, tpmutil.Handle(fActivateHandle),
		tpmutil.Handle(fActivateProtector), *fActivatePassword,
		*fActivateProtectorPassword, cred, secret)
	if err != nil {
		return fmt.Errorf("failed to activate credential: %v", err)
	}

	// Output the credential.
	os.Stdout.Write(cred)

	return nil
}
