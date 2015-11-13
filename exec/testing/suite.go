// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"
)

type StubSuite struct {
	Stub     *testing.Stub
	StubExec *StubExec
	StubCmd  *StubCommand
	StubProc *StubProcess
}

func (s *StubSuite) SetUpTest(c *gc.C) {
	s.Stub = &testing.Stub{}
	s.StubExec = s.NewStubExec()
	s.StubCmd = s.NewStubCommand()
	s.StubProc = s.NewStubProcess()
}

func (s *StubSuite) SetFailure() error {
	failure := errors.New("<failure>")
	s.Stub.SetErrors(failure)
	return failure
}

// TODO(ericsnow) Add CheckNoCalls and CheckCall to testing.Stub?

func (s *StubSuite) CheckNoCalls(c *gc.C) {
	s.Stub.CheckCalls(c, nil)
}

func (s *StubSuite) CheckCall(c *gc.C, name string, args ...interface{}) {
	s.Stub.CheckCalls(c, []testing.StubCall{{
		FuncName: name,
		Args:     args,
	}})
}

func (s *StubSuite) NewStubExec() *StubExec {
	return NewStubExec(s.Stub)
}

func (s *StubSuite) NewStubCommand() *StubCommand {
	return NewStubCommand(s.Stub)
}

func (s *StubSuite) NewStubProcess() *StubProcess {
	return NewStubProcess(s.Stub)
}
