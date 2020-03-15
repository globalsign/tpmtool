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

	data, err := ioutil.ReadFile(*fCreatePrimaryTemplate)
	if err != nil {
		log.Fatalf("failed to read template: %v", err)
	}

	var tmpl pgtpm.PublicTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		log.Fatalf("failed to unmarshal template: %v", err)
	}

	t := getTPM(*fCreatePrimaryTPM)
	defer t.Close()

	handle, _, err := tpm2.CreatePrimary(t, tpm2.HandleOwner, tpm2.PCRSelection{},
		*fCreatePrimaryOwnerPassword, *fCreatePrimaryPassword, tmpl.ToPublic())
	if err != nil {
		log.Fatalf("failed to create primary key: %v", err)
	}

	err = tpm2.EvictControl(t, *fCreatePrimaryOwnerPassword, tpm2.HandleOwner,
		handle, tpmutil.Handle(fCreatePrimaryPersistent))
	if err != nil {
		tpm2.FlushContext(t, handle)

		log.Fatalf("failed to evict object: %v", err)
	}

	if err := tpm2.FlushContext(t, handle); err != nil {
		log.Fatalf("failed to flush object: %v", err)
	}
}
