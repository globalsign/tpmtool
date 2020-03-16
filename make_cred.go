package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// makeCred makes an activation credential.
func makeCred() error {
	err := ensureAllPassed(fMakeCredSet, handleFlagName, publicAreaFlagName,
		credOutFlagName, secretOutFlagName)
	if err != nil {
		return err
	}

	// Read public area and extract name.
	data, err := ioutil.ReadFile(*fMakeCredPublicArea)
	if err != nil {
		return fmt.Errorf("failed to read public area: %v", err)
	}

	pub, err := tpm2.DecodePublic(data)
	if err != nil {
		return fmt.Errorf("failed to decode public area: %v", err)
	}

	name, err := pub.Name()
	if err != nil {
		return fmt.Errorf("failed to compute name from public area: %v", err)
	}

	nameBytes, err := name.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode name: %v", err)
	}

	// Per TPM Spec Part 1 section 11.4.9.2, contextU and contextV are passed
	// to KDFa as sized buffers, but the size fields are not used in the
	// computation. Accordingly, we strip the first two bytes (the size field)
	// from the name bytes.
	if len(nameBytes) < 2 {
		return fmt.Errorf("size of name bytes is %d, expected at least 2", len(nameBytes))
	}
	nameBytes = nameBytes[2:]

	// Read credential value to be encrypted.
	var f *os.File

	if *fMakeCredIn == "" {
		f = os.Stdout
	} else {
		f, err = os.Open(*fMakeCredIn)
		if err != nil {
			return fmt.Errorf("failed to open credential file: %v", err)
		}
		defer f.Close()
	}

	cred, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read credential: %v", err)
	}

	// Make the credential blob and encrypted secret.
	t, err := getTPM(*fMakeCredTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	credBlob, secret, err := tpm2.MakeCredential(t,
		tpmutil.Handle(fMakeCredHandle), cred, nameBytes)
	if err != nil {
		return fmt.Errorf("failed to make credential: %v", err)
	}

	// Output the credential blob and encrypted secret.
	credFile, err := os.Create(*fMakeCredCredOut)
	if err != nil {
		return fmt.Errorf("failed to create credential output file: %v", err)
	}
	defer credFile.Close()

	if _, err := credFile.Write(credBlob); err != nil {
		return fmt.Errorf("failed to write credential blob: %v", err)
	}

	secretFile, err := os.Create(*fMakeCredSecretOut)
	if err != nil {
		return fmt.Errorf("failed to create encrypted secret output file: %v", err)
	}
	defer secretFile.Close()

	if _, err := secretFile.Write(secret); err != nil {
		return fmt.Errorf("failed to write encrypted secret: %v", err)
	}

	return nil
}
