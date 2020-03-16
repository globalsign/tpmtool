package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/go-tpm/tpm2"

	"github.com/paulgriffiths/pgtpm"
)

// getTPM opens the TPM associated with the named device. If the named device
// cannot be found, and the name is in the form hostname:port, an attempt will
// be made to open a connection with a Microsoft TPM 2.0 Simulator listening on
// that port.
func getTPM(name string) (io.ReadWriteCloser, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) && len(strings.Split(name, ":")) == 2 {
			s, err := pgtpm.NewMSSimulator(name)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize MS TPM simulator: %v", err)
			}

			return s, nil
		}

		return nil, fmt.Errorf("failed to locate TPM device: %v", err)
	}

	t, err := tpm2.OpenTPM(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open TPM device: %v", err)
	}

	return t, nil
}
