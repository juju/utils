// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"io"

	"github.com/juju/errors"
	"github.com/juju/testing"

	"github.com/juju/utils/exec"
)

type StubCommand struct {
	stub *testing.Stub

	ReturnInfo       exec.CommandInfo
	ReturnStdinPipe  io.WriteCloser
	ReturnStdoutPipe io.ReadCloser
	ReturnStderrPipe io.ReadCloser
	ReturnStart      exec.Process
}

func NewStubCommand(stub *testing.Stub) *StubCommand {
	return &StubCommand{
		stub: stub,
	}
}

func (s *StubCommand) Info() exec.CommandInfo {
	s.stub.AddCall("Info")
	s.stub.NextErr() // pop one off

	return s.ReturnInfo
}

func (s *StubCommand) SetStdio(stdio exec.Stdio) error {
	s.stub.AddCall("SetStdio", stdio)
	if err := s.stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *StubCommand) StdinPipe() (io.WriteCloser, error) {
	s.stub.AddCall("StdinPipe")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnStdinPipe, nil
}

func (s *StubCommand) StdoutPipe() (io.ReadCloser, error) {
	s.stub.AddCall("StdoutPipe")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnStdoutPipe, nil
}

func (s *StubCommand) StderrPipe() (io.ReadCloser, error) {
	s.stub.AddCall("StderrPipe")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnStderrPipe, nil
}

func (s *StubCommand) Start() (exec.Process, error) {
	s.stub.AddCall("Start")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnStart, nil
}
