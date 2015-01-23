// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"
	"path"
)

// SimpleOps is an implementation of Operations that has its own
// rudimentary in-memory filesystem.
type SimpleOps struct {
	Files map[string]*File
	Mode  os.FileMode
}

// NewCachedFileOperations initializes a new CachedOps and returns it
func NewSimpleOps() *SimpleOps {
	return &SimpleOps{
		Files: make(map[string]*File),
		Mode:  0644,
	}
}

// Exists implements Operations.
func (so *SimpleOps) Exists(name string) (bool, error) {
	_, exists := so.Files[name]
	return exists, nil
}

// MkdirAll implements Operations.
func (so *SimpleOps) MkdirAll(dirname string, mode os.FileMode) error {
	if exists, _ := so.Exists(dirname); exists {
		return nil
	}

	dirs := []string{dirname}
	parent := path.Dir(dirname)
	for len(parent) != 0 {
		if exists, _ := so.Exists(parent); exists {
			break
		}
		dirs = append(dirs, parent)
		parent = path.Dir(parent)
	}

	// Traverse in reverse order (most root to least).
	for i := len(dirs); i > 0; {
		i -= 1
		// TODO(ericsnow) Pull this out into a helper (newDir?).
		name := dirs[i]
		so.Files[name] = &File{Info: FileInfo{
			Name:  name,
			Mode:  mode,
			IsDir: true,
		}}
	}

	return nil
}

// ReadFile implements Operations.
func (so *SimpleOps) ReadFile(filename string) ([]byte, error) {
	file, ok := so.Files[filename]
	if !ok {
		return nil, os.ErrNotExist
	}
	data := make([]byte, len(file.Data))
	copy(data, file.Data)
	return data, nil
}

// CreateFile implements Operations.
func (so *SimpleOps) CreateFile(filename string) (io.WriteCloser, error) {
	if exists, _ := so.Exists(filename); exists {
		return nil, os.ErrExist
	}

	file := &File{Info: FileInfo{
		Name: filename,
		Mode: so.Mode,
	}}
	so.Files[filename] = file
	return file.Open()
}

// RemoveAll implements Operations.
func (so *SimpleOps) RemoveAll(name string) error {
	file, exists := so.Files[name]
	if !exists {
		return nil
	}

	// TODO(ericsnow) delete in order (least root to most)?

	delete(so.Files, name)
	if !file.IsDir() {
		return nil
	}

	pattern := name + "/*"
	for child, _ := range so.Files {
		// The pattern is okay, so we can ignore the error.
		if matched, _ := path.Match(pattern, child); matched {
			delete(so.Files, child)
		}
	}

	return nil
}

// Chmod implements Operations.
func (so *SimpleOps) Chmod(name string, mode os.FileMode) error {
	file, exists := so.Files[name]
	if !exists {
		return os.ErrNotExist
	}

	file.Info.Mode = mode
	return nil
}
