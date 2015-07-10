// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filepath_test

import (
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filepath"
)

var _ = gc.Suite(&commonSuite{})

type commonSuite struct {
	testing.IsolationSuite
}

func (s commonSuite) TestSplitSuffixHasSuffix(c *gc.C) {
	path, suffix := filepath.SplitSuffix("spam.ext")

	c.Check(path, gc.Equals, "spam")
	c.Check(suffix, gc.Equals, ".ext")
}

func (s commonSuite) TestSplitSuffixNoSuffix(c *gc.C) {
	path, suffix := filepath.SplitSuffix("spam")

	c.Check(path, gc.Equals, "spam")
	c.Check(suffix, gc.Equals, "")
}

func (s commonSuite) TestSplitSuffixEmpty(c *gc.C) {
	path, suffix := filepath.SplitSuffix("")

	c.Check(path, gc.Equals, "")
	c.Check(suffix, gc.Equals, "")
}

func (s commonSuite) TestSplitSuffixDotFilePlain(c *gc.C) {
	path, suffix := filepath.SplitSuffix(".spam")

	c.Check(path, gc.Equals, ".spam")
	c.Check(suffix, gc.Equals, "")
}

func (s commonSuite) TestSplitSuffixDofileWithSuffix(c *gc.C) {
	path, suffix := filepath.SplitSuffix(".spam.ext")

	c.Check(path, gc.Equals, ".spam")
	c.Check(suffix, gc.Equals, ".ext")
}
