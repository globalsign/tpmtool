package main

import (
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// evictObject evicts an object from persistent storage.
func evictObject() error {
	err := ensureAllPassed(fEvictSet, handleFlagName)
	if err != nil {
		return err
	}

	t, err := getTPM(*fEvictTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	handle := tpmutil.Handle(fEvictHandle)

	err = tpm2.EvictControl(t, *fEvictOwnerPassword, tpm2.HandleOwner, handle, handle)
	if err != nil {
		return fmt.Errorf("failed to evict object: %v", err)
	}

	return nil
}
