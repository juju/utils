package symlink_test

import (
	"log"
	"os"
	"path/filepath"

	gc "launchpad.net/gocheck"

	"github.com/juju/utils/symlink"
)

type PathSuite struct {
	Target string
	Link   string
}

var _ = gc.Suite(&PathSuite{})

func (s *PathSuite) SetUpTest(c *gc.C) {
	s.Target = c.MkDir()
	s.Link = "symlink"
}

func (s *PathSuite) TearDownTest(c *gc.C) {
	os.Remove(s.Link)
}

func (s *PathSuite) TestCreateSymLink(c *gc.C) {
	target := filepath.FromSlash(s.Target)
	target = filepath.FromSlash(target)

	err := symlink.New(target, s.Link)
	if err != nil {
		log.Print(err)
	}
	compare, err := symlink.Read(s.Link)
	if err != nil {
		log.Print(err)
	}

	c.Assert(err, gc.IsNil)
	c.Assert(err, gc.IsNil)
	c.Assert(compare, gc.Equals, target)
}
