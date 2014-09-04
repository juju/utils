// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"

	"github.com/juju/utils/filestorage"
)

type metadataStorage struct {
	filestorage.MetadataDocStorage
	docStor *docStorage
}

// NewMetadataStorage provides a simple memory-backed MetadataStorage.
func NewMetadataStorage() filestorage.MetadataStorage {
	docStor := NewDocStorage()
	stor := metadataStorage{
		MetadataDocStorage: filestorage.MetadataDocStorage{docStor},
		docStor:            docStor.(*docStorage),
	}
	return &stor
}

// SetStored implements MetadataStorage.SetStored.
func (s *metadataStorage) SetStored(id string) error {
	doc, err := s.docStor.lookUp(id)
	if err != nil {
		return errors.Trace(err)
	}
	meta, ok := doc.(filestorage.Metadata)
	if !ok {
		return errors.Errorf("doc wasn't Metadata (got %v)", doc)
	}
	meta.SetStored()
	return nil
}
