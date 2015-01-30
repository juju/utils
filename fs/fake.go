// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"

	"github.com/juju/testing"
)

// FakeOpsReturns holds all the return values for the FakeOps methods.
type FakeOpsReturns struct {
	// Exists is the value that will be returned from Exists().
	Exists bool

	// Info is the value that will be returned from Info().
	Info os.FileInfo

	// DirEntries is the value that will be returned by ListDir().
	DirEntries []os.FileInfo

	// Data is the value that will be returned by ReadFile().
	Data []byte

	// Filename is the value that will be returned by Readlink().
	Filename string

	// File is the value that will be returned by CreateFile().
	File io.WriteCloser

	// NWritten is the value that will be returned by Write().
	NWritten int
}

// FakeOps keeps track of calls to Operation methods and allows direct
// control of what those methods return. This is useful in tests.  The
// calls are tracked in testing.Fake. The error return values are also
// managed there.
type FakeOps struct {
	testing.Fake
	Returns FakeOpsReturns
}

// NewFakeOps returns a FakeOps with File initially set to the FakeOps.
// That way it will also keep track of operations on the io.ReadWriter
// returned by CreateFile.
func NewFakeOps() *FakeOps {
	fake := &FakeOps{}
	fake.Returns.File = fake
	return fake
}

// Exists implements Operations.
func (ff *FakeOps) Exists(name string) (bool, error) {
	ff.AddCall("Exists", testing.FakeCallArgs{
		"name": name,
	})
	return ff.Returns.Exists, ff.Err()
}

// Info implements Operations.
func (ff *FakeOps) Info(name string) (os.FileInfo, error) {
	ff.AddCall("Info", testing.FakeCallArgs{
		"name": name,
	})
	return ff.Returns.Info, ff.Err()
}

// MkdirAll implements Operations.
func (ff *FakeOps) MkdirAll(dirname string, perm os.FileMode) error {
	ff.AddCall("MkdirAll", testing.FakeCallArgs{
		"dirname": dirname,
		"perm":    perm,
	})
	return ff.Err()
}

// ListDir implements Operations.
func (ff *FakeOps) ListDir(dirname string) ([]os.FileInfo, error) {
	ff.AddCall("ReadDir", testing.FakeCallArgs{
		"dirname": dirname,
	})
	return ff.Returns.DirEntries, ff.Err()
}

// ReadFile implements Operations.
func (ff *FakeOps) ReadFile(filename string) ([]byte, error) {
	ff.AddCall("ReadFile", testing.FakeCallArgs{
		"filename": filename,
	})
	return ff.Returns.Data, ff.Err()
}

// CreateFile implements Operations.
func (ff *FakeOps) CreateFile(filename string) (io.WriteCloser, error) {
	ff.AddCall("CreateFile", testing.FakeCallArgs{
		"filename": filename,
	})
	return ff.Returns.File, ff.Err()
}

// WriteFile implements Operations.
func (ff *FakeOps) WriteFile(filename string, data []byte, perm os.FileMode) error {
	ff.AddCall("WriteFile", testing.FakeCallArgs{
		"filename": filename,
		"data":     data,
		"perm":     perm,
	})
	return ff.Err()
}

// RemoveAll implements Operations.
func (ff *FakeOps) RemoveAll(name string) error {
	ff.AddCall("RemoveAll", testing.FakeCallArgs{
		"name": name,
	})
	return ff.Err()
}

// Chmod implements Operations.
func (ff *FakeOps) Chmod(name string, perm os.FileMode) error {
	ff.AddCall("Chmod", testing.FakeCallArgs{
		"name": name,
		"perm": perm,
	})
	return ff.Err()
}

// Symlink implements Operations.
func (ff *FakeOps) Symlink(oldName, newName string) error {
	ff.AddCall("Symlink", testing.FakeCallArgs{
		"oldName": oldName,
		"newName": newName,
	})
	return ff.Err()
}

// ReadLink implements Operations.
func (ff *FakeOps) Readlink(name string) (string, error) {
	ff.AddCall("Readlink", testing.FakeCallArgs{
		"name": name,
	})
	return ff.Returns.Filename, ff.Err()
}

// Write Implements io.Writer.
func (ff *FakeOps) Write(data []byte) (int, error) {
	ff.AddCall("Write", testing.FakeCallArgs{
		"data": data,
	})
	return ff.Returns.NWritten, ff.Err()
}

// Write Implements io.Closer.
func (ff *FakeOps) Close() error {
	ff.AddCall("Close", nil)
	return ff.Err()
}
