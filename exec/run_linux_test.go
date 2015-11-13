// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

type execSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&execSuite{})

func (*execSuite) TestRunCommands(c *gc.C) {
	newDir := c.MkDir()

	for i, test := range []struct {
		message     string
		commands    string
		workingDir  string
		environment []string
		stdout      string
		stderr      string
		code        int
	}{
		{
			message:  "test stdout capture",
			commands: "echo testing stdout",
			stdout:   "testing stdout\n",
		}, {
			message:  "test stderr capture",
			commands: "echo testing stderr >&2",
			stderr:   "testing stderr\n",
		}, {
			message:  "test return code",
			commands: "exit 42",
			code:     42,
		}, {
			message:    "test working dir",
			commands:   "pwd",
			workingDir: newDir,
			stdout:     newDir + "\n",
		}, {
			message:     "test environment",
			commands:    "echo $OMG_IT_WORKS",
			environment: []string{"OMG_IT_WORKS=like magic"},
			stdout:      "like magic\n",
		}, {
			message:  "multiple commands",
			commands: "cat\necho 123",
			stdout:   "123\n",
		},
	} {
		c.Logf("%v: %s", i, test.message)

		params := exec.RunParams{
			Commands:    test.commands,
			WorkingDir:  test.workingDir,
			Environment: test.environment,
		}

		result, err := exec.RunCommands(params)
		c.Assert(err, gc.IsNil)
		c.Assert(string(result.Stdout), gc.Equals, test.stdout)
		c.Assert(string(result.Stderr), gc.Equals, test.stderr)
		c.Assert(result.Code, gc.Equals, test.code)

		err = params.Run()
		c.Assert(err, gc.IsNil)
		c.Assert(params.Process(), gc.Not(gc.IsNil))
		result, err = params.Wait()
		c.Assert(err, gc.IsNil)
		c.Assert(string(result.Stdout), gc.Equals, test.stdout)
		c.Assert(string(result.Stderr), gc.Equals, test.stderr)
		c.Assert(result.Code, gc.Equals, test.code)

	}
}

func (*execSuite) TestExecUnknownCommand(c *gc.C) {
	result, err := exec.RunCommands(
		exec.RunParams{
			Commands: "unknown-command",
		},
	)
	c.Assert(err, gc.IsNil)
	c.Assert(result.Stdout, gc.HasLen, 0)
	c.Assert(string(result.Stderr), jc.Contains, "unknown-command: command not found")
	// 127 is a special bash return code meaning command not found.
	c.Assert(result.Code, gc.Equals, 127)
}
