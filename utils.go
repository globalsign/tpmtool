package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// hexEncodeBytes returns a string contains the hex-encoding of the provided
// slice of bytes.
func hexEncodeBytes(b []byte) string {
	dst := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(dst, b)

	return string(dst)
}

// outputBigInt outputs a big integer on multiple lines, with the label
// only on the first line.
func outputBigInt(label string, fw int, n *big.Int) {
	b := n.Bytes()
	const bytesPerLine = 16

	for i := 0; i < len(b); i += bytesPerLine {
		if i != 0 {
			label = ""
		}

		var end = i + bytesPerLine
		if end > len(b) {
			end = len(b)
		}

		fmt.Printf("%-*s: %s\n", fw, label, hexEncodeBytes(b[i:end]))
	}
}
