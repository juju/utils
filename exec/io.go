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

// RawStdio exposes low-level machinery for managing stdio.
type RawStdio interface {
	// SetStdio sets the stdio to use.
	SetStdio(values Stdio) error

	// StdinPipe, StdoutPipe, and StderrPipe each create a pipe for
	// the corresponding stdio stream.
	StdinPipe() (io.WriteCloser, io.Reader, error)
	StdoutPipe() (io.ReadCloser, io.Writer, error)
	StderrPipe() (io.ReadCloser, io.Writer, error)
}

// CmdStdio is a basic StdioSetter implementation for a single set
// of stdio streams.
type CmdStdio struct {
	// Raw provides the underlying functionality.
	Raw RawStdio

	// Stdio holds the Stdio information.
	Stdio *Stdio
}

// SetStdio implements StdioSetter.
func (s *CmdStdio) SetStdio(values Stdio) error {
	stdio, err := s.Stdio.WithInitial(values)
	if err != nil {
		return errors.Trace(err)
	}

	if err := s.Raw.SetStdio(stdio); err != nil {
		return errors.Trace(err)
	}

	s.Stdio = &stdio
	return nil
}

// StdinPipe implements StdioSetter.
func (s *CmdStdio) StdinPipe() (io.WriteCloser, error) {
	w, r, err := s.Raw.StdinPipe()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := s.SetStdio(Stdio{In: r}); err != nil {
		if err := w.Close(); err != nil {
			logger.Errorf("while closing stdin pipe: %v", err)
		}
		return nil, errors.Trace(err)
	}

	return w, nil
}

// StdoutPipe implements StdioSetter.
func (s *CmdStdio) StdoutPipe() (io.ReadCloser, error) {
	r, w, err := s.Raw.StdoutPipe()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := s.SetStdio(Stdio{Out: w}); err != nil {
		if err := r.Close(); err != nil {
			logger.Errorf("while closing stdout pipe: %v", err)
		}
		return nil, errors.Trace(err)
	}

	return r, nil
}

// StderrPipe implements StdioSetter.
func (s *CmdStdio) StderrPipe() (io.ReadCloser, error) {
	r, w, err := s.Raw.StderrPipe()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := s.SetStdio(Stdio{Err: w}); err != nil {
		if err := r.Close(); err != nil {
			logger.Errorf("while closing stderr pipe: %v", err)
		}
		return nil, errors.Trace(err)
	}

	return r, nil
}
