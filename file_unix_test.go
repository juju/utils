// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build !windows

package utils_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type unixFileSuite struct {
}

var _ = gc.Suite(&unixFileSuite{})

func (s *unixFileSuite) TestEnsureBaseDir(c *gc.C) {
	c.Assert(utils.EnsureBaseDir(`/a`, `/b/c`), gc.Equals, `/a/b/c`)
	c.Assert(utils.EnsureBaseDir(`/`, `/b/c`), gc.Equals, `/b/c`)
	c.Assert(utils.EnsureBaseDir(``, `/b/c`), gc.Equals, `/b/c`)
}
