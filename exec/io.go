// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"io"

	"github.com/juju/errors"
)

// Stdio holds the 3 stdio streams for an execution context.
type Stdio struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// WithInitial returns a copy with the provided streams set, if not set
// already. Provided nil values are ignored. Any non-nil value for which
// the corresponding current value is non-nil results in an error.
func (s Stdio) WithInitial(values Stdio) (Stdio, error) {
	// TODO(ericsnow) Do not fail if collision is with same pointer?

	if values.In == nil {
		values.In = s.In
	} else if s.In != nil {
		return values, errors.NewNotValid(nil, "stdin already set")
	}

	if values.Out == nil {
		values.Out = s.Out
	} else if s.Out != nil {
		return values, errors.NewNotValid(nil, "stdout already set")
	}

	if values.Err == nil {
		values.Err = s.Err
	} else if s.Err != nil {
		return values, errors.NewNotValid(nil, "stderr already set")
	}

	return values, nil
}

// StdioSetter exposes the functionality for setting stdio streams.
type StdioSetter interface {
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
}
