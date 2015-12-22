// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"io"

	"github.com/juju/errors"

	"github.com/juju/utils"
)

func newCmd(info CommandInfo, rawStdio RawStdio) *Cmd {
	cmd := &Cmd{
		data: info,
		// Starter is not set.
	}
	cmd.CmdStdio = CmdStdio{
		Raw:   rawStdio,
		Stdio: &cmd.data.Stdio,
	}
	return cmd
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

// TODO(ericsnow) Expose this as utils.Waiter?

// RawProcessControl exposes low-level process control.
type RawProcessControl interface {
	Wait() error
}

// ProcControl is a ProcessControl implementation that
// wraps a RawProcessControl.
type ProcControl struct {
	// Data holds the proc's data.
	Data ProcessData

	// Raw holds the proc's functionality.
	Raw RawProcessControl
}

// Wait implements Process.
func (p ProcControl) Wait() (ProcessState, error) {
	err := p.Raw.Wait()
	state, stErr := p.Data.State()
	if err != nil {
		return state, errors.Trace(err)
	}
	if stErr != nil {
		return nil, errors.Trace(err)
	}
	return state, nil
}

// Kill implements Process.
func (p ProcControl) Kill() error {
	if err := utils.KillIfSupported(p.Raw); err != nil {
		return errors.Trace(err)
	}
	return nil
}
