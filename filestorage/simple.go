// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/juju/errors"
	"github.com/juju/utils"
)

func NewSimpleStorage(dirname string) (FileStorage, error) {
	meta := NewMetadataStorage()
	files, err := NewRawFileStorage(dirname)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return NewFileStorage(meta, files), nil
}

//---------------------------
// metadata

type metadataStorage struct {
	metadata map[string]Metadata
}

func NewMetadataStorage() MetadataStorage {
	stor := metadataStorage{
		metadata: make(map[string]Metadata),
	}
	return &stor
}

func (s *metadataStorage) Doc(id string) (interface{}, error) {
	meta, err := s.Metadata(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return interface{}(meta), nil
}

func (s *metadataStorage) Metadata(id string) (Metadata, error) {
	meta, ok := s.metadata[id]
	if !ok {
		return nil, errors.NotFoundf(id)
	}
	return meta, nil
}

func (s *metadataStorage) ListDocs() ([]interface{}, error) {
	list, err := s.ListMetadata()
	if err != nil {
		return nil, errors.Trace(err)
	}
	docs := []interface{}{}
	for _, doc := range list {
		if doc == nil {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (s *metadataStorage) ListMetadata() ([]Metadata, error) {
	list := []Metadata{}
	for _, meta := range s.metadata {
		if meta == nil {
			continue
		}
		list = append(list, meta)
	}
	return list, nil
}

func (s *metadataStorage) AddDoc(doc interface{}) (string, error) {
	meta, ok := doc.(Metadata)
	if !ok {
		return "", errors.Errorf("doc must be a Metadata")
	}

	uuid, err := utils.NewUUID()
	if err != nil {
		return "", errors.Annotate(err, "error while creating ID")
	}
	id := uuid.String()
	alreadySet := meta.SetID(id)
	if alreadySet {
		return "", errors.AlreadyExistsf("ID already set (tried %q)", id)
	}

	s.metadata[id] = meta
	return id, nil
}

func (s *metadataStorage) RemoveDoc(id string) error {
	if _, ok := s.metadata[id]; !ok {
		return errors.NotFoundf(id)
	}
	delete(s.metadata, id)
	return nil
}

func (s *metadataStorage) New() Metadata {
	return &FileMetadata{timestamp: time.Now().UTC()}
}

func (s *metadataStorage) SetStored(meta Metadata) error {
	meta.SetStored()
	return nil
}

//---------------------------
// raw files

type fsStorage struct {
	dirname string
}

func NewRawFileStorage(dirname string) (RawFileStorage, error) {
	stor := fsStorage{
		dirname: dirname,
	}
	if err := os.MkdirAll(dirname, 0777); err != nil {
		return nil, errors.Annotate(err, "error while creating directory")
	}
	return &stor, nil
}

func (s *fsStorage) File(id string) (io.ReadCloser, error) {
	filename := filepath.Join(s.dirname, id)
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.Annotate(err, "error while opening file")
	}
	return file, nil
}

func (s *fsStorage) AddFile(id string, file io.Reader, size int64) error {
	filename := filepath.Join(s.dirname, id)
	target, err := os.Create(filename)
	if err != nil {
		return errors.Annotate(err, "error while creating file")
	}
	defer target.Close()
	_, err = io.Copy(target, file)
	if err != nil {
		return errors.Annotate(err, "error while writing to file")
	}
	return nil
}

func (s *fsStorage) RemoveFile(id string) error {
	filename := filepath.Join(s.dirname, id)
	err := os.Remove(filename)
	if os.IsNotExist(err) {
		return errors.NotFoundf(id)
	} else if err != nil {
		return errors.Annotate(err, "error removing file")
	}
	return nil
}
