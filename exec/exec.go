// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// exec provides utilities for executing commands through the OS.
package exec

import (
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("juju.utils.exec")

// Local is a shortcut for the Exec implementation that wraps os/exec.
var Local = NewOSExec()

// Exec exposes the functionality of a command execution system.
type Exec interface {
	// Command returns a Command related to the system for the given info.
	Command(info CommandInfo) (Command, error)

	// TODO(ericsnow) Consider adding:
	//  - List() ([]Process, error)
	//  - Get(pid int) (Process, error)
}
