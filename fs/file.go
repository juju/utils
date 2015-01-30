// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"
	"time"

	"github.com/juju/errors"
)

// FileInfo holds the information exposed by the os.FileInfo interface.
type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
}

// File holds information about a filesystem node. It implements
// os.FileInfo. It may also hold the file's data. That data is exposed
// through the Open method as a separate io.ReadWriteCloser.
//
// File is useful for testing and for ad hoc in-memory filesystems.
type File struct {
	Info FileInfo
	Data []byte
}

// NewFile builds a new (regular) File from the provided information.
func NewFile(filename string, perm os.FileMode, data []byte) *File {
	// TODO(ericsnow) Fail if perm.IsRegular() returns false?
	return newFile(filename, perm, data)
}

func newFile(name string, mode os.FileMode, data []byte) *File {
	info := FileInfo{
		Name:    name,
		Size:    int64(len(data)),
		Mode:    mode,
		ModTime: time.Now(),
	}
	return &File{
		Info: info,
		Data: data,
	}
}

// NewDir builds a new directory File from the provided information.
func NewDir(dirname string, perm os.FileMode) *File {
	// TODO(ericsnow) Fail if perm.IsRegular() returns false?
	return newFile(dirname, perm|os.ModeDir, nil)
}

// NewSymlink builds a new symlink File from the provided information.
func NewSymlink(oldName, newName string) *File {
	perm := os.ModePerm
	return newFile(newName, perm|os.ModeSymlink, []byte(oldName))
}

var _ os.FileInfo = (*File)(nil)

// TODO(ericsnow) special-case directory operations vs. file ops? Split
// directories into own type?

// Name implements os.FileInfo.
func (f File) Name() string {
	return f.Info.Name
}

// Size implements os.FileInfo.
func (f File) Size() int64 {
	return f.Info.Size
}

// Mode implements os.FileInfo.
func (f File) Mode() os.FileMode {
	return f.Info.Mode
}

// ModTime implements os.FileInfo.
func (f File) ModTime() time.Time {
	return f.Info.ModTime
}

// IsDir implements os.FileInfo.
func (f File) IsDir() bool {
	return f.Info.Mode.IsDir()
}

// Sys implements os.FileInfo.
func (f File) Sys() interface{} {
	// This is not implemented.
	return nil
}

// SetData updates the file's data and associated file info.
func (f *File) SetData(data []byte) {
	// TODO(ericsnow) Restrict to regular files only?
	f.Data = data
	f.Info.Size = int64(len(data))
	f.Info.ModTime = time.Now()
}

// TODO(ericsnow) Support other file modes in Open (or beside it)? Or
// follow the lead of the os package with separate Create and OpenFile
// methods?

func (f *File) Open() (*FileData, error) {
	file := &FileData{
		file: f,
	}
	return file, nil
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
	file *File

	pos uint64
	// TODO(ericsnow) current will suffer from synchronization issues
	// when File.Data gets resized externally (e,g, some other
	// FileData). This should be addressable somehow.
	current []byte
	closed  bool
}

var _ io.ReadWriteCloser = (*FileData)(nil)

func newFileData(file *File) *FileData {
	return &FileData{
		file:    file,
		current: file.Data,
	}
}

var errFileClosed = errors.New("already closed")

func newFileClosed(filename string) error {
	return errors.Annotatef(errFileClosed, "file %s", filename)
}

// Read implements io.Reader.
func (fd *FileData) Read(buf []byte) (int, error) {
	if fd.closed {
		return 0, newFileClosed(fd.file.Name())
	}

	size := len(buf)
	if size == 0 {
		return 0, nil
	}

	//numBytes := copy(buf, fd.file.Data[fd.pos:])
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
		return 0, newFileClosed(fd.file.Name())
	}

	// TODO(ericsnow) This won't work if File.Data got resized somehow.

	numBytes := len(data)
	size := len(fd.current)

	if size == 0 {
		fd.file.Data = append(fd.file.Data, data...)
	} else if size < numBytes {
		fd.file.Data = append(fd.file.Data[:fd.pos], data...)
		fd.current = nil
	} else {
		copy(fd.current, data)
		fd.current = fd.current[numBytes:]
	}
	fd.pos += uint64(numBytes)
	fd.file.Info.Size = int64(len(fd.file.Data))

	return numBytes, nil
}

// Close implements io.Closer.
func (fd *FileData) Close() error {
	if fd.closed {
		return newFileClosed(fd.file.Name())
	}

	fd.closed = true
	return nil
}
