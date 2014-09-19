// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"github.com/juju/errors"
)

func NewSimpleStorage(dirname string) (FileStorage, error) {
	meta := NewMetadataStorage()
	files, err := NewRawFileStorage(dirname)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return NewFileStorage(meta, files), nil
}
