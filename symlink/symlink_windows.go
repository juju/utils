// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.
// Author: Robert Tingirica

package symlink

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"unicode/utf16"

	"github.com/juju/utils/path"
)

const (
	SYMBOLIC_LINK_FLAG_DIRECTORY = 1
	// This is the equivalent of syscall.GENERIC_EXECUTION
	// Using syscall.GENERIC_EXECUTION results in an "Access denied" error
	GENERIC_EXECUTION = 33554432
)

//sys createSymbolicLink(symlinkname *uint16, targetname *uint16, flags uint32) (err error) = CreateSymbolicLinkW
//sys getFinalPathNameByHandle(handle Handle, buf *uint16, buflen uint32, flags uint32) (n uint32, err error) = GetFinalPathNameByHandleW

// New creates newname as a symbolic link to oldname.
// If there is an error, it will be of type *LinkError.
func New(oldname, newname string) error {
	fi, err := os.Stat(oldname)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}
	var flag uint32
	if fi.IsDir() {
		flag = SYMBOLIC_LINK_FLAG_DIRECTORY
	}

	targetp, err := path.GetLongPathAsString(oldname)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}

	linkp, err := syscall.UTF16PtrFromString(newname)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}

	targetpUtf, err := syscall.UTF16FromString(targetp)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}

	err = createSymbolicLink(linkp, &targetpUtf[0], flag)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}
	return nil
}

// Read returns the destination of the named symbolic link.
// If there is an error, it will be of type *PathError.
func Read(link string) (string, error) {
	linkp, err := path.GetLongPathAsString(link)
	if err != nil {
		return "", err
	}

	linkpUtf, err := syscall.UTF16FromString(linkp)
	if err != nil {
		return "", &os.PathError{"readlink", link, err}
	}

	h, err := syscall.CreateFile(
		&linkpUtf[0],
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ,
		nil,
		syscall.OPEN_EXISTING,
		GENERIC_EXECUTION,
		0)
	if err != nil {
		return "", &os.PathError{"readlink", link, err}
	}
	defer syscall.CloseHandle(h)

	pathw := make([]uint16, syscall.MAX_PATH)
	n, err := getFinalPathNameByHandle(h, &pathw[0], uint32(len(pathw)), 0)
	if err != nil {
		return "", &os.PathError{"readlink", link, err}
	}
	if n > uint32(len(pathw)) {
		pathw = make([]uint16, n)
		n, err = getFinalPathNameByHandle(h, &pathw[0], uint32(len(pathw)), 0)
		if err != nil {
			return "", &os.PathError{"readlink", link, err}
		}
		if n > uint32(len(pathw)) {
			return "", &os.PathError{"readlink", link, errors.New("link length too long")}
		}
	}
	ret := string(utf16.Decode(pathw[0:n]))

	if strings.HasPrefix(ret, `\\?\`) {
		return ret[4:], nil
	}

	retp, err := path.GetLongPathAsString(ret)
	if err != nil {
		return "", &os.PathError{"readlink", link, err}
	}
	return retp, nil
}
