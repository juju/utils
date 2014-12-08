// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build !windows

package utils

import (
	"os"
	"os/user"
	"strings"

	"github.com/juju/errors"
)

func homeDir(userName string) (string, error) {
	u, err := user.Lookup(userName)
	if err != nil {
		return "", errors.NewUserNotFound(err, "no such user")
	}
	return u.HomeDir, nil
}

// ReplaceFile atomically replaces the destination file or directory
// with the source. The errors that are returned are identical to
// those returned by os.Rename.
func ReplaceFile(source, destination string) error {
	return os.Rename(source, destination)
}

// MakeFileURL returns a file URL if a directory is passed in else it does nothing
func MakeFileURL(in string) string {
	if strings.HasPrefix(in, "/") {
		return "file://" + in
	}
	return in
}
