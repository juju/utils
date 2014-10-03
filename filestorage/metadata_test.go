// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filestorage"
)

var _ = gc.Suite(&FileMetadataSuite{})

type FileMetadataSuite struct {
	testing.IsolationSuite
}

func (s *FileMetadataSuite) TestNewMetadata(c *gc.C) {
	meta := filestorage.NewMetadata(nil)

	c.Check(meta.Raw.Size, gc.Equals, int64(0))
	c.Check(meta.Raw.Checksum, gc.Equals, "")
	c.Check(meta.Raw.ChecksumFormat, gc.Equals, "")
}

func (s *FileMetadataSuite) TestSetFile(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	c.Assert(meta.Size(), gc.Equals, int64(0))
	c.Assert(meta.Checksum(), gc.Equals, "")
	c.Assert(meta.ChecksumFormat(), gc.Equals, "")
	meta.SetFile(10, "some sum", "SHA-1")

	c.Check(meta.Size(), gc.Equals, int64(10))
	c.Check(meta.Checksum(), gc.Equals, "some sum")
	c.Check(meta.ChecksumFormat(), gc.Equals, "SHA-1")
}

func (s *FileMetadataSuite) TestSetFileNotStored(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	c.Assert(meta.Stored(), gc.IsNil)
	meta.SetFile(10, "some sum", "SHA-1")

	c.Check(meta.Stored(), gc.IsNil)
}

func (s *FileMetadataSuite) TestCopy(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")
	doc := meta.Copy("")
	copied, ok := doc.(filestorage.Metadata)
	c.Assert(ok, jc.IsTrue)

	c.Check(copied, gc.Not(gc.Equals), meta)
	c.Check(copied, gc.DeepEquals, meta)
}
