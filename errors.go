// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"fmt"
)

// RcPassthroughError indicates that a Juju plugin command exited with a
// non-zero exit code. This error is used to exit with the return code.
type RcPassthroughError struct {
	Code int
}

// Error implements error.
func (e *RcPassthroughError) Error() string {
	return fmt.Sprintf("subprocess encountered error code %v", e.Code)
}

// IsRcPassthroughError returns whether the error is an RcPassthroughError.
func IsRcPassthroughError(err error) bool {
	_, ok := err.(*RcPassthroughError)
	return ok
}

// NewRcPassthroughError creates an error that will have the code used at the
// return code from the cmd.Main function rather than the default of 1 if
// there is an error.
func NewRcPassthroughError(code int) error {
	return &RcPassthroughError{code}
}
