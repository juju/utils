// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"
)

// CachedOps is an implementation of Operations that records each
// requested operation. It also maintains a simple in-memory filesystem.
// Both features make CachedOps useful in testing.
type CachedOps struct {
	ops    Operations
	Events *OpEvents
}

// NewCachedOps initializes a new CachedOps and returns it.
func NewCachedOps(ops Operations) *CachedOps {
	return &CachedOps{
		ops:    ops,
		Events: &OpEvents{},
	}
}

// Exists implements Operations.
func (co *CachedOps) Exists(name string) (bool, error) {
	co.Events.Record(OpExists, name)
	return co.ops.Exists(name)
}

// Info implements Operations.
func (co *CachedOps) Info(name string) (os.FileInfo, error) {
	co.Events.Record(OpInfo, name)
	return co.ops.Info(name)
}

// MkdirAll implements Operations.
func (co *CachedOps) MkdirAll(dirname string, perm os.FileMode) error {
	event := co.Events.Record(OpMkdirAll, dirname)
	event.Permissions = perm
	return co.ops.MkdirAll(dirname, perm)
}

// ListDir implements Operations.
func (co *CachedOps) ListDir(dirname string) ([]os.FileInfo, error) {
	co.Events.Record(OpListDir, dirname)
	return co.ops.ListDir(dirname)
}

// ReadFile implements Operations.
func (co *CachedOps) ReadFile(filename string) ([]byte, error) {
	co.Events.Record(OpReadFile, filename)
	return co.ops.ReadFile(filename)
}

// CreateFile implements Operations.
func (co *CachedOps) CreateFile(filename string) (io.WriteCloser, error) {
	co.Events.Record(OpCreateFile, filename)
	return co.ops.CreateFile(filename)
}

// WriteFile implements Operations.
func (co *CachedOps) WriteFile(filename string, data []byte, perm os.FileMode) error {
	event := co.Events.Record(OpWriteFile, filename)
	event.Permissions = perm
	return co.ops.WriteFile(filename, data, perm)
}

// RemoveAll implements Operations.
func (co *CachedOps) RemoveAll(name string) error {
	co.Events.Record(OpRemoveAll, name)
	return co.ops.RemoveAll(name)
}

// Chmod implements Operations.
func (co *CachedOps) Chmod(name string, perm os.FileMode) error {
	event := co.Events.Record(OpChmod, name)
	event.Permissions = perm
	return co.ops.Chmod(name, perm)
}

// Symlink implements Operations.
func (co *CachedOps) Symlink(oldName, newName string) error {
	event := co.Events.Record(OpSymlink, newName)
	event.Source = oldName
	return co.ops.Symlink(oldName, newName)
}
