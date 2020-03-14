package main

import (
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

// readPublic reads a TPM object's public area.
func readPublic() {
	var pub tpm2.Public
	var nameAlg pgtpm.Algorithm
	var qnameAlg pgtpm.Algorithm
	var nameHash []byte
	var qnameHash []byte

	ensureExactlyOnePassed(fReadPublicSet, inFlagName, handleFlagName)

	// Read a public area from a file, or from a TPM.
	if *fReadPublicIn == "" {
		if fReadPublicHandle == 0 {
			log.Fatalf("no handle specified")
		}

		var handle = pgtpm.Handle(fReadPublicHandle)

		t := getTPM(*fReadPublicTPM)
		defer t.Close()

		var err error
		pub, nameHash, qnameHash, err = tpm2.ReadPublic(t, tpmutil.Handle(handle))
		if err != nil {
			log.Fatalf("failed to read public area: %v", err)
		}

		nameAlg = pgtpm.Algorithm(binary.BigEndian.Uint16(nameHash))
		nameHash = nameHash[2:]

		qnameAlg = pgtpm.Algorithm(binary.BigEndian.Uint16(qnameHash))
		qnameHash = qnameHash[2:]
	} else {
		data, err := ioutil.ReadFile(*fReadPublicIn)
		if err != nil {
			log.Fatalf("failed to read public area: %v", err)
		}

		pub, err = tpm2.DecodePublic(data)
		if err != nil {
			log.Fatalf("failed to decode public area: %v", err)
		}

		name, err := pub.Name()
		if err != nil {
			log.Fatalf("failed to get name from public area: %v", err)
		}

		if name.Digest != nil {
			nameHash = name.Digest.Value
			nameAlg = pgtpm.Algorithm(name.Digest.Alg)
		}
	}

	// Write the raw public area, if requested.
	if *fReadPublicOut != "" || (!*fReadPublicText && !*fReadPublicPubOut) {
		var f *os.File
		var err error

		if *fReadPublicOut != "" {
			f, err = os.Create(*fReadPublicOut)
			if err != nil {
				log.Fatalf("failed to create output file: %v", err)
			}
			defer f.Close()
		} else {
			f = os.Stdout
		}

		data, err := pub.Encode()
		if err != nil {
			log.Fatalf("failed to encode public area: %v", err)
		}

		_, err = f.Write(data)
		if err != nil {
			log.Fatalf("failed to write public area: %v", err)
		}
	}

	// Write the public area as text, if requested.
	if *fReadPublicText {
		const fw = 20

		fmt.Printf("%-*s: %s\n", fw, "Type", pgtpm.Algorithm(pub.Type).String())
		fmt.Printf("%-*s: %s\n", fw, "Name algorithm", pgtpm.Algorithm(pub.NameAlg).String())

		if nameHash != nil {
			fmt.Printf("%-*s: %s (%s)\n", fw, "Name", hexEncodeBytes(nameHash[2:]), nameAlg.String())
		}

		if qnameHash != nil {
			fmt.Printf("%-*s: %s (%s)\n", fw, "Qualified nameHash", hexEncodeBytes(qnameHash[2:]), qnameAlg.String())
		}

		if pub.Attributes != 0 {
			var first = true

			for _, a := range []pgtpm.ObjectAttribute{
				pgtpm.TPMA_OBJECT_FIXEDTPM,
				pgtpm.TPMA_OBJECT_STCLEAR,
				pgtpm.TPMA_OBJECT_FIXEDPARENT,
				pgtpm.TPMA_OBJECT_SENSITIVEDATAORIGIN,
				pgtpm.TPMA_OBJECT_USERWITHAUTH,
				pgtpm.TPMA_OBJECT_ADMINWITHPOLICY,
				pgtpm.TPMA_OBJECT_NODA,
				pgtpm.TPMA_OBJECT_ENCRYPTEDDUPLICATION,
				pgtpm.TPMA_OBJECT_RESTRICTED,
				pgtpm.TPMA_OBJECT_DECRYPT,
				pgtpm.TPMA_OBJECT_SIGN_ENCRYPT,
			} {
				if pgtpm.ObjectAttribute(pub.Attributes)&a != 0 {
					var label string
					if first {
						label = "Attributes"
						first = false
					}
					fmt.Printf("%-*s: %s\n", fw, label, a.String())
				}
			}
		}

		if len(pub.AuthPolicy) > 0 {
			fmt.Printf("%-*s: %s\n", fw, "Auth policy", hexEncodeBytes([]byte(pub.AuthPolicy)))
		}

		switch {
		case pub.RSAParameters != nil:
			if pub.RSAParameters.Symmetric != nil {
				fmt.Printf("%-*s: %s\n", fw, "Symmetric algorithm", pgtpm.Algorithm(pub.RSAParameters.Symmetric.Alg).String())
				fmt.Printf("%-*s: %d\n", fw, "Symmetric key bits", pub.RSAParameters.Symmetric.KeyBits)
				fmt.Printf("%-*s: %s\n", fw, "Symmetric mode", pgtpm.Algorithm(pub.RSAParameters.Symmetric.Mode).String())
			}

			if pub.RSAParameters.Sign != nil {
				fmt.Printf("%-*s: %s\n", fw, "Signature algorithm", pgtpm.Algorithm(pub.RSAParameters.Sign.Alg).String())
				fmt.Printf("%-*s: %s\n", fw, "Signature hash", pgtpm.Algorithm(pub.RSAParameters.Sign.Hash).String())
			}

			fmt.Printf("%-*s: %d\n", fw, "Key bits", pub.RSAParameters.KeyBits)

			var e = pub.RSAParameters.Exponent()
			fmt.Printf("%-*s: %d (0x%x)\n", fw, "Exponent", e, e)

			outputBigInt("Modulus", fw, pub.RSAParameters.Modulus())
		}
	}

	// Write the PEM-encoded public key, if requested.
	if *fReadPublicPubOut {
		key, err := pub.Key()
		if err != nil {
			log.Fatalf("failed to get public key from public area: %v", err)
		}

		der, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			log.Fatalf("failed to marshal public key: %v", err)
		}

		fmt.Printf("%s", pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: der,
		}))
	}
}
