package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// makeCred makes an activation credential.
func makeCred() {
	ensureAllPassed(fMakeCredSet, handleFlagName, publicAreaFlagName,
		credOutFlagName, secretOutFlagName)

	// Read public area and extract name.
	data, err := ioutil.ReadFile(*fMakeCredPublicArea)
	if err != nil {
		log.Fatalf("failed to read public area: %v", err)
	}

	pub, err := tpm2.DecodePublic(data)
	if err != nil {
		log.Fatalf("failed to decode public area: %v", err)
	}

	name, err := pub.Name()
	if err != nil {
		log.Fatalf("failed to compute name from public area: %v", err)
	}

	if name.Digest == nil {
		log.Fatalf("failed to compute name digest from public area")
	}

	// Read credential value to be encrypted.
	var f *os.File

	if *fMakeCredIn == "" {
		f = os.Stdout
	} else {
		f, err = os.Open(*fMakeCredIn)
		if err != nil {
			log.Fatalf("failed to open credential file: %v", err)
		}
		defer f.Close()
	}

	cred, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("failed to read credential: %v", err)
	}

	// Make the credential blob and encrypted secret.
	t := getTPM(*fMakeCredTPM)
	defer t.Close()

	credBlob, secret, err := tpm2.MakeCredential(t,
		tpmutil.Handle(fMakeCredHandle), cred, []byte(name.Digest.Value))
	if err != nil {
		log.Fatalf("failed to make credential: %v", err)
	}

	// Output the credential blob and encrypted secret.
	credFile, err := os.Create(*fMakeCredCredOut)
	if err != nil {
		log.Fatalf("failed to create credential output file: %v", err)
	}
	defer credFile.Close()

	if _, err := credFile.Write(credBlob); err != nil {
		log.Fatalf("failed to write credential blob: %v", err)
	}

	secretFile, err := os.Create(*fMakeCredSecretOut)
	if err != nil {
		log.Fatalf("failed to create encrypted secret output file: %v", err)
	}
	defer secretFile.Close()

	if _, err := secretFile.Write(secret); err != nil {
		log.Fatalf("failed to write encrypted secret: %v", err)
	}
}
