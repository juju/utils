// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package basic_test

import (
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/document"
	"github.com/juju/utils/document/basic"
)

var _ = gc.Suite(&DocStorageSuite{})

type DocStorageSuite struct {
	testing.IsolationSuite
	original document.Document
	stor     document.DocumentStorage
}

func (s *DocStorageSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.original = &document.Doc{}
	s.stor = basic.NewDocStorage()
}

func (s *DocStorageSuite) TestNewDocStorage(c *gc.C) {
	var stor document.DocumentStorage = basic.NewDocStorage()

	c.Check(stor, gc.NotNil)
}

func (s *DocStorageSuite) TestDoc(c *gc.C) {
	id, err := s.stor.AddDoc(s.original)
	c.Assert(err, gc.IsNil)

	doc, err := s.stor.Doc(id)
	c.Assert(err, gc.IsNil)
	meta, ok := doc.(document.Document)
	c.Assert(ok, gc.Equals, true)
	s.original.SetID(id)
	c.Check(meta, gc.DeepEquals, s.original)
}

func (s *DocStorageSuite) TestListDocs(c *gc.C) {
	id, err := s.stor.AddDoc(s.original)
	c.Assert(err, gc.IsNil)

	list, err := s.stor.ListDocs()
	c.Assert(err, gc.IsNil)
	c.Assert(list, gc.HasLen, 1)
	c.Assert(list[0], gc.NotNil)
	meta, ok := list[0].(document.Document)
	c.Assert(ok, gc.Equals, true)
	c.Check(meta.ID(), gc.Equals, id)
}

func (s *DocStorageSuite) TestAddDoc(c *gc.C) {
	list, err := s.stor.ListDocs()
	c.Assert(err, gc.IsNil)
	c.Assert(list, gc.HasLen, 0)

	id, err := s.stor.AddDoc(s.original)

	meta, err := s.stor.Doc(id)
	c.Assert(err, gc.IsNil)
	s.original.SetID(id)
	c.Check(meta, gc.DeepEquals, s.original)
}

func (s *DocStorageSuite) TestRemoveDoc(c *gc.C) {
	id, err := s.stor.AddDoc(s.original)
	c.Assert(err, gc.IsNil)

	err = s.stor.RemoveDoc(id)
	c.Assert(err, gc.IsNil)
	_, err = s.stor.Doc(id)
	c.Assert(err, gc.NotNil)
}
