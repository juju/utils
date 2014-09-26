// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"

	"github.com/juju/utils/filestorage"
)

type metadataStorage struct {
	filestorage.MetadataDocStorage
}

// NewMetadataStorage provides a simple memory-backed MetadataStorage.
func NewMetadataStorage() filestorage.MetadataStorage {
	stor := metadataStorage{
		MetadataDocStorage: filestorage.MetadataDocStorage{NewDocStorage()},
	}
	return &stor
}

// SetStored implements MetadataStorage.SetStored.
func (s *metadataStorage) SetStored(meta filestorage.Metadata) error {
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
