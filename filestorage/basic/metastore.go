// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"

	docbasics "github.com/juju/utils/document/basic"
	"github.com/juju/utils/filestorage"
)

type metadataStorage struct {
	filestorage.MetadataDocStorage
	docStor *docbasics.DocStorage
}

// NewMetadataStorage provides a simple memory-backed MetadataStorage.
func NewMetadataStorage() filestorage.MetadataStorage {
	docStor := docbasics.NewDocStorage()
	stor := metadataStorage{
		MetadataDocStorage: filestorage.MetadataDocStorage{docStor},
		docStor:            docStor.(*docbasics.DocStorage),
	}
	return &stor
}

// SetStored implements MetadataStorage.SetStored.
func (s *metadataStorage) SetStored(id string) error {
	doc, err := s.docStor.LookUp(id)
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
