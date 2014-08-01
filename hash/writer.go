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
// written to it.
type HashingWriter struct {
	wrapped io.Writer
	hasher  hash.Hash
}

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
