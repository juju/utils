// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filepath_test

import (
	gofilepath "path/filepath"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filepath"
)

var _ = gc.Suite(&unixSuite{})
var _ = gc.Suite(&unixThinWrapperSuite{})

type unixBaseSuite struct {
	testing.IsolationSuite

	path     string
	renderer *filepath.UnixRenderer
}

func (s *unixBaseSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.path = "/a/b/c.xyz"
	s.renderer = &filepath.UnixRenderer{}
}

func (s *unixBaseSuite) matchesRuntime() bool {
	return gofilepath.Separator == filepath.UnixSeparator
}

type unixSuite struct {
	unixBaseSuite
}

func (s unixSuite) TestIsAbs(c *gc.C) {
	isAbs := s.renderer.IsAbs(s.path)

	c.Check(isAbs, jc.IsTrue)
	if s.matchesRuntime() {
		c.Check(isAbs, gc.Equals, gofilepath.IsAbs(s.path))
	}
}

func (s unixSuite) TestSplitList(c *gc.C) {
	list := s.renderer.SplitList("/a:b:/c/d")

	c.Check(list, jc.DeepEquals, []string{"/a", "b", "/c/d"})
	if s.matchesRuntime() {
		golist := gofilepath.SplitList("/a:b:/c/d")
		c.Check(list, jc.DeepEquals, golist)
	}
}

func (s unixSuite) TestVolumeName(c *gc.C) {
	volumeName := s.renderer.VolumeName(s.path)

	c.Check(volumeName, gc.Equals, "")
}

func (s unixSuite) TestNormCaseLower(c *gc.C) {
	normalized := s.renderer.NormCase("spam")

	c.Check(normalized, gc.Equals, "spam")
}

func (s unixSuite) TestNormCaseUpper(c *gc.C) {
	normalized := s.renderer.NormCase("SPAM")

	c.Check(normalized, gc.Equals, "SPAM")
}

func (s unixSuite) TestNormCaseMixed(c *gc.C) {
	normalized := s.renderer.NormCase("sPaM")

	c.Check(normalized, gc.Equals, "sPaM")
}

func (s unixSuite) TestNormCaseCapitalized(c *gc.C) {
	normalized := s.renderer.NormCase("Spam")

	c.Check(normalized, gc.Equals, "Spam")
}

func (s unixSuite) TestNormCasePunctuation(c *gc.C) {
	normalized := s.renderer.NormCase("spam-eggs.ext")

	c.Check(normalized, gc.Equals, "spam-eggs.ext")
}

func (s unixSuite) TestSplitSuffix(c *gc.C) {
	// This is just a sanity check. The splitSuffix tests are more
	// comprehensive.
	path, suffix := s.renderer.SplitSuffix("spam.ext")

	c.Check(path, gc.Equals, "spam")
	c.Check(suffix, gc.Equals, ".ext")
}

// unixThinWrapperSuite contains test methods for UnixRenderer methods
// that are just thin wrappers around the corresponding helpers in the
// filepath package. As such the test coverage is minimal (more of a
// sanity check).
type unixThinWrapperSuite struct {
	unixBaseSuite
}

func (s unixThinWrapperSuite) TestBase(c *gc.C) {
	path := s.renderer.Base(s.path)

	c.Check(path, gc.Equals, "c.xyz")
	if s.matchesRuntime() {
		gopath := gofilepath.Base(s.path)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s unixThinWrapperSuite) TestClean(c *gc.C) {
	// TODO(ericsnow) Add more cases.
	originals := map[string]string{
		s.path: s.path,
	}
	for original, expected := range originals {
		c.Logf("checking %q", original)
		path := s.renderer.Clean(original)

		c.Check(path, gc.Equals, expected)
		if s.matchesRuntime() {
			gopath := gofilepath.Clean(original)
			c.Check(path, gc.Equals, gopath)
		}
	}
}

func (s unixThinWrapperSuite) TestDir(c *gc.C) {
	path := s.renderer.Dir(s.path)

	c.Check(path, gc.Equals, "/a/b")
	if s.matchesRuntime() {
		gopath := gofilepath.Dir(s.path)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s unixThinWrapperSuite) TestExt(c *gc.C) {
	ext := s.renderer.Ext(s.path)

	c.Check(ext, gc.Equals, ".xyz")
	if s.matchesRuntime() {
		goext := gofilepath.Ext(s.path)
		c.Check(ext, gc.Equals, goext)
	}
}

func (s unixThinWrapperSuite) TestFromSlash(c *gc.C) {
	original := "/a/b/c.xyz"
	path := s.renderer.FromSlash(original)

	c.Check(path, gc.Equals, s.path)
	if s.matchesRuntime() {
		gopath := gofilepath.FromSlash(original)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s unixThinWrapperSuite) TestJoin(c *gc.C) {
	path := s.renderer.Join("a", "b", "c.xyz")

	c.Check(path, gc.Equals, s.path[1:])
	if s.matchesRuntime() {
		gopath := gofilepath.Join("a", "b", "c.xyz")
		c.Check(path, gc.Equals, gopath)
	}
}

func (s unixThinWrapperSuite) TestSplit(c *gc.C) {
	dir, base := s.renderer.Split(s.path)

	c.Check(dir, gc.Equals, "/a/b/")
	c.Check(base, gc.Equals, "c.xyz")
	if s.matchesRuntime() {
		godir, gobase := gofilepath.Split(s.path)
		c.Check(dir, gc.Equals, godir)
		c.Check(base, gc.Equals, gobase)
	}
}

func (s unixThinWrapperSuite) TestToSlash(c *gc.C) {
	path := s.renderer.ToSlash(s.path)

	c.Check(path, gc.Equals, "/a/b/c.xyz")
	if s.matchesRuntime() {
		gopath := gofilepath.ToSlash(s.path)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s unixThinWrapperSuite) TestMatchTrue(c *gc.C) {
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
		matched, err := s.renderer.Match(pattern, name)
		c.Assert(err, jc.ErrorIsNil)

		c.Check(matched, jc.IsTrue)
		if s.matchesRuntime() {
			gomatched, err := gofilepath.Match(pattern, name)
			c.Assert(err, jc.ErrorIsNil)
			c.Check(matched, gc.Equals, gomatched)
		}
	}
}

func (s unixThinWrapperSuite) TestMatchFalse(c *gc.C) {
	tests := map[string]string{
		"abc": "xyz",
		"":    "abc",
		"a*c": "a",
		"?":   "",
		"a?c": "ac",
	}
	for pattern, name := range tests {
		c.Logf("- checking pattern %q against %q -", pattern, name)
		matched, err := s.renderer.Match(pattern, name)
		c.Assert(err, jc.ErrorIsNil)

		c.Check(matched, jc.IsFalse)
		if s.matchesRuntime() {
			gomatched, err := gofilepath.Match(pattern, name)
			c.Assert(err, jc.ErrorIsNil)
			c.Check(matched, gc.Equals, gomatched)
		}
	}
}

func (s unixThinWrapperSuite) TestMatchBadPattern(c *gc.C) {
	tests := map[string]string{
		"ab[":    "abc",
		"ab[-c]": "abc",
		"ab[]":   "abc",
	}
	for pattern, name := range tests {
		c.Logf("- checking pattern %q against %q -", pattern, name)
		_, err := s.renderer.Match(pattern, name)

		c.Check(err, gc.Equals, gofilepath.ErrBadPattern)
		if s.matchesRuntime() {
			_, goerr := gofilepath.Match(pattern, name)
			c.Check(err, gc.Equals, goerr)
		}
	}
}
