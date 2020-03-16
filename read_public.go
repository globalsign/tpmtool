package main

import (
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/paulgriffiths/pgtpm"
)

// readPublic reads a TPM object's public area.
func readPublic() error {
	var pub tpm2.Public
	var nameAlg pgtpm.Algorithm
	var qnameAlg pgtpm.Algorithm
	var nameHash []byte
	var qnameHash []byte

	err := ensureExactlyOnePassed(fReadPublicSet, inFlagName, handleFlagName)
	if err != nil {
		return err
	}

	// Read a public area from a file, or from a TPM.
	if *fReadPublicIn == "" {
		var handle = pgtpm.Handle(fReadPublicHandle)

		t, err := getTPM(*fReadPublicTPM)
		if err != nil {
			return err
		}
		defer t.Close()

		pub, nameHash, qnameHash, err = tpm2.ReadPublic(t, tpmutil.Handle(handle))
		if err != nil {
			return fmt.Errorf("failed to read public area: %v", err)
		}

		nameAlg = pgtpm.Algorithm(binary.BigEndian.Uint16(nameHash))
		nameHash = nameHash[2:]

		qnameAlg = pgtpm.Algorithm(binary.BigEndian.Uint16(qnameHash))
		qnameHash = qnameHash[2:]
	} else {
		data, err := ioutil.ReadFile(*fReadPublicIn)
		if err != nil {
			return fmt.Errorf("failed to read public area: %v", err)
		}

		pub, err = tpm2.DecodePublic(data)
		if err != nil {
			return fmt.Errorf("failed to decode public area: %v", err)
		}

		name, err := pub.Name()
		if err != nil {
			return fmt.Errorf("failed to get name from public area: %v", err)
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
				return fmt.Errorf("failed to create output file: %v", err)
			}
			defer f.Close()
		} else {
			f = os.Stdout
		}

		data, err := pub.Encode()
		if err != nil {
			return fmt.Errorf("failed to encode public area: %v", err)
		}

		_, err = f.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write public area: %v", err)
		}
	}

	// Write the public area as text, if requested.
	if *fReadPublicText {
		const fw = 21

		fmt.Printf("%-*s: %s\n", fw, "Type", pgtpm.Algorithm(pub.Type).String())
		fmt.Printf("%-*s: %s\n", fw, "Name algorithm", pgtpm.Algorithm(pub.NameAlg).String())

		if nameHash != nil {
			fmt.Printf("%-*s: %s (%s)\n", fw, "Name", hexEncodeBytes(nameHash[2:]), nameAlg.String())
		}

		if qnameHash != nil {
			fmt.Printf("%-*s: %s (%s)\n", fw, "Qualified name", hexEncodeBytes(qnameHash[2:]), qnameAlg.String())
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
			param := pub.RSAParameters

			if sym := param.Symmetric; sym != nil {
				fmt.Printf("%-*s: %s\n", fw, "Symmetric algorithm", pgtpm.Algorithm(sym.Alg).String())
				fmt.Printf("%-*s: %d\n", fw, "Symmetric key bits", sym.KeyBits)
				fmt.Printf("%-*s: %s\n", fw, "Symmetric mode", pgtpm.Algorithm(sym.Mode).String())
			}

			if sig := param.Sign; sig != nil {
				fmt.Printf("%-*s: %s\n", fw, "Signature algorithm", pgtpm.Algorithm(sig.Alg).String())
				fmt.Printf("%-*s: %s\n", fw, "Signature hash", pgtpm.Algorithm(sig.Hash).String())
			}

			fmt.Printf("%-*s: %d\n", fw, "Key bits", param.KeyBits)

			var e = param.Exponent()
			fmt.Printf("%-*s: %d (0x%x)\n", fw, "Exponent", e, e)

			outputBigInt("Modulus", fw, param.Modulus())

		case pub.ECCParameters != nil:
			param := pub.ECCParameters

			if sym := param.Symmetric; sym != nil {
				fmt.Printf("%-*s: %s\n", fw, "Symmetric algorithm", pgtpm.Algorithm(sym.Alg).String())
				fmt.Printf("%-*s: %d\n", fw, "Symmetric key bits", sym.KeyBits)
				fmt.Printf("%-*s: %s\n", fw, "Symmetric mode", pgtpm.Algorithm(sym.Mode).String())
			}

			if sig := param.Sign; sig != nil {
				fmt.Printf("%-*s: %s\n", fw, "Signature algorithm", pgtpm.Algorithm(sig.Alg).String())
				fmt.Printf("%-*s: %s\n", fw, "Signature hash", pgtpm.Algorithm(sig.Hash).String())

				if param.Sign.Alg.UsesCount() {
					fmt.Printf("%-*s: %d\n", fw, "Signature count", pgtpm.Algorithm(sig.Count))
				}
			}

			fmt.Printf("%-*s: %s\n", fw, "Elliptic curve", pgtpm.EllipticCurve(param.CurveID).String())

			if kdf := param.KDF; kdf != nil {
				fmt.Printf("%-*s: %s\n", fw, "KDF scheme algorithm", pgtpm.Algorithm(kdf.Alg).String())
				fmt.Printf("%-*s: %s\n", fw, "KDF scheme hash", pgtpm.Algorithm(kdf.Hash).String())
			}

			outputBigInt("X point", fw, param.Point.X())
			outputBigInt("Y point", fw, param.Point.Y())

		case pub.SymCipherParameters != nil:
			param := pub.SymCipherParameters

			if sym := param.Symmetric; sym != nil {
				fmt.Printf("%-*s: %s\n", fw, "Symmetric algorithm", pgtpm.Algorithm(sym.Alg).String())
				fmt.Printf("%-*s: %d\n", fw, "Symmetric key bits", sym.KeyBits)
				fmt.Printf("%-*s: %s\n", fw, "Symmetric mode", pgtpm.Algorithm(sym.Mode).String())
			}

			if uniq := param.Unique; len(uniq) > 0 {
				fmt.Printf("%-*s: %s\n", fw, "Unique", hexEncodeBytes(uniq))
			}

		case pub.KeyedHashParameters != nil:
			param := pub.KeyedHashParameters

			fmt.Printf("%-*s: %s\n", fw, "Keyed hash algorithm", pgtpm.Algorithm(param.Alg).String())
			fmt.Printf("%-*s: %s\n", fw, "Keyed hash hash", pgtpm.Algorithm(param.Hash).String())
			fmt.Printf("%-*s: %s\n", fw, "Keyed hash KDF", pgtpm.Algorithm(param.KDF).String())

			if uniq := param.Unique; len(uniq) > 0 {
				fmt.Printf("%-*s: %s\n", fw, "Unique", hexEncodeBytes(uniq))
			}
		}
	}

	// Write the PEM-encoded public key, if requested.
	if *fReadPublicPubOut {
		key, err := pub.Key()
		if err != nil {
			return fmt.Errorf("failed to get public key from public area: %v", err)
		}

		der, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			return fmt.Errorf("failed to marshal public key: %v", err)
		}

		fmt.Printf("%s", pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: der,
		}))
	}

	return nil
}
