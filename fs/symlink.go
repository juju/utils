// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

// SymlinkNode is a filesystem node for a symbolic link.
type SymlinkNode struct {
	NodeInfo

	// Target is the path to the symlinked dir entry.
	Target string
}

// NewSymlinkNode initializes a new symlink node with the provided
// information and returns it.
func NewSymlinkNode(target string) *SymlinkNode {
	// Symlinks do not need their permissions set.
	node := &SymlinkNode{
		NodeInfo: newNode(NodeKindSymlink),
	}
	node.SetTarget(target)
	return node
}

// SetTarget updates the node based on the provided target path.
func (sn *SymlinkNode) SetTarget(target string) {
	sn.Target = target
	sn.Size = int64(len(target))
}
