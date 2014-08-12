// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"bytes"
	"io/ioutil"

	"github.com/juju/testing"
	gc "launchpad.net/gocheck"

	"github.com/juju/utils/filestorage"
)

var _ = gc.Suite(&WrapperSuite{})

type WrapperSuite struct {
	testing.IsolationSuite
	rawstor  filestorage.RawFileStorage
	metastor filestorage.MetadataStorage
	original filestorage.Metadata
}

func (s *WrapperSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	var err error
	s.rawstor, err = filestorage.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)
	s.metastor = filestorage.NewMetadataStorage()
	s.original = filestorage.NewMetadata(nil)
	s.original.SetFile(10, "", "")
}

func (s *WrapperSuite) TestFileStorageNewFileStorage(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)

	c.Check(stor, gc.NotNil)
}

func (s *WrapperSuite) TestFileStorageMetadata(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	id, err := stor.Add(s.original, nil)
	c.Assert(err, gc.IsNil)
	meta, err := stor.Metadata(id)
	c.Check(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, s.original)
}

func (s *WrapperSuite) TestFileStorageGet(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	data := bytes.NewBufferString("spam")
	id, err := stor.Add(s.original, data)
	c.Assert(err, gc.IsNil)
	meta, file, err := stor.Get(id)
	c.Check(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)

	c.Check(meta, gc.DeepEquals, s.original)
	c.Check(string(content), gc.Equals, "spam")
}

func (s *WrapperSuite) TestFileStorageListEmpty(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	list, err := stor.List()
	c.Check(err, gc.IsNil)

	c.Check(list, gc.HasLen, 0)
}

func (s *WrapperSuite) TestFileStorageListOne(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	id, err := stor.Add(s.original, nil)
	c.Assert(err, gc.IsNil)
	list, err := stor.List()
	c.Check(err, gc.IsNil)

	c.Check(list, gc.HasLen, 1)
	c.Assert(list[0], gc.NotNil)
	c.Check(list[0].ID(), gc.Equals, id)
}

func (s *WrapperSuite) TestFileStorageListTwo(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	original1 := filestorage.NewMetadata(nil)
	id1, err := stor.Add(original1, nil)
	c.Assert(err, gc.IsNil)
	original2 := filestorage.NewMetadata(nil)
	id2, err := stor.Add(original2, nil)
	c.Assert(err, gc.IsNil)
	list, err := stor.List()
	c.Check(err, gc.IsNil)

	c.Assert(list, gc.HasLen, 2)
	c.Assert(list[0], gc.NotNil)
	c.Assert(list[1], gc.NotNil)
	if list[0].ID() == id1 {
		c.Check(list[1].ID(), gc.Equals, id2)
	} else {
		c.Check(list[1].ID(), gc.Equals, id1)
	}
}

func (s *WrapperSuite) TestFileStorageAddMeta(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	id, err := stor.Add(s.original, nil)
	c.Check(err, gc.IsNil)

	meta, err := stor.Metadata(id)
	c.Assert(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, s.original)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *WrapperSuite) TestFileStorageAddFile(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	data := bytes.NewBufferString("spam")
	id, err := stor.Add(s.original, data)
	c.Check(err, gc.IsNil)

	meta, file, err := stor.Get(id)
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, s.original)
	c.Check(string(content), gc.Equals, "spam")
	c.Check(meta.Stored(), gc.Equals, true)
}

func (s *WrapperSuite) TestFileStorageAddIDAlreadySet(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	s.original.SetID("eggs")
	_, err := stor.Add(s.original, nil)

	c.Check(err, gc.ErrorMatches, "ID already set .*")
}

func (s *WrapperSuite) TestFileStorageSetFile(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	id, err := stor.Add(s.original, nil)
	c.Assert(err, gc.IsNil)

	_, _, err = stor.Get(id)
	c.Assert(err, gc.NotNil)
	meta, err := stor.Metadata(id)
	c.Assert(err, gc.IsNil)
	c.Check(meta.Stored(), gc.Equals, false)

	data := bytes.NewBufferString("spam")
	err = stor.SetFile(id, data)
	meta, file, err := stor.Get(id)
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)
	c.Check(meta.Stored(), gc.Equals, true)
	c.Check(string(content), gc.Equals, "spam")
}

func (s *WrapperSuite) TestFileStorageRemove(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)
	id, err := stor.Add(s.original, nil)
	c.Assert(err, gc.IsNil)
	_, err = stor.Metadata(id)
	c.Assert(err, gc.IsNil)

	err = stor.Remove(id)
	c.Check(err, gc.IsNil)
	_, err = stor.Metadata(id)
	c.Check(err, gc.NotNil)
}
