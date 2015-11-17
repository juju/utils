// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"time"
)

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

	// SysUsage returns system-dependent resource usage information
	// about the exited process.
	SysUsage() Rusage

	// SystemTime returns the system CPU time of the exited process
	// and its children.
	SystemTime() time.Duration

	// UserTime returns the user CPU time of the exited process
	// and its children.
	UserTime() time.Duration
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
