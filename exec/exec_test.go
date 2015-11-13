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

var _ = gc.Suite(&execSuite{})

type execSuite struct {
	testing.IsolationSuite

	stub *testing.Stub
	cmd  *exectesting.StubCommand
}

func (s *execSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.stub = &testing.Stub{}
}

func (s *execSuite) setProcs(e exec.Exec, pids ...int) {
	var processes []exec.Process
	for _, pid := range pids {
		process := exectesting.NewStubProcess(s.stub)
		process.ReturnPID = pid
		processes = append(processes, process)
	}
	exec.TestingSetExec(e, processes...)
}

func (s *execSuite) command(info exec.CommandInfo) (exec.Command, error) {
	s.stub.AddCall("command", info)
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.cmd, nil
}

func (s *execSuite) TestNewExecOkay(c *gc.C) {
	e := exec.NewExec(s.command)
	processes := exec.TestingExposeExec(e)

	c.Check(e, gc.NotNil)
	c.Check(processes, gc.HasLen, 0)
}

func (s *execSuite) TestNewExecNoFunc(c *gc.C) {
	e := exec.NewExec(nil)
	processes := exec.TestingExposeExec(e)

	c.Check(e, gc.NotNil)
	c.Check(processes, gc.HasLen, 0)
}

func (s *execSuite) TestCommandOkay(c *gc.C) {
	s.cmd = exectesting.NewStubCommand(s.stub)
	e := exec.NewExec(s.command)
	info := exec.CommandInfo{
		Path: "spam",
	}

	cmd, err := e.Command(info)
	c.Assert(err, jc.ErrorIsNil)
	raw, processes := exec.TestingExposeExecCommand(cmd)

	c.Check(raw, gc.Equals, s.cmd)
	c.Check(processes, gc.HasLen, 0)
	s.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "command",
		Args: []interface{}{
			info,
		},
	}})
}

func (s *execSuite) TestCommandNilCommand(c *gc.C) {
	s.cmd = nil
	e := exec.NewExec(s.command)
	info := exec.CommandInfo{
		Path: "spam",
	}

	cmd, err := e.Command(info)
	c.Assert(err, jc.ErrorIsNil)
	raw, _ := exec.TestingExposeExecCommand(cmd)

	c.Check(raw, gc.Equals, s.cmd)
	s.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "command",
		Args: []interface{}{
			info,
		},
	}})
}

func (s *execSuite) TestCommandError(c *gc.C) {
	s.cmd = exectesting.NewStubCommand(s.stub)
	failure := errors.New("<failure>")
	s.stub.SetErrors(failure)
	e := exec.NewExec(s.command)
	info := exec.CommandInfo{
		Path: "spam",
	}

	_, err := e.Command(info)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "command",
		Args: []interface{}{
			info,
		},
	}})
}

func (s *execSuite) TestListOkay(c *gc.C) {
	e := exec.NewExec(s.command)
	proc1 := exectesting.NewStubProcess(s.stub)
	proc2 := exectesting.NewStubProcess(s.stub)
	exec.TestingSetExec(e, proc1, proc2)

	processes, err := e.List()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(processes, jc.DeepEquals, []exec.Process{
		proc1,
		proc2,
	})
	s.stub.CheckCalls(c, nil)
}

func (s *execSuite) TestListOne(c *gc.C) {
	e := exec.NewExec(s.command)
	process := exectesting.NewStubProcess(s.stub)
	exec.TestingSetExec(e, process)

	processes, err := e.List()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(processes, jc.DeepEquals, []exec.Process{
		process,
	})
	s.stub.CheckCalls(c, nil)
}

func (s *execSuite) TestListNone(c *gc.C) {
	e := exec.NewExec(s.command)

	processes, err := e.List()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(processes, gc.HasLen, 0)
	s.stub.CheckCalls(c, nil)
}

func (s *execSuite) TestGetOkay(c *gc.C) {
	e := exec.NewExec(s.command)
	pid1 := 7
	pid2 := 15
	s.setProcs(e, pid1, pid2)

	process, err := e.Get(pid1)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.Equals, process)
	s.stub.CheckCallNames(c, "PID")
}

func (s *execSuite) TestGetNotFirst(c *gc.C) {
	e := exec.NewExec(s.command)
	pid1 := 7
	pid2 := 15
	s.setProcs(e, pid1, pid2)

	process, err := e.Get(pid2)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.Equals, process)
	s.stub.CheckCallNames(c, "PID", "PID")
}

func (s *execSuite) TestGetNone(c *gc.C) {
	e := exec.NewExec(s.command)

	_, err := e.Get(10)

	c.Check(err, jc.Satisfies, errors.IsNotFound)
	s.stub.CheckCalls(c, nil)
}

func (s *execSuite) TestGetNotFound(c *gc.C) {
	e := exec.NewExec(s.command)
	pid1 := 7
	pid2 := 15
	s.setProcs(e, pid1, pid2)

	_, err := e.Get(10)

	c.Check(err, jc.Satisfies, errors.IsNotFound)
	s.stub.CheckCallNames(c, "PID", "PID")
}

func (s *execSuite) TestStartOkay(c *gc.C) {
	s.cmd = exectesting.NewStubCommand(s.stub)
	expected := exectesting.NewStubProcess(s.stub)
	s.cmd.ReturnStart = expected
	e := exec.NewExec(s.command)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.stub.ResetCalls()

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.Equals, expected)
	s.stub.CheckCallNames(c, "Start")
}

func (s *execSuite) TestStartExecUpdatedOkay(c *gc.C) {
	s.cmd = exectesting.NewStubCommand(s.stub)
	expected := exectesting.NewStubProcess(s.stub)
	s.cmd.ReturnStart = expected
	e := exec.NewExec(s.command)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.stub.ResetCalls()

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)
	processes := exec.TestingExposeExec(e)

	c.Check(process, gc.Equals, expected)
	c.Check(processes, jc.DeepEquals, []exec.Process{
		process,
	})
	s.stub.CheckCallNames(c, "Start")
}

func (s *execSuite) TestStartExecUpdatedNotFirst(c *gc.C) {
	s.cmd = exectesting.NewStubCommand(s.stub)
	expected := exectesting.NewStubProcess(s.stub)
	s.cmd.ReturnStart = expected
	e := exec.NewExec(s.command)
	s.setProcs(e, 1001, 8, 23)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.stub.ResetCalls()

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)
	processes := exec.TestingExposeExec(e)

	c.Check(process, gc.Equals, expected)
	c.Check(processes, gc.HasLen, 4)
	c.Check(processes[len(processes)-1], gc.Equals, expected)
	s.stub.CheckCallNames(c, "Start")
}

func (s *execSuite) TestStartError(c *gc.C) {
	s.cmd = exectesting.NewStubCommand(s.stub)
	e := exec.NewExec(s.command)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.stub.ResetCalls()
	failure := errors.New("<failure>")
	s.stub.SetErrors(failure)

	_, err = cmd.Start()

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.stub.CheckCallNames(c, "Start")
}
