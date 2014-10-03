// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"time"

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
	timestamp := time.Now().UTC()
	meta := filestorage.NewMetadata(&timestamp)

	c.Check(meta.ID(), gc.Equals, "")
	c.Check(meta.Size(), gc.Equals, int64(0))
	c.Check(meta.Checksum(), gc.Equals, "")
	c.Check(meta.ChecksumFormat(), gc.Equals, "")
	c.Check(meta.Timestamp(), gc.Equals, timestamp)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *FileMetadataSuite) TestNewMetadataDefaultTimestamp(c *gc.C) {
	meta := filestorage.NewMetadata(nil)

	c.Check(meta.ID(), gc.Equals, "")
	c.Check(meta.Timestamp(), gc.NotNil)
}

func (s *FileMetadataSuite) TestDoc(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")
	doc := meta.Doc()

	c.Check(doc, gc.Equals, meta)
}

func (s *FileMetadataSuite) TestSetIDInitial(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")
	c.Assert(meta.ID(), gc.Equals, "")

	success := meta.SetID("some id")
	c.Check(success, gc.Equals, false)
	c.Check(meta.ID(), gc.Equals, "some id")
}

func (s *FileMetadataSuite) TestSetIDAlreadySetSame(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")
	success := meta.SetID("some id")
	c.Assert(success, gc.Equals, false)

	success = meta.SetID("some id")
	c.Check(success, gc.Equals, true)
	c.Check(meta.ID(), gc.Equals, "some id")
}

func (s *FileMetadataSuite) TestSetIDAlreadySetDifferent(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")
	success := meta.SetID("some id")
	c.Assert(success, gc.Equals, false)

	success = meta.SetID("another id")
	c.Check(success, gc.Equals, true)
	c.Check(meta.ID(), gc.Equals, "some id")
}

func (s *FileMetadataSuite) TestSetFile(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	c.Assert(meta.Size(), gc.Equals, int64(0))
	c.Assert(meta.Checksum(), gc.Equals, "")
	c.Assert(meta.ChecksumFormat(), gc.Equals, "")
	c.Assert(meta.Stored(), gc.Equals, false)
	meta.SetFile(10, "some sum", "SHA-1")

	c.Check(meta.Size(), gc.Equals, int64(10))
	c.Check(meta.Checksum(), gc.Equals, "some sum")
	c.Check(meta.ChecksumFormat(), gc.Equals, "SHA-1")
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *FileMetadataSuite) TestSetStored(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")
	c.Assert(meta.Stored(), gc.Equals, false)

	meta.SetStored()
	c.Check(meta.Stored(), gc.Equals, true)
}

func (s *FileMetadataSuite) TestSetStoredIdempotent(c *gc.C) {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "some sum", "SHA-1")

	meta.SetStored()
	meta.SetStored()
	c.Check(meta.Stored(), gc.Equals, true)
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
