// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package storage

import (
	"time"

	"github.com/juju/utils/document"
)

// RawMetadata holds the data exposed by the Metadata interface.
type RawMetadata struct {
	// Stored indicates when the item was stored.
	Stored *time.Time
}

// Ensure StorageMetadata implements Metadata.
var _ Metadata = (*StorageMetadata)(nil)

// StorageMetadata contains the metadata for a single stored item.
type StorageMetadata struct {
	document.Doc

	// Raw holds the raw data backing the doc.
	Raw RawMetadata
}

// NewMetadata returns a new Metadata for an item.  ID is left unset (use
// SetID() for that).  Likewise for Stored.  If no timestamp is provided,
// the current one is used.
func NewMetadata(created *time.Time) *StorageMetadata {
	doc := document.NewDocument(created)
	meta := StorageMetadata{
		Doc: *doc,
	}
	return &meta
}

// Stored implements Metadata.Stored.
func (m *StorageMetadata) Stored() *time.Time {
	return m.Raw.Stored
}

// SetStored implements Metadata.SetStored.  If Stored is already set,
// SetStored() will return true (false otherwise).
func (m *StorageMetadata) SetStored(timestamp *time.Time) bool {
	if m.Raw.Stored != nil {
		return true
	}
	if timestamp == nil {
		now := time.Now().UTC()
		timestamp = &now
	}
	m.Raw.Stored = timestamp
	return false
}

// Copy implements Doc.Copy.
func (m *StorageMetadata) Copy(id string) document.Document {
	copied := StorageMetadata{
		Doc: *(m.Doc.Copy(id).(*document.Doc)),
		Raw: m.Raw,
	}
	return &copied
}
