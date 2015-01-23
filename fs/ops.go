// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"io/ioutil"
	"os"
)

// Operations exposes various key file system operations as methods
// on a consolidated type.
type Operations interface {
	// Exists returns true if the named file or directory exists and
	// false otherwise. This is a replacement for calling os.Stat and
	// checking the error with os.IsNotExist.
	Exists(name string) (bool, error)

	// MkdirAll is a replacement for os.MkdirAll.
	MkdirAll(dirname string) error

	// ReadFile is a replacement for ioutil.ReadFile.
	ReadFile(filename string) ([]byte, error)

	// CreateFile is a replacement for os.Create.
	CreateFile(filename string) (io.WriteCloser, error)

	// RemoveAll is a replacement for os.RemoveAll.
	RemoveAll(name string) error

	// Chmod is a replacement for os.Chmod.
	Chmod(name string, mode os.FileMode) error
}

// Ops satisfies the FileOperations interface, wrapping the
// equivalent functionality out of the Go stdlib (e.g os.MkdirAll
// for FileOperations.MkdirAll).
type Ops struct{}

// Exists implements Operations.
func (Ops) Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MkdirAll implements Operations.
func (Ops) MkdirAll(dirname string, mode os.FileMode) error {
	return os.MkdirAll(dirname, mode)
}

// ReadFile implements Operations.
func (Ops) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// CreateFile implements Operations.
func (Ops) CreateFile(filename string) (io.WriteCloser, error) {
	return os.Create(filename)
}

// RemoveAll implements Operations.
func (Ops) RemoveAll(name string) error {
	return os.RemoveAll(name)
}

// Chmod implements Operations.
func (Ops) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}
