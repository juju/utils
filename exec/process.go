// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

// Process supports interacting with a running command.
//
// See os.Process.
type Process interface {
	// Command returns the information used to start the command.
	Command() CommandInfo

	ProcessData
	ProcessControl
}

// ProcessData provides a mechanism to get data about a process.
type ProcessData interface {
	// State returns the current state of the process.
	State() (ProcessState, error)

	// PID returns the PID of the process.
	PID() int
}

// ProcessControl exposes functionality to control a running process.
type ProcessControl interface {
	// Wait waits for the command to exit.
	Wait() (ProcessState, error)

	// Kill causes the Process to exit immediately.
	Kill() error
}

// Proc is a basic Process implementation.
type Proc struct {
	ProcessData
	ProcessControl

	// Info holds the process's original command info.
	Info CommandInfo
}

// Command implements Process.
func (p Proc) Command() CommandInfo {
	return p.Info
}

// NewProcess returns a new Proc that wraps the provided data
// and RawProcessControl.
func NewProcess(info CommandInfo, data ProcessData, raw RawProcessControl) *Proc {
	control := &ProcControl{
		Data: data,
		Raw:  raw,
	}
	return &Proc{
		ProcessData:    data,
		ProcessControl: control,
		Info:           info,
	}
}
