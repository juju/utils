// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

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
	hash
	wrapped io.Writer
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
		hash: hash{
			raw: hasher,
		},
		wrapped: writer,
	}
	return &hashingWriter
}

// Write writes to both the wrapped file and the hash.
func (h *HashingWriter) Write(data []byte) (int, error) {
	n, err := h.wrapped.Write(data)
	if err != nil {
		return n, err
	}
	return h.raw.Write(data[:n])
}

// HashingReader wraps an io.Reader, providing the checksum of all data
// written to it.  A HashingReader may be used in place of the reader it
// wraps.
type HashingReader struct {
	hash
	wrapped io.Reader
}

// NewHashingReader returns a new HashingReader that wraps the provided
// reader and the hasher.
//
// Example:
//   hw := NewHashingReader(w, sha1.New())
//   io.Copy(writer, hw)
//   hash := hw.Base64Sum()
func NewHashingReader(reader io.Reader, hasher hash.Hash) *HashingReader {
	hashingReader := HashingReader{
		hash: hash{
			raw: hasher,
		},
		wrapped: reader,
	}
	return &hashingReader
}

// Write writes to both the wrapped file and the hash.
func (h *HashingReader) Read(data []byte) (int, error) {
	n, err := h.wrapped.Read(data)
	if err != nil {
		return n, err
	}
	return h.raw.Write(data[:n])
}
