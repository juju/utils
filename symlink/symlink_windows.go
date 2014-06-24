// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.
// Author: Robert Tingirica

package symlink

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"unicode/utf16"
)

type Handle uintptr
const InvalidHandle = ^Handle(0)

//sys createSymbolicLink(symlinkname *uint16, targetname *uint16, flags uint32) (err error) = CreateSymbolicLinkW
//sys getFinalPathNameByHandle(handle Handle, buf *uint16, buflen uint32, flags uint32) (n uint32, err error) = GetFinalPathNameByHandleW
//sys createFile(name *uint16, access uint32, mode uint32, sa *syscall.SecurityAttributes, createmode uint32, attrs uint32, templatefile int32) (handle Handle, err error) [failretval==InvalidHandle] = CreateFileW
//sys CloseHandle(handle Handle) (err error)

// New creates newname as a symbolic link to oldname.
// If there is an error, it will be of type *LinkError.
func New(oldname, newname string) error {
	fi, err := os.Stat(oldname)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}
	var flag uint32
	if fi.IsDir() {
		flag = 1
	}

	targetp, err := syscall.UTF16PtrFromString(oldname)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}

	linkp, err := syscall.UTF16PtrFromString(newname)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}

	err = createSymbolicLink(linkp, targetp, flag)
	if err != nil {
		return &os.LinkError{"symlink", oldname, newname, err}
	}
	return nil
}

// Read returns the destination of the named symbolic link.
// If there is an error, it will be of type *PathError.
func Read(link string) (string, error) {
	linkp, err := syscall.UTF16PtrFromString(link)
	if err != nil {
		return "", err
	}
	h, err := createFile(
		linkp,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ,
		nil,
		syscall.OPEN_EXISTING,
		33554432, // for some reason, syscall.GENERIC_EXECUTE results in "Access Denied" error
		0)
	if err != nil {
		return "", &os.PathError{"readlink", link, err}
	}
	defer CloseHandle(h)

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
	return ret, nil
}

func GetLongPath(path string) (string, error) {
    p, err := syscall.UTF16FromString(path)
    if err != nil {
        return "", err
    }
    b := p
    n, err := syscall.GetLongPathName(&p[0], &b[0], uint32(len(b)))
    if err != nil {
        return "", err
    }
    if n > uint32(len(b)) {
        b = make([]uint16, n)
        n, err = syscall.GetLongPathName(&p[0], &b[0], uint32(len(b)))
        if err != nil {
            return "", err
        }
    }
    b = b[:n]
    return syscall.UTF16ToString(b), nil
}
