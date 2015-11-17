// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	osexec "os/exec"
)

type TestingExposer struct{}

func (e TestingExposer) ExposeOSCommand(cmd Command) *osexec.Cmd {
	return cmd.(*Cmd).CmdStdio.Raw.(*osRawStdio).Cmd
}

func (e TestingExposer) ExposeOSProcess(process Process) *osexec.Cmd {
	return process.(*Proc).ProcessData.(*osProcessData).info
}
