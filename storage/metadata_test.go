// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package storage_test

import (
	"time"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/storage"
)

var _ = gc.Suite(&MetadataSuite{})

type MetadataSuite struct {
	testing.IsolationSuite
}

func (s *MetadataSuite) TestNewMetadata(c *gc.C) {
	timestamp := time.Now().UTC()
	meta := storage.NewMetadata(&timestamp)

	c.Check(meta.ID(), gc.Equals, "")
	c.Check(meta.Created(), gc.Equals, timestamp)
	c.Check(meta.Stored(), gc.IsNil)
}

func (s *MetadataSuite) TestStored(c *gc.C) {
	timestamp := time.Now().UTC()
	meta := storage.NewMetadata(nil)
	before := meta.Stored()
	meta.Raw.Stored = &timestamp
	after := meta.Stored()

	c.Check(before, gc.IsNil)
	c.Check(after, gc.Equals, &timestamp)
}

func (s *MetadataSuite) TestSetStoredInitial(c *gc.C) {
	timestamp := time.Now().UTC()
	meta := storage.NewMetadata(nil)
	c.Assert(meta.Raw.Stored, gc.IsNil)
	already := meta.SetStored(&timestamp)

	c.Check(already, gc.Equals, false)
	c.Check(meta.Raw.Stored, gc.Equals, &timestamp)
}

func (s *MetadataSuite) TestSetStoredAlreadySetSame(c *gc.C) {
	timestamp := time.Now().UTC()
	meta := storage.NewMetadata(nil)
	meta.Raw.Stored = &timestamp
	already := meta.SetStored(&timestamp)

	c.Check(already, gc.Equals, true)
	c.Check(meta.Raw.Stored, gc.Equals, &timestamp)
}

func (s *MetadataSuite) TestSetStoredAlreadySetDifferent(c *gc.C) {
	timestamp := time.Now().UTC()
	meta := storage.NewMetadata(nil)
	meta.Raw.Stored = &timestamp
	nextTS := timestamp.Add(time.Minute)
	already := meta.SetStored(&nextTS)

	c.Check(already, gc.Equals, true)
	c.Check(meta.Raw.Stored, gc.Equals, &timestamp)
}

func (s *MetadataSuite) TestCopy(c *gc.C) {
	meta := storage.NewMetadata(nil)
	doc := meta.Copy("")
	copied, ok := doc.(storage.Metadata)
	c.Assert(ok, jc.IsTrue)

	c.Check(copied, gc.Not(gc.Equals), meta)
	c.Check(copied.Stored(), gc.Equals, meta.Raw.Stored)
}
