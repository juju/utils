// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hash

import (
	"crypto/sha1"
	"io"
)

// NewSHA1Proxy returns a HashingWriter with an underlying SHA-1 hash.
func NewSHA1Proxy(writer io.Writer) *HashingWriter {
	proxy := HashingWriter{
		wrapped: writer,
		hasher:  sha1.New(),
	}
	return &proxy
}
