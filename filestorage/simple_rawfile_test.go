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

var _ = gc.Suite(&RawFileSuite{})

type RawFileSuite struct {
	testing.IsolationSuite
}

func (s *RawFileSuite) TestRawFileStorageNewRawFileStorage(c *gc.C) {
	stor, err := filestorage.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)

	c.Check(stor, gc.NotNil)
}

func (s *RawFileSuite) TestRawFileStorageFile(c *gc.C) {
	stor, err := filestorage.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)
	data := bytes.NewBufferString("spam")
	err = stor.AddFile("eggs", data, 4)
	c.Assert(err, gc.IsNil)

	file, err := stor.File("eggs")
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)
	c.Check(string(content), gc.Equals, "spam")
}

func (s *RawFileSuite) TestRawFileStorageAddFile(c *gc.C) {
	stor, err := filestorage.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)
	data := bytes.NewBufferString("spam")

	_, err = stor.File("eggs")
	c.Check(err, gc.NotNil)

	err = stor.AddFile("eggs", data, 4)
	c.Assert(err, gc.IsNil)
	file, err := stor.File("eggs")
	c.Assert(err, gc.IsNil)
	content, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)
	c.Check(string(content), gc.Equals, "spam")
}

func (s *RawFileSuite) TestRawFileStorageRemoveFile(c *gc.C) {
	stor, err := filestorage.NewRawFileStorage(c.MkDir())
	c.Assert(err, gc.IsNil)
	data := bytes.NewBufferString("spam")
	err = stor.AddFile("eggs", data, 4)
	c.Assert(err, gc.IsNil)

	err = stor.RemoveFile("eggs")
	c.Check(err, gc.IsNil)
	_, err = stor.File("eggs")
	c.Check(err, gc.NotNil)
}
