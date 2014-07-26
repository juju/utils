// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hash

import (
	"crypto/sha1"
	"io"
)

func NewSHA1Proxy(file io.Writer) *HashingWriter {
	proxy := HashingWriter{
		file:   file,
		hasher: sha1.New(),
	}
	return &proxy
}
