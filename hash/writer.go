// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hash

import (
	"encoding/base64"
	"fmt"
	"hash"
	"io"
)

// HashingWriter wraps an io.Writer, providing the checksum of all data
// written to it.  A HashingWriter may be used in place of the writer it
// wraps.
type HashingWriter struct {
	wrapped io.Writer
	hasher  hash.Hash
}

// NewHashingWriter returns a new HashingWriter that wraps the provided
// writer and the hasher.
//
// Example:
//   hw := NewHashingWriter(w, sha1.New())
//   io.Copy(hw, reader)
//   hash := hw.Base64Sum()
func NewHashingWriter(writer io.Writer, hasher hash.Hash) *HashingWriter {
	hashingWriter := HashingWriter{
		wrapped: writer,
		hasher:  hasher,
	}
	return &hashingWriter
}

// Write writes to both the wrapped file and the hash.
func (h *HashingWriter) Write(data []byte) (int, error) {
	h.hasher.Write(data)
	return h.wrapped.Write(data)
}

// Sum returns the raw checksum.
func (h *HashingWriter) Sum() []byte {
	return h.hasher.Sum(nil)
}

// Base64Sum returns the base64 encoded hash.
func (h *HashingWriter) Base64Sum() string {
	raw := h.hasher.Sum(nil)
	return base64.StdEncoding.EncodeToString(raw)
}

// HexSum returns the hex-ified checksum.
func (h *HashingWriter) HexSum() string {
	raw := h.hasher.Sum(nil)
	return fmt.Sprintf("%x", raw)
}
