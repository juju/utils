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

var _ = gc.Suite(&windowsSuite{})
var _ = gc.Suite(&windowsThinWrapperSuite{})

type windowsBaseSuite struct {
	testing.IsolationSuite

	path     string
	renderer *filepath.WindowsRenderer
}

func (s *windowsBaseSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.path = `c:\a\b\c.xyz`
	s.renderer = &filepath.WindowsRenderer{}
}

func (s *windowsBaseSuite) matchesRuntime() bool {
	return gofilepath.Separator == filepath.WindowsSeparator
}

type windowsSuite struct {
	windowsBaseSuite
}

func (s windowsSuite) TestIsAbs(c *gc.C) {
	isAbs := s.renderer.IsAbs(s.path)

	c.Check(isAbs, jc.IsTrue)
	if s.matchesRuntime() {
		c.Check(isAbs, gc.Equals, gofilepath.IsAbs(s.path))
	}
}

func (s windowsSuite) TestSplitList(c *gc.C) {
	list := s.renderer.SplitList(`\a;b;\c\d`)

	c.Check(list, jc.DeepEquals, []string{`\a`, "b", `\c\d`})
	if s.matchesRuntime() {
		golist := gofilepath.SplitList(`\a;b;\c\d`)
		c.Check(list, jc.DeepEquals, golist)
	}
}

func (s windowsSuite) TestVolumeName(c *gc.C) {
	volumeName := s.renderer.VolumeName(s.path)

	c.Check(volumeName, gc.Equals, "c:")
	if s.matchesRuntime() {
		goresult := gofilepath.VolumeName(s.path)
		c.Check(volumeName, gc.Equals, goresult)
	}
}

func (s windowsSuite) TestNormCaseLower(c *gc.C) {
	normalized := s.renderer.NormCase("spam")

	c.Check(normalized, gc.Equals, "spam")
}

func (s windowsSuite) TestNormCaseUpper(c *gc.C) {
	normalized := s.renderer.NormCase("SPAM")

	c.Check(normalized, gc.Equals, "spam")
}

func (s windowsSuite) TestNormCaseMixed(c *gc.C) {
	normalized := s.renderer.NormCase("sPaM")

	c.Check(normalized, gc.Equals, "spam")
}

func (s windowsSuite) TestNormCaseCapitalized(c *gc.C) {
	normalized := s.renderer.NormCase("Spam")

	c.Check(normalized, gc.Equals, "spam")
}

func (s windowsSuite) TestNormCasePunctuation(c *gc.C) {
	normalized := s.renderer.NormCase("spam-eggs.ext")

	c.Check(normalized, gc.Equals, "spam-eggs.ext")
}

func (s windowsSuite) TestSplitSuffix(c *gc.C) {
	// This is just a sanity check. The splitSuffix tests are more
	// comprehensive.
	path, suffix := s.renderer.SplitSuffix("spam.ext")

	c.Check(path, gc.Equals, "spam")
	c.Check(suffix, gc.Equals, ".ext")
}

// windowsThinWrapperSuite contains test methods for WindowsRenderer methods
// that are just thin wrappers around the corresponding helpers in the
// filepath package. As such the test coverage is minimal (more of a
// sanity check).
type windowsThinWrapperSuite struct {
	windowsBaseSuite
}

func (s windowsThinWrapperSuite) TestBase(c *gc.C) {
	path := s.renderer.Base(s.path)

	c.Check(path, gc.Equals, "c.xyz")
	if s.matchesRuntime() {
		gopath := gofilepath.Base(s.path)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s windowsThinWrapperSuite) TestClean(c *gc.C) {
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

func (s windowsThinWrapperSuite) TestDir(c *gc.C) {
	path := s.renderer.Dir(s.path)

	c.Check(path, gc.Equals, `c:\a\b`)
	if s.matchesRuntime() {
		gopath := gofilepath.Dir(s.path)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s windowsThinWrapperSuite) TestExt(c *gc.C) {
	ext := s.renderer.Ext(s.path)

	c.Check(ext, gc.Equals, ".xyz")
	if s.matchesRuntime() {
		goext := gofilepath.Ext(s.path)
		c.Check(ext, gc.Equals, goext)
	}
}

func (s windowsThinWrapperSuite) TestFromSlash(c *gc.C) {
	original := "/a/b/c.xyz"
	path := s.renderer.FromSlash(original)

	c.Check(path, gc.Equals, s.path[2:])
	if s.matchesRuntime() {
		gopath := gofilepath.FromSlash(original)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s windowsThinWrapperSuite) TestJoin(c *gc.C) {
	path := s.renderer.Join("a", "b", "c.xyz")

	c.Check(path, gc.Equals, s.path[3:])
	if s.matchesRuntime() {
		gopath := gofilepath.Join("a", "b", "c.xyz")
		c.Check(path, gc.Equals, gopath)
	}
}

func (s windowsThinWrapperSuite) TestSplit(c *gc.C) {
	dir, base := s.renderer.Split(s.path)

	c.Check(dir, gc.Equals, `c:\a\b\`)
	c.Check(base, gc.Equals, "c.xyz")
	if s.matchesRuntime() {
		godir, gobase := gofilepath.Split(s.path)
		c.Check(dir, gc.Equals, godir)
		c.Check(base, gc.Equals, gobase)
	}
}

func (s windowsThinWrapperSuite) TestToSlash(c *gc.C) {
	path := s.renderer.ToSlash(s.path)

	c.Check(path, gc.Equals, "c:/a/b/c.xyz")
	if s.matchesRuntime() {
		gopath := gofilepath.ToSlash(s.path)
		c.Check(path, gc.Equals, gopath)
	}
}

func (s windowsThinWrapperSuite) TestMatchTrue(c *gc.C) {
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

func (s windowsThinWrapperSuite) TestMatchFalse(c *gc.C) {
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

func (s windowsThinWrapperSuite) TestMatchBadPattern(c *gc.C) {
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
