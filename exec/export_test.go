// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	osexec "os/exec"
)

type TestingExposer struct{}

func (e TestingExposer) FakeOSCommand(raw *osexec.Cmd, start func(*osexec.Cmd) error) *OSCommand {
	if start == nil {
		return newOSCommand(raw)
	}
	return &OSCommand{
		Cmd:   raw,
		start: start,
	}
}

func (e TestingExposer) FakeOSProcess(info *osexec.Cmd, wait func() error, kill func() error) *OSProcess {
	if wait == nil && kill == nil {
		return newOSProcess(info)
	}
	return &OSProcess{
		info: info,
		wait: wait,
		kill: kill,
	}
}

func (e TestingExposer) ExposeOSProcess(process Process) *osexec.Cmd {
	return process.(*OSProcess).info
}
