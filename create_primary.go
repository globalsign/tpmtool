package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

// createPrimary creates a primary object.
func createPrimary() {
	ensureAllPassed(fCreatePrimarySet, templateFlagName, persistentFlagName)

	// Read template.
	data, err := ioutil.ReadFile(*fCreatePrimaryTemplate)
	if err != nil {
		log.Fatalf("failed to read template: %v", err)
	}

	var tmpl pgtpm.PublicTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		log.Fatalf("failed to unmarshal template: %v", err)
	}

	// Create primary object.
	t := getTPM(*fCreatePrimaryTPM)
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
		log.Fatalf("failed to create primary object: %v", err)
	}

	// Make primary object persistent.
	err = tpm2.EvictControl(t, *fCreatePrimaryOwnerPassword, tpm2.HandleOwner,
		handle, tpmutil.Handle(fCreatePrimaryPersistent))
	if err != nil {
		tpm2.FlushContext(t, handle)

		log.Fatalf("failed to evict primary object: %v", err)
	}

	if err := tpm2.FlushContext(t, handle); err != nil {
		log.Fatalf("failed to flush primary object: %v", err)
	}
}
