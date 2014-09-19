// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"io"
	"testing"

	"github.com/juju/errors"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filestorage"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

// FakeMetadataStorage is used for testing.
type FakeMetadataStorage struct {
	calls []string

	id       string
	meta     filestorage.Metadata
	metaList []filestorage.Metadata
	err      error

	idArg   string
	metaArg filestorage.Metadata
}

func (s *FakeMetadataStorage) Doc(id string) (filestorage.Doc, error) {
	s.calls = append(s.calls, "Doc")
	s.idArg = id
	if s.err != nil {
		return nil, s.err
	}
	return s.meta, nil
}

func (s *FakeMetadataStorage) ListDocs() ([]filestorage.Doc, error) {
	s.calls = append(s.calls, "ListDoc")
	if s.err != nil {
		return nil, s.err
	}
	var docs []filestorage.Doc
	for _, doc := range s.metaList {
		docs = append(docs, doc)
	}
	return docs, nil
}

func (s *FakeMetadataStorage) AddDoc(doc filestorage.Doc) (string, error) {
	s.calls = append(s.calls, "AddDoc")
	meta, err := filestorage.Convert(doc)
	if err != nil {
		return "", errors.Trace(err)
	}
	s.metaArg = meta
	return s.id, nil
}

func (s *FakeMetadataStorage) RemoveDoc(id string) error {
	s.calls = append(s.calls, "RemoveDoc")
	s.idArg = id
	return s.err
}

func (s *FakeMetadataStorage) Metadata(id string) (filestorage.Metadata, error) {
	s.calls = append(s.calls, "Metadata")
	s.idArg = id
	if s.err != nil {
		return nil, s.err
	}
	return s.meta, nil
}

func (s *FakeMetadataStorage) ListMetadata() ([]filestorage.Metadata, error) {
	s.calls = append(s.calls, "ListMetadata")
	if s.err != nil {
		return nil, s.err
	}
	return s.metaList, nil
}

func (s *FakeMetadataStorage) AddMetadata(meta filestorage.Metadata) (string, error) {
	s.calls = append(s.calls, "AddMetadata")
	s.metaArg = meta
	if s.err != nil {
		return "", s.err
	}
	return s.id, nil
}

func (s *FakeMetadataStorage) RemoveMetadata(id string) error {
	s.calls = append(s.calls, "RemoveMetadata")
	s.idArg = id
	return s.err
}

func (s *FakeMetadataStorage) SetStored(meta filestorage.Metadata) error {
	s.calls = append(s.calls, "SetStored")
	s.metaArg = meta
	return s.err
}

// FakeRawFileStorage is used for testing.
type FakeRawFileStorage struct {
	calls []string

	file io.ReadCloser
	err  error

	idArg   string
	fileArg io.Reader
	sizeArg int64
}

func (s *FakeRawFileStorage) File(id string) (io.ReadCloser, error) {
	s.calls = append(s.calls, "File")
	s.idArg = id
	if s.err != nil {
		return nil, s.err
	}
	return s.file, nil
}

func (s *FakeRawFileStorage) AddFile(id string, file io.Reader, size int64) error {
	s.calls = append(s.calls, "AddFile")
	s.idArg = id
	s.fileArg = file
	s.sizeArg = size
	return s.err
}

func (s *FakeRawFileStorage) RemoveFile(id string) error {
	s.calls = append(s.calls, "RemoveFile")
	s.idArg = id
	return s.err
}
