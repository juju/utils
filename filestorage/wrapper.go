// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"

	"github.com/juju/errors"
)

type DocStorage interface {
	Doc(id string) (interface{}, error)
	ListDocs() ([]interface{}, error)
	AddDoc(doc interface{}) (string, error)
	RemoveDoc(id string) error
}

type RawFileStorage interface {
	File(id string) (io.ReadCloser, error)
	AddFile(id string, file io.Reader, size int64) error
	RemoveFile(id string) error
}

type MetadataStorage interface {
	DocStorage
	Metadata(id string) (Metadata, error)
	ListMetadata() ([]Metadata, error)
	New() Metadata
	SetStored(meta Metadata) error
}

//---------------------------
// wrapper implementation

type fileStorage struct {
	metadata MetadataStorage
	files    RawFileStorage
}

func NewFileStorage(meta MetadataStorage, files RawFileStorage) FileStorage {
	stor := fileStorage{
		metadata: meta,
		files:    files,
	}
	return &stor
}

func (s *fileStorage) Metadata(id string) (Metadata, error) {
	meta, err := s.metadata.Metadata(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return meta, nil
}

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

func (s *fileStorage) Add(meta Metadata, file io.Reader) (string, error) {
	id, err := s.metadata.AddDoc(meta)
	if err != nil {
		return "", errors.Trace(err)
	}
	meta.SetID(id)

	if file != nil {
		err = s.addFile(meta, file)
		if err != nil {
			return "", errors.Trace(err)
		}
	}

	return id, nil
}

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
