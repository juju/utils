// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"os"

	"github.com/juju/errors"
)

// ListSubdirectories extracts the names of all subdirectories of the
// specified directory and returns that list.
func ListSubdirectories(dirname string) ([]string, error) {
	return ListSubdirectoriesOp(dirname, &Ops{})
}

// ListSubdirectoriesOp extracts the names of all subdirectories of the
// specified directory and returns that list. The provided Operations
// is used to make the filesystem calls.
func ListSubdirectoriesOp(dirname string, fops Operations) ([]string, error) {
	entries, err := fops.ListDir(dirname)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var dirnames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirnames = append(dirnames, entry.Name())
	}
	return dirnames, nil
}

// TODO(ericsnow) Should we touch a directory node whenever any of its
// files get touched?

// DirNode is a filesystem node for a directory.
type DirNode struct {
	NodeInfo

	// Content is the list of nodes in the directory.
	Content map[string]Node
}

// NewDir node initializes a new dir node and returns it.
func NewDirNode() *DirNode {
	return &DirNode{
		NodeInfo: newNode(NodeKindDir),
		Content:  make(map[string]Node),
	}
}

// NewDir builds a new directory File from the provided information.
func NewDir(dirname string, perm os.FileMode) os.FileInfo {
	// TODO(ericsnow) Fail if perm.IsDir() returns false?
	node := NewDirNode()
	node.SetPermissions(perm)
	return node.FileInfo(dirname)
}
