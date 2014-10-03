// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"time"

	"github.com/juju/errors"

	"github.com/juju/utils/document"
)

// RawMetadata holds the data exposed by the Metadata interface.
type RawMetadata struct {
	// Size is the size of the stored file.
	Size int64
	// Checksum is the checksum of the stored file.
	Checksum string
	// ChecksumFormat describes the format of the checksum.
	ChecksumFormat string
	// Timestamp is when the file was created.
	Timestamp time.Time
	// Stored indicates that the file has been stored.
	Stored bool
}

// Ensure FileMetadata implements Metadata.
var _ Metadata = (*FileMetadata)(nil)

type doc struct{ document.Doc } // ...to avoid a name conflict.

// FileMetadata contains the metadata for a single stored file.
type FileMetadata struct {
	doc

	// Raw holds the raw data backing the doc.
	Raw RawMetadata
}

// NewMetadata returns a new Metadata for a file.  ID is left unset (use
// SetID() for that).  Size, Checksum, and ChecksumFormat are left unset
// (use SetFile() for those).  If no timestamp is provided, the
// current one is used.
func NewMetadata(timestamp *time.Time) *FileMetadata {
	meta := FileMetadata{}
	if timestamp == nil {
		meta.Raw.Timestamp = time.Now().UTC()
	} else {
		meta.Raw.Timestamp = *timestamp
	}
	return &meta
}

func (m *FileMetadata) Size() int64 {
	return m.Raw.Size
}

func (m *FileMetadata) Checksum() string {
	return m.Raw.Checksum
}

func (m *FileMetadata) ChecksumFormat() string {
	return m.Raw.ChecksumFormat
}

func (m *FileMetadata) Timestamp() time.Time {
	return m.Raw.Timestamp
}

func (m *FileMetadata) Stored() bool {
	return m.Raw.Stored
}

func (m *FileMetadata) Doc() interface{} {
	return m
}

func (m *FileMetadata) SetFile(size int64, checksum, format string) error {
	// Fall back to existing values.
	if size == 0 {
		size = m.Raw.Size
	}
	if checksum == "" {
		checksum = m.Raw.Checksum
	}
	if format == "" {
		format = m.Raw.ChecksumFormat
	}
	if checksum != "" {
		if format == "" {
			return errors.Errorf("missing checksum format")
		}
	} else if format != "" {
		return errors.Errorf("missing checksum")
	}
	// Only allow setting once.
	if m.Raw.Size != 0 && size != m.Raw.Size {
		return errors.Errorf("file information (size) already set")
	}
	if m.Raw.Checksum != "" && checksum != m.Raw.Checksum {
		return errors.Errorf("file information (checksum) already set")
	}
	if m.Raw.ChecksumFormat != "" && format != m.Raw.ChecksumFormat {
		return errors.Errorf("file information (checksum format) already set")
	}
	// Set the values.
	m.Raw.Size = size
	m.Raw.Checksum = checksum
	m.Raw.ChecksumFormat = format
	return nil
}

func (m *FileMetadata) SetStored() {
	m.Raw.Stored = true
}

// Copy implements Doc.Copy.
func (m *FileMetadata) Copy(id string) document.Document {
	copied := FileMetadata{
		doc: doc{*(m.doc.Doc.Copy(id).(*document.Doc))},
		Raw: m.Raw,
	}
	return &copied
}
