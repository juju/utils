// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package manager_test

import (
	"os"
	"os/exec"

	"github.com/juju/testing"
	"github.com/juju/utils"
	"github.com/juju/utils/packaging/manager"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(&UtilsSuite{})

type UtilsSuite struct {
	testing.IsolationSuite
}

func (s *UtilsSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
}

func (s *UtilsSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
}

func (s *UtilsSuite) TearDownTest(c *gc.C) {
	s.IsolationSuite.TearDownTest(c)
}

func (s *UtilsSuite) TearDownSuite(c *gc.C) {
	s.IsolationSuite.TearDownSuite(c)
}

type mockExitStatuser int

func (es mockExitStatuser) ExitStatus() int {
	return int(es)
}

func (s *UtilsSuite) TestRunCommandWithRetryDoesNotCallCombinedOutputTwice(c *gc.C) {
	const minRetries = 3
	var calls int
	state := os.ProcessState{}
	cmdError := &exec.ExitError{&state}
	s.PatchValue(&manager.AttemptStrategy, utils.AttemptStrategy{Min: minRetries})
	s.PatchValue(&manager.ProcessStateSys, func(*os.ProcessState) interface{} {
		return mockExitStatuser(100) // retry each time.
	})
	s.PatchValue(&manager.CommandOutput, func(cmd *exec.Cmd) ([]byte, error) {
		calls++
		// Replace the command path and args so it's a no-op.
		cmd.Path = ""
		cmd.Args = []string{"version"}
		// Call the real cmd.CombinedOutput to simulate better what
		// happens in production. See also http://pad.lv/1394524.
		output, err := cmd.CombinedOutput()
		if _, ok := err.(*exec.Error); err != nil && !ok {
			c.Check(err, gc.ErrorMatches, "exec: Stdout already set")
			c.Fatalf("CommandOutput called twice unexpectedly")
		}
		return output, cmdError
	})

	apt := manager.NewAptPackageManager()

	err := apt.Install(testedPackageName)
	c.Check(err, gc.ErrorMatches, "packaging command failed: exit status.*")
	c.Check(calls, gc.Equals, minRetries)

	// reset calls and re-test for Yum calls:
	calls = 0
	yum := manager.NewYumPackageManager()
	err = yum.Install(testedPackageName)
	c.Check(err, gc.ErrorMatches, "packaging command failed: exit status.*")
	c.Check(calls, gc.Equals, minRetries)
}
