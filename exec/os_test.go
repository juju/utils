// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"bytes"
	"io"
	"os"
	osexec "os/exec"

	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var (
	_ = gc.Suite(&osExecSuite{})
	_ = gc.Suite(&osCommandSuite{})
	_ = gc.Suite(&osCommandFunctionalSuite{})
	_ = gc.Suite(&osProcessSuite{})
	_ = gc.Suite(&osProcessFunctionalSuite{})
	_ = gc.Suite(&osProcessStateSuite{})
	_ = gc.Suite(&osWaitStatusSuite{})
)

type osExecSuite struct {
	BaseSuite
}

func (s *osExecSuite) TestInterface(c *gc.C) {
	e := exec.NewOSExec()

	var t exec.Exec
	c.Check(e, gc.Implements, &t)
}

func (s *osExecSuite) TestNewOSExec(c *gc.C) {
	exec.NewOSExec()
}

type osCommandSuite struct {
	BaseSuite
}

func (s *osCommandSuite) newRaw(in io.Reader, out, err io.Writer) (*osexec.Cmd, exec.CommandInfo) {
	args := []string{"spam", "eggs"}
	env := []string{"X=y"}
	dir := "/x/y/z"

	raw := &osexec.Cmd{
		Path:   args[0],
		Args:   args,
		Env:    env,
		Dir:    dir,
		Stdin:  in,
		Stdout: out,
		Stderr: err,
	}
	info := exec.CommandInfo{
		Path: args[0],
		Args: args,
		Context: exec.Context{
			Env:    env,
			Dir:    dir,
			Stdin:  in,
			Stdout: out,
			Stderr: err,
		},
	}
	return raw, info
}

func (s *osCommandSuite) TestInterface(c *gc.C) {
	var cmd exec.OSCommand

	var t exec.Command
	c.Check(&cmd, gc.Implements, &t)
}

func (s *osCommandSuite) TestInfo(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, expected := s.newRaw(&stdin, &stdout, &stderr)
	cmd := s.NewOSCommand(raw, nil)
	info := cmd.Info()

	c.Check(info, jc.DeepEquals, expected)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, _ := s.newRaw(nil, nil, nil)
	expected := *raw // copied
	expected.Stdin = &stdin
	expected.Stdout = &stdout
	expected.Stderr = &stderr
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Out: &stdout,
		Err: &stderr,
	})
	c.Assert(err, jc.ErrorIsNil)

	c.Check(raw, jc.DeepEquals, &expected)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioErrorAlreadyStdin(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	var existing bytes.Buffer
	raw, _ := s.newRaw(&existing, nil, nil)
	orig := *raw // copied
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Out: &stdout,
		Err: &stderr,
	})

	c.Check(err, jc.Satisfies, errors.IsNotValid)
	// Ensure raw did not get changed.
	c.Check(raw, jc.DeepEquals, &orig)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioErrorAlreadyStdout(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	var existing bytes.Buffer
	raw, _ := s.newRaw(nil, &existing, nil)
	orig := *raw // copied
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Out: &stdout,
		Err: &stderr,
	})

	c.Check(err, jc.Satisfies, errors.IsNotValid)
	// Ensure raw did not get changed.
	c.Check(raw, jc.DeepEquals, &orig)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioErrorAlreadyStderr(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	var existing bytes.Buffer
	raw, _ := s.newRaw(nil, nil, &existing)
	orig := *raw // copied
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Out: &stdout,
		Err: &stderr,
	})

	c.Check(err, jc.Satisfies, errors.IsNotValid)
	// Ensure raw did not get changed.
	c.Check(raw, jc.DeepEquals, &orig)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioAlreadyStdinOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, _ := s.newRaw(&stdin, nil, nil)
	expected := *raw // copied
	expected.Stdin = &stdin
	expected.Stdout = &stdout
	expected.Stderr = &stderr
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		Out: &stdout,
		Err: &stderr,
	})
	c.Assert(err, jc.ErrorIsNil)

	c.Check(raw, jc.DeepEquals, &expected)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioAlreadyStdoutOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, _ := s.newRaw(nil, &stdout, nil)
	expected := *raw // copied
	expected.Stdin = &stdin
	expected.Stdout = &stdout
	expected.Stderr = &stderr
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Err: &stderr,
	})
	c.Assert(err, jc.ErrorIsNil)

	c.Check(raw, jc.DeepEquals, &expected)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioAlreadyStderrOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, _ := s.newRaw(nil, nil, &stderr)
	expected := *raw // copied
	expected.Stdin = &stdin
	expected.Stdout = &stdout
	expected.Stderr = &stderr
	cmd := s.NewOSCommand(raw, nil)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Out: &stdout,
	})
	c.Assert(err, jc.ErrorIsNil)

	c.Check(raw, jc.DeepEquals, &expected)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestSetStdioNil(c *gc.C) {
	var cmd exec.OSCommand
	err := cmd.SetStdio(exec.Stdio{})

	c.Check(err, gc.ErrorMatches, `command not initialized`)
	s.Stub.CheckNoCalls(c)
}

// TODO(ericsnow) Add tests for Std*Pipe()?

func (s *osCommandSuite) TestStartOkay(c *gc.C) {
	var orig osexec.Cmd
	cmd := s.NewOSCommand(&orig, func(*osexec.Cmd) error {
		s.Stub.AddCall("Start")
		return s.Stub.NextErr()
	})

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.NotNil)
	raw := s.ExposeOSProcess(process)
	c.Check(raw, jc.DeepEquals, &orig)
	c.Check(raw, gc.Not(gc.Equals), &orig)
	s.Stub.CheckCallNames(c, "Start")
}

func (s *osCommandSuite) TestStartError(c *gc.C) {
	var raw osexec.Cmd
	failure := s.SetFailure()
	cmd := s.NewOSCommand(&raw, func(*osexec.Cmd) error {
		s.Stub.AddCall("Start")
		return s.Stub.NextErr()
	})

	_, err := cmd.Start()

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Start")
}

func (s *osCommandSuite) TestStartNil(c *gc.C) {
	var cmd exec.OSCommand
	_, err := cmd.Start()

	c.Check(err, gc.ErrorMatches, `command not initialized`)
	s.Stub.CheckNoCalls(c)
}

type osCommandFunctionalSuite struct {
	BaseSuite
}

func (s *osCommandFunctionalSuite) TestStart(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	// cmd := s.NewOSCommand(raw, nil)
}

type osProcessSuite struct {
	BaseSuite
}

func (s *osProcessSuite) TestInterface(c *gc.C) {
	var process exec.OSProcess

	var t exec.Process
	c.Check(&process, gc.Implements, &t)
}

func (s *osProcessSuite) TestCommandOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw := &osexec.Cmd{
		Path:   "spam",
		Args:   []string{"spam", "eggs"},
		Env:    []string{"X=y"},
		Dir:    "/x/y/z",
		Stdin:  &stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	process := s.NewOSProcess(raw, nil, nil)
	info := process.Command()

	c.Check(info, jc.DeepEquals, exec.CommandInfo{
		Path: "spam",
		Args: []string{"spam", "eggs"},
		Context: exec.Context{
			Env:    []string{"X=y"},
			Dir:    "/x/y/z",
			Stdin:  &stdin,
			Stdout: &stdout,
			Stderr: &stderr,
		},
	})
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestCommandNil(c *gc.C) {
	var process exec.OSProcess
	info := process.Command()

	c.Check(info, jc.DeepEquals, exec.CommandInfo{})
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestStateOkay(c *gc.C) {
	raw := &os.ProcessState{}
	info := &osexec.Cmd{
		ProcessState: raw,
	}
	process := s.NewOSProcess(info, nil, nil)
	state, err := process.State()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(state, jc.DeepEquals, &exec.OSProcessState{raw})
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestStateNil(c *gc.C) {
	var process exec.OSProcess
	_, err := process.State()

	c.Check(err, gc.ErrorMatches, `process not initialized`)
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestPIDOkay(c *gc.C) {
	raw := &osexec.Cmd{
		Process: &os.Process{
			Pid: 5,
		},
	}
	process := s.NewOSProcess(raw, nil, nil)
	pid := process.PID()

	c.Check(pid, gc.Equals, 5)
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestPIDNil(c *gc.C) {
	var process exec.OSProcess
	pid := process.PID()

	c.Check(pid, gc.Equals, 0)
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestWaitOkay(c *gc.C) {
	raw := &os.ProcessState{}
	info := &osexec.Cmd{
		ProcessState: raw,
	}
	wait := func() error {
		s.Stub.AddCall("wait")
		return s.Stub.NextErr()
	}
	process := s.NewOSProcess(info, wait, nil)
	state, err := process.Wait()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(state, jc.DeepEquals, &exec.OSProcessState{raw})
	s.Stub.CheckCallNames(c, "wait")
}

func (s *osProcessSuite) TestWaitError(c *gc.C) {
	raw := &os.ProcessState{}
	info := &osexec.Cmd{
		ProcessState: raw,
	}
	failure := s.SetFailure()
	wait := func() error {
		s.Stub.AddCall("wait")
		return s.Stub.NextErr()
	}
	process := s.NewOSProcess(info, wait, nil)
	state, err := process.Wait()

	c.Check(state, jc.DeepEquals, &exec.OSProcessState{raw})
	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "wait")
}

func (s *osProcessSuite) TestWaitNil(c *gc.C) {
	var process exec.OSProcess
	_, err := process.Wait()

	c.Check(err, gc.ErrorMatches, `process not initialized`)
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestKillOkay(c *gc.C) {
	var info osexec.Cmd
	kill := func() error {
		s.Stub.AddCall("kill")
		return s.Stub.NextErr()
	}
	process := s.NewOSProcess(&info, nil, kill)
	err := process.Kill()
	c.Assert(err, jc.ErrorIsNil)

	s.Stub.CheckCallNames(c, "kill")
}

func (s *osProcessSuite) TestKillError(c *gc.C) {
	var info osexec.Cmd
	failure := s.SetFailure()
	kill := func() error {
		s.Stub.AddCall("kill")
		return s.Stub.NextErr()
	}
	process := s.NewOSProcess(&info, nil, kill)
	err := process.Kill()

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "kill")
}

func (s *osProcessSuite) TestKillNil(c *gc.C) {
	var process exec.OSProcess
	err := process.Kill()

	c.Check(err, gc.ErrorMatches, `process not initialized`)
	s.Stub.CheckNoCalls(c)
}

type osProcessFunctionalSuite struct {
	BaseSuite
}

func (s *osProcessFunctionalSuite) TestWait(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//process := s.NewOSProcess(cmd, nil, nil)
}

func (s *osProcessFunctionalSuite) TestKillOkay(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//process := s.NewOSProcess(cmd, nil, nil)
}

type osProcessStateSuite struct {
	BaseSuite
}

func (s *osProcessStateSuite) TestInterface(c *gc.C) {
	var state exec.OSProcessState

	var t exec.ProcessState
	c.Check(&state, gc.Implements, &t)
}

func (s *osProcessStateSuite) TestSysOkay(c *gc.C) {
	state := exec.OSProcessState{new(os.ProcessState)}
	sys := state.Sys()

	c.Check(sys, gc.NotNil)
}

func (s *osProcessStateSuite) TestSysNil(c *gc.C) {
	state := exec.OSProcessState{nil}
	sys := state.Sys()

	c.Check(sys, gc.IsNil)
}

func (s *osProcessStateSuite) TestSysUsageOkay(c *gc.C) {
	state := exec.OSProcessState{new(os.ProcessState)}
	sys := state.SysUsage()

	c.Check(sys, gc.IsNil)
}

func (s *osProcessStateSuite) TestSysUsageNil(c *gc.C) {
	state := exec.OSProcessState{nil}
	sys := state.Sys()

	c.Check(sys, gc.IsNil)
}

type osWaitStatusSuite struct {
	BaseSuite
}

func (s *osWaitStatusSuite) TestInterface(c *gc.C) {
	var ws exec.OSWaitStatus

	var t exec.WaitStatus
	c.Check(&ws, gc.Implements, &t)
}
