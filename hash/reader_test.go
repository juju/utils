// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash_test

import (
	"bytes"
	"io/ioutil"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/filetesting"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/hash"
)

var _ = gc.Suite(&ReaderSuite{})

type ReaderSuite struct {
	testing.IsolationSuite

	stub    *testing.Stub
	rBuffer *bytes.Buffer
	reader  *filetesting.StubReader
	hBuffer *bytes.Buffer
	hash    *filetesting.StubHash
}

func (s *ReaderSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.stub = &testing.Stub{}
	s.rBuffer = new(bytes.Buffer)
	s.reader = &filetesting.StubReader{
		Stub:       s.stub,
		ReturnRead: s.rBuffer,
	}
	s.hBuffer = new(bytes.Buffer)
	s.hash = filetesting.NewStubHash(s.stub, s.hBuffer)
}

func (s *ReaderSuite) TestHashingReaderReadEmpty(c *gc.C) {
	r := hash.NewHashingReader(s.reader, s.hash)
	data, err := ioutil.ReadAll(r)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Read")
	c.Check(string(data), gc.HasLen, 0)
	c.Check(s.hBuffer.String(), gc.Equals, "")
}

func (s *ReaderSuite) TestHashingReaderReadSmall(c *gc.C) {
	_, err := s.rBuffer.WriteString("spam")
	c.Assert(err, jc.ErrorIsNil)
	r := hash.NewHashingReader(s.reader, s.hash)
	data, err := ioutil.ReadAll(r)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Read", "Write", "Read")
	c.Check(string(data), gc.Equals, "spam")
	c.Check(s.hBuffer.String(), gc.Equals, "spam")
}

func (s *ReaderSuite) TestHashingReaderReadFileError(c *gc.C) {
	r := hash.NewHashingReader(s.reader, s.hash)
	failure := errors.New("<failed>")
	s.stub.SetErrors(failure)

	_, err := ioutil.ReadAll(r)

	s.stub.CheckCallNames(c, "Read")
	c.Check(errors.Cause(err), gc.Equals, failure)
	c.Check(s.hBuffer.String(), gc.Equals, "")
}

func (s *ReaderSuite) TestHashingReaderSum(c *gc.C) {
	s.hash.ReturnSum = []byte("spam")
	w := hash.NewHashingReader(s.reader, s.hash)
	sum := string(w.Sum())

	s.stub.CheckCallNames(c, "Sum")
	c.Check(sum, gc.Equals, "spam")
}

func (s *ReaderSuite) TestHashingReaderBase64Sum(c *gc.C) {
	s.hash.ReturnSum = []byte("spam")
	w := hash.NewHashingReader(s.reader, s.hash)
	b64sum := w.Base64Sum()

	s.stub.CheckCallNames(c, "Sum")
	c.Check(b64sum, gc.Equals, "c3BhbQ==")
}

func (s *ReaderSuite) TestHashingReaderHexSum(c *gc.C) {
	s.hash.ReturnSum = []byte("spam")
	w := hash.NewHashingReader(s.reader, s.hash)
	hexSum := w.HexSum()

	s.stub.CheckCallNames(c, "Sum")
	c.Check(hexSum, gc.Equals, "7370616d")
}
