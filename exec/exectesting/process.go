// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exectesting

import (
	"github.com/juju/errors"
	"github.com/juju/testing"

	"github.com/juju/utils/exec"
)

type StubProcess struct {
	stub *testing.Stub

	ReturnCommand exec.CommandInfo
	ReturnState   exec.ProcessState
	ReturnPID     int
	ReturnWait    exec.ProcessState
}

func NewStubProcess(stub *testing.Stub) *StubProcess {
	return &StubProcess{
		stub: stub,
	}
}

func (s *StubProcess) Command() exec.CommandInfo {
	s.stub.AddCall("Command")
	s.stub.PopNoErr()

	return s.ReturnCommand
}

func (s *StubProcess) State() (exec.ProcessState, error) {
	s.stub.AddCall("State")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnState, nil
}

func (s *StubProcess) PID() int {
	s.stub.AddCall("PID")
	s.stub.PopNoErr()

	return s.ReturnPID
}

func (s *StubProcess) Wait() (exec.ProcessState, error) {
	s.stub.AddCall("Wait")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnWait, nil
}

func (s *StubProcess) Kill() error {
	s.stub.AddCall("Kill")
	if err := s.stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

type FakeProcess struct {
	exec.Process

	HandleWait func(exec.ProcessState, error) (exec.ProcessState, error)
}

func NewFakeProcess(raw exec.Process) *FakeProcess {
	return &FakeProcess{
		Process: raw,
	}
}

func (f *FakeProcess) Wait() (exec.ProcessState, error) {
	state, err := f.Process.Wait()
	if f.HandleWait != nil {
		return f.HandleWait(state, err)
	}
	return state, err
}
