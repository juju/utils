// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package symlink_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	gc "gopkg.in/check.v1"

	"github.com/juju/utils/symlink"
)

func (*SymlinkSuite) TestLongPath(c *gc.C) {
	programFiles := `C:\PROGRA~1`
	longProg := `C:\Program Files`
	target, err := symlink.GetLongPathAsString(programFiles)
	c.Assert(err, gc.IsNil)
	c.Assert(target, gc.Equals, longProg)
}

func (*SymlinkSuite) TestCreateSymLink(c *gc.C) {
	target, err := symlink.GetLongPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)

	link := filepath.Join(target, "link")

	_, err = os.Stat(target)
	c.Assert(err, gc.IsNil)

	err = symlink.New(target, link)
	c.Assert(err, gc.IsNil)

	link, err = symlink.Read(link)
	c.Assert(err, gc.IsNil)
	c.Assert(link, gc.Equals, filepath.FromSlash(target))
}

func (*SymlinkSuite) TestReadData(c *gc.C) {
	dir := c.MkDir()
	sub := filepath.Join(dir, "sub")

	err := os.Mkdir(sub, 0700)
	c.Assert(err, gc.IsNil)

	oldname := filepath.Join(sub, "foo")
	data := []byte("data")

	err = ioutil.WriteFile(oldname, data, 0644)
	c.Assert(err, gc.IsNil)

	newname := filepath.Join(dir, "bar")
	err = symlink.New(oldname, newname)
	c.Assert(err, gc.IsNil)

	b, err := ioutil.ReadFile(newname)
	c.Assert(err, gc.IsNil)

	c.Assert(string(b), gc.Equals, string(data))
}
