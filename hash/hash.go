// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash

import (
	"encoding/base64"
	"fmt"
	"hash"
	"io"
)

type hash struct {
	raw hash.Hash
}

// Sum returns the raw checksum.
func (h hash) Sum() []byte {
	return h.raw.Sum(nil)
}

// Base64Sum returns the base64 encoded hash.
func (h hash) Base64Sum() string {
	raw := h.raw.Sum(nil)
	return base64.StdEncoding.EncodeToString(raw)
}

// HexSum returns the hex-ified checksum.
func (h hash) HexSum() string {
	raw := h.raw.Sum(nil)
	return fmt.Sprintf("%x", raw)
}
