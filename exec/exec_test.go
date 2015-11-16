// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&execSuite{})

type execSuite struct {
	BaseSuite
}

func (s *execSuite) TestLocal(c *gc.C) {
	c.Check(exec.Local, gc.FitsTypeOf, &exec.OSExec{})
}
