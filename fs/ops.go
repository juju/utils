// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/juju/utils/symlink"
)

// TODO(ericsnow) Either support CWD or require that all paths be
// absolute. If the former then the following methods should be added
// to genericOperations:
//  os.Chdir
//  os.Getwd

type genericOperations interface {
	// Exists returns true if the named file or directory exists and
	// false otherwise. This is a replacement for calling os.Stat and
	// checking the error with os.IsNotExist.
	Exists(name string) (bool, error)

	// Info is a replacement for os.Lstat. However, use Exists if you
	// are only checking the error return (e.g. os.IsNotExist).
	Info(name string) (os.FileInfo, error)

	// RemoveAll is a replacement for os.RemoveAll.
	RemoveAll(name string) error

	// Chmod is a replacement for os.Chmod.
	Chmod(name string, mode os.FileMode) error

	// TODO(ericsnow) Candidates:
	// filepath.Abs
	// filepath.Clean
	// filepath.IsAbs
	// os.Lchown (don't need os.Chown)
	// os.Chtimes
	// os.Link
	// Remove
	// Rename
}

type fileOperations interface {
	// ReadFile is a replacement for ioutil.ReadFile.
	ReadFile(filename string) ([]byte, error)

	// CreateFile is a replacement for os.Create.
	CreateFile(filename string) (io.WriteCloser, error)

	// WriteFile is a replacement for ioutil.WriteFile.
	WriteFile(filename string, data []byte, perm os.FileMode) error

	// TODO(ericsnow) Candidates:
	// os.Open
	// os.OpenFile
	// os.Truncate
	// ioutil.TempFile
}

type dirOperations interface {
	// MkdirAll is a replacement for os.MkdirAll.
	MkdirAll(dirname string, perm os.FileMode) error

	// ListDir is a replacement for ioutil.ReadDir.
	ListDir(dirname string) ([]os.FileInfo, error)

	// TODO(ericsnow) Candidates:
	// filepath.Glob
	// filepath.Walk
	// os.MkDir
	// ioutil.TempDir
}

type symlinkOperations interface {
	// Symlink is a replacement for os.Symlink and utils/symlink.New.
	Symlink(oldName, newName string) error

	// Readlink is a replacement for os.Readlink and utils/symlink.Read.
	Readlink(name string) (string, error)

	// TODO(ericsnow) Candidates:
	// filepath.EvalSymlinks
}

// Operations exposes various key file system operations as methods
// on a consolidated type.
type Operations interface {
	genericOperations
	fileOperations
	dirOperations
	symlinkOperations

	// TODO(ericsnow) Candidates:
	// os.Pipe
}

// TODO(ericsnow) Add a helper to resolve a symbolic link so that
// the equivalent of os.Stat can be achieved more easily.

// Ops satisfies the Operations interface, wrapping the
// equivalent functionality out of the Go stdlib (e.g os.MkdirAll
// for FileOperations.MkdirAll) and of other relevant packages.
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

// Info implements Operations.
func (Ops) Info(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll implements Operations.
func (Ops) MkdirAll(dirname string, perm os.FileMode) error {
	return os.MkdirAll(dirname, perm)
}

// ListDir implements Operations.
func (Ops) ListDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

// ReadFile implements Operations.
func (Ops) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// CreateFile implements Operations.
func (Ops) CreateFile(filename string) (io.WriteCloser, error) {
	return os.Create(filename)
}

// WriteFile implements Operations.
func (Ops) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

// RemoveAll implements Operations.
func (Ops) RemoveAll(name string) error {
	return os.RemoveAll(name)
}

// Chmod implements Operations.
func (Ops) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

// Symlink implements Operations.
func (Ops) Symlink(oldName, newName string) error {
	return symlink.New(oldName, newName)
}

// Readlink implements Operations.
func (Ops) Readlink(name string) (string, error) {
	return symlink.Read(name)
}
