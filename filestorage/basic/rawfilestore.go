// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"io"
	"os"
	"path/filepath"

	"github.com/juju/errors"

	"github.com/juju/utils/filestorage"
)

type fsStorage struct {
	dirname string
}

func NewRawFileStorage(dirname string) (filestorage.RawFileStorage, error) {
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

// Close implements io.Closer.Close.
func (s *fsStorage) Close() error {
	return nil
}
