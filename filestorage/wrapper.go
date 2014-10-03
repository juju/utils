// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"

	"github.com/juju/errors"
)

// Ensure fileStorage implements FileStorage.
var _ = FileStorage((*fileStorage)(nil))

type fileStorage struct {
	metaStor MetadataStorage
	fileStor RawFileStorage
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
		metaStor: meta,
		fileStor: files,
	}
	return &stor
}

// Metadata returns the matching metadata.  Failure to find it (see
// errors.IsNotFound) or any other problem results in an error.
func (s *fileStorage) Metadata(id string) (Metadata, error) {
	meta, err := s.metaStor.Metadata(id)
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
	file, err := s.fileStor.File(id)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return meta, file, nil
}

// List returns a list of the metadata for all files in the storage.
func (s *fileStorage) List() ([]Metadata, error) {
	return s.metaStor.ListMetadata()
}

func (s *fileStorage) addFile(id string, size int64, file io.Reader) error {
	err := s.fileStor.AddFile(id, file, size)
	if err != nil {
		return errors.Trace(err)
	}
	err = s.metaStor.SetStored(id)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Add adds the file to the storage.  It returns the unique ID generated
// by the storage for the file.  If no file is provided, only the
// metadata is stored.  While the passed-in "meta" is not modified, the
// new ID and "stored" flag will be saved in metadata storage.  Feel
// free to explicitly call meta.SetID() and meta.SetStored() afterward.
//
// The metadata is added first, so if storing the raw file fails the
// metadata will still be stored.  A non-empty returned ID indicates
// that the metadata was stored successfully.
func (s *fileStorage) Add(meta Metadata, file io.Reader) (string, error) {
	id, err := s.metaStor.AddMetadata(meta)
	if err != nil {
		return "", errors.Trace(err)
	}

	if file != nil {
		err = s.addFile(id, meta.Size(), file)
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
	err = s.addFile(id, meta.Size(), file)
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
	err := s.fileStor.RemoveFile(id)
	if err != nil && !errors.IsNotFound(err) {
		return errors.Trace(err)
	}
	err = s.metaStor.RemoveMetadata(id)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Close implements io.Closer.Close.
func (s *fileStorage) Close() error {
	ferr := s.fileStor.Close()
	merr := s.metaStor.Close()
	if ferr == nil {
		return errors.Trace(merr)
	} else if merr == nil {
		return errors.Trace(ferr)
	} else {
		msg := "closing both failed: metadata (%v) and files (%v)"
		return errors.Errorf(msg, merr, ferr)
	}
}
