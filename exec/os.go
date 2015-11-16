// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"os"
	osexec "os/exec"
	"syscall"

	"github.com/juju/errors"
)

// OSExec is an Exec implementation that wraps os/exec.
type OSExec struct{}

// NewOSExec returns a new OSExec.
func NewOSExec() Exec {
	return &OSExec{}
}

func (e OSExec) Command(info CommandInfo) (Command, error) {
	// TODO(ericsnow) Ensure that raw.Path and raw.Args are not empty?

	env := info.Env
	if env != nil {
		env = make([]string, len(info.Env))
		copy(env, info.Env)
	}

	raw := osexec.Command(info.Args[0], info.Args[1:]...)
	raw.Env = env
	raw.Dir = info.Dir
	raw.Stdin = info.Stdin
	raw.Stdout = info.Stdout
	raw.Stderr = info.Stderr

	cmd := newOSCommand(raw)
	return cmd, nil
}

// OSCommand is a Command implementation that wraps os/exec.Cmd.
type OSCommand struct {
	*osexec.Cmd
	start func(*osexec.Cmd) error
}

func newOSCommand(raw *osexec.Cmd) *OSCommand {
	return &OSCommand{
		Cmd: raw,
		start: func(cmd *osexec.Cmd) error {
			return cmd.Start()
		},
	}
}

// Info implements Command.
func (o OSCommand) Info() CommandInfo {
	return osCommandInfo(o.Cmd)
}

// SetStdio implements Command.
func (o OSCommand) SetStdio(stdio Stdio) error {
	if o.Cmd == nil {
		return errors.New("command not initialized")
	}

	// TODO(ericsnow) Do not fail if collision is with same pointer?

	stdin := stdio.In
	if stdin == nil {
		stdin = o.Cmd.Stdin
	} else if o.Cmd.Stdin != nil {
		return errors.NewNotValid(nil, "stdin already set")
	}

	stdout := stdio.Out
	if stdout == nil {
		stdout = o.Cmd.Stdout
	} else if o.Cmd.Stdout != nil {
		return errors.NewNotValid(nil, "stdout already set")
	}

	stderr := stdio.Err
	if stderr == nil {
		stderr = o.Cmd.Stderr
	} else if o.Cmd.Stderr != nil {
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
		return nil, errors.New("command not initialized")
	}
	raw := *o.Cmd // make a copy

	if err := o.start(&raw); err != nil {
		return nil, errors.Trace(err)
	}

	process := newOSProcess(&raw)
	return process, nil
}

// OSProcess is a Process implementation that wraps os/exec.Cmd.
type OSProcess struct {
	info *osexec.Cmd
	wait func() error
	kill func() error
}

func newOSProcess(cmd *osexec.Cmd) *OSProcess {
	return &OSProcess{
		info: cmd,
		wait: cmd.Wait,
		kill: func() error {
			return cmd.Process.Kill()
		},
	}
}

// Command implements Process.
func (o OSProcess) Command() CommandInfo {
	return osCommandInfo(o.info)
}

func osCommandInfo(raw *osexec.Cmd) CommandInfo {
	if raw == nil {
		return CommandInfo{}
	}
	// TODO(ericsnow) Ensure that raw.Path and raw.Args are not empty?

	args := make([]string, len(raw.Args))
	copy(args, raw.Args)

	env := raw.Env
	if env != nil {
		env = make([]string, len(raw.Env))
		copy(env, raw.Env)
	}

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
	if o.info == nil {
		return nil, errors.New("process not initialized")
	}

	// TODO(ericsnow) Fail if o.info.ProcessState is nil?

	state := &OSProcessState{o.info.ProcessState}
	return state, nil
}

// PID implements Process.
func (o OSProcess) PID() int {
	if o.info == nil {
		return 0
	}
	return o.info.Process.Pid
}

// Wait implements Process.
func (o OSProcess) Wait() (ProcessState, error) {
	if o.info == nil {
		return nil, errors.New("process not initialized")
	}

	err := o.wait()
	if err != nil {
		err = errors.Trace(err)
	}
	state := &OSProcessState{o.info.ProcessState}
	return state, err
}

// Kill implements Process.
func (o OSProcess) Kill() error {
	if o.info == nil {
		return errors.New("process not initialized")
	}

	if err := o.kill(); err != nil {
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
