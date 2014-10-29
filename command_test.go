// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type EnvironmentPatcher interface {
	PatchEnvironment(name, value string)
}

func patchExecutable(patcher EnvironmentPatcher, dir, execName, script string) {
	patcher.PatchEnvironment("PATH", dir)
	filename := filepath.Join(dir, execName)
	ioutil.WriteFile(filename, []byte(script), 0755)
}

type commandSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&commandSuite{})

func (s *commandSuite) TestRunCommandCombinesOutput(c *gc.C) {
	var content string
	var cmdName string
	var expect string
	if runtime.GOOS != "windows" {
		content = `#!/bin/bash --norc
echo stdout
echo stderr 1>&2
`
		cmdName = "test-output"
		expect = "stdout\nstderr\n"
	} else {
		content = `@echo off
echo stdout
echo stderr 1>&2
`
		cmdName = "test-output.bat"
		expect = "stdout\r\nstderr \r\n"
	}
	patchExecutable(s, c.MkDir(), cmdName, content)
	output, err := utils.RunCommand("test-output")
	c.Assert(err, gc.IsNil)
	c.Assert(output, gc.Equals, expect)
}

func (s *commandSuite) TestRunCommandNonZeroExit(c *gc.C) {
	var content string
	var cmdName string
	var expect string
	if runtime.GOOS != "windows" {
		content = `#!/bin/bash --norc
echo stdout
exit 42
`
		cmdName = "test-output"
		expect = "stdout\n"
	} else {
		content = `@echo off
echo stdout
exit 42
`
		cmdName = "test-output.bat"
		expect = "stdout\r\n"
	}
	patchExecutable(s, c.MkDir(), cmdName, content)
	output, err := utils.RunCommand("test-output")
	c.Assert(err, gc.ErrorMatches, `exit status 42`)
	c.Assert(output, gc.Equals, expect)
}
