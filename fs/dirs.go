// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"github.com/juju/errors"
)

// ListSubdirectories extracts the names of all subdirectories of the
// specified directory and returns that list.
func ListSubdirectories(dirname string) ([]string, error) {
	return ListSubdirectoriesOp(dirname, &Ops{})
}

// ListSubdirectoriesOp extracts the names of all subdirectories of the
// specified directory and returns that list. The provided Operations
// is used to make the filesystem calls.
func ListSubdirectoriesOp(dirname string, fops Operations) ([]string, error) {
	entries, err := fops.ListDir(dirname)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var dirnames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirnames = append(dirnames, entry.Name())
	}
	return dirnames, nil
}
