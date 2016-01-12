// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash

import (
	"hash"
	"io"
)

// HashingWriter wraps an io.Writer, providing the checksum of all data
// written to it.  A HashingWriter may be used in place of the writer it
// wraps.
type HashingWriter struct {
	hashSum
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
	return &HashingWriter{
		hashSum: hashSum{
			raw: hasher,
		},
		wrapped: writer,
	}
}

// Write writes to both the wrapped file and the hash.
func (hw *HashingWriter) Write(data []byte) (int, error) {
	n, err := hw.wrapped.Write(data)
	if err != nil {
		return n, err
	}
	return hw.raw.Write(data[:n])
}

// HashingReader wraps an io.Reader, providing the checksum of all data
// written to it.  A HashingReader may be used in place of the reader it
// wraps.
type HashingReader struct {
	hashSum
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
	return &HashingReader{
		hashSum: hashSum{
			raw: hasher,
		},
		wrapped: reader,
	}
}

// Write writes to both the wrapped file and the hash.
func (hr *HashingReader) Read(data []byte) (int, error) {
	n, err := hr.wrapped.Read(data)
	if err != nil {
		return n, err
	}
	return hr.raw.Write(data[:n])
}
