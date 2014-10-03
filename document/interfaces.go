// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package document

import (
	"io"
	"time"
)

// Document represents a uniquely identifiable structured document.
type Document interface {
	// ID is the unique ID assigned by the storage system.
	ID() string
	// Created records when the doc was created.
	Created() time.Time

	// SetID sets the ID of the Doc.  If the ID is already set, SetID()
	// should return true (false otherwise).
	SetID(id string) (alreadySet bool)
	// Copy returns a new copy of the metadata updated with the given ID.
	Copy(id string) Document
	// Dump writes the doc to the writer in the specified serialization
	// format.  If the format is not supported, errors.NotSupported is
	// returned.
	Dump(w io.Writer, format string) error
	// Load deserializes the doc, using the specified format, from the
	// reader.  Values from reader overwrite the corresponding values
	// that may already be set on the doc.  If the format is not
	// supported, errors.NotSupported is returned.
	Load(w io.Reader, format string) error
	// DefaultID returns an ID string derived from the doc that may be
	// used for the doc.  If the doc does not support a default ID,
	// errors.NotSupported is returned.
	DefaultID() (string, error)
	// Validate checks that the doc is populated to the specified level
	// and that the populated values are valid.  At least 2 levels are
	// always supported: "full" and "initialized".  These correspond to
	// the highest and lowest levels of validation, respectively.  If
	// the level is unrecognized, errors.NotSupported is returned.  If
	// the doc is not valid, errors.NotValid is returned.
	Validate(level string) error
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
