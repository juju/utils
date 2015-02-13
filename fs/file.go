// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"

	"github.com/juju/errors"
)

// FileNode is a filesystem node for regular files.
type FileNode struct {
	NodeInfo

	// Data is the file's content.
	Data []byte
}

// NewFileNode initializes a new FileNode with the provided data and
// returns it.
func NewFileNode(data []byte) *FileNode {
	node := &FileNode{
		NodeInfo: newNode(NodeKindFile),
	}
	node.SetData(data)
	return node
}

// NewFile builds a new (regular) File from the provided information.
func NewFile(filename string, perm os.FileMode, data []byte) os.FileInfo {
	// TODO(ericsnow) Fail if perm.IsRegular() returns false?
	node := NewFileNode(data)
	node.SetPermissions(perm)
	return node.FileInfo(filename)
}

// SetData updates the file's data and associated file info.
func (fn *FileNode) SetData(data []byte) {
	fn.Data = data
	fn.Size = int64(len(data))
	fn.Touch()
}

// TODO(ericsnow) Support other file modes in Open (or beside it)? Or
// follow the lead of the os package with separate Create and OpenFile
// methods?

// Open creates a new io.ReadWriteCloser that wraps the file's data.
func (fn *FileNode) Open(filename string) (*FileData, error) {
	return newFileData(fn, filename), nil
}

// TODO(ericsnow) Use a channel (stored on File) to syncronize writes
// when more than one FileData wraps the same File.
// TODO(ericsnow) Implement buffering (and a Flush method) on FileData
// (perhaps using bufio.ReadWriter)?
// TODO(ericsnow) Implement the other methods of os.File on File.

// FileData exposes the data in a File via the io.ReadWriteCloser
// interface. It keep track of the file position, so reads and writes
// will behave accordingly.
//
// FileData is a simple wrapper around File.Data for a File. There is no
// buffering so all operations are immediately passed through to the
// underlying File.Data. One consequence is that multiple FileData that
// wrap the same File may behave in unexpected ways.
type FileData struct {
	filename string
	node     *FileNode

	pos uint64
	// TODO(ericsnow) current will suffer from synchronization issues
	// when File.Data gets resized externally (e,g, some other
	// FileData). This should be addressable somehow.
	current []byte
	closed  bool
}

var _ io.ReadWriteCloser = (*FileData)(nil)

func newFileData(node *FileNode, filename string) *FileData {
	return &FileData{
		filename: filename,
		node:     node,
		current:  node.Data,
	}
}

var errFileClosed = errors.New("already closed")

func newErrFileClosed(filename string) error {
	return errors.Annotatef(errFileClosed, "file %s", filename)
}

// Read implements io.Reader.
func (fd *FileData) Read(buf []byte) (int, error) {
	if fd.closed {
		return 0, newErrFileClosed(fd.filename)
	}

	size := len(buf)
	if size == 0 {
		return 0, nil
	}

	//numBytes := copy(buf, fd.node.Data[fd.pos:])
	numBytes := copy(buf, fd.current)
	fd.pos += uint64(numBytes)
	// TODO(ericsnow) This won't work if File.Data got resized somehow.
	fd.current = fd.current[numBytes:]

	if numBytes < size {
		return numBytes, io.EOF
	}
	return numBytes, nil
}

// Write implements io.Writer.
func (fd *FileData) Write(data []byte) (int, error) {
	if fd.closed {
		return 0, newErrFileClosed(fd.filename)
	}

	// TODO(ericsnow) This won't work if FileNode.Data got resized somehow.

	numBytes := len(data)
	size := len(fd.current)

	if size == 0 {
		fd.node.Data = append(fd.node.Data, data...)
	} else if size < numBytes {
		fd.node.Data = append(fd.node.Data[:fd.pos], data...)
		fd.current = nil
	} else {
		copy(fd.current, data)
		fd.current = fd.current[numBytes:]
	}
	fd.pos += uint64(numBytes)

	// Update the size and timestamps.
	fd.node.SetData(fd.node.Data)

	return numBytes, nil
}

// Close implements io.Closer.
func (fd *FileData) Close() error {
	if fd.closed {
		return newErrFileClosed(fd.filename)
	}

	fd.closed = true
	return nil
}
