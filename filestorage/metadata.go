// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"time"

	"github.com/juju/errors"
)

// RawDoc is a basic, uniquely identifiable document.
type RawDoc struct {
	// ID is the unique identifier for the document.
	ID string
}

// DocWrapper wraps a document in the Document interface.
type DocWrapper struct {
	Raw *RawDoc
}

// ID returns the document's unique identifier.
func (d *DocWrapper) ID() string {
	return d.Raw.ID
}

// SetID sets the document's unique identifier.  If the ID is already
// set, SetID() returns true (false otherwise).
func (d *DocWrapper) SetID(id string) bool {
	if d.Raw.ID != "" {
		return true
	}
	d.Raw.ID = id
	return false
}

// Copy returns a copy of the document.
func (d *DocWrapper) Copy() Document {
	copied := *d.Raw
	copied.ID = ""
	return &DocWrapper{&copied}
}

// FileMetadata contains the metadata for a single stored file.
type FileMetadata struct {
	DocWrapper
	size           int64
	checksum       string
	checksumFormat string
	stored         *time.Time
}

// NewMetadata returns a new Metadata for a file.  ID is left unset (use
// SetID() for that).  Size, Checksum, and ChecksumFormat are left unset
// (use SetFile() for those).  If no timestamp is provided, the
// current one is used.
func NewMetadata() *FileMetadata {
	meta := FileMetadata{}
	meta.DocWrapper.Raw = &RawDoc{}
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

func (m *FileMetadata) Stored() *time.Time {
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

func (m *FileMetadata) SetStored(timestamp *time.Time) {
	if timestamp == nil {
		now := time.Now().UTC()
		m.stored = &now
	} else {
		m.stored = timestamp
	}
}

// Copy returns a copy of the document.
func (m *FileMetadata) Copy() Document {
	copied := *m
	doc := m.DocWrapper.Copy().(*DocWrapper)
	copied.DocWrapper = *doc
	return &copied
}
