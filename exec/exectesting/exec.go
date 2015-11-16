// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exectesting

import (
	"github.com/juju/errors"
	"github.com/juju/testing"

	"github.com/juju/utils/exec"
)

type StubExec struct {
	stub *testing.Stub

	ReturnCommand exec.Command
	ReturnList    []exec.Process
	ReturnGet     exec.Process
}

func NewStubExec(stub *testing.Stub) *StubExec {
	return &StubExec{
		stub: stub,
	}
}

func (s *StubExec) Command(info exec.CommandInfo) (exec.Command, error) {
	s.stub.AddCall("Command", info)
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnCommand, nil
}

func (s *StubExec) List() ([]exec.Process, error) {
	s.stub.AddCall("List")
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnList, nil
}

func (s *StubExec) Get(pid int) (exec.Process, error) {
	s.stub.AddCall("Get", pid)
	if err := s.stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return s.ReturnGet, nil
}
