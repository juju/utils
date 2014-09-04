// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"bytes"
	"io"
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
	stor     filestorage.FileStorage
}

func (s *WrapperSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	var err error
	s.rawstor, err = filestorage.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)
	s.metastor = filestorage.NewMetadataStorage()
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
	c.Assert(meta.ID(), gc.Equals, id)
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
	c.Check(err, gc.IsNil)
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

func (s *WrapperSuite) TestFileStorageAdd(c *gc.C) {
	original := s.metadata()
	data := bytes.NewBufferString("spam")
	id, err := s.stor.Add(original, data)
	c.Check(err, gc.IsNil)

	meta, file, err := s.stor.Get(id)
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, original)
	c.Check(string(content), gc.Equals, "spam")
	c.Check(meta.Stored(), gc.Equals, true)
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

type fakeRawStorage struct {
	err string
}

func (s *fakeRawStorage) Error() string {
	return s.err
}

func (s *fakeRawStorage) File(string) (io.ReadCloser, error) {
	return nil, s
}

func (s *fakeRawStorage) AddFile(string, io.Reader, int64) error {
	return s
}

func (s *fakeRawStorage) RemoveFile(string) error {
	return s
}

func (s *WrapperSuite) TestFileStorageAddFileFailureDropsMetadata(c *gc.C) {
	raw := &fakeRawStorage{"error!"}
	stor := filestorage.NewFileStorage(s.metastor, raw)
	original := s.metadata()
	_, err := stor.Add(original, &bytes.Buffer{})

	c.Check(err, gc.ErrorMatches, "error!")

	metalist, metaErr := s.metastor.ListMetadata()
	c.Assert(metaErr, gc.IsNil)
	c.Check(metalist, gc.HasLen, 0)
	c.Check(original.ID(), gc.Equals, "")
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
