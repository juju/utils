// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TODO(ericsnow) Ensure paths are absolute?
// TODO(ericsnow) Use filepath.Clean on paths?
// TODO(ericsnow) Track CWD (and add related methods)?

// SimpleOps is an implementation of Operations that has its own
// rudimentary in-memory filesystem.
type SimpleOps struct {
	Files       map[string]*File
	Permissions os.FileMode
}

// NewCachedFileOperations initializes a new CachedOps and returns it
func NewSimpleOps() *SimpleOps {
	return &SimpleOps{
		Files:       make(map[string]*File),
		Permissions: 0644,
	}
}

// Exists implements Operations.
func (so *SimpleOps) Exists(name string) (bool, error) {
	_, exists := so.Files[name]
	return exists, nil
}

// Info implements Operations.
func (so *SimpleOps) Info(name string) (os.FileInfo, error) {
	info, exists := so.Files[name]
	if !exists {
		return nil, &os.PathError{
			Op:   "stat",
			Path: name,
			Err:  os.ErrNotExist,
		}
	}
	return info, nil
}

// MkdirAll implements Operations.
func (so *SimpleOps) MkdirAll(dirname string, perm os.FileMode) error {
	if exists, _ := so.Exists(dirname); exists {
		return nil
	}

	// Build the list of missing directories.
	dirs := []string{dirname}
	parent := filepath.Dir(dirname)
	for len(parent) != 0 {
		if exists, _ := so.Exists(parent); exists {
			break
		}
		dirs = append(dirs, parent)
		parent = filepath.Dir(parent)
	}

	// Traverse the list of missing directories (most root to least).
	for i := len(dirs); i > 0; {
		i -= 1
		// TODO(ericsnow) Pull this out into a helper (newDir?).
		name := dirs[i]
		so.Files[name] = NewDir(name, perm)
	}

	return nil
}

// ListDir implements Operations.
func (so *SimpleOps) ListDir(dirname string) ([]os.FileInfo, error) {
	var result []os.FileInfo
	for name, info := range so.Files {
		if filepath.Dir(name) == dirname {
			result = append(result, info)
		}
	}
	return result, nil
}

// ReadFile implements Operations.
func (so *SimpleOps) ReadFile(filename string) ([]byte, error) {
	file, ok := so.Files[filename]
	if !ok {
		return nil, &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  os.ErrNotExist,
		}
	}
	data := make([]byte, len(file.Data))
	copy(data, file.Data)
	return data, nil
}

// CreateFile implements Operations.
func (so *SimpleOps) CreateFile(filename string) (io.WriteCloser, error) {
	if exists, _ := so.Exists(filename); exists {
		return nil, &os.PathError{
			Op:   "open", // TODO(ericsnow) or is it "create"?
			Path: filename,
			Err:  os.ErrNotExist,
		}
	}

	file := NewFile(filename, so.Permissions, nil)
	so.Files[filename] = file
	return file.Open()
}

// CreateFile implements Operations.
func (so *SimpleOps) WriteFile(filename string, data []byte, perm os.FileMode) error {
	file, ok := so.Files[filename]

	// Handle the new file case.
	if !ok {
		so.Files[filename] = NewFile(filename, perm, data)
		return nil
	}

	// Handle the directory case.
	if file.IsDir() {
		return &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  errors.New("is a directory"),
		}
	}

	// Handle the existing file case.
	file.SetData(data)

	return nil
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

	dirname := name + string(os.PathSeparator)
	for child, _ := range so.Files {
		if strings.HasPrefix(child, dirname) {
			delete(so.Files, child)
		}
	}

	return nil
}

// Chmod implements Operations.
func (so *SimpleOps) Chmod(name string, mode os.FileMode) error {
	file, exists := so.Files[name]
	if !exists {
		return &os.PathError{
			Op:   "chmod",
			Path: name,
			Err:  os.ErrNotExist,
		}
	}

	file.Info.Mode = mode
	return nil
}

// Symlink implements Operations.
func (so *SimpleOps) Symlink(oldName, newName string) error {
	err := so.symlink(oldName, newName)
	if err != nil {
		return &os.LinkError{"symlink", oldName, newName, err}
	}
	return nil
}

func (so *SimpleOps) symlink(oldName, newName string) error {
	if exists, _ := so.Exists(newName); exists {
		return os.ErrExist
	}
	if exists, _ := so.Exists(oldName); !exists {
		return os.ErrNotExist
	}

	so.Files[newName] = NewSymlink(oldName, newName)
	return nil
}
