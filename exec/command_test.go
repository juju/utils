// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
	exectesting "github.com/juju/utils/exec/testing"
)

var _ = gc.Suite(&commandSuite{})

type commandSuite struct {
	testing.IsolationSuite

	stub *testing.Stub
}

func (s *commandSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.stub = &testing.Stub{}
}

func (s *commandSuite) TestNewCommandOkay(c *gc.C) {
	e := exectesting.NewStubExec(s.stub)
	expected := exectesting.NewStubCommand(s.stub)
	e.ReturnCommand = expected

	cmd, err := exec.NewCommand(e, "/x/y/z/spam", "--ham", "eggs")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(cmd, gc.Equals, expected)
	s.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "Command",
		Args: []interface{}{
			exec.CommandInfo{
				Path: "/x/y/z/spam",
				Args: []string{
					"/x/y/z/spam",
					"--ham",
					"eggs",
				},
			},
		},
	}})
}

func (s *commandSuite) TestNewCommandError(c *gc.C) {
	e := exectesting.NewStubExec(s.stub)
	expected := exectesting.NewStubCommand(s.stub)
	e.ReturnCommand = expected
	failure := errors.New("<failure>")
	s.stub.SetErrors(failure)

	_, err := exec.NewCommand(e, "/x/y/z/spam", "--ham", "eggs")

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.stub.CheckCallNames(c, "Command")
}

func (s *commandSuite) TestNewCommandInfo(c *gc.C) {
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
