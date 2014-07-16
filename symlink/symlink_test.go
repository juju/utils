// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package symlink_test

import (
	"os"
	"path/filepath"
	"testing"

	gc "launchpad.net/gocheck"

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
