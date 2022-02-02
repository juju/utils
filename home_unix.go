// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.
//go:build !windows
// +build !windows

package utils

import (
	"os"
)

// Home returns the os-specific home path.
// Always returns the "real" home, not the
// confined home that is used when running
// inside a strictly confined snap.
func Home() string {
	// Used when running inside a confined snap.
	realHome := os.Getenv("SNAP_REAL_HOME")
	if realHome != "" {
		return realHome
	}
	return os.Getenv("HOME")
}

// SetHome sets the os-specific home path in the environment.
func SetHome(s string) error {
	return os.Setenv("SNAP_REAL_HOME", s)
}
