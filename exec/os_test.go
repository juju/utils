// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"bytes"
	"fmt"
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
	_ = gc.Suite(&osExecFunctionalSuite{})
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
	e := exec.NewOSExec()

	c.Check(e, gc.NotNil)
}

func (s *osExecSuite) TestCommand(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
}

type osExecFunctionalSuite struct {
	BaseSuite
}

func (s *osExecFunctionalSuite) TestCommandOkay(c *gc.C) {
	resolved := s.AddScript(c, "ls", "/bin/ls $@")
	args := []string{"ls", "-la", "."}
	env := []string{"X=y"}
	dir := "/x/y/z"
	var stdin, stdout, stderr bytes.Buffer
	e := exec.NewOSExec()

	cmd, err := e.Command(exec.CommandInfo{
		Path: "ls",
		Args: args,
		Context: exec.Context{
			Env: env,
			Dir: dir,
			Stdio: exec.Stdio{
				In:  &stdin,
				Out: &stdout,
				Err: &stderr,
			},
		},
	})
	c.Assert(err, jc.ErrorIsNil)

	raw := cmd.(*exec.Cmd).CmdStdio.Raw.(*exec.OSRawStdio).Cmd
	c.Check(raw, jc.DeepEquals, &osexec.Cmd{
		Path:   resolved,
		Args:   args,
		Env:    env,
		Dir:    dir,
		Stdin:  &stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	})
}

func (s *osExecFunctionalSuite) TestCommandBasic(c *gc.C) {
	args := []string{"ls"}
	e := exec.NewOSExec()

	cmd, err := e.Command(exec.CommandInfo{
		Path: "ls",
		Args: args,
	})
	c.Assert(err, jc.ErrorIsNil)

	raw := cmd.(*exec.Cmd).CmdStdio.Raw.(*exec.OSRawStdio).Cmd
	expected := osexec.Command("ls") // sets expected.err
	expected.Path = "ls"
	expected.Args = args
	expected.Env = nil
	expected.Dir = ""
	expected.Stdin = nil
	expected.Stdout = nil
	expected.Stderr = nil
	c.Check(raw, jc.DeepEquals, expected)
	c.Check(raw.Env, gc.IsNil)
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
			Env: env,
			Dir: dir,
			Stdio: exec.Stdio{
				In:  in,
				Out: out,
				Err: err,
			},
		},
	}
	return raw, info
}

func (s *osCommandSuite) TestInfoOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, expected := s.newRaw(&stdin, &stdout, &stderr)
	cmd := exec.NewOSCommand(raw)
	info := cmd.Info()

	c.Check(info, jc.DeepEquals, expected)
	s.Stub.CheckNoCalls(c)
}

func (s *osCommandSuite) TestInfoBasic(c *gc.C) {
	cmd := exec.NewOSCommand(&osexec.Cmd{
		Path: "/bin/ls",
		Args: []string{"ls"},
	})
	info := cmd.Info()

	c.Check(info, jc.DeepEquals, exec.CommandInfo{
		Path: "/bin/ls",
		Args: []string{"ls"},
		Context: exec.Context{
			Env: nil,
			Dir: "",
			Stdio: exec.Stdio{
				In:  nil,
				Out: nil,
				Err: nil,
			},
		},
	})
	c.Check(info.Env, gc.IsNil)
}

func (s *osCommandSuite) TestSetStdioOkay(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	raw, _ := s.newRaw(nil, nil, nil)
	expected := *raw // copied
	expected.Stdin = &stdin
	expected.Stdout = &stdout
	expected.Stderr = &stderr
	cmd := exec.NewOSCommand(raw)
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
	cmd := exec.NewOSCommand(raw)
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
	cmd := exec.NewOSCommand(raw)
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
	cmd := exec.NewOSCommand(raw)
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
	cmd := exec.NewOSCommand(raw)
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
	cmd := exec.NewOSCommand(raw)
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
	cmd := exec.NewOSCommand(raw)
	err := cmd.SetStdio(exec.Stdio{
		In:  &stdin,
		Out: &stdout,
	})
	c.Assert(err, jc.ErrorIsNil)

	c.Check(raw, jc.DeepEquals, &expected)
	s.Stub.CheckNoCalls(c)
}

// TODO(ericsnow) Add tests for Std*Pipe()?

func (s *osCommandSuite) TestStartOkay(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//process, err := cmd.Start()
}

func (s *osCommandSuite) TestStartError(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//_, err := cmd.Start()
}

func (s *osCommandSuite) TestStartNil(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//_, err := cmd.Start()
	//
	//c.Check(err, gc.ErrorMatches, `command not initialized`)
	//s.Stub.CheckNoCalls(c)
}

type osCommandFunctionalSuite struct {
	BaseSuite
}

func (s *osCommandFunctionalSuite) TestStart(c *gc.C) {
	var stdin, stdout, stderr bytes.Buffer
	dirname := c.MkDir()
	path := s.AddScript(c, "dump-call", `#!/bin/bash
    echo $0 $@
    pwd
    unset SHLVL
    unset PWD
    unset _
    env | sort | grep -v '^_='
    `)
	orig := &osexec.Cmd{
		Path:   path,
		Args:   []string{"dump-call", "-xy", "z"},
		Env:    []string{"SPAM=eggs"},
		Dir:    dirname,
		Stdin:  &stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cmd := exec.NewOSCommand(orig)

	process, err := cmd.Start()
	c.Assert(err, jc.ErrorIsNil)
	_, err = process.Wait()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(process, gc.NotNil)
	raw := process.(*exec.Proc).ProcessData.(exec.OSProcessData).Cmd
	c.Check(raw, gc.Not(gc.Equals), orig)
	c.Check(orig.Process, gc.IsNil)
	c.Check(orig.ProcessState, gc.IsNil)
	c.Check(raw.Process, gc.NotNil)
	c.Check(raw.ProcessState, gc.NotNil)
	c.Check(raw.Path, gc.Equals, orig.Path)
	c.Check(raw.Args, jc.DeepEquals, orig.Args)
	c.Check(raw.Env, jc.DeepEquals, orig.Env)
	c.Check(raw.Dir, gc.Equals, orig.Dir)
	c.Check(raw.Stdin, gc.Equals, orig.Stdin)
	c.Check(raw.Stdout, gc.Equals, orig.Stdout)
	c.Check(raw.Stderr, gc.Equals, orig.Stderr)
	c.Check(stdout.String(), gc.Equals, fmt.Sprintf(`
%s -xy z
%s
SPAM=eggs
`[1:], path, dirname))
	c.Check(stderr.String(), gc.Equals, "")
}

type osProcessSuite struct {
	BaseSuite
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
	process := exec.NewOSProcess(raw)
	info := process.Command()

	c.Check(info, jc.DeepEquals, exec.CommandInfo{
		Path: "spam",
		Args: []string{"spam", "eggs"},
		Context: exec.Context{
			Env: []string{"X=y"},
			Dir: "/x/y/z",
			Stdio: exec.Stdio{
				In:  &stdin,
				Out: &stdout,
				Err: &stderr,
			},
		},
	})
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestStateOkay(c *gc.C) {
	raw := &os.ProcessState{}
	info := &osexec.Cmd{
		ProcessState: raw,
	}
	process := exec.NewOSProcess(info)
	state, err := process.State()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(state, jc.DeepEquals, &exec.OSProcessState{raw})
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestPIDOkay(c *gc.C) {
	raw := &osexec.Cmd{
		Process: &os.Process{
			Pid: 5,
		},
	}
	process := exec.NewOSProcess(raw)
	pid := process.PID()

	c.Check(pid, gc.Equals, 5)
	s.Stub.CheckNoCalls(c)
}

func (s *osProcessSuite) TestWaitOkay(c *gc.C) {
	raw := &os.ProcessState{}
	info := &osexec.Cmd{
		ProcessState: raw,
	}
	process := exec.NewOSProcess(info)
	process.(*exec.Proc).ProcessControl.(*exec.ProcControl).Raw = s.NewStubWaiter()

	state, err := process.Wait()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(state, jc.DeepEquals, &exec.OSProcessState{raw})
	s.Stub.CheckCallNames(c, "Wait")
}

func (s *osProcessSuite) TestWaitError(c *gc.C) {
	raw := &os.ProcessState{}
	info := &osexec.Cmd{
		ProcessState: raw,
	}
	failure := s.SetFailure()
	process := exec.NewOSProcess(info)
	process.(*exec.Proc).ProcessControl.(*exec.ProcControl).Raw = s.NewStubWaiter()

	state, err := process.Wait()

	c.Check(state, jc.DeepEquals, &exec.OSProcessState{raw})
	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Wait")
}

func (s *osProcessSuite) TestKillOkay(c *gc.C) {
	var info osexec.Cmd
	process := exec.NewOSProcess(&info)
	process.(*exec.Proc).ProcessControl.(*exec.ProcControl).Raw = s.NewStubRawProcessControl()

	err := process.Kill()
	c.Assert(err, jc.ErrorIsNil)

	s.Stub.CheckCallNames(c, "Kill")
}

func (s *osProcessSuite) TestKillError(c *gc.C) {
	var info osexec.Cmd
	failure := s.SetFailure()
	process := exec.NewOSProcess(&info)
	process.(*exec.Proc).ProcessControl.(*exec.ProcControl).Raw = s.NewStubRawProcessControl()

	err := process.Kill()

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Kill")
}

type osProcessFunctionalSuite struct {
	BaseSuite
}

func (s *osProcessFunctionalSuite) TestWait(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//process := exec.NewOSProcess(raw)
}

func (s *osProcessFunctionalSuite) TestKillOkay(c *gc.C) {
	c.Skip("not implemented")
	// TODO(ericsnow) Finish!
	//process := exec.NewOSProcess(raw)
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
