// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.
// Author: Robert Tingirica

package symlink

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func fileOrFolder(target string) (dwFlag int, err error) {
	f, err := os.Open(target)
	if err != nil {
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		dwFlag = 1
	case mode.IsRegular():
		dwFlag = 0
	}
	return dwFlag, err
}

func CreateSymLink(link, target string) error {
	dwFlag, err := fileOrFolder(target)
	if err != nil {
		return err
	}
	var (
		kernel32, _            = syscall.LoadLibrary("kernel32.dll")
		CreateSymbolicLinkW, _ = syscall.GetProcAddress(kernel32, "CreateSymbolicLinkW")
	)
	var nargs uintptr = 3
	_, _, callErr := syscall.Syscall(
		uintptr(CreateSymbolicLinkW), nargs,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(link))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(target))),
		uintptr(dwFlag))
	if callErr != 0 {
		return errors.New(fmt.Sprintf("CreateSymbolicLinkW Error: %v", callErr))
	}
	defer syscall.FreeLibrary(kernel32)
	return nil
}

func Readlink(link string) (string, error) {
	var (
		kernel                        = syscall.MustLoadDLL("kernel32.dll")
		CreateFile                    = kernel.MustFindProc("CreateFileW")
		GetFinalPathNameByHandleW     = kernel.MustFindProc("GetFinalPathNameByHandleW")
		buf_size                  int = 526
		buf                       [526]byte
		target                    string
	)

	handle, _, callErr := CreateFile.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(link))),
		uintptr(syscall.GENERIC_READ),
		syscall.FILE_SHARE_READ,
		0,
		syscall.OPEN_EXISTING,
		uintptr(33554432),
		0)
	if callErr.Error() != "The operation completed successfully." {
		return "", errors.New(fmt.Sprintf("ReadFile Error: %v", callErr))
	}

	_, _, callErr = GetFinalPathNameByHandleW.Call(
		uintptr(unsafe.Pointer(handle)),
		uintptr(unsafe.Pointer(&buf)),
		uintptr(buf_size), 0)
	if callErr.Error() != "The operation completed successfully." {
		return "", errors.New(fmt.Sprintf("GetFinalPathNameByHandleW Error: %v", callErr))
	}

	for i, _ := range buf {
		if buf[i] != 0 {
			target += string(buf[i])
		}
	}
	if len(target) > 4 {
		if target[:4] == `\\?\` {
			target = target[4:]
		}
	}
	return target, nil
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

func Symlink(oldname, newname string) error {
	return CreateSymLink(newname, oldname)
}
