// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"
)

// FileStorage is an abstraction of a system that can be used for the
// storage of files.  The type exposes the essential capabilities of
// such a system.
type FileStorage interface {
	Metadata(id string) (Metadata, error)
	Get(id string) (Metadata, io.ReadCloser, error)
	List() ([]Metadata, error)
	Add(meta Metadata, archive io.Reader) (string, error)
	SetFile(id string, file io.Reader) error
	Remove(id string) error
}
