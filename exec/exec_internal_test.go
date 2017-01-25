// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type execSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&execSuite{})

func (*execSuite) TestShellAndArgsNoUserSpecified(c *gc.C) {
	if runtime.GOOS == "windows" {
		c.Skip("non-windows only test")
	}

	dir := c.MkDir()
	stat, err := os.Stat(dir)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(stat.Mode().Perm(), gc.Equals, os.FileMode(0700))

	cmd, args, err := shellAndArgs(dir, "env", "")
	c.Assert(err, jc.ErrorIsNil)

	scriptFile := filepath.Join(dir, "script.sh")

	c.Assert(cmd, gc.Equals, "/bin/bash")
	c.Assert(args, jc.DeepEquals, []string{scriptFile})
}

func (*execSuite) TestShellAndArgsAsUser(c *gc.C) {
	if runtime.GOOS == "windows" {
		c.Skip("non-windows only test")
	}

	dir := c.MkDir()
	stat, err := os.Stat(dir)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(stat.Mode().Perm(), gc.Equals, os.FileMode(0700))

	cmd, args, err := shellAndArgs(dir, "env", "ubuntu")
	c.Assert(err, jc.ErrorIsNil)

	scriptFile := filepath.Join(dir, "script.sh")

	c.Assert(cmd, gc.Equals, "/bin/su")
	command := "/bin/bash " + scriptFile
	c.Assert(args, jc.DeepEquals, []string{"ubuntu", "--login", "--command", command})

	// The directory is now readable by everyone.
	stat, err = os.Stat(dir)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(stat.Mode().Perm(), gc.Equals, os.FileMode(0755))
	// And the file is world readable
	stat, err = os.Stat(scriptFile)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(stat.Mode().Perm(), gc.Equals, os.FileMode(0644))
}
