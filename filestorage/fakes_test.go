// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"io"

	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filestorage"
)

// FakeMetadataStorage is used as a DocStorage and MetadataStorage for
// testing purposes.
type FakeMetadataStorage struct {
	calls []string

	id       string
	meta     filestorage.Metadata
	metaList []filestorage.Metadata
	err      error

	idArg   string
	metaArg filestorage.Metadata
}

// Check verfies the state of the fake.
func (s *FakeMetadataStorage) Check(c *gc.C, id string, meta filestorage.Metadata, calls ...string) {
	c.Check(s.calls, jc.DeepEquals, calls)
	c.Check(s.idArg, gc.Equals, id)
	c.Check(s.metaArg, gc.Equals, meta)
}

func (s *FakeMetadataStorage) Doc(id string) (filestorage.Document, error) {
	s.calls = append(s.calls, "Doc")
	s.idArg = id
	if s.err != nil {
		return nil, s.err
	}
	return s.meta, nil
}

func (s *FakeMetadataStorage) ListDocs() ([]filestorage.Document, error) {
	s.calls = append(s.calls, "ListDoc")
	if s.err != nil {
		return nil, s.err
	}
	var docs []filestorage.Document
	for _, doc := range s.metaList {
		docs = append(docs, doc)
	}
	return docs, nil
}

func (s *FakeMetadataStorage) AddDoc(doc filestorage.Document) (string, error) {
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

func (s *FakeMetadataStorage) Close() error {
	s.calls = append(s.calls, "Close")
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

func (s *FakeMetadataStorage) SetStored(id string) error {
	s.calls = append(s.calls, "SetStored")
	s.idArg = id
	return s.err
}

// FakeRawFileStorage is used in testing as a RawFileStorage.
type FakeRawFileStorage struct {
	calls []string

	file io.ReadCloser
	err  error

	idArg   string
	fileArg io.Reader
	sizeArg int64
}

// Check verfies the state of the fake.
func (s *FakeRawFileStorage) Check(c *gc.C, id string, file io.Reader, size int64, calls ...string) {
	c.Check(s.calls, jc.DeepEquals, calls)
	c.Check(s.idArg, gc.Equals, id)
	c.Check(s.fileArg, gc.Equals, file)
	c.Check(s.sizeArg, gc.Equals, size)
}

// CheckNotUsed verifies that the fake was not used.
func (s *FakeRawFileStorage) CheckNotUsed(c *gc.C) {
	s.Check(c, "", nil, 0)
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

func (s *FakeRawFileStorage) Close() error {
	s.calls = append(s.calls, "Close")
	return s.err
}
