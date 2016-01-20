// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"io"
	"io/ioutil"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/filetesting"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

var _ = gc.Suite(&sizeSuite{})

type sizeSuite struct {
	testing.IsolationSuite
}

func (*sizeSuite) TestParseSize(c *gc.C) {
	type test struct {
		in  string
		out uint64
		err string
	}
	tests := []test{{
		in:  "",
		err: `expected a non-negative number, got ""`,
	}, {
		in:  "-1",
		err: `expected a non-negative number, got "-1"`,
	}, {
		in:  "1MZ",
		err: `invalid multiplier suffix "MZ", expected one of MGTPEZY`,
	}, {
		in:  "0",
		out: 0,
	}, {
		in:  "123",
		out: 123,
	}, {
		in:  "1M",
		out: 1,
	}, {
		in:  "0.5G",
		out: 512,
	}, {
		in:  "0.5GB",
		out: 512,
	}, {
		in:  "0.5GiB",
		out: 512,
	}, {
		in:  "0.5T",
		out: 524288,
	}, {
		in:  "0.5P",
		out: 536870912,
	}, {
		in:  "0.0009765625E",
		out: 1073741824,
	}, {
		in:  "1Z",
		out: 1125899906842624,
	}, {
		in:  "1Y",
		out: 1152921504606846976,
	}}
	for i, test := range tests {
		c.Logf("test %d: %+v", i, test)
		size, err := utils.ParseSize(test.in)
		if test.err != "" {
			c.Assert(err, gc.NotNil)
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, gc.IsNil)
			c.Assert(size, gc.Equals, test.out)
		}
	}
}

func (*sizeSuite) TestSizingReaderOkay(c *gc.C) {
	expected := "some data"
	stub := &testing.Stub{}
	reader := filetesting.NewStubReader(stub, expected)

	var st utils.SizeTracker
	sizingReader := io.TeeReader(reader, &st)
	data, err := ioutil.ReadAll(sizingReader)
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "Read", "Read")
	c.Check(string(data), gc.Equals, expected)
	c.Check(st.Size(), gc.Equals, int64(len(expected)))
}

func (*sizeSuite) TestSizingReaderMixedEOF(c *gc.C) {
	expected := "some data"
	stub := &testing.Stub{}
	reader := &filetesting.StubReader{
		Stub: stub,
		ReturnRead: &fakeStream{
			data: expected,
		},
	}

	var st utils.SizeTracker
	sizingReader := io.TeeReader(reader, &st)
	data, err := ioutil.ReadAll(sizingReader)
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "Read") // The EOF was mixed with the data.
	c.Check(string(data), gc.Equals, expected)
	c.Check(st.Size(), gc.Equals, int64(len(expected)))
}

func (*sizeSuite) TestSizingWriter(c *gc.C) {
	expected := "some data"
	stub := &testing.Stub{}
	writer, buffer := filetesting.NewStubWriter(stub)

	var st utils.SizeTracker
	sizingWriter := io.MultiWriter(writer, &st)
	n, err := sizingWriter.Write([]byte(expected))
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "Write")
	c.Check(n, gc.Equals, len(expected))
	c.Check(buffer.String(), gc.Equals, expected)
	c.Check(st.Size(), gc.Equals, int64(len(expected)))
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
