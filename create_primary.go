package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

// createPrimary creates a primary object.
func createPrimary() (err error) {
	err = ensureAllPassed(fCreatePrimarySet, templateFlagName, persistentFlagName)
	if err != nil {
		return err
	}

	// Read template.
	data, err := ioutil.ReadFile(*fCreatePrimaryTemplate)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	var tmpl pgtpm.PublicTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("failed to unmarshal template: %v", err)
	}

	// Create primary object.
	t, err := getTPM(*fCreatePrimaryTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	owner := tpm2.HandleOwner
	if *fCreatePrimaryEndorsement {
		owner = tpm2.HandleEndorsement
	} else if *fCreatePrimaryPlatform {
		owner = tpm2.HandlePlatform
	}

	handle, _, err := tpm2.CreatePrimary(t, owner, tpm2.PCRSelection{},
		*fCreatePrimaryOwnerPassword, *fCreatePrimaryPassword, tmpl.ToPublic())
	if err != nil {
		return fmt.Errorf("failed to create primary object: %v", err)
	}
	defer func() {
		if ferr := tpm2.FlushContext(t, handle); ferr != nil {
			if err == nil {
				err = fmt.Errorf("failed to flush primary object: %v", ferr)
			} else {
				log.Printf("failed to flush primary object: %v", ferr)
			}
		}
	}()

	// Make primary object persistent.
	err = tpm2.EvictControl(t, *fCreatePrimaryOwnerPassword, tpm2.HandleOwner,
		handle, tpmutil.Handle(fCreatePrimaryPersistent))
	if err != nil {
		return fmt.Errorf("failed to evict primary object: %v", err)
	}

	return nil
}
