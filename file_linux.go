// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.
// +build !windows

package utils

import (
	"os"
)

// ReplaceFile atomically replaces the destination file or directory
// with the source. The errors that are returned are identical to
// those returned by os.Rename.
func ReplaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
