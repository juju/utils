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
	realHome, exists := os.LookupEnv("SNAP_REAL_HOME")
	if exists {
		return realHome
	}
	return os.Getenv("HOME")
}

// SetHome sets the os-specific home path in the environment.
func SetHome(s string) error {
	if _, exists := os.LookupEnv("SNAP_REAL_HOME"); exists {
		return os.Setenv("SNAP_REAL_HOME", s)
	}
	return os.Setenv("HOME", s)
}
