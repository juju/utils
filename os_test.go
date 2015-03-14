// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

var _ = gc.Suite(&osSuite{})

type osSuite struct {
	testing.IsolationSuite
}

func (osSuite) TestOSIsUnixKnown(c *gc.C) {
	for _, os := range utils.OSUnix {
		c.Logf("checking %q", os)
		isUnix := utils.OSIsUnix(os)

		c.Check(isUnix, jc.IsTrue)
	}
}

func (osSuite) TestOSIsUnixWindows(c *gc.C) {
	isUnix := utils.OSIsUnix("windows")

	c.Check(isUnix, jc.IsFalse)
}

func (osSuite) TestOSIsUnixUnknown(c *gc.C) {
	isUnix := utils.OSIsUnix("<unknown OS>")

	c.Check(isUnix, jc.IsFalse)
}
