package main

import (
	"io"
	"log"

	"github.com/google/go-tpm-tools/simulator"
	"github.com/google/go-tpm/tpm2"

	"github.com/paulgriffiths/pgtpm"
)

// getTPM opens the TPM associated with the named device, or creates a
// simulated TPM with the specified seed, if it is not zero, or connects
// to a Microsoft TPM 2.0 Simulator, if one was specified.
func getTPM(name string, seed int64, mssim string) io.ReadWriteCloser {
	if mssim != "" {
		s, err := pgtpm.NewMSSimulator(mssim)
		if err != nil {
			log.Fatalf("couldn't initialize MS TPM simulator: %v", err)
		}

		return s
	}

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
