// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package symlink_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
	"github.com/juju/utils/symlink"
)

type SymlinkSuite struct{}

var _ = gc.Suite(&SymlinkSuite{})

func Test(t *testing.T) {
	gc.TestingT(t)
}

func (*SymlinkSuite) TestReplace(c *gc.C) {
	target, err := symlink.GetLongPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)
	target_second, err := symlink.GetLongPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)
	link := filepath.Join(target, "link")

	_, err = os.Stat(target)
	c.Assert(err, gc.IsNil)
	_, err = os.Stat(target_second)
	c.Assert(err, gc.IsNil)

	err = symlink.New(target, link)
	c.Assert(err, gc.IsNil)

	link_target, err := symlink.Read(link)
	c.Assert(err, gc.IsNil)
	c.Assert(link_target, gc.Equals, filepath.FromSlash(target))

	err = symlink.Replace(link, target_second)
	c.Assert(err, gc.IsNil)

	link_target, err = symlink.Read(link)
	c.Assert(err, gc.IsNil)
	c.Assert(link_target, gc.Equals, filepath.FromSlash(target_second))
}

func (*SymlinkSuite) TestIsSymlinkFile(c *gc.C) {
	dir, err := symlink.GetLongPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)

	target := filepath.Join(dir, "file")
	err = ioutil.WriteFile(target, []byte("TOP SECRET"), 0644)
	c.Assert(err, gc.IsNil)

	link := filepath.Join(dir, "link")

	_, err = os.Stat(target)
	c.Assert(err, gc.IsNil)

	err = symlink.New(target, link)
	c.Assert(err, gc.IsNil)

	isSymlink, err := symlink.IsSymlink(link)
	c.Assert(err, gc.IsNil)
	c.Assert(isSymlink, jc.IsTrue)
}

func (*SymlinkSuite) TestIsSymlinkFolder(c *gc.C) {
	target, err := symlink.GetLongPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)

	link := filepath.Join(target, "link")

	_, err = os.Stat(target)
	c.Assert(err, gc.IsNil)

	err = symlink.New(target, link)
	c.Assert(err, gc.IsNil)

	isSymlink, err := symlink.IsSymlink(link)
	c.Assert(err, gc.IsNil)
	c.Assert(isSymlink, jc.IsTrue)
}

func (*SymlinkSuite) TestIsSymlinkFalseFile(c *gc.C) {
	dir := c.MkDir()

	target := filepath.Join(dir, "file")
	err := ioutil.WriteFile(target, []byte("TOP SECRET"), 0644)
	c.Assert(err, gc.IsNil)

	_, err = os.Stat(target)
	c.Assert(err, gc.IsNil)

	isSymlink, err := symlink.IsSymlink(target)
	c.Assert(err, gc.IsNil)
	c.Assert(isSymlink, jc.IsFalse)
}

func (*SymlinkSuite) TestIsSymlinkFalseFolder(c *gc.C) {
	target, err := symlink.GetLongPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)

	_, err = os.Stat(target)
	c.Assert(err, gc.IsNil)

	isSymlink, err := symlink.IsSymlink(target)
	c.Assert(err, gc.IsNil)
	c.Assert(isSymlink, jc.IsFalse)
}

func (*SymlinkSuite) TestIsSymlinkFileDoesNotExist(c *gc.C) {
	dir := c.MkDir()

	target := filepath.Join(dir, "file")

	isSymlink, err := symlink.IsSymlink(target)
	c.Assert(err, gc.ErrorMatches, ".*"+utils.NoSuchFileErrRegexp)
	c.Assert(isSymlink, jc.IsFalse)
}
