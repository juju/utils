// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"io"

	"github.com/juju/errors"
)

// Command exposes the functionality of a command.
//
// See os/exec.Cmd.
type Command interface {
	// Info returns the CommandInfo defining this Command.
	Info() CommandInfo

	StdioSetter
	Starter
}

// Starter describes a command that may be started.
type Starter interface {
	// Start starts execution of the command.
	Start() (Process, error)
}

// NewCommand returns a new Command for the given Exec and command.
func NewCommand(e Exec, path string, args ...string) (Command, error) {
	info := NewCommandInfo(path, args...)
	cmd, err := e.Command(info)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return cmd, nil
}

// CommandInfo holds the definition of a command's execution.
//
// See os/exec.Cmd.
type CommandInfo struct {
	// Path is the path to the command's executable.
	Path string

	// Args is the list of arguments to execute. Path must be Args[0].
	// If Args is not set then []string{Path} is used.
	Args []string

	Context
}

// NewCommandInfo returns a new CommandInfo for the given command. None
// of the command's context is set.
func NewCommandInfo(path string, args ...string) CommandInfo {
	return CommandInfo{
		Path: path,
		Args: append([]string{path}, args...),
	}
}

// Context describes the context in which a command will run.
type Context struct {
	// Env is the list of environment variables to use. If Env is nil
	// then the current environment is used. If it is empty then
	// commands will run with no environment set.
	Env []string

	// Dir is the directory in which the command will be run. If omitted
	// then the current directory is used.
	Dir string

	// Stdio holds the stdio streams for the context.
	Stdio Stdio
}

// SetStdio sets the stdio this command will use. Nil values are
// ignored. Any non-nil value for which the corresponding current
// value is non-nil results in an error.
func (c Context) SetStdio(values Stdio) error {
	stdio, err := c.Stdio.WithInitial(values)
	if err != nil {
		return errors.Trace(err)
	}

	c.Stdio = stdio
	return nil
}
