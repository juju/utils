// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"
	"time"
)

// FileStorage is an abstraction of a system that can be used for the
// storage of files.  The type exposes the essential capabilities of
// such a system.
type FileStorage interface {
	// Metadata returns a file's metadata.
	Metadata(id string) (Metadata, error)
	// Get returns a file and its metadata.
	Get(id string) (Metadata, io.ReadCloser, error)
	// List returns the metadata for each stored file.
	List() ([]Metadata, error)
	// Add stores a file and its metadata.
	Add(meta Metadata, archive io.Reader) (string, error)
	// SetFile stores a file for an existing metadata entry.
	SetFile(id string, file io.Reader) error
	// Remove removes a file from storage.
	Remove(id string) error
}

// Doc is an identifiable document.
type Doc interface {
	// ID is the unique ID assigned by the storage system.
	ID() string
	// SetID sets the ID of the Doc.  If the ID is already set,
	// SetID() should return true (false otherwise).
	SetID(id string) (alreadySet bool)
}

// Metadata is the meta information for a stored file.
type Metadata interface {
	Doc
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
	// SetFile sets the file info on the metadata.
	SetFile(size int64, checksum, checksumFormat string) error
	// SetStored sets Stored to true on the metadata.
	SetStored()
}

// DocStorage is an abstraction for a system that can store docs (structs).
// The system is expected to generate its own unique ID for each doc.
type DocStorage interface {
	// Doc returns the doc that matches the ID.  If there is no match,
	// an error is returned (see errors.IsNotFound).  Any other problem
	// also results in an error.
	Doc(id string) (Doc, error)
	// ListDocs returns a list of all the docs in the storage.
	ListDocs() ([]Doc, error)
	// AddDoc adds the doc to the storage.  If successful, the storage-
	// generated ID for the doc is returned.  Otherwise an error is
	// returned.
	AddDoc(doc Doc) (string, error)
	// RemoveDoc removes the matching doc from the storage.  If there
	// is no match an error is returned (see errors.IsNotFound).  Any
	// other problem also results in an error.
	RemoveDoc(id string) error
}

// RawFileStorage is an abstraction around a system that can store files.
// The system is expected to rely on the user for unique IDs.
type RawFileStorage interface {
	// File returns the matching file.  If there is no match an error is
	// returned (see errors.IsNotFound).  Any other problem also results
	// in an error.
	File(id string) (io.ReadCloser, error)
	// AddFile adds the file to the storage.  If it fails to do so,
	// it returns an error.  If a file is already stored for the ID,
	// AddFile() fails (see errors.IsAlreadyExists).
	AddFile(id string, file io.Reader, size int64) error
	// RemoveFile removes the matching file from the storage.  It fails
	// if there is no error (see errors.IsNotFound).  Any other problem
	// also results in an error.
	RemoveFile(id string) error
}

// MetadataStorage is an extension of DocStorage adapted to file metadata.
type MetadataStorage interface {
	DocStorage
	// Metadata returns the matching Metadata.  It fails if there is no
	// match (see errors.IsNotFound).  Any other problems likewise
	// results in an error.
	Metadata(id string) (Metadata, error)
	// ListMetadata returns a list of all metadata in the storage.
	ListMetadata() ([]Metadata, error)
	// SetStored updates the stored metadata to indicate that the
	// associated file has been successfully stored in a RawFileStorage
	// system.  It will also call SetStored() on the metadata.  If it
	// does not find a stored metadata with the matching ID, it will
	// return an error (see errors.IsNotFound).  It also returns an
	// error if it fails to update the stored metadata.
	SetStored(meta Metadata) error
}
