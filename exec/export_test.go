// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

type TestingExposer struct{}

func (TestingExposer) SetExec(e Exec, processes ...Process) {
	e.(*exec).processes = processes
}

func (TestingExposer) ExposeExec(e Exec) []Process {
	return e.(*exec).processes
}

func (e TestingExposer) ExposeExecCommand(cmd Command) (Command, []Process) {
	ecmd := cmd.(*execCommand)
	return ecmd.Command, e.ExposeExec(ecmd.exec)
}
