package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

// createObject creates an object.
func createObject() {
	ensureAllPassed(fCreateSet, templateFlagName, parentFlagName)

	// Read template.
	data, err := ioutil.ReadFile(*fCreateTemplate)
	if err != nil {
		log.Fatalf("failed to read template: %v", err)
	}

	var tmpl pgtpm.PublicTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		log.Fatalf("failed to unmarshal template: %v", err)
	}

	// Create object.
	t := getTPM(*fCreateTPM)
	defer t.Close()

	parentHandle := tpmutil.Handle(fCreateParent)

	private, public, _, _, _, err := tpm2.CreateKey(t, parentHandle, tpm2.PCRSelection{},
		*fCreateParentPassword, *fCreatePassword, tmpl.ToPublic())
	if err != nil {
		log.Fatalf("failed to create object: %v", err)
	}

	// Make object persistent, if requested.
	if fCreatePersistent != 0 {
		handle, _, err := tpm2.Load(t, parentHandle, *fCreateParentPassword, public, private)
		if err != nil {
			log.Fatalf("failed to load object: %v", err)
		}

		err = tpm2.EvictControl(t, *fCreateOwnerPassword, tpm2.HandleOwner,
			handle, tpmutil.Handle(fCreatePersistent))
		if err != nil {
			tpm2.FlushContext(t, handle)

			log.Fatalf("failed to evict object: %v", err)
		}

		if err := tpm2.FlushContext(t, handle); err != nil {
			log.Fatalf("failed to flush object: %v", err)
		}
	}

	// Output public area, if requested.
	if *fCreatePublicOut != "" {
		f, err := os.Create(*fCreatePublicOut)
		if err != nil {
			log.Fatalf("failed to create public area file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(public); err != nil {
			log.Fatalf("failed to write public area: %v", err)
		}
	}

	// Output private area, if requested.
	if *fCreatePrivateOut != "" {
		f, err := os.Create(*fCreatePrivateOut)
		if err != nil {
			log.Fatalf("failed to create private area file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(private); err != nil {
			log.Fatalf("failed to write private area: %v", err)
		}
	}
}
