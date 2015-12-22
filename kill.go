// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"github.com/juju/errors"
)

// Killer exposes the functionality to kill something.
type Killer interface {
	// Kill causes value to end immediately.
	Kill() error
}

// KillIfSupported calls Kill() on the provided value if it has the method.
func KillIfSupported(v interface{}) error {
	k, ok := v.(Killer)
	if !ok {
		return nil
	}

	if err := k.Kill(); err != nil {
		return errors.Trace(err)
	}
	return nil
}
