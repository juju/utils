// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"github.com/juju/errors"
)

type metadataStorage struct {
	MetadataDocStorage
}

// NewMetadataStorage provides a simple memory-backed MetadataStorage.
func NewMetadataStorage() MetadataStorage {
	stor := metadataStorage{
		MetadataDocStorage: MetadataDocStorage{&docStorage{}},
	}
	return &stor
}

// SetStored implements MetadataStorage.SetStored.
func (s *metadataStorage) SetStored(meta Metadata) error {
	id := meta.ID()
	if id == "" {
		return errors.NotFoundf("metadata missing ID")
	}
	found, err := s.Metadata(id)
	if err != nil {
		return errors.Trace(err)
	}

	found.SetStored()
	meta.SetStored()
	return nil
}
