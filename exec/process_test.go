// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"bytes"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&ProcSuite{})

type ProcSuite struct {
	BaseSuite
}

func (s *ProcSuite) TestCommand(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	expected := exec.CommandInfo{
		Path: "spam",
		Args: []string{"spam", "eggs"},
		Context: exec.Context{
			Env: []string{"X=y"},
			Dir: "/x/y/z",
			Stdio: exec.Stdio{
				In:  &stdin,
				Out: &stdout,
				Err: &stderr,
			},
		},
	}
	p := exec.Proc{Info: expected}

	info := p.Command()

	c.Check(info, jc.DeepEquals, expected)
}
