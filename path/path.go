// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

// +build !windows

package path

// GetLongPathAsString returns a string representation of a long path.
func GetLongPathAsString(path string) (string, error) {
	return path, nil
}
