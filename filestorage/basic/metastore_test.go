// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package basic_test

import (
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filestorage"
	"github.com/juju/utils/filestorage/basic"
)

var _ = gc.Suite(&MetadataStorageSuite{})

type MetadataStorageSuite struct {
	testing.IsolationSuite
	original filestorage.Metadata
	stor     filestorage.MetadataStorage
}

func (s *MetadataStorageSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.original = filestorage.NewMetadata(nil)
	s.original.SetFile(0, "", "")
	s.stor = basic.NewMetadataStorage()
}

func (s *MetadataStorageSuite) TestMetadataStorageNewMetadataStorage(c *gc.C) {
	var stor filestorage.MetadataStorage = basic.NewMetadataStorage()

	c.Check(stor, gc.NotNil)
}

func (s *MetadataStorageSuite) TestMetadata(c *gc.C) {
	id, err := s.stor.AddMetadata(s.original)
	c.Assert(err, gc.IsNil)

	meta, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)
	s.original.SetID(id)
	c.Check(meta, gc.DeepEquals, s.original)
}

func (s *MetadataStorageSuite) TestListMetadata(c *gc.C) {
	id, err := s.stor.AddMetadata(s.original)
	c.Assert(err, gc.IsNil)

	list, err := s.stor.ListMetadata()
	c.Assert(err, gc.IsNil)
	c.Assert(list, gc.HasLen, 1)
	c.Assert(list[0], gc.NotNil)
	c.Check(list[0].ID(), gc.Equals, id)
}

func (s *MetadataStorageSuite) TestSetStored(c *gc.C) {
	id, err := s.stor.AddMetadata(s.original)
	c.Assert(err, gc.IsNil)
	meta, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)
	c.Check(meta.Stored(), gc.Equals, false)

	err = s.stor.SetStored(id)
	c.Assert(err, gc.IsNil)
	c.Check(meta.Stored(), gc.Equals, false)

	stored, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)
	c.Check(stored.Stored(), gc.Equals, true)
}
