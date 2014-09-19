// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"time"

	"github.com/juju/errors"
)

var _ Doc = (*doc)(nil)

type doc struct {
	id string
}

// ID implements Doc.ID.
func (d *doc) ID() string {
	return d.id
}

// SetID implements Doc.SetID.
func (d *doc) SetID(id string) bool {
	if d.id != "" {
		return true
	}
	d.id = id
	return false
}

// Ensure FileMetadata implements Metadata.
var _ Metadata = (*FileMetadata)(nil)

// FileMetadata contains the metadata for a single stored file.
type FileMetadata struct {
	doc
	size           int64
	checksum       string
	checksumFormat string
	timestamp      time.Time
	stored         bool
}

// NewMetadata returns a new Metadata for a file.  ID is left unset (use
// SetID() for that).  Size, Checksum, and ChecksumFormat are left unset
// (use SetFile() for those).  If no timestamp is provided, the
// current one is used.
func NewMetadata(timestamp *time.Time) *FileMetadata {
	meta := FileMetadata{}
	if timestamp == nil {
		meta.timestamp = time.Now().UTC()
	} else {
		meta.timestamp = *timestamp
	}
	return &meta
}

func (m *FileMetadata) Size() int64 {
	return m.size
}

func (m *FileMetadata) Checksum() string {
	return m.checksum
}

func (m *FileMetadata) ChecksumFormat() string {
	return m.checksumFormat
}

func (m *FileMetadata) Timestamp() time.Time {
	return m.timestamp
}

func (m *FileMetadata) Stored() bool {
	return m.stored
}

func (m *FileMetadata) Doc() interface{} {
	return m
}

func (m *FileMetadata) SetFile(size int64, checksum, format string) error {
	// Fall back to existing values.
	if size == 0 {
		size = m.size
	}
	if checksum == "" {
		checksum = m.checksum
	}
	if format == "" {
		format = m.checksumFormat
	}
	if checksum != "" {
		if format == "" {
			return errors.Errorf("missing checksum format")
		}
	} else if format != "" {
		return errors.Errorf("missing checksum")
	}
	// Only allow setting once.
	if m.size != 0 && size != m.size {
		return errors.Errorf("file information (size) already set")
	}
	if m.checksum != "" && checksum != m.checksum {
		return errors.Errorf("file information (checksum) already set")
	}
	if m.checksumFormat != "" && format != m.checksumFormat {
		return errors.Errorf("file information (checksum format) already set")
	}
	// Set the values.
	m.size = size
	m.checksum = checksum
	m.checksumFormat = format
	return nil
}

func (m *FileMetadata) SetStored() {
	m.stored = true
}
