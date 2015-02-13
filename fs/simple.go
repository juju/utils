// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
)

// TODO(ericsnow) Ensure paths are absolute?
// TODO(ericsnow) Use filepath.Clean on paths?
// TODO(ericsnow) Track CWD (and add related methods)?

// SimpleOps is an implementation of Operations that has its own
// rudimentary in-memory filesystem.
type SimpleOps struct {
	// Files holds the binding of filenames to nodes.
	Files map[string]Node

	// Permissions is the default permissions to use for files.
	Permissions os.FileMode
}

// NewCachedFileOperations initializes a new CachedOps and returns it
func NewSimpleOps() *SimpleOps {
	return &SimpleOps{
		Files:       make(map[string]Node),
		Permissions: 0644,
	}
}

// Exists implements Operations.
func (so *SimpleOps) Exists(name string) (bool, error) {
	_, exists := so.Files[name]
	return exists, nil
}

// Info implements Operations.
func (so *SimpleOps) Info(path string) (os.FileInfo, error) {
	node, exists := so.Files[path]
	if !exists {
		return nil, &os.PathError{
			Op:   "stat",
			Path: path,
			Err:  os.ErrNotExist,
		}
	}
	return node.FileInfo(path), nil
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
		path := dirs[i]
		node := NewDirNode()
		node.SetPermissions(perm)
		so.Files[path] = node
	}

	return nil
}

// ListDir implements Operations.
func (so *SimpleOps) ListDir(dirname string) ([]os.FileInfo, error) {
	var result []os.FileInfo
	for path, node := range so.Files {
		if filepath.Dir(path) == dirname {
			result = append(result, node.FileInfo(path))
		}
	}
	return result, nil
}

// ReadFile implements Operations.
func (so *SimpleOps) ReadFile(filename string) ([]byte, error) {
	node, ok := so.Files[filename]
	if !ok {
		return nil, &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  os.ErrNotExist,
		}
	}
	switch node := node.(type) {
	case *FileNode:
		data := make([]byte, len(node.Data))
		copy(data, node.Data)
		return data, nil
	case *DirNode:
		return nil, &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  errors.New("is a directory"),
		}
	default:
		return nil, &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  errors.New("is not a regular file"),
		}
	}
}

// CreateFile implements Operations.
func (so *SimpleOps) CreateFile(filename string) (io.WriteCloser, error) {
	if exists, _ := so.Exists(filename); exists {
		return nil, &os.PathError{
			Op:   "open", // TODO(ericsnow) or is it "create"?
			Path: filename,
			Err:  os.ErrExist,
		}
	}

	node := NewFileNode(nil)
	node.SetPermissions(so.Permissions)
	so.Files[filename] = node
	return node.Open(filename)
}

// CreateFile implements Operations.
func (so *SimpleOps) WriteFile(filename string, data []byte, perm os.FileMode) error {
	node, ok := so.Files[filename]

	// Handle the new file case.
	if !ok {
		node := NewFileNode(data)
		node.SetPermissions(perm)
		so.Files[filename] = node
		return nil
	}

	switch node := node.(type) {
	case *FileNode:
		node.SetData(data)
		return nil
	case *DirNode:
		return &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  errors.New("is a directory"),
		}
	default:
		return &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  errors.New("is not a regular file"),
		}
	}
}

// RemoveAll implements Operations.
func (so *SimpleOps) RemoveAll(path string) error {
	node, exists := so.Files[path]
	if !exists {
		return nil
	}

	// TODO(ericsnow) delete in order (least root to most)?

	delete(so.Files, path)
	if _, ok := node.(*DirNode); !ok {
		return nil
	}

	dirname := path + string(os.PathSeparator)
	for child, _ := range so.Files {
		if strings.HasPrefix(child, dirname) {
			delete(so.Files, child)
		}
	}

	return nil
}

// Chmod implements Operations.
func (so *SimpleOps) Chmod(path string, mode os.FileMode) error {
	node, exists := so.Files[path]
	if !exists {
		return &os.PathError{
			Op:   "chmod",
			Path: path,
			Err:  os.ErrNotExist,
		}
	}

	// TODO(ericsnow) allow setting more than permissions?
	node.SetPermissions(mode)
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

	so.Files[newName] = NewSymlinkNode(oldName)
	return nil
}

// Readlink implements Operations.
func (so *SimpleOps) Readlink(path string) (string, error) {
	oldName, err := so.readlink(path)
	if err != nil {
		err = &os.PathError{
			Op:   "readlink",
			Path: path,
			Err:  err,
		}
	}
	return oldName, err
}

func (so *SimpleOps) readlink(path string) (string, error) {
	node, ok := so.Files[path]
	if !ok {
		return "", os.ErrNotExist
	}

	symlinkNode, ok := node.(*SymlinkNode)
	if !ok {
		return "", errors.New("not a symbolic link")
	}

	return symlinkNode.Target, nil
}
