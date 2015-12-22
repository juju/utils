// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exectesting

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
)

type StubWaiter struct {
	*testing.Stub
}

func NewStubWaiter(stub *testing.Stub) *StubWaiter {
	return &StubWaiter{
		Stub: stub,
	}
}

func (s *StubWaiter) Wait() error {
	s.Stub.AddCall("Wait")
	if err := s.Stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

type StubKiller struct {
	*testing.Stub
}

func NewStubKiller(stub *testing.Stub) *StubKiller {
	return &StubKiller{
		Stub: stub,
	}
}

func (s *StubKiller) Kill() error {
	s.Stub.AddCall("Kill")
	if err := s.Stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}
