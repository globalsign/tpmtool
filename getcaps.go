package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// outputCaps outputs selected TPM capabilities.
func outputCaps() {
	t := getTPM(*fCapsDevice, *fCapsSeed)
	defer t.Close()

	for _, c := range []struct {
		process bool
		id      uint32
		cmdFunc func(io.ReadWriteCloser)
	}{
		{*fCapsAlgs, TPM2_CAP_ALGS, outputCapsAlgorithms},
		{*fCapsHandles, TPM2_CAP_HANDLES, outputCapsHandles},
	} {
		if c.process || *fCapsAll {
			fmt.Printf("%s:\n", capToString[c.id])
			c.cmdFunc(t)
			fmt.Println()
		}
	}
}

// outputCapsAlgorithms outputs the algorithms supported by the TPM.
func outputCapsAlgorithms(t io.ReadWriteCloser) {
	vals, _, err := tpm2.GetCapability(t, tpm2.CapabilityAlgs, 1000, uint32(0))
	if err != nil {
		log.Fatalf("failed to get TPM algorithms: %v", err)
	}

	for _, val := range vals {
		ad := val.(tpm2.AlgorithmDescription)

		desc, ok := algToString[ad.ID]
		if !ok {
			desc = "<unknown>"
		}

		var props []string
		for p, d := range algPropToString {
			if ad.Attributes&p != 0 {
				props = append(props, strings.TrimPrefix(d, "TPMA_ALGORITHM_"))
			}
		}

		fmt.Printf("  %-*s %s\n", 24, desc, strings.Join(props, " | "))
	}
}

// outputCapsHandles outputs the handles currently active in the TPM.
func outputCapsHandles(t io.ReadWriteCloser) {
	for _, next := range []uint32{
		TPM2_HT_PCR,
		TPM2_HT_NV_INDEX,
		TPM2_HT_HMAC_SESSION,
		TPM2_HT_POLICY_SESSION,
		TPM2_HT_PERMANENT,
		TPM2_HT_TRANSIENT,
		TPM2_HT_PERSISTENT,
	} {
		vals, _, err := tpm2.GetCapability(t, tpm2.CapabilityHandles, 1000, next<<24)
		if err != nil {
			log.Fatalf("failed to get TPM handles: %v", err)
		}

		for _, val := range vals {
			fmt.Printf("  0x%08x (%s)\n", val.(tpmutil.Handle), handleTypeToString[next])
		}
	}
}
