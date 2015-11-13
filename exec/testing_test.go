// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
	exectesting "github.com/juju/utils/exec/testing"
)

type BaseSuite struct {
	testing.IsolationSuite
	exectesting.StubSuite
	exec.TestingExposer
}

func (s *BaseSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.StubSuite.SetUpTest(c)
}

func (s *BaseSuite) SetExecPIDs(e exec.Exec, pids ...int) {
	var processes []exec.Process
	for _, pid := range pids {
		process := s.NewStubProcess()
		process.ReturnPID = pid
		processes = append(processes, process)
	}
	s.SetExec(e, processes...)
}
