// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"
)

// CachedOps is an implementation of Operations that records each
// requested operation. It also maintains a simple in-memory filesystem.
type CachedOps struct {
	*SimpleOps
	Events *OpEvents
}

// NewCachedOps initializes a new CachedOps and returns it.
func NewCachedOps() *CachedOps {
	return &CachedOps{
		SimpleOps: NewSimpleOps(),
		Events:    &OpEvents{},
	}
}

// Exists implements Operations.
func (co *CachedOps) Exists(name string) (bool, error) {
	co.Events.Add(OpExists, name)
	return co.SimpleOps.Exists(name)
}

// MkdirAll implements Operations.
func (co *CachedOps) MkdirAll(dirname string, mode os.FileMode) error {
	event := co.Events.Add(OpExists, dirname)
	event.Mode = mode
	return co.SimpleOps.MkdirAll(dirname, mode)
}

// ReadFile implements Operations.
func (co *CachedOps) ReadFile(filename string) ([]byte, error) {
	co.Events.Add(OpExists, filename)
	return co.SimpleOps.ReadFile(filename)
}

// CreateFile implements Operations.
func (co *CachedOps) CreateFile(filename string) (io.WriteCloser, error) {
	co.Events.Add(OpExists, filename)
	return co.SimpleOps.CreateFile(filename)
}

// RemoveAll implements Operations.
func (co *CachedOps) RemoveAll(name string) error {
	co.Events.Add(OpExists, name)
	return co.SimpleOps.RemoveAll(name)
}

// Chmod implements Operations.
func (co *CachedOps) Chmod(name string, mode os.FileMode) error {
	event := co.Events.Add(OpChmod, name)
	event.Mode = mode
	return co.SimpleOps.Chmod(name, mode)
}
