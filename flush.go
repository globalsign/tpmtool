package main

import (
	"log"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// flushContext flushes a transient object from the TPM.
func flushContext() {
	ensureAllPassed(fFlushSet, handleFlagName)

	t := getTPM(*fFlushTPM)
	defer t.Close()

	handle := tpmutil.Handle(fFlushHandle)

	err := tpm2.FlushContext(t, handle)
	if err != nil {
		log.Fatalf("failed to flush object: %v", err)
	}
}
