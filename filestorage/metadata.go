// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"time"
)

type Metadata interface {
	// ID is the unique ID assigned by the storage system.
	ID() string
	// Size is the size of the file (in bytes).
	Size() int64
	// Checksum is the checksum for the file.
	Checksum() string
	// ChecksumFormat is the kind (and encoding) of checksum.
	ChecksumFormat() string
	// Timestamp records when the file was created.
	Timestamp() time.Time
	// Stored indicates whether or not the file has been stored.
	Stored() bool

	// Doc returns a storable copy of the metadata.
	Doc() interface{}
	// SetID sets the ID of the metadata.  If the ID is already set,
	// SetID() should return true (false otherwise).
	SetID(id string) (alreadySet bool)
	// SetStored sets Stored to true the metadata.
	SetStored()
}

// Ensure FileMetadata implements Metadata.
var _ = Metadata(&FileMetadata{})

// FileMetadata contains the metadata for a single stored file.
type FileMetadata struct {
	id             string
	size           int64
	checksum       string
	checksumFormat string
	timestamp      time.Time
	stored         bool
}

// NewMetadata returns a new Metadata for a file.  ID is left unset (use
// SetID() for that).  If no timestamp is provided, the current one is
// used.  Everything else should be provided.
func NewMetadata(
	size int64, checksum, checksumFormat string, timestamp *time.Time,
) Metadata {
	meta := FileMetadata{
		// id is omitted.
		size:           size,
		checksum:       checksum,
		checksumFormat: checksumFormat,
	}
	if timestamp == nil {
		meta.timestamp = time.Now().UTC()
	} else {
		meta.timestamp = *timestamp
	}
	return &meta
}

func (m *FileMetadata) ID() string {
	return m.id
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

func (m *FileMetadata) SetID(id string) bool {
	if m.id != "" {
		return true
	}
	m.id = id
	return false
}

func (m *FileMetadata) SetStored() {
	m.stored = true
}
