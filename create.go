package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

// createObject creates an object.
func createObject() (err error) {
	err = ensureAllPassed(fCreateSet, templateFlagName, parentFlagName)
	if err != nil {
		return err
	}

	// Read template.
	data, err := ioutil.ReadFile(*fCreateTemplate)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	var tmpl pgtpm.PublicTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("failed to unmarshal template: %v", err)
	}

	// Create object.
	t, err := getTPM(*fCreateTPM)
	if err != nil {
		return err
	}
	defer t.Close()

	parentHandle := tpmutil.Handle(fCreateParent)

	private, public, _, _, _, err := tpm2.CreateKey(t, parentHandle, tpm2.PCRSelection{},
		*fCreateParentPassword, *fCreatePassword, tmpl.ToPublic())
	if err != nil {
		return fmt.Errorf("failed to create object: %v", err)
	}

	// Make object persistent, if requested.
	if fCreatePersistent != 0 {
		handle, _, err := tpm2.Load(t, parentHandle, *fCreateParentPassword, public, private)
		if err != nil {
			return fmt.Errorf("failed to load object: %v", err)
		}
		defer func() {
			if ferr := tpm2.FlushContext(t, handle); ferr != nil {
				if err == nil {
					err = fmt.Errorf("failed to flush object: %v", ferr)
				} else {
					log.Printf("failed to flush object: %v", ferr)
				}
			}
		}()

		err = tpm2.EvictControl(t, *fCreateOwnerPassword, tpm2.HandleOwner,
			handle, tpmutil.Handle(fCreatePersistent))
		if err != nil {
			return fmt.Errorf("failed to evict object: %v", err)
		}
	}

	// Output public area, if requested.
	if *fCreatePublicOut != "" {
		f, err := os.Create(*fCreatePublicOut)
		if err != nil {
			return fmt.Errorf("failed to create public area file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(public); err != nil {
			return fmt.Errorf("failed to write public area: %v", err)
		}
	}

	// Output private area, if requested.
	if *fCreatePrivateOut != "" {
		f, err := os.Create(*fCreatePrivateOut)
		if err != nil {
			return fmt.Errorf("failed to create private area file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(private); err != nil {
			return fmt.Errorf("failed to write private area: %v", err)
		}
	}

	return nil
}
