// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io/ioutil"

	"github.com/juju/errors"
)

// ListSubdirectories extracts the names of all subdirectories of the
// specified directory and returns that list.
func ListSubdirectories(dirname string) ([]string, error) {
	return listSubdirectories(dirname, &Ops{})
}

func listSubdirectories(dirname string, fops Operations) ([]string, error) {
	entries, err := ioutil.ReadDir(dirname)
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
