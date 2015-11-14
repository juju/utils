// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"time"

	"github.com/juju/testing"

	"github.com/juju/utils/exec"
)

type StubProcessState struct {
	stub *testing.Stub

	ReturnExited     bool
	ReturnPid        int
	ReturnSuccess    bool
	ReturnSys        exec.WaitStatus
	ReturnSysUsage   exec.Rusage
	ReturnSystemTime time.Duration
	ReturnUserTime   time.Duration
}

func NewStubProcessState(stub *testing.Stub) *StubProcessState {
	return &StubProcessState{
		stub: stub,
	}
}

func (s *StubProcessState) Exited() bool {
	s.stub.AddCall("Exited")
	s.stub.NextErr() // pop one off

	return s.ReturnExited
}

func (s *StubProcessState) Pid() int {
	s.stub.AddCall("Pid")
	s.stub.NextErr() // pop one off

	return s.ReturnPid
}

func (s *StubProcessState) Success() bool {
	s.stub.AddCall("Success")
	s.stub.NextErr() // pop one off

	return s.ReturnSuccess
}

func (s *StubProcessState) Sys() exec.WaitStatus {
	s.stub.AddCall("Sys")
	s.stub.NextErr() // pop one off

	return s.ReturnSys
}

func (s *StubProcessState) SysUsage() exec.Rusage {
	s.stub.AddCall("SysUsage")
	s.stub.NextErr() // pop one off

	return s.ReturnSysUsage
}

func (s *StubProcessState) SystemTime() time.Duration {
	s.stub.AddCall("SystemTime")
	s.stub.NextErr() // pop one off

	return s.ReturnSystemTime
}

func (s *StubProcessState) UserTime() time.Duration {
	s.stub.AddCall("UserTime")
	s.stub.NextErr() // pop one off

	return s.ReturnUserTime
}

type StubWaitStatus struct {
	stub *testing.Stub

	ReturnExitStatus int
	ReturnExited     bool
}

func NewStubWaitStatus(stub *testing.Stub) *StubWaitStatus {
	return &StubWaitStatus{
		stub: stub,
	}
}

func (s *StubWaitStatus) ExitStatus() int {
	s.stub.AddCall("ExitStatus")
	s.stub.NextErr() // pop one off

	return s.ReturnExitStatus
}

func (s *StubWaitStatus) Exited() bool {
	s.stub.AddCall("Exited")
	s.stub.NextErr() // pop one off

	return s.ReturnExited
}
