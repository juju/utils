// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build linux darwin

package symlink

import (
	"os"
)

// New is a wrapper function for os.Symlink() on Linux
func New(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// Read is a wrapper for os.Readlink() on Linux
func Read(link string) (string, error) {
	return os.Readlink(link)
}

// getLongPathAsString does nothing on linux. Its here for compatibillity
// with the windows implementation
func getLongPathAsString(path string) (string, error) {
	return path, nil
}
