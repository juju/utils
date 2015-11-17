// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"github.com/juju/errors"

	"github.com/juju/utils"
)

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

// TODO(ericsnow) Expose this as utils.Waiter?

// RawProcessControl exposes low-level process control.
type RawProcessControl interface {
	Wait() error
}

// ProcControl is a ProcessControl implementation that
// wraps a RawProcessControl.
type ProcControl struct {
	// Data holds the proc's data.
	Data ProcessData

	// Raw holds the proc's functionality.
	Raw RawProcessControl
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

// Wait implements Process.
func (p ProcControl) Wait() (ProcessState, error) {
	err := p.Raw.Wait()
	state, stErr := p.Data.State()
	if err != nil {
		return state, errors.Trace(err)
	}
	if stErr != nil {
		return nil, errors.Trace(err)
	}
	return state, nil
}

// Kill implements Process.
func (p ProcControl) Kill() error {
	if err := utils.KillIfSupported(p.Raw); err != nil {
		return errors.Trace(err)
	}
	return nil
}
