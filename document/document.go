// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package document

import (
	"time"
)

// RawDoc holds the data exposed by the Document interface.
type RawDoc struct {
	// ID is the document's ID.
	ID string
	// Created is when the document was created.
	Created time.Time
}

var _ Document = (*Doc)(nil)

// Doc is the obvious implementation of Document.  While perhaps useful
// on its own, it is most useful for embedding in other types.
type Doc struct {
	// Raw holds the raw data backing doc.
	Raw RawDoc
}

// NewDocument returns a new Document.  ID is left unset (use SetID()
// for that).  If no created is provided, the current one is used.
func NewDocument(created *time.Time) *Doc {
	doc := Doc{}
	if created == nil {
		doc.Raw.Created = time.Now().UTC()
	} else {
		doc.Raw.Created = *created
	}
	return &doc
}

// ID implements Doc.ID.
func (d *Doc) ID() string {
	return d.Raw.ID
}

// Created implements Doc.Created.
func (d *Doc) Created() time.Time {
	return d.Raw.Created
}

// SetID implements Doc.SetID.  If the ID is already set, SetID()
// will return true (false otherwise).
func (d *Doc) SetID(id string) bool {
	if d.Raw.ID != "" {
		return true
	}
	d.Raw.ID = id
	return false
}

// Copy returns a new Doc with Raw set to a shallow copy of the current
// value.  The raw ID is set to the one passed in.
func (d *Doc) Copy(id string) Document {
	copied := Doc{Raw: d.Raw}
	copied.Raw.ID = id
	return &copied
}
