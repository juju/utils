// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"
)

type FileStorage interface {
	Metadata(id string) (Metadata, error)
	Get(id string) (Metadata, io.ReadCloser, error)
	List() ([]Metadata, error)
	Add(meta Metadata, archive io.Reader) (string, error)
	SetFile(id string, file io.Reader) error
	Remove(id string) error
}
