// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.
// +build !windows

package utils_test

import (
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type homeSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&homeSuite{})

func (s *homeSuite) TestHomeLinux(c *gc.C) {
	h := "/home/foo/bar"
	s.PatchEnvironment("HOME", h)
	c.Check(utils.Home(), gc.Equals, h)
}
