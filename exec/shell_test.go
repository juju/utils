// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&shellSuite{})

type shellSuite struct {
	BaseSuite
}

func (s *shellSuite) TestRunWithStdinStringOkay(c *gc.C) {
	var input string
	cmd := s.newStdioCommand(&input,
		"abc",
		"!xyz",
	)
	script := "do-something\ndo-more\ndone!"

	output, err := exec.RunWithStdinString(cmd, script)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(input, gc.Equals, script)
	c.Check(output, gc.Equals, "abc")
	s.Stub.CheckCallNames(c, "SetStdio", "Start", "Wait")
}

func (s *shellSuite) TestRunWithStdinStringError(c *gc.C) {
	var input string
	cmd := s.newStdioCommand(&input,
		"abc",
		"!xyz",
	)
	script := "do-something\ndo-more\ndone!"
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, nil, failure)

	_, err := exec.RunWithStdinString(cmd, script)

	c.Check(input, gc.Equals, script)
	c.Check(err, gc.ErrorMatches, ".*xyz.*")
	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio", "Start", "Wait")
}

func (s *shellSuite) TestRunBashScript(c *gc.C) {
	e := s.NewStubExec()
	var input string
	e.ReturnCommand = s.newStdioCommand(&input,
		"abc",
		"!xyz",
	)
	script := "do-something\ndo-more\ndone!"

	data, err := exec.RunBashScript(e, script)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(string(data), gc.Equals, "abc")
	s.Stub.CheckCallNames(c, "Command", "SetStdio", "Start", "Wait")
}

func (s *shellSuite) TestBashCommand(c *gc.C) {
	e := s.NewStubExec()
	expected := s.NewStubCommand()
	e.ReturnCommand = expected

	cmd, err := exec.BashCommand(e)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(cmd, gc.Equals, expected)
	s.Stub.CheckCallNames(c, "Command")
	s.Stub.CheckCall(c, 0, "Command", exec.CommandInfo{
		Path: "/bin/bash",
		Args: []string{"/bin/bash"},
	})
}
