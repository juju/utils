// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hash_test

import (
	"bytes"
	"errors"

	"github.com/juju/testing"
	gc "launchpad.net/gocheck"

	"github.com/juju/utils/hash"
)

var _ = gc.Suite(&WriterSuite{})

type WriterSuite struct {
	testing.IsolationSuite
}

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(data []byte) (int, error) {
	return 0, w.err
}

type fakeHasher struct {
	bytes.Buffer
	sum []byte
}

func (h *fakeHasher) Sum(b []byte) []byte {
	return h.sum
}

// Not used:
func (h *fakeHasher) Reset()         {}
func (h *fakeHasher) Size() int      { return -1 }
func (h *fakeHasher) BlockSize() int { return -1 }

func (s *WriterSuite) TestHashingWriterWriteEmpty(c *gc.C) {
	var buf bytes.Buffer
	hasher := fakeHasher{}
	w := hash.NewHashingWriter(&buf, &hasher)
	n, err := w.Write(nil)

	c.Check(err, gc.IsNil)
	c.Check(n, gc.Equals, 0)
	c.Check(buf.String(), gc.Equals, "")
	c.Check(hasher.String(), gc.Equals, "")
}

func (s *WriterSuite) TestHashingWriterWriteSmall(c *gc.C) {
	var buf bytes.Buffer
	hasher := fakeHasher{}
	w := hash.NewHashingWriter(&buf, &hasher)
	n, err := w.Write([]byte("spam"))

	c.Check(err, gc.IsNil)
	c.Check(n, gc.Equals, 4)
	c.Check(buf.String(), gc.Equals, "spam")
	c.Check(hasher.String(), gc.Equals, "spam")
}

func (s *WriterSuite) TestHashingWriterWriteFileError(c *gc.C) {
	file := errorWriter{err: errors.New("failed!")}
	hasher := fakeHasher{}
	w := hash.NewHashingWriter(&file, &hasher)
	_, err := w.Write([]byte("spam"))

	c.Check(err, gc.ErrorMatches, "failed!")
}

func (s *WriterSuite) TestHashingWriterSum(c *gc.C) {
	var buf bytes.Buffer
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&buf, &hasher)
	b64hash := string(w.Sum())

	c.Check(b64hash, gc.Equals, "spam")
}

func (s *WriterSuite) TestHashingWriterBase64Sum(c *gc.C) {
	var buf bytes.Buffer
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&buf, &hasher)
	b64hash := w.Base64Sum()

	c.Check(b64hash, gc.Equals, "c3BhbQ==")
}

func (s *WriterSuite) TestHashingWriterHexSum(c *gc.C) {
	var buf bytes.Buffer
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&buf, &hasher)
	rawhash := w.HexSum()

	c.Check(rawhash, gc.Equals, "7370616d")
}
