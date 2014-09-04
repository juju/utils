// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filestorage"
	"github.com/juju/utils/filestorage/basic"
)

var _ = gc.Suite(&WrapperSuite{})

type WrapperSuite struct {
	testing.IsolationSuite
	rawstor  filestorage.RawFileStorage
	metastor filestorage.MetadataStorage
	stor     filestorage.FileStorage
}

func (s *WrapperSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	var err error
	s.rawstor, err = basic.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)
	s.metastor = basic.NewMetadataStorage()
	s.stor = filestorage.NewFileStorage(s.metastor, s.rawstor)
}

func (s *WrapperSuite) metadata() filestorage.Metadata {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "", "")
	return meta
}

func (s *WrapperSuite) addmeta(c *gc.C, meta filestorage.Metadata, file io.Reader) string {
	id, err := s.stor.Add(meta, file)
	c.Assert(err, gc.IsNil)
	alreadySet := meta.SetID(id)
	c.Assert(alreadySet, jc.IsFalse)
	c.Assert(meta.ID(), gc.Equals, id)
	if file != nil {
		meta.SetStored()
	}
	return id
}

func (s *WrapperSuite) add(c *gc.C, file io.Reader) (string, filestorage.Metadata) {
	meta := s.metadata()
	id := s.addmeta(c, meta, file)
	return id, meta
}

func (s *WrapperSuite) TestFileStorageNewFileStorage(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)

	c.Check(stor, gc.NotNil)
}

func (s *WrapperSuite) TestFileStorageMetadata(c *gc.C) {
	id, original := s.add(c, nil)
	meta, err := s.stor.Metadata(id)
	c.Check(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, original)
}

func (s *WrapperSuite) TestFileStorageGet(c *gc.C) {
	data := bytes.NewBufferString("spam")
	id, original := s.add(c, data)
	meta, file, err := s.stor.Get(id)
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)

	c.Check(meta, gc.DeepEquals, original)
	c.Check(string(content), gc.Equals, "spam")
}

func (s *WrapperSuite) TestFileStorageListEmpty(c *gc.C) {
	list, err := s.stor.List()
	c.Check(err, gc.IsNil)

	c.Check(list, gc.HasLen, 0)
}

func (s *WrapperSuite) TestFileStorageListOne(c *gc.C) {
	id, _ := s.add(c, nil)
	list, err := s.stor.List()
	c.Check(err, gc.IsNil)

	c.Check(list, gc.HasLen, 1)
	c.Assert(list[0], gc.NotNil)
	c.Check(list[0].ID(), gc.Equals, id)
}

func (s *WrapperSuite) TestFileStorageListTwo(c *gc.C) {
	id1, _ := s.add(c, nil)
	id2, _ := s.add(c, nil)
	list, err := s.stor.List()
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
	original := s.metadata()
	c.Assert(original.ID(), gc.Equals, "")

	id, err := s.stor.Add(original, nil)
	c.Check(err, gc.IsNil)

	meta, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)

	c.Check(original.ID(), gc.Equals, "")

	original.SetID(id)
	c.Check(meta, gc.DeepEquals, original)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *WrapperSuite) TestFileStorageAddFile(c *gc.C) {
	original := s.metadata()
	data := bytes.NewBufferString("spam")
	id, err := s.stor.Add(original, data)
	c.Assert(err, gc.IsNil)

	meta, file, err := s.stor.Get(id)
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)

	c.Check(original.ID(), gc.Equals, "")
	c.Check(original.Stored(), jc.IsFalse)

	original.SetID(id)
	original.SetStored()
	c.Check(meta, gc.DeepEquals, original)

	c.Check(string(content), gc.Equals, "spam")
	c.Check(meta.Stored(), gc.Equals, true)
}

func (s *WrapperSuite) TestFileStorageAddIDNotSet(c *gc.C) {
	original := s.metadata()
	c.Assert(original.ID(), gc.Equals, "")
	_, err := s.stor.Add(original, nil)
	c.Check(err, gc.IsNil)

	c.Check(original.ID(), gc.Equals, "")
}

func (s *WrapperSuite) TestFileStorageAddMetaOnly(c *gc.C) {
	id, original := s.add(c, nil)

	meta, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, original)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *WrapperSuite) TestFileStorageAddIDAlreadySet(c *gc.C) {
	original := s.metadata()
	original.SetID("eggs")
	_, err := s.stor.Add(original, nil)

	c.Check(err, gc.ErrorMatches, "ID already set .*")
}

func (s *WrapperSuite) TestFileStorageSetFile(c *gc.C) {
	id, _ := s.add(c, nil)
	_, _, err := s.stor.Get(id)
	c.Assert(err, gc.NotNil)
	meta, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)
	c.Check(meta.Stored(), gc.Equals, false)

	data := bytes.NewBufferString("spam")
	err = s.stor.SetFile(id, data)
	meta, file, err := s.stor.Get(id)
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)
	c.Check(meta.Stored(), gc.Equals, true)
	c.Check(string(content), gc.Equals, "spam")
}

func (s *WrapperSuite) TestFileStorageRemove(c *gc.C) {
	id, _ := s.add(c, nil)
	_, err := s.stor.Metadata(id) // Ensure it is there.
	c.Assert(err, gc.IsNil)

	err = s.stor.Remove(id)
	c.Check(err, gc.IsNil)
	_, err = s.stor.Metadata(id) // Ensure it isn't there.
	c.Check(err, gc.NotNil)
}

func (s *WrapperSuite) TestClose(c *gc.C) {
	metaStor := &FakeMetadataStorage{}
	fileStor := &FakeRawFileStorage{}
	stor := filestorage.NewFileStorage(metaStor, fileStor)
	err := stor.Close()
	c.Assert(err, gc.IsNil)

	c.Check(metaStor.calls, gc.DeepEquals, []string{"Close"})
	c.Check(fileStor.calls, gc.DeepEquals, []string{"Close"})
}
