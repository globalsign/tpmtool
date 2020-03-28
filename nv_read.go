package main

import (
	"fmt"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// nvRead reads the value from an NV Index.
func nvRead() error {
	err := ensureAllPassed(fNVReadSet, handleFlagName)
	if err != nil {
		return err
	}

	t, err := getTPM(*fNVReadTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	data, err := tpm2.NVReadEx(t, tpmutil.Handle(fNVReadHandle), tpm2.HandleOwner, *fNVReadPassword, 0)
	if err != nil {
		return fmt.Errorf("failed to read from NV index: %v", err)
	}

	var f *os.File
	if *fNVReadOut != "" {
		f, err = os.Create(*fNVReadOut)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
	} else {
		f = os.Stdout
	}

	f.Write(data)

	return nil
}
