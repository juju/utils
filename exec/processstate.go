// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

// ProcessState describes the state of a started command.
//
// See os.ProcessState.
type ProcessState interface {
	// Exited reports whether the program has exited.
	Exited() bool

	// Pid returns the process id of the exited process.
	Pid() int

	// Success reports whether the program exited successfully.
	Success() bool

	// Sys return system-dependent exit information about the process.
	Sys() WaitStatus

	// TODO(ericsnow) Add SysUsage, SystemTime, and UserTime methods?
}

// WaitStatus exposes system-dependent exit information about a process.
//
// See syscall.WaitStatus
type WaitStatus interface {
	// ExitStatus returns the exit code for the process.
	ExitStatus() int

	// Exited reports whether the program has exited.
	Exited() bool

	// For now we don't worry about any others.
}

// Rusage exposes system-dependent resource information.
type Rusage interface {
	// For now we don't worry about it.
}
