package main

import (
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// flushContext flushes a transient object from the TPM.
func flushContext() error {
	err := ensureAllPassed(fFlushSet, handleFlagName)
	if err != nil {
		return err
	}

	t, err := getTPM(*fFlushTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	handle := tpmutil.Handle(fFlushHandle)

	err = tpm2.FlushContext(t, handle)
	if err != nil {
		return fmt.Errorf("failed to flush object: %v", err)
	}

	return nil
}
