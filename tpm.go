package main

import (
	"io"
	"log"

	"github.com/google/go-tpm-tools/simulator"
	"github.com/google/go-tpm/tpm2"
)

// getTPM opens the TPM associated with the named device, or creates a
// simulated TPM with the specified seed, if it is not zero.
func getTPM(name string, seed int64) io.ReadWriteCloser {
	if seed != 0 {
		t, err := simulator.GetWithFixedSeedInsecure(seed)
		if err != nil {
			log.Fatalf("couldn't get simulated TPM: %v", err)
		}

		return t
	}

	t, err := tpm2.OpenTPM(name)
	if err != nil {
		log.Fatalf("couldn't get TPM device: %v", err)
	}

	return t
}
