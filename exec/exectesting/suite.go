// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exectesting

import (
	"bytes"

	"github.com/juju/errors"
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
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

func (s *StubSuite) NewStubExec() *StubExec {
	return NewStubExec(s.Stub)
}

func (s *StubSuite) NewStubCommand() *StubCommand {
	return NewStubCommand(s.Stub)
}

func (s *StubSuite) NewFakeCommand() (*FakeCommand, *StubCommand) {
	stub := s.NewStubCommand()
	fake := NewFakeCommand(stub)
	return fake, stub
}

func (s *StubSuite) NewStdioCommand(handleStdio func(exec.Stdio, error) error) exec.Command {
	cmd, stub := s.NewFakeCommand()
	stub.ReturnStart = s.NewStubProcess()
	cmd.HandleStart = func(stdio exec.Stdio, raw exec.Process, err error) (exec.Process, error) {
		if err != nil {
			return raw, err
		}
		process := NewFakeProcess(raw)
		process.HandleWait = func(state exec.ProcessState, err error) (exec.ProcessState, error) {
			if stdio.In == nil {
				stdio.In = &bytes.Buffer{}
			}
			if stdio.Out == nil {
				stdio.Out = &bytes.Buffer{}
			}
			if stdio.Err == nil {
				stdio.Err = &bytes.Buffer{}
			}
			err = handleStdio(stdio, err)
			return state, err
		}
		return process, nil
	}
	return cmd
}

func (s *StubSuite) NewStubProcess() *StubProcess {
	return NewStubProcess(s.Stub)
}

func (s *StubSuite) NewStubRawProcessControl() *StubRawProcessControl {
	return NewStubRawProcessControl(s.Stub)
}

func (s *StubSuite) NewFakeProcess() (*FakeProcess, *StubProcess) {
	stub := s.NewStubProcess()
	fake := NewFakeProcess(stub)
	return fake, stub
}

func (s *StubSuite) NewStubProcessState() *StubProcessState {
	return NewStubProcessState(s.Stub)
}

func (s *StubSuite) NewStubWaiter() *StubWaiter {
	return NewStubWaiter(s.Stub)
}

func (s *StubSuite) NewStubKiller() *StubKiller {
	return NewStubKiller(s.Stub)
}
