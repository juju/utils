// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package path

import (
	"syscall"
)

// getLongPath converts short style paths (c:\Progra~1\foo) to full
// long paths.
func getLongPath(path string) ([]uint16, error) {
	pathp, err := syscall.UTF16FromString(path)
	if err != nil {
		return nil, err
	}

	longp := pathp
	n, err := syscall.GetLongPathName(&pathp[0], &longp[0], uint32(len(longp)))
	if err != nil {
		return nil, err
	}
	if n > uint32(len(longp)) {
		longp = make([]uint16, n)
		n, err = syscall.GetLongPathName(&pathp[0], &longp[0], uint32(len(longp)))
		if err != nil {
			return nil, err
		}
	}
	longp = longp[:n]

	return longp, nil
}

// GetLongPathAsString returns a string representation of a long path.
func GetLongPathAsString(path string) (string, error) {
	longp, err := getLongPath(path)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(longp), nil
}
