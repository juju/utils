// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&execSuite{})

type execSuite struct {
	BaseSuite
}

/*
func (s *execSuite) SetExecPIDs(e exec.Exec, pids ...int) {
	var processes []exec.Process
	for _, pid := range pids {
		process := exectesting.NewStubProcess(s.Stub)
		process.ReturnPID = pid
        processes = append(processes, process)
	}
	s.SetExec(e, processes...)
}
*/

func (s *execSuite) newCommandFn() (func(exec.CommandInfo) (exec.Command, error), exec.Command, exec.Process) {
	cmd := s.NewStubCommand()
	process := s.NewStubProcess()
	cmd.ReturnStart = process

	commandFn := func(info exec.CommandInfo) (exec.Command, error) {
		s.Stub.AddCall("commandFn", info)
		if err := s.Stub.NextErr(); err != nil {
			return nil, errors.Trace(err)
		}

		return cmd, nil
	}
	return commandFn, cmd, process
}

func (s *execSuite) dummyCommandFn(info exec.CommandInfo) (exec.Command, error) {
	s.Stub.AddCall("commandFn", info)
	if err := s.Stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return nil, nil
}

func (s *execSuite) TestNewExecOkay(c *gc.C) {
	commandFn, _, _ := s.newCommandFn()

	e := exec.NewExec(commandFn)
	processes := s.ExposeExec(e)

	c.Check(e, gc.NotNil)
	c.Check(processes, gc.HasLen, 0)
}

func (s *execSuite) TestNewExecNoFunc(c *gc.C) {
	e := exec.NewExec(nil)
	processes := s.ExposeExec(e)

	c.Check(e, gc.NotNil)
	c.Check(processes, gc.HasLen, 0)
}

func (s *execSuite) TestCommandOkay(c *gc.C) {
	commandFn, expected, _ := s.newCommandFn()
	e := exec.NewExec(commandFn)
	info := exec.CommandInfo{
		Path: "spam",
	}

	cmd, err := e.Command(info)
	c.Assert(err, jc.ErrorIsNil)
	raw, processes := s.ExposeExecCommand(cmd)

	c.Check(raw, gc.Equals, expected)
	c.Check(processes, gc.HasLen, 0)
	s.CheckCall(c, "commandFn", info)
}

func (s *execSuite) TestCommandNilCommand(c *gc.C) {
	e := exec.NewExec(func(info exec.CommandInfo) (exec.Command, error) {
		s.Stub.AddCall("commandFn", info)
		return nil, nil
	})
	info := exec.CommandInfo{
		Path: "spam",
	}

	cmd, err := e.Command(info)
	c.Assert(err, jc.ErrorIsNil)
	raw, _ := s.ExposeExecCommand(cmd)

	c.Check(raw, gc.IsNil)
	s.CheckCall(c, "commandFn", info)
}

func (s *execSuite) TestCommandError(c *gc.C) {
	failure := s.SetFailure()
	e := exec.NewExec(s.dummyCommandFn)
	info := exec.CommandInfo{
		Path: "spam",
	}

	_, err := e.Command(info)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.CheckCall(c, "commandFn", info)
}

func (s *execSuite) TestListOkay(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)
	proc1 := s.NewStubProcess()
	proc2 := s.NewStubProcess()
	s.SetExec(e, proc1, proc2)

	processes, err := e.List()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(processes, jc.DeepEquals, []exec.Process{
		proc1,
		proc2,
	})
	s.CheckNoCalls(c)
}

func (s *execSuite) TestListOne(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)
	process := s.NewStubProcess()
	s.SetExec(e, process)

	processes, err := e.List()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(processes, jc.DeepEquals, []exec.Process{
		process,
	})
	s.CheckNoCalls(c)
}

func (s *execSuite) TestListNone(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)

	processes, err := e.List()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(processes, gc.HasLen, 0)
	s.CheckNoCalls(c)
}

func (s *execSuite) TestGetOkay(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)
	pid1 := 7
	pid2 := 15
	s.SetExecPIDs(e, pid1, pid2)

	process, err := e.Get(pid1)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.Equals, process)
	s.Stub.CheckCallNames(c, "PID")
}

func (s *execSuite) TestGetNotFirst(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)
	pid1 := 7
	pid2 := 15
	s.SetExecPIDs(e, pid1, pid2)

	process, err := e.Get(pid2)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.Equals, process)
	s.Stub.CheckCallNames(c, "PID", "PID")
}

func (s *execSuite) TestGetNone(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)

	_, err := e.Get(10)

	c.Check(err, jc.Satisfies, errors.IsNotFound)
	s.Stub.CheckCalls(c, nil)
}

func (s *execSuite) TestGetNotFound(c *gc.C) {
	e := exec.NewExec(s.dummyCommandFn)
	s.SetExecPIDs(e, 7, 15)

	_, err := e.Get(10)

	c.Check(err, jc.Satisfies, errors.IsNotFound)
	s.Stub.CheckCallNames(c, "PID", "PID")
}

func (s *execSuite) TestStartOkay(c *gc.C) {
	commandFn, _, expected := s.newCommandFn()
	e := exec.NewExec(commandFn)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.Stub.ResetCalls()

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.Equals, expected)
	s.Stub.CheckCallNames(c, "Start")
}

func (s *execSuite) TestStartExecUpdatedOkay(c *gc.C) {
	commandFn, _, expected := s.newCommandFn()
	e := exec.NewExec(commandFn)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.Stub.ResetCalls()

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)
	processes := s.ExposeExec(e)

	c.Check(process, gc.Equals, expected)
	c.Check(processes, jc.DeepEquals, []exec.Process{
		process,
	})
	s.Stub.CheckCallNames(c, "Start")
}

func (s *execSuite) TestStartExecUpdatedNotFirst(c *gc.C) {
	commandFn, _, expected := s.newCommandFn()
	e := exec.NewExec(commandFn)
	s.SetExecPIDs(e, 1001, 8, 23)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.Stub.ResetCalls()

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)
	processes := s.ExposeExec(e)

	c.Check(process, gc.Equals, expected)
	c.Check(processes, gc.HasLen, 4)
	c.Check(processes[len(processes)-1], gc.Equals, process)
	s.Stub.CheckCallNames(c, "Start")
}

func (s *execSuite) TestStartError(c *gc.C) {
	commandFn, _, _ := s.newCommandFn()
	e := exec.NewExec(commandFn)
	cmd, err := e.Command(exec.CommandInfo{})
	c.Assert(err, jc.ErrorIsNil)
	s.Stub.ResetCalls()
	failure := s.SetFailure()

	_, err = cmd.Start()

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Start")
}
