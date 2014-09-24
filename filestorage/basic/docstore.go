// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"
	"github.com/juju/utils"

	"github.com/juju/utils/filestorage"
)

type docStorage struct {
	docs map[string]filestorage.Doc
}

// NewDocStorage returns a simple memory-backed DocStorage.
func NewDocStorage() filestorage.DocStorage {
	return &docStorage{}
}

// Doc implements DocStorage.Doc.
func (s *docStorage) Doc(id string) (filestorage.Doc, error) {
	doc, ok := s.docs[id]
	if !ok {
		return nil, errors.NotFoundf(id)
	}
	return doc, nil
}

// ListDocs implements DocStorage.ListDocs.
func (s *docStorage) ListDocs() ([]filestorage.Doc, error) {
	var list []filestorage.Doc
	for _, doc := range s.docs {
		if doc == nil {
			continue
		}
		list = append(list, doc)
	}
	return list, nil
}

// AddDoc implements DocStorage.AddDoc.
func (s *docStorage) AddDoc(doc filestorage.Doc) (string, error) {
	uuid, err := utils.NewUUID()
	if err != nil {
		return "", errors.Annotate(err, "error while creating ID")
	}
	id := uuid.String()

	alreadySet := doc.SetID(id)
	if alreadySet {
		return "", errors.AlreadyExistsf("ID already set (tried %q)", id)
	}

	if s.docs == nil {
		s.docs = make(map[string]filestorage.Doc)
	}
	s.docs[id] = doc
	return id, nil
}

// RemoveDoc implements DocStorage.RemoveDoc.
func (s *docStorage) RemoveDoc(id string) error {
	if _, ok := s.docs[id]; !ok {
		return errors.NotFoundf(id)
	}
	delete(s.docs, id)
	return nil
}

// Close implements io.Closer.Close.
func (s *docStorage) Close() error {
	return nil
}
