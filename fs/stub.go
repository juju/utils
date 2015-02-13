// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"

	"github.com/juju/testing"
)

// StubFile keeps track of calls to io.ReadWriteCloser methods and
// allows direct control of what those methods return. This is useful
// in tests. The calls are tracked in testing.Stub. The error return
// values are also managed there.
type StubFile struct {
	// This is a pointer so it may be shared between different stubs.
	*testing.Stub

	// NWritten is the value that will be returned by Read() and Write().
	NWritten int
}

// NewStubFile builds a new StubFile and returns it.
func NewStubFile() *StubFile {
	stub := &StubFile{
		Stub: &testing.Stub{},
	}
	return stub
}

// Read Implements io.Reader.
func (ff *StubFile) Read(buf []byte) (int, error) {
	ff.AddCall("Read", buf)
	return ff.NWritten, ff.NextErr()
}

// Write Implements io.Writer.
func (ff *StubFile) Write(data []byte) (int, error) {
	ff.AddCall("Write", data)
	return ff.NWritten, ff.NextErr()
}

// Close Implements io.Closer.
func (ff *StubFile) Close() error {
	ff.AddCall("Close")
	return ff.NextErr()
}

// StubOpsReturns holds all the return values for the StubOps methods.
type StubOpsReturns struct {
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
}

// StubOps keeps track of calls to Operation methods and allows direct
// control of what those methods return. This is useful in tests. The
// calls are tracked in testing.Stub. The error return values are also
// managed there.
type StubOps struct {
	// This is a pointer so it may be shared between different stubs.
	*testing.Stub

	// Returns holds the stub's (non-error) return values.
	Returns StubOpsReturns
}

// NewStubOps builds a new StubOps and returns it.
func NewStubOps() *StubOps {
	stub := &StubOps{
		Stub: &testing.Stub{},
	}
	return stub
}

// Exists implements Operations.
func (fo *StubOps) Exists(name string) (bool, error) {
	fo.AddCall("Exists", name)
	return fo.Returns.Exists, fo.NextErr()
}

// Info implements Operations.
func (fo *StubOps) Info(name string) (os.FileInfo, error) {
	fo.AddCall("Info", name)
	return fo.Returns.Info, fo.NextErr()
}

// MkdirAll implements Operations.
func (fo *StubOps) MkdirAll(dirname string, perm os.FileMode) error {
	fo.AddCall("MkdirAll", dirname, perm)
	return fo.NextErr()
}

// ListDir implements Operations.
func (fo *StubOps) ListDir(dirname string) ([]os.FileInfo, error) {
	fo.AddCall("ReadDir", dirname)
	return fo.Returns.DirEntries, fo.NextErr()
}

// ReadFile implements Operations.
func (fo *StubOps) ReadFile(filename string) ([]byte, error) {
	fo.AddCall("ReadFile", filename)
	return fo.Returns.Data, fo.NextErr()
}

// CreateFile implements Operations.
func (fo *StubOps) CreateFile(filename string) (io.WriteCloser, error) {
	fo.AddCall("CreateFile", filename)
	return fo.Returns.File, fo.NextErr()
}

// WriteFile implements Operations.
func (fo *StubOps) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fo.AddCall("WriteFile", filename, data, perm)
	return fo.NextErr()
}

// RemoveAll implements Operations.
func (fo *StubOps) RemoveAll(name string) error {
	fo.AddCall("RemoveAll", name)
	return fo.NextErr()
}

// Chmod implements Operations.
func (fo *StubOps) Chmod(name string, perm os.FileMode) error {
	fo.AddCall("Chmod", name, perm)
	return fo.NextErr()
}

// Symlink implements Operations.
func (fo *StubOps) Symlink(oldName, newName string) error {
	fo.AddCall("Symlink", oldName, newName)
	return fo.NextErr()
}

// ReadLink implements Operations.
func (fo *StubOps) Readlink(name string) (string, error) {
	fo.AddCall("Readlink", name)
	return fo.Returns.Filename, fo.NextErr()
}
