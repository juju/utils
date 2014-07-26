// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hash_test

import (
	"errors"

	"github.com/juju/testing"
	gc "launchpad.net/gocheck"

	"github.com/juju/utils/hash"
)

var _ = gc.Suite(&WriterSuite{})

type WriterSuite struct {
	testing.IsolationSuite
}

type fakeFile struct {
	written []byte
	err     error
}

func (f *fakeFile) Write(data []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.written == nil {
		f.written = make([]byte, 0)
	}
	f.written = append(f.written, data...)
	return len(data), nil
}

type fakeHasher struct {
	fakeFile
	sum []byte
}

func (h *fakeHasher) Sum(b []byte) []byte {
	return h.sum
}

// Not used:
func (h *fakeHasher) Reset()         {}
func (h *fakeHasher) Size() int      { return -1 }
func (h *fakeHasher) BlockSize() int { return -1 }

//---------------------------
// tests

func (s *WriterSuite) TestHashingWriterWriteEmpty(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{}
	w := hash.NewHashingWriter(&file, &hasher)
	n, err := w.Write([]byte(""))

	c.Check(err, gc.IsNil)
	c.Check(n, gc.Equals, 0)
	c.Check(string(file.written), gc.Equals, "")
	c.Check(string(hasher.written), gc.Equals, "")
}

func (s *WriterSuite) TestHashingWriterWriteSmall(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{}
	w := hash.NewHashingWriter(&file, &hasher)
	n, err := w.Write([]byte("spam"))

	c.Check(err, gc.IsNil)
	c.Check(n, gc.Equals, 4)
	c.Check(string(file.written), gc.Equals, "spam")
	c.Check(string(hasher.written), gc.Equals, "spam")
}

func (s *WriterSuite) TestHashingWriterWriteFileError(c *gc.C) {
	file := fakeFile{err: errors.New("failed!")}
	hasher := fakeHasher{}
	w := hash.NewHashingWriter(&file, &hasher)
	_, err := w.Write([]byte("spam"))

	c.Check(err, gc.ErrorMatches, "failed!")
}

func (s *WriterSuite) TestHashingWriterWriteHasherError(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{}
	hasher.err = errors.New("failed!")
	w := hash.NewHashingWriter(&file, &hasher)
	_, err := w.Write([]byte("spam"))

	c.Check(err, gc.ErrorMatches, "failed!")
}

func (s *WriterSuite) TestHashingWriterHash(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&file, &hasher)
	b64hash := w.Hash()

	c.Check(b64hash, gc.Equals, "c3BhbQ==")
}

func (s *WriterSuite) TestHashingWriterRawHash(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&file, &hasher)
	rawhash := w.RawHash()

	c.Check(rawhash, gc.Equals, "7370616d")
}
