// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

func TestingExposeExec(e Exec) []Process {
	return e.(*exec).processes
}

func TestingExposeExecCommand(cmd Command) (Command, []Process) {
	ecmd := cmd.(*execCommand)
	return ecmd.Command, TestingExposeExec(ecmd.exec)
}

func TestingSetExec(e Exec, processes ...Process) {
	e.(*exec).processes = processes
}
