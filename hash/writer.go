// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hash

import (
	"encoding/base64"
	"fmt"
	"hash"
	"io"
)

type HashingWriter struct {
	file   io.Writer
	hasher hash.Hash
	multiw io.Writer
}

func NewHashingWriter(file io.Writer, hasher hash.Hash) *HashingWriter {
	writer := HashingWriter{
		file:   file,
		hasher: hasher,
	}
	return &writer
}

func (h *HashingWriter) Write(data []byte) (int, error) {
	if h.multiw == nil {
		h.multiw = io.MultiWriter(h.file, h.hasher)
	}
	return h.multiw.Write(data)
}

func (h *HashingWriter) Hash() string {
	raw := h.hasher.Sum(nil)
	return base64.StdEncoding.EncodeToString(raw)
}

func (h *HashingWriter) RawHash() string {
	raw := h.hasher.Sum(nil)
	return fmt.Sprintf("%x", raw)
}
