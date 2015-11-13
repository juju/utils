// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// exec provides utilities for executing commands through the OS.
package exec

import (
	"github.com/juju/errors"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("juju.utils.exec")

// Exec exposes the functionality of a command execution system.
type Exec interface {
	// Command returns a Command related to the system for the given info.
	Command(info CommandInfo) (Command, error)

	// List returns a list of all processes started by the system.
	List() ([]Process, error)

	// Get returns the process corresponding to the given process ID.
	// If it wasn't started through this system then errors.NotFound
	// is returned.
	Get(pid int) (Process, error)
}

// exec is a simple implementation of Exec that keeps track of commands
// started through the system.
type exec struct {
	command   func(CommandInfo) (Command, error)
	processes []Process
}

// NewExec returns an Exec that uses the provided function to produce
// new Commands. All processes started through the system are tracked.
func NewExec(command func(CommandInfo) (Command, error)) Exec {
	return &exec{
		command: command,
	}
}

// Command implements Exec.
func (e *exec) Command(info CommandInfo) (Command, error) {
	cmd, err := e.command(info)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ecmd := &execCommand{
		Command: cmd,
		exec:    e,
	}
	return ecmd, nil
}

// List implements Exec.
func (e exec) List() ([]Process, error) {
	copied := make([]Process, len(e.processes))
	copy(copied, e.processes)
	return copied, nil
}

// Get implements Exec.
func (e exec) Get(pid int) (Process, error) {
	for _, process := range e.processes {
		if process.PID() == pid {
			return process, nil
		}
	}
	return nil, errors.NotFoundf("process with PID %d", pid)
}

type execCommand struct {
	Command
	exec *exec
}

// Start implements Command.
func (c execCommand) Start() (Process, error) {
	process, err := c.Command.Start()
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.exec.processes = append(c.exec.processes, process)
	return process, nil
}
