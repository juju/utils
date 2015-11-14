// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"io"

	"github.com/juju/errors"
)

// Command expoes the functionality of a command.
//
// See os/exec.Cmd.
type Command interface {
	// Info returns the CommandInfo defining this Command.
	Info() CommandInfo

	// SetStdio sets the stdio this command will use. Nil values are
	// ignored. Any non-nil value for which the corresponding current
	// value is non-nil results in an error.
	SetStdio(stdio Stdio) error

	// StdinPipe returns a pipe that will be connected to the command's
	// standard input when the command starts.
	StdinPipe() (io.WriteCloser, error)

	// StdoutPipe returns a pipe that will be connected to the command's
	// standard output when the command starts.
	StdoutPipe() (io.ReadCloser, error)

	// StderrPipe returns a pipe that will be connected to the command's
	// standard error when the command starts.
	StderrPipe() (io.ReadCloser, error)

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
	// Env is the list of environment variables to use. If None are set
	// then the current environment is used.
	Env []string

	// Dir is the directory in which the command will be run. If omitted
	// then the current directory is used.
	Dir string

	// Stdin, Stdout, and Stderr are the 3 stdio streams for the command.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Stdio holds the 3 stdio streams for an execution context.
type Stdio struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}
