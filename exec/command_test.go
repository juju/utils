// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&CommandSuite{})

type CommandSuite struct {
	BaseSuite
}

func (s *CommandSuite) TestNewCommandOkay(c *gc.C) {
	expected := s.NewStubCommand()
	s.StubExec.ReturnCommand = expected

	cmd, err := exec.NewCommand(s.StubExec, "/x/y/z/spam", "--ham", "eggs")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(cmd, gc.Equals, expected)
	s.Stub.CheckCallNames(c, "Command")
	s.Stub.CheckCall(c, 0, "Command",
		exec.CommandInfo{
			Path: "/x/y/z/spam",
			Args: []string{
				"/x/y/z/spam",
				"--ham",
				"eggs",
			},
		},
	)
}

func (s *CommandSuite) TestNewCommandError(c *gc.C) {
	expected := s.NewStubCommand()
	s.StubExec.ReturnCommand = expected
	failure := s.SetFailure()

	_, err := exec.NewCommand(s.StubExec, "/x/y/z/spam", "--ham", "eggs")

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Command")
}

func (s *CommandSuite) TestNewCommandInfo(c *gc.C) {
	info := exec.NewCommandInfo("/x/y/z/spam", "--ham", "eggs")

	c.Check(info, jc.DeepEquals, exec.CommandInfo{
		Path: "/x/y/z/spam",
		Args: []string{
			"/x/y/z/spam",
			"--ham",
			"eggs",
		},
	})
}
