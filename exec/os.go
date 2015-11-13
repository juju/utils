// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"os"
	osexec "os/exec"
	"syscall"

	"github.com/juju/errors"
)

// NewOSExec returns an Exec that wraps os/exec.
func NewOSExec() Exec {
	return NewExec(osExecCommand)
}

func osExecCommand(info CommandInfo) (Command, error) {
	args := make([]string, len(info.Args))
	copy(args, info.Args)
	env := make([]string, len(info.Env))
	copy(env, info.Env)
	raw := &osexec.Cmd{
		Path:   info.Path,
		Args:   info.Args,
		Env:    env,
		Dir:    info.Dir,
		Stdin:  info.Stdin,
		Stdout: info.Stdout,
		Stderr: info.Stderr,
	}
	return &OSCommand{raw}, nil
}

// OSCommand is a Command implementation that wraps os/exec.Cmd.
type OSCommand struct {
	*osexec.Cmd
}

// Info implements Command.
func (o OSCommand) Info() CommandInfo {
	return osCommandInfo(o.Cmd)
}

// SetStdio implements Command.
func (o OSCommand) SetStdio(stdio Stdio) error {
	if o.Cmd == nil {
		return errors.New("command not initalized")
	}

	stdin := stdio.In
	if stdin != nil && o.Cmd.Stdin != nil {
		return errors.NewNotValid(nil, "stdin already set")
	}

	stdout := stdio.Out
	if stdout != nil && o.Cmd.Stdout != nil {
		return errors.NewNotValid(nil, "stdout already set")
	}

	stderr := stdio.Err
	if stderr != nil && o.Cmd.Stderr != nil {
		return errors.NewNotValid(nil, "stderr already set")
	}

	o.Cmd.Stdin = stdin
	o.Cmd.Stderr = stdout
	o.Cmd.Stdout = stderr
	return nil
}

// Start implements Command.
func (o OSCommand) Start() (Process, error) {
	if o.Cmd == nil {
		return nil, errors.New("command not initalized")
	}
	raw := *o.Cmd // make a copy

	if err := raw.Start(); err != nil {
		return nil, errors.Trace(err)
	}

	process := &OSProcess{
		raw: &raw,
	}
	return process, nil
}

// OSProcess is a Process implementation that wraps os/exec.Cmd.
type OSProcess struct {
	raw *osexec.Cmd
}

// Command implements Process.
func (o OSProcess) Command() CommandInfo {
	return osCommandInfo(o.raw)
}

func osCommandInfo(raw *osexec.Cmd) CommandInfo {
	if raw == nil {
		return CommandInfo{}
	}

	args := make([]string, len(raw.Args))
	copy(args, raw.Args)
	env := make([]string, len(raw.Env))
	copy(env, raw.Env)
	return CommandInfo{
		Path: raw.Path,
		Args: args,
		Context: Context{
			Env:    env,
			Dir:    raw.Dir,
			Stdin:  raw.Stdin,
			Stdout: raw.Stdout,
			Stderr: raw.Stderr,
		},
	}
}

// State implements Process.
func (o OSProcess) State() (ProcessState, error) {
	if o.raw == nil {
		return nil, errors.New("process not initialized")
	}

	state := &OSProcessState{o.raw.ProcessState}
	return state, nil
}

// PID implements Process.
func (o OSProcess) PID() int {
	if o.raw == nil {
		return 0
	}
	return o.raw.Process.Pid
}

// Wait implements Process.
func (o OSProcess) Wait() (ProcessState, error) {
	if o.raw == nil {
		return nil, errors.New("process not initialized")
	}

	err := o.raw.Wait()
	if err != nil {
		err = errors.Trace(err)
	}
	state := &OSProcessState{o.raw.ProcessState}
	return state, err
}

// Kill implements Process.
func (o OSProcess) Kill() error {
	if o.raw == nil {
		return errors.New("process not initialized")
	}

	if err := o.raw.Process.Kill(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// OSProcessState is a ProcessState implementation that wrapps os.ProcessState.
type OSProcessState struct {
	*os.ProcessState
}

// Sys implements ProcessState.
func (o OSProcessState) Sys() WaitStatus {
	if o.ProcessState == nil {
		return nil
	}

	ws, ok := o.ProcessState.Sys().(*syscall.WaitStatus)
	if !ok {
		// TODO(ericsnow) Do something else?
		ws = nil
	}
	return &OSWaitStatus{ws}
}

// SysUsage implements ProcessState.
func (o OSProcessState) SysUsage() Rusage {
	if o.ProcessState == nil {
		return nil
	}

	// For now we don't worry about it.
	return nil
}

// OSWaitStatus is a WaitState implementation that wraps syscall.WaitStatus.
type OSWaitStatus struct {
	*syscall.WaitStatus
}
