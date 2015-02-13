// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"os"
	"path"
	"time"
)

// These are the different node kinds supported in this package.
const (
	NodeKindFile    = "regular"
	NodeKindDir     = "dir"
	NodeKindSymlink = "symlink"
)

const (
	// ModeUnknown is the mode used to indicate an unknown kind of node.
	ModeUnknown = ^os.FileMode(0)
)

var (
	nodeKindModes = map[string]os.FileMode{
		NodeKindFile:    0,
		NodeKindDir:     os.ModeDir,
		NodeKindSymlink: os.ModeSymlink,
	}
)

// Node represents a unique "file" in a filesystem.
type Node interface {
	// FileInfo returns an os.FileInfo that exposes a copy of the
	// Node's data for the given filename.
	FileInfo(path string) os.FileInfo

	// Touch updates the node's access and modication timestamps.
	Touch() time.Time

	// SetPermissions updates the permissions part of the node's mode.
	SetPermissions(perm os.FileMode)
}

// NodeInfo is a Node implementation that holds the information about a
// filesystem node.
type NodeInfo struct {
	// Size is the size of the node's content.
	Size int64

	// Mode holds the node's permissions and other mode data.
	Mode os.FileMode

	// ModTime is when the node's data was last modified.
	ModTime time.Time

	// AccessTime is when the node's data was last accessed.
	AccessTime time.Time

	// TODO(ericsnow) Add other posix inode data (e.g. user and group IDs).
}

func newNode(kind string) NodeInfo {
	mode, ok := nodeKindModes[kind]
	if !ok {
		mode = ModeUnknown
	}

	info := NodeInfo{
		Mode: mode,
	}
	info.Touch()

	return info
}

// FileInfo implements Node.
func (ni NodeInfo) FileInfo(path string) os.FileInfo {
	return &fileInfo{
		path: path,
		node: ni, // copies the value
	}
}

// Touch implements Node.
func (ni *NodeInfo) Touch() time.Time {
	now := time.Now()
	ni.ModTime = now
	ni.AccessTime = now
	return now
}

// SetPermissions implements Node.
func (ni *NodeInfo) SetPermissions(perm os.FileMode) {
	ni.Mode = (^os.ModePerm & ni.Mode) | (os.ModePerm & perm)
}

// fileInfo implements os.FileInfo for a single NodeInfo.
type fileInfo struct {
	path string
	node NodeInfo
}

// Path is the absolute path to the "file".
func (fi fileInfo) Path() string {
	return fi.path
}

// Name implements os.FileInfo.
func (fi fileInfo) Name() string {
	return path.Base(fi.path)
}

// Size implements os.FileInfo.
func (fi fileInfo) Size() int64 {
	return fi.node.Size
}

// Mode implements os.FileInfo.
func (fi fileInfo) Mode() os.FileMode {
	return fi.node.Mode
}

// ModTime implements os.FileInfo.
func (fi fileInfo) ModTime() time.Time {
	return fi.node.ModTime
}

// IsDir implements os.FileInfo.
func (fi fileInfo) IsDir() bool {
	return fi.node.Mode.IsDir()
}

// Sys implements os.FileInfo.
func (fi fileInfo) Sys() interface{} {
	// This is not implemented.
	return nil
}
