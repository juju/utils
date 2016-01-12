// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash

import (
	"encoding/base64"
	"fmt"
	"hash"
)

type hashSum struct {
	raw hash.Hash
}

// Sum returns the raw checksum.
func (hs hashSum) Sum() []byte {
	return hs.raw.Sum(nil)
}

// Base64Sum returns the base64 encoded hash.
func (hs hashSum) Base64Sum() string {
	raw := hs.raw.Sum(nil)
	return base64.StdEncoding.EncodeToString(raw)
}

// HexSum returns the hex-ified checksum.
func (hs hashSum) HexSum() string {
	raw := hs.raw.Sum(nil)
	return fmt.Sprintf("%x", raw)
}
