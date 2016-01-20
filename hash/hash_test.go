// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/filetesting"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/hash"
)

var _ = gc.Suite(&HashSuite{})

type HashSuite struct {
	testing.IsolationSuite
}

func (s *HashSuite) TestHashingWriter(c *gc.C) {
	data := "some data"
	newHash, _ := hash.SHA384()
	expected, err := hash.GenerateFingerprint(strings.NewReader(data), newHash)
	c.Assert(err, jc.ErrorIsNil)
	var writer bytes.Buffer

	h := newHash()
	hashingWriter := io.MultiWriter(&writer, h)
	_, err = hashingWriter.Write([]byte(data))
	c.Assert(err, jc.ErrorIsNil)
	fp := hash.NewValidFingerprint(h)

	c.Check(fp, jc.DeepEquals, expected)
	c.Check(writer.String(), gc.Equals, data)
}

func (s *HashSuite) TestHashingReader(c *gc.C) {
	expected := "some data"
	stub := &testing.Stub{}
	reader := &filetesting.StubReader{
		Stub: stub,
		ReturnRead: &fakeStream{
			data: expected,
		},
	}

	newHash, validate := hash.SHA384()
	h := newHash()
	hashingReader := io.TeeReader(reader, h)
	data, err := ioutil.ReadAll(hashingReader)
	c.Assert(err, jc.ErrorIsNil)
	fp := hash.NewValidFingerprint(h)
	hexSum := fp.Hex()
	fpAgain, err := hash.ParseHexFingerprint(hexSum, validate)
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "Read") // The EOF was mixed with the data.
	c.Check(string(data), gc.Equals, expected)
	c.Check(fpAgain, jc.DeepEquals, fp)
}

type fakeStream struct {
	data string
	pos  uint64
}

func (f *fakeStream) Read(data []byte) (int, error) {
	n := copy(data, f.data[f.pos:])
	f.pos += uint64(n)
	if f.pos >= uint64(len(f.data)) {
		return n, io.EOF
	}
	return n, nil
}
