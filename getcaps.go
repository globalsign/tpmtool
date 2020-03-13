package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

const (
	capRequestSize = 16
)

// outputCaps outputs selected TPM capabilities.
func outputCaps() {
	t := getTPM(*fCapsDevice, *fCapsSeed)
	defer t.Close()

	for _, c := range []struct {
		process bool
		id      pgtpm.Capability
		cmdFunc func(io.ReadWriteCloser)
	}{
		{*fCapsAlgs, pgtpm.TPM2_CAP_ALGS, outputCapsAlgorithms},
		{*fCapsHandles, pgtpm.TPM2_CAP_HANDLES, outputCapsHandles},
	} {
		if c.process || *fCapsAll {
			fmt.Printf("%s:\n", c.id.String())
			c.cmdFunc(t)
			fmt.Println()
		}
	}
}

// outputCapsAlgorithms outputs the algorithms supported by the TPM.
func outputCapsAlgorithms(t io.ReadWriteCloser) {
	var vals []interface{}
	var more = true
	var err error
	var next uint32 = 0

	for more {
		vals, more, err = tpm2.GetCapability(t, tpm2.CapabilityAlgs, capRequestSize, next)
		if err != nil {
			log.Fatalf("failed to get algorithms: %v", err)
		}

		for _, val := range vals {
			ad := val.(tpm2.AlgorithmDescription)
			next = uint32(pgtpm.Algorithm(ad.ID) + 1)

			var props []string
			for _, p := range []pgtpm.AlgorithmAttribute{
				pgtpm.TPMA_ALGORITHM_ASYMMETRIC,
				pgtpm.TPMA_ALGORITHM_SYMMETRIC,
				pgtpm.TPMA_ALGORITHM_HASH,
				pgtpm.TPMA_ALGORITHM_OBJECT,
				pgtpm.TPMA_ALGORITHM_SIGNING,
				pgtpm.TPMA_ALGORITHM_ENCRYPTING,
				pgtpm.TPMA_ALGORITHM_METHOD,
			} {
				if pgtpm.AlgorithmAttribute(ad.Attributes)&p != 0 {
					props = append(props, strings.TrimPrefix(p.String(), "TPMA_ALGORITHM_"))
				}
			}

			fmt.Printf("  %-*s %s\n", 24, pgtpm.Algorithm(ad.ID).String(), strings.Join(props, " | "))
		}
	}
}

// outputCapsHandles outputs the handles currently active in the TPM.
func outputCapsHandles(t io.ReadWriteCloser) {
	for _, ht := range []pgtpm.HandleType{
		pgtpm.TPM2_HT_PCR,
		pgtpm.TPM2_HT_NV_INDEX,
		pgtpm.TPM2_HT_HMAC_SESSION,
		pgtpm.TPM2_HT_POLICY_SESSION,
		pgtpm.TPM2_HT_PERMANENT,
		pgtpm.TPM2_HT_TRANSIENT,
		pgtpm.TPM2_HT_PERSISTENT,
	} {
		var vals []interface{}
		var more = true
		var err error
		var next = uint32(ht.First())

		for more {
			vals, more, err = tpm2.GetCapability(t, tpm2.CapabilityHandles, capRequestSize, next)
			if err != nil {
				log.Fatalf("failed to get handles: %v", err)
			}

			for _, val := range vals {
				fmt.Printf("  0x%08X  %s\n", val.(tpmutil.Handle), ht.String())
				next = uint32(val.(tpmutil.Handle)) + 1
			}
		}
	}
}
