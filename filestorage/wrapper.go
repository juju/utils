// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"

	"github.com/juju/errors"
)

// DocStorage is an abstraction for a system that can store docs (structs).
// The system is expected to generate its own unique ID for each doc.
type DocStorage interface {
	// Doc returns the doc that matches the ID.  If there is no match,
	// an error is returned (see errors.IsNotFound).  Any other problem
	// also results in an error.
	Doc(id string) (interface{}, error)
	// ListDocs returns a list of all the docs in the storage.
	ListDocs() ([]interface{}, error)
	// AddDoc adds the doc to the storage.  If successful, the storage-
	// generated ID for the doc is returned.  Otherwise an error is
	// returned.
	AddDoc(doc interface{}) (string, error)
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
	// New returns a new Metadata value, initialized by the storage.
	// This value is not added to the storage until explicitly done so
	// by a call to AddDoc().
	New() Metadata
	// SetStored updates the stored metadata to indicate that the
	// associated file has been successfully stored in a RawFileStorage
	// system.  It will also call SetStored() on the metadata.  If it
	// does not find a stored metadata with the matching ID, it will
	// return an error (see errors.IsNotFound).  It also returns an
	// error if it fails to update the stored metadata.
	SetStored(meta Metadata) error
}

// Ensure fileStorage implements FileStorage.
var _ = FileStorage((*fileStorage)(nil))

type fileStorage struct {
	metadata MetadataStorage
	files    RawFileStorage
}

// NewFileStorage returns a new FileStorage value that wraps a
// MetadataStorage and a RawFileStorage.  It coordinates the two even
// though they may not be designed to be compatible (or the two may be
// the same value).
//
// A stored file will always have a metadata value stored.  However, it
// is not required to have a raw file stored.
func NewFileStorage(meta MetadataStorage, files RawFileStorage) FileStorage {
	stor := fileStorage{
		metadata: meta,
		files:    files,
	}
	return &stor
}

// Metadata returns the matching metadata.  Failure to find it (see
// errors.IsNotFound) or any other problem results in an error.
func (s *fileStorage) Metadata(id string) (Metadata, error) {
	meta, err := s.metadata.Metadata(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return meta, nil
}

// Get returns the matching file and its associated metadata.  If there
// is no match (see errors.IsNotFound) or any other problem, it returns
// an error.  Both the metadata and file must have been stored for the
// file to be considered found.
func (s *fileStorage) Get(id string) (Metadata, io.ReadCloser, error) {
	meta, err := s.Metadata(id)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	if !meta.Stored() {
		return nil, nil, errors.NotFoundf("no file stored for %q", id)
	}
	file, err := s.files.File(id)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return meta, file, nil
}

// List returns a list of the metadata for all files in the storage.
func (s *fileStorage) List() ([]Metadata, error) {
	return s.metadata.ListMetadata()
}

func (s *fileStorage) addFile(meta Metadata, file io.Reader) error {
	err := s.files.AddFile(meta.ID(), file, meta.Size())
	if err != nil {
		return errors.Trace(err)
	}
	err = s.metadata.SetStored(meta)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Add adds the file to the storage.  It returns the unique ID generated
// by the storage for the file.  Any problem (including an existing
// file, see errors.IsAlreadyExists) results in an error.
//
// The metadata is added first, so if storing the raw file fails the
// metadata will still be stored.  A non-empty returned ID indicates
// that the metadata was stored successfully.
func (s *fileStorage) Add(meta Metadata, file io.Reader) (string, error) {
	id, err := s.metadata.AddDoc(meta)
	if err != nil {
		return "", errors.Trace(err)
	}
	meta.SetID(id)

	if file != nil {
		err = s.addFile(meta, file)
		if err != nil {
			return id, errors.Trace(err)
		}
	}

	return id, nil
}

// SetFile stores the raw file for an existing metadata.  If there is no
// matching stored metadata an error is returned (see errors.IsNotFound).
// If a file has already been stored an error is returned (see
// errors.IsAlreadyExists).  Any other failure to add the file also
// results in an error.
func (s *fileStorage) SetFile(id string, file io.Reader) error {
	meta, err := s.Metadata(id)
	if err != nil {
		return errors.Trace(err)
	}
	err = s.addFile(meta, file)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Remove removes both the metadata and raw file from the storage.  If
// there is no match an error is returned (see errors.IsNotFound).
//
// The raw file is removed first.  Thus if there is any problem after
// removing the raw file, the metadata will still be stored.  However,
// in that case the stored metadata is not guaranteed to accurately
// represent that there is no corresponding raw file in storage.
func (s *fileStorage) Remove(id string) error {
	err := s.files.RemoveFile(id)
	if err != nil && !errors.IsNotFound(err) {
		return errors.Trace(err)
	}
	err = s.metadata.RemoveDoc(id)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
