// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filepath_test

import (
	gofilepath "path/filepath"
	"runtime"
	"strings"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filepath"
)

// The tests here are mostly just sanity checks against the behavior
// of the stdlib path/filepath. We are not trying for high coverage levels.

type stdlibSuite struct {
	testing.IsolationSuite

	path       string
	volumeName func(string) string
}

var _ = gc.Suite(&stdlibSuite{})

func (s *stdlibSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	switch runtime.GOOS {
	case "windows":
		s.path = `C:\a\b\c.xyz`
		s.volumeName = func(path string) string {
			return "C:"
		}
	default:
		s.path = "/a/b/c.xyz"
		s.volumeName = func(string) string { return "" }
	}
}

func (s stdlibSuite) TestBase(c *gc.C) {
	path := filepath.Base(gofilepath.Separator, s.volumeName, s.path)

	gopath := gofilepath.Base(s.path)
	c.Check(path, gc.Equals, gopath)
	c.Check(path, gc.Equals, "c.xyz")
}

func (s stdlibSuite) TestClean(c *gc.C) {
	// TODO(ericsnow) Add more cases.
	originals := map[string]string{
		s.path: s.path,
	}
	for original, expected := range originals {
		c.Logf("checking %q", original)
		path := filepath.Clean(gofilepath.Separator, s.volumeName, original)

		gopath := gofilepath.Clean(original)
		c.Check(path, gc.Equals, gopath)
		c.Check(path, gc.Equals, expected)
	}
}

func (s stdlibSuite) TestDir(c *gc.C) {
	path := filepath.Dir(gofilepath.Separator, s.volumeName, s.path)

	gopath := gofilepath.Dir(s.path)
	c.Check(path, gc.Equals, gopath)
	switch runtime.GOOS {
	case "windows":
		c.Check(path, gc.Equals, `\a\b`)
	default:
		c.Check(path, gc.Equals, "/a/b")
	}
}

func (s stdlibSuite) TestExt(c *gc.C) {
	ext := filepath.Ext(gofilepath.Separator, s.path)

	goext := gofilepath.Ext(s.path)
	c.Check(ext, gc.Equals, goext)
	c.Check(ext, gc.Equals, ".xyz")
}

func (s stdlibSuite) TestFromSlash(c *gc.C) {
	original := "/a/b/c.xyz"
	path := filepath.FromSlash(gofilepath.Separator, original)

	gopath := gofilepath.FromSlash(original)
	c.Check(path, gc.Equals, gopath)
	c.Check(path, gc.Equals, s.path)
}

func (s stdlibSuite) TestJoin(c *gc.C) {
	path := filepath.Join(gofilepath.Separator, s.volumeName, "a", "b", "c.xyz")

	gopath := gofilepath.Join("a", "b", "c.xyz")
	c.Check(path, gc.Equals, gopath)
	expected := s.path[strings.Index(s.path, string(gofilepath.Separator))+1:]
	c.Check(path, gc.Equals, expected)
}

func (s stdlibSuite) TestSplit(c *gc.C) {
	dir, base := filepath.Split(gofilepath.Separator, s.volumeName, s.path)

	godir, gobase := gofilepath.Split(s.path)
	c.Check(dir, gc.Equals, godir)
	c.Check(base, gc.Equals, gobase)
	switch runtime.GOOS {
	case "windows":
		c.Check(dir, gc.Equals, `\a\b\`)
	default:
		c.Check(dir, gc.Equals, "/a/b/")
	}
	c.Check(base, gc.Equals, "c.xyz")
}

func (s stdlibSuite) TestToSlash(c *gc.C) {
	path := filepath.ToSlash(gofilepath.Separator, s.path)

	gopath := gofilepath.ToSlash(s.path)
	c.Check(path, gc.Equals, gopath)
	c.Check(path, gc.Equals, "/a/b/c.xyz")
}

func (s stdlibSuite) TestMatchTrue(c *gc.C) {
	tests := map[string]string{
		"abc":   "abc",
		"ab[c]": "abc",
		"":      "",
		"*":     "abc",
		"a*c":   "abc",
		"?":     "a",
		"a?c":   "abc",
	}
	for pattern, name := range tests {
		c.Logf("- checking pattern %q against %q -", pattern, name)
		matched, err := filepath.Match(gofilepath.Separator, pattern, name)
		c.Assert(err, jc.ErrorIsNil)

		gomatched, err := gofilepath.Match(pattern, name)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(matched, gc.Equals, gomatched)
		c.Check(matched, jc.IsTrue)
	}
}

func (s stdlibSuite) TestMatchFalse(c *gc.C) {
	tests := map[string]string{
		"abc": "xyz",
		"":    "abc",
		"a*c": "a",
		"?":   "",
		"a?c": "ac",
	}
	for pattern, name := range tests {
		c.Logf("- checking pattern %q against %q -", pattern, name)
		matched, err := filepath.Match(gofilepath.Separator, pattern, name)
		c.Assert(err, jc.ErrorIsNil)

		gomatched, err := gofilepath.Match(pattern, name)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(matched, gc.Equals, gomatched)
		c.Check(matched, jc.IsFalse)
	}
}

func (s stdlibSuite) TestMatchBadPattern(c *gc.C) {
	tests := map[string]string{
		"ab[":    "abc",
		"ab[-c]": "abc",
		"ab[]":   "abc",
	}
	for pattern, name := range tests {
		c.Logf("- checking pattern %q against %q -", pattern, name)
		_, err := filepath.Match(gofilepath.Separator, pattern, name)

		_, goerr := gofilepath.Match(pattern, name)
		c.Check(err, gc.Equals, goerr)
		c.Check(err, gc.Equals, gofilepath.ErrBadPattern)
	}
}
