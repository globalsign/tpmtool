package main

import (
	"log"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// evictObject evicts an object from persistent storage.
func evictObject() {
	ensureAllPassed(fEvictSet, handleFlagName)

	t := getTPM(*fEvictTPM)
	defer t.Close()

	handle := tpmutil.Handle(fEvictHandle)

	err := tpm2.EvictControl(t, *fEvictOwnerPassword, tpm2.HandleOwner, handle, handle)
	if err != nil {
		log.Fatalf("failed to evict object: %v", err)
	}
}
