// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package document

import (
	"io"
)

// Document represents a uniquely identifiable structured document.
type Document interface {
	// ID is the unique ID assigned by the storage system.
	ID() string

	// SetID sets the ID of the Doc.  If the ID is already set, SetID()
	// should return true (false otherwise).
	SetID(id string) (alreadySet bool)
	// Copy returns a new copy of the metadata updated with the given ID.
	Copy(id string) Document
}

// DocumentStorage is an abstraction for a system that can store docs
// (structs).  The system is expected to generate its own unique ID for
// each doc.
type DocumentStorage interface {
	io.Closer

	// Doc returns the doc that matches the ID.  If there is no match,
	// an error is returned (see errors.IsNotFound).  Any other problem
	// also results in an error.
	Doc(id string) (Document, error)
	// ListDocs returns a list of all the docs in the storage.
	ListDocs() ([]Document, error)
	// AddDoc adds the doc to the storage.  If successful, the storage-
	// generated ID for the doc is returned.  Otherwise an error is
	// returned.
	AddDoc(doc Document) (string, error)
	// RemoveDoc removes the matching doc from the storage.  If there
	// is no match an error is returned (see errors.IsNotFound).  Any
	// other problem also results in an error.
	RemoveDoc(id string) error
}
