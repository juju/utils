package symlink_test

import (
	"io/ioutil"
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

func (*SymlinkSuite) TestCreateSymLink(c *gc.C) {
	target := c.MkDir()

	link := filepath.Join(target, "link")
	c.Logf("Making link %q to %q", link, target)

	_, err := os.Stat(target)
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
