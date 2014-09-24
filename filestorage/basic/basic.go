// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"

	"github.com/juju/utils/filestorage"
)

func NewSimpleStorage(dirname string) (filestorage.FileStorage, error) {
	meta := NewMetadataStorage()
	files, err := NewRawFileStorage(dirname)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return filestorage.NewFileStorage(meta, files), nil
}
