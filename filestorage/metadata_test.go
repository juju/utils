// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"time"

	"github.com/juju/testing"
	gc "launchpad.net/gocheck"

	"github.com/juju/utils/filestorage"
)

var _ = gc.Suite(&MetadataSuite{})

type MetadataSuite struct {
	testing.IsolationSuite
}

func (s *MetadataSuite) TestFileMetadataNewMetadata(c *gc.C) {
	timestamp := time.Now().UTC()
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", &timestamp)

	c.Check(meta.ID(), gc.Equals, "")
	c.Check(meta.Size(), gc.Equals, int64(10))
	c.Check(meta.Checksum(), gc.Equals, "some sum")
	c.Check(meta.ChecksumFormat(), gc.Equals, "SHA-1")
	c.Check(meta.Timestamp(), gc.Equals, timestamp)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *MetadataSuite) TestFileMetadataNewMetadataDefaults(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)

	c.Check(meta.ID(), gc.Equals, "")
	c.Check(meta.Size(), gc.Equals, int64(10))
	c.Check(meta.Checksum(), gc.Equals, "some sum")
	c.Check(meta.ChecksumFormat(), gc.Equals, "SHA-1")
	c.Check(meta.Timestamp(), gc.NotNil)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *MetadataSuite) TestFileMetadataDoc(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)
	doc := meta.Doc()

	c.Check(doc, gc.Equals, meta)
}

func (s *MetadataSuite) TestFileMetadataSetIDInitial(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)
	c.Assert(meta.ID(), gc.Equals, "")

	success := meta.SetID("some id")
	c.Check(success, gc.Equals, false)
	c.Check(meta.ID(), gc.Equals, "some id")
}

func (s *MetadataSuite) TestFileMetadataSetIDAlreadySetSame(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)
	success := meta.SetID("some id")
	c.Assert(success, gc.Equals, false)

	success = meta.SetID("some id")
	c.Check(success, gc.Equals, true)
	c.Check(meta.ID(), gc.Equals, "some id")
}

func (s *MetadataSuite) TestFileMetadataSetIDAlreadySetDifferent(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)
	success := meta.SetID("some id")
	c.Assert(success, gc.Equals, false)

	success = meta.SetID("another id")
	c.Check(success, gc.Equals, true)
	c.Check(meta.ID(), gc.Equals, "some id")
}

func (s *MetadataSuite) TestFileMetadataSetStored(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)
	c.Assert(meta.Stored(), gc.Equals, false)

	meta.SetStored()
	c.Check(meta.Stored(), gc.Equals, true)
}

func (s *MetadataSuite) TestFileMetadataSetStoredIdempotent(c *gc.C) {
	meta := filestorage.NewMetadata(10, "some sum", "SHA-1", nil)

	meta.SetStored()
	meta.SetStored()
	c.Check(meta.Stored(), gc.Equals, true)
}
