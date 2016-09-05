// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"io"
	"io/ioutil"
	"strings"
	"testing/iotest"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
)

type multiReaderSeekerSuite struct{}

var _ = gc.Suite(&multiReaderSeekerSuite{})

func (*multiReaderSeekerSuite) TestSequentialRead(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	r := newMultiStringReader(parts)
	data, err := ioutil.ReadAll(r)
	c.Assert(err, gc.IsNil)
	c.Assert(string(data), gc.Equals, strings.Join(parts, ""))
}

func (*multiReaderSeekerSuite) TestSeekStart(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	all := strings.Join(parts, "")
	for off := int64(0); off <= int64(len(all)); off++ {
		c.Logf("-- offset %d", off)
		r := newMultiStringReader(parts)
		gotOff, err := r.Seek(off, 0)
		c.Assert(err, gc.IsNil)
		c.Assert(gotOff, gc.Equals, off)

		data, err := ioutil.ReadAll(r)
		c.Assert(err, gc.IsNil)
		c.Assert(string(data), gc.Equals, all[off:])
	}
}

func (*multiReaderSeekerSuite) TestSeekEnd(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	all := strings.Join(parts, "")
	for off := int64(0); off <= int64(len(all)); off++ {
		r := newMultiStringReader(parts)
		expectOff := int64(len(all)) - off
		gotOff, err := r.Seek(-off, 2)
		c.Assert(err, gc.IsNil)
		c.Assert(gotOff, gc.Equals, expectOff)

		data, err := ioutil.ReadAll(r)
		c.Assert(err, gc.IsNil)
		c.Assert(string(data), gc.Equals, all[expectOff:])
	}
}

func (*multiReaderSeekerSuite) TestSeekCur(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	all := strings.Join(parts, "")
	for off := int64(0); off <= int64(len(all)); off++ {
		for newOff := int64(0); newOff <= int64(len(all)); newOff++ {
			readers := make([]io.ReadSeeker, len(parts))
			for i, part := range parts {
				readers[i] = strings.NewReader(part)
			}
			r := utils.NewMultiReaderSeeker(readers...)
			gotOff, err := r.Seek(off, 0)
			c.Assert(gotOff, gc.Equals, off)
			c.Assert(err, jc.ErrorIsNil)

			diff := newOff - off
			gotNewOff, err := r.Seek(diff, 1)
			c.Assert(err, gc.IsNil)
			c.Assert(gotNewOff, gc.Equals, newOff)

			data, err := ioutil.ReadAll(r)
			c.Assert(err, gc.IsNil)
			c.Assert(string(data), gc.Equals, all[newOff:])
		}
	}
}

func (*multiReaderSeekerSuite) TestSeekAfterRead(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	all := strings.Join(parts, "")
	r := newMultiStringReader(parts)
	data, err := ioutil.ReadAll(iotest.OneByteReader(r))
	c.Assert(err, gc.IsNil)
	c.Assert(string(data), gc.Equals, all)

	off, err := r.Seek(-8, 2)
	c.Assert(err, gc.IsNil)
	c.Assert(off, gc.Equals, int64(len(all)-8))

	data, err = ioutil.ReadAll(r)
	c.Assert(err, gc.IsNil)
	c.Assert(string(data), gc.Equals, "hreefour")
}

func (*multiReaderSeekerSuite) TestSeekNegative(c *gc.C) {
	r := newMultiStringReader([]string{"one", "two"})

	_, err := r.Seek(-1, 0)
	c.Assert(err, gc.ErrorMatches, "negative position")

	n, err := r.Seek(0, 0)
	c.Assert(err, gc.IsNil)
	c.Assert(n, gc.Equals, int64(0))

	_, err = r.Seek(-7, 2)
	c.Assert(err, gc.ErrorMatches, "negative position")

	n, err = r.Seek(0, 0)
	c.Assert(err, gc.IsNil)
	c.Assert(n, gc.Equals, int64(0))

	_, err = r.Seek(-1, 1)
	c.Assert(err, gc.ErrorMatches, "negative position")

	n, err = r.Seek(0, 0)
	c.Assert(err, gc.IsNil)
	c.Assert(n, gc.Equals, int64(0))
}

func (*multiReaderSeekerSuite) TestSeekPastEnd(c *gc.C) {
	r := newMultiStringReader([]string{"one", "two"})

	n, err := r.Seek(100, 0)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(n, gc.Equals, int64(100))

	nr, err := r.Read(make([]byte, 1))
	c.Assert(nr, gc.Equals, 0)
	c.Assert(err, gc.Equals, io.EOF)

	n, err = r.Seek(-5, 1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(n, gc.Equals, int64(95))

	nr, err = r.Read(make([]byte, 1))
	c.Assert(nr, gc.Equals, 0)
	c.Assert(err, gc.Equals, io.EOF)

	n, err = r.Seek(-94, 1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(n, gc.Equals, int64(1))

	data, err := ioutil.ReadAll(r)
	c.Assert(err, gc.IsNil)
	c.Assert(string(data), gc.Equals, "netwo")
}

type multiReaderAtSuite struct{}

var _ = gc.Suite(&multiReaderAtSuite{})

func (*multiReaderAtSuite) TestReadComplete(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	all := strings.Join(parts, "")
	r := newMultistringReaderAt(parts)

	buf := make([]byte, len(all))
	n, err := r.ReadAt(buf, 0)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(n, gc.Equals, len(buf))
	c.Assert(string(buf), gc.Equals, all)
}

func (*multiReaderAtSuite) TestReadPartial(c *gc.C) {
	parts := []string{
		"one",
		"two",
		"three",
		"four",
	}
	all := strings.Join(parts, "")
	r := newMultistringReaderAt(parts)

	buf := make([]byte, len(all)-4)
	n, err := r.ReadAt(buf, 2)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(n, gc.Equals, len(buf))
	c.Assert(string(buf), gc.Equals, "etwothreefo")
}

func newMultiStringReader(parts []string) io.ReadSeeker {
	readers := make([]io.ReadSeeker, len(parts))
	for i, part := range parts {
		readers[i] = strings.NewReader(part)
	}
	return utils.NewMultiReaderSeeker(readers...)
}

type stringReader struct {
	*strings.Reader
}

// This method is implemented in later versions
// of Go's StringReader but not prior to Go 1.5.
func (r stringReader) Size() int64 {
	return int64(r.Len())
}

func newMultistringReaderAt(parts []string) io.ReaderAt {
	readers := make([]utils.SizeReaderAt, len(parts))
	for i, part := range parts {
		readers[i] = stringReader{strings.NewReader(part)}
	}
	return utils.NewMultiReaderAt(readers...)
}
