// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"path/filepath"
	"strings"
	"syscall"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

type execSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&execSuite{})

// longPath is copied over from the symlink package. This should be removed
// if we add it to gc or in some other convenience package
func longPath(path string) ([]uint16, error) {
	pathp, err := syscall.UTF16FromString(path)
	if err != nil {
		return nil, err
	}

	longp := pathp
	n, err := syscall.GetLongPathName(&pathp[0], &longp[0], uint32(len(longp)))
	if err != nil {
		return nil, err
	}
	if n > uint32(len(longp)) {
		longp = make([]uint16, n)
		n, err = syscall.GetLongPathName(&pathp[0], &longp[0], uint32(len(longp)))
		if err != nil {
			return nil, err
		}
	}
	longp = longp[:n]

	return longp, nil
}

func longPathAsString(path string) (string, error) {
	longp, err := longPath(path)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(longp), nil
}

func (*execSuite) TestRunCommands(c *gc.C) {
	newDir, err := longPathAsString(c.MkDir())
	c.Assert(err, gc.IsNil)
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
			commands: "echo 'testing stdout'",
			stdout:   "testing stdout\r\n",
		}, {
			message:  "test stderr capture",
			commands: "Write-Error 'testing stderr'",
			stderr:   "testing stderr\r\n",
		}, {
			message:  "test return code",
			commands: "exit 42",
			code:     42,
		}, {
			message:    "test working dir",
			commands:   "(pwd).Path",
			workingDir: newDir,
			stdout:     filepath.FromSlash(newDir) + "\r\n",
		}, {
			message:     "test environment",
			commands:    "echo $env:OMG_IT_WORKS",
			environment: []string{"OMG_IT_WORKS=like magic"},
			stdout:      "like magic\r\n",
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
		c.Assert(string(result.Stderr), jc.Contains, test.stderr)
		c.Assert(result.Code, gc.Equals, test.code)

		err = params.Run()
		c.Assert(err, gc.IsNil)
		c.Assert(params.Process(), gc.Not(gc.IsNil))
		result, err = params.Wait()
		c.Assert(err, gc.IsNil)
		c.Assert(string(result.Stdout), gc.Equals, test.stdout)
		c.Assert(string(result.Stderr), jc.Contains, test.stderr)
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
	stderr := strings.Replace(string(result.Stderr), "\r\n", "", -1)
	c.Assert(stderr, jc.Contains, "is not recognized as the name of a cmdlet")
	// 1 is returned by RunCommands when powershell commands throw exceptions
	c.Assert(result.Code, gc.Equals, 1)
}
