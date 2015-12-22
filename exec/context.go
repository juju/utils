// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"github.com/juju/errors"
)

// Context describes the context in which a command will run.
type Context struct {
	// Env is the list of environment variables to use. If Env is nil
	// then the current environment is used. If it is empty then
	// commands will run with no environment set.
	Env []string

	// Dir is the directory in which the command will be run. If omitted
	// then the current directory is used.
	Dir string

	// Stdio holds the stdio streams for the context.
	Stdio Stdio
}

// SetStdio sets the stdio this command will use. Nil values are
// ignored. Any non-nil value for which the corresponding current
// value is non-nil results in an error.
func (c Context) SetStdio(values Stdio) error {
	stdio, err := c.Stdio.WithInitial(values)
	if err != nil {
		return errors.Trace(err)
	}

	c.Stdio = stdio
	return nil
}
