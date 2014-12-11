// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build windows

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	movefile_replace_existing = 0x1
	movefile_write_through    = 0x8
)

var LineEnding = "\r\n"
var noSuchUser = "No mapping between account names and security IDs was done."

//sys moveFileEx(lpExistingFileName *uint16, lpNewFileName *uint16, dwFlags uint32) (err error) = MoveFileExW

// ReplaceFile atomically replaces the destination file or directory with the source.
// The errors that are returned are identical to those returned by os.Rename.
func ReplaceFile(source, destination string) error {
	src, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return &os.LinkError{"replace", source, destination, err}
	}
	dest, err := syscall.UTF16PtrFromString(destination)
	if err != nil {
		return &os.LinkError{"replace", source, destination, err}
	}

	// see http://msdn.microsoft.com/en-us/library/windows/desktop/aa365240(v=vs.85).aspx
	if err := moveFileEx(src, dest, movefile_replace_existing|movefile_write_through); err != nil {
		return &os.LinkError{"replace", source, destination, err}
	}
	return nil
}

// MakeFileURL returns a proper file URL for the given path/directory
func MakeFileURL(in string) string {
	var volumeName string
	if strings.HasPrefix(in, "file://") {
		volumeName = filepath.VolumeName(in[len("file://"):])
	} else {
		volumeName = filepath.VolumeName(in)
	}

	if volumeName != "" {

        // Strip colon
        volumeName := string(volumeName[:len(volumeName) - 1])

        // Do not apply function twice
	if strings.HasPrefix(in, "file://\\\\localhost\\") {
		return in
	}

	in = filepath.ToSlash(in)
	prefix := "file://\\\\localhost\\" + volumeName + "$"
        if strings.HasPrefix(in, volumeName) {
		return prefix + in[len(volumeName) + 1:]
	}
	if strings.HasPrefix(in, "file://" + volumeName) {
		return prefix + in[len("file://" + volumeName) + 1:]
	}

	}

	return in
}

func getUserSID(username string) (string, error) {
	sid, _, _, e := syscall.LookupSID("", username)
	if e != nil {
		return "", e
	}
	sidStr, err := sid.String()
	return sidStr, err
}

func readRegString(h syscall.Handle, key string) (value string, err error) {
	var typ uint32
	var buf uint32

	// Get size of registry key
	err = syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr(key), nil, &typ, nil, &buf)
	if err != nil {
		return value, err
	}

	n := make([]uint16, buf/2+1)
	err = syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr(key), nil, &typ, (*byte)(unsafe.Pointer(&n[0])), &buf)
	if err != nil {
		return value, err
	}
	return syscall.UTF16ToString(n[:]), err
}

func homeFromRegistry(sid string) (string, error) {
	var h syscall.Handle
	// This key will exist on all platforms we support the agent on (windows server 2008 and above)
	keyPath := fmt.Sprintf("Software\\Microsoft\\Windows NT\\CurrentVersion\\ProfileList\\%s", sid)
	err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE,
		syscall.StringToUTF16Ptr(keyPath),
		0, syscall.KEY_READ, &h)
	if err != nil {
		return "", err
	}
	defer syscall.RegCloseKey(h)
	str, err := readRegString(h, "ProfileImagePath")
	if err != nil {
		return "", err
	}
	return str, nil
}

// homeDir returns a local user home dir on Windows
// user.Lookup() does not populate Gid and HomeDir on Windows,
// so we get it from the registry
func homeDir(user string) (string, error) {
	u, err := getUserSID(user)
	if err != nil {
		return "", err
	}
	return homeFromRegistry(u)
}
