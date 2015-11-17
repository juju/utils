// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"io"
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

	raw := &osexec.Cmd{}
	if info.Args[0] == info.Path {
		raw = osexec.Command(info.Args[0], info.Args[1:]...)
	} else {
		raw.Path = info.Path

		args := make([]string, len(info.Args))
		copy(args, info.Args)
		raw.Args = args
	}
	raw.Env = env
	raw.Dir = info.Dir
	raw.Stdin = info.Stdio.In
	raw.Stdout = info.Stdio.Out
	raw.Stderr = info.Stdio.Err

	cmd := NewOSCommand(raw)
	return cmd, nil
}

// CauseIsCommandNotFound determines if the error means the command
// wasn't found.
func CauseIsCommandNotFound(err error) bool {
	_, execError := err.(*osexec.Error)
	// exec.Error results are for when the executable isn't found, in
	// those cases, drop through.
	// TODO(ericsnow) Is that actually true?
	return execError
}

// NewOSCommand returns a new Command that wraps the provided osexec.Cmd.
func NewOSCommand(raw *osexec.Cmd) Command {
	info := OSCommandInfo(raw)
	rawStdio := &OSRawStdio{raw}
	cmd := newCmd(info, rawStdio)
	cmd.Starter = OSCommandStarter{raw}
	return cmd
}

type OSCommandStarter struct {
	*osexec.Cmd
}

// Start implements Starter.
func (o OSCommandStarter) Start() (Process, error) {
	if o.Cmd == nil {
		return nil, errors.New("command not initialized")
	}
	raw := *o.Cmd // make a copy

	if err := raw.Start(); err != nil {
		return nil, errors.Trace(err)
	}

	process := NewOSProcess(&raw)
	return process, nil
}

type OSRawStdio struct {
	*osexec.Cmd
}

// SetStdio implements RawStdio.
func (o OSRawStdio) SetStdio(values Stdio) error {
	o.Cmd.Stdin = values.In
	o.Cmd.Stderr = values.Out
	o.Cmd.Stdout = values.Err
	return nil
}

// StdinPipe implements RawStdio.
func (o OSRawStdio) StdinPipe() (io.WriteCloser, io.Reader, error) {
	w, err := o.Cmd.StdinPipe()
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return w, o.Cmd.Stdin, nil
}

// StdoutPipe implements RawStdio.
func (o OSRawStdio) StdoutPipe() (io.ReadCloser, io.Writer, error) {
	r, err := o.Cmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return r, o.Cmd.Stdout, nil
}

// StderrPipe implements RawStdio.
func (o OSRawStdio) StderrPipe() (io.ReadCloser, io.Writer, error) {
	r, err := o.Cmd.StderrPipe()
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return r, o.Cmd.Stderr, nil
}

// NewOSProcess returns a Process that wraps an os/exec.Cmd.
func NewOSProcess(raw *osexec.Cmd) Process {
	info := OSCommandInfo(raw)
	data := OSProcessData{raw}
	control := OSRawProcessControl{raw}
	return NewProcess(info, data, control)
}

type OSProcessData struct {
	*osexec.Cmd
}

// State implements ProcessData.
func (o OSProcessData) State() (ProcessState, error) {
	if o.Cmd == nil {
		return nil, errors.New("process not initialized")
	}

	// TODO(ericsnow) Fail if o.info.ProcessState is nil?

	state := &OSProcessState{o.Cmd.ProcessState}
	return state, nil
}

// PID implements ProcessData.
func (o OSProcessData) PID() int {
	if o.Cmd == nil || o.Cmd.Process == nil {
		return 0
	}
	return o.Cmd.Process.Pid
}

type OSRawProcessControl struct {
	raw *osexec.Cmd
}

// Kill implements utils.Killer.
func (o OSRawProcessControl) Wait() error {
	if o.raw == nil {
		return errors.New("process not initialized")
	}

	if err := o.raw.Wait(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Kill implements utils.Killer.
func (o OSRawProcessControl) Kill() error {
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

// OSCommandInfo converts an os/exec.Cmd into a CommandInfo.
func OSCommandInfo(raw *osexec.Cmd) CommandInfo {
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
			Env: env,
			Dir: raw.Dir,
			Stdio: Stdio{
				In:  raw.Stdin,
				Out: raw.Stdout,
				Err: raw.Stderr,
			},
		},
	}
}
