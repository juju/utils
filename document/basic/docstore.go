// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"
	"github.com/juju/utils"

	"github.com/juju/utils/document"
)

type DocStorage struct {
	docs map[string]document.Document
}

// NewDocStorage returns a simple memory-backed DocumentStorage.
func NewDocStorage() document.DocumentStorage {
	storage := DocStorage{
		docs: make(map[string]document.Document),
	}
	return &storage
}

// LookUp returns the stored doc (not a copy).
func (s *DocStorage) LookUp(id string) (document.Document, error) {
	doc, ok := s.docs[id]
	if !ok {
		return nil, errors.NotFoundf(id)
	}
	return doc, nil
}

// Doc implements DocStorage.Doc.
func (s *DocStorage) Doc(id string) (document.Document, error) {
	raw, err := s.LookUp(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return raw.Copy(id), nil
}

// ListDocs implements DocStorage.ListDocs.
func (s *DocStorage) ListDocs() ([]document.Document, error) {
	var list []document.Document
	for _, doc := range s.docs {
		if doc == nil {
			continue
		}
		list = append(list, doc)
	}
	return list, nil
}

// AddDoc implements DocStorage.AddDoc.
func (s *DocStorage) AddDoc(doc document.Document) (string, error) {
	if doc.ID() != "" {
		return "", errors.AlreadyExistsf("ID already set")
	}

	uuid, err := utils.NewUUID()
	if err != nil {
		return "", errors.Annotate(err, "error while creating ID")
	}
	id := uuid.String()
	// We let the caller call meta.SetID() if they so desire.

	s.docs[id] = doc.Copy(id)
	return id, nil
}

// RemoveDoc implements DocStorage.RemoveDoc.
func (s *DocStorage) RemoveDoc(id string) error {
	if _, ok := s.docs[id]; !ok {
		return errors.NotFoundf(id)
	}
	delete(s.docs, id)
	return nil
}

// Close implements io.Closer.Close.
func (s *DocStorage) Close() error {
	return nil
}
