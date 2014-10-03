// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package storage

import (
	"time"

	"github.com/juju/utils/document"
)

// Metadata is a document containing information about something that
// may be stored.
type Metadata interface {
	document.Document

	// Stored records when the doc stored.  If the doc has not been
	// stored, nil is returned.
	Stored() *time.Time

	// SetStored sets Stored to when the doc was stored.  If no
	// timestamp is provided, the current timestamp is used.  If Stored
	// is already set, SetStored() will return true (false otherwise).
	SetStored(timestamp *time.Time) bool
}
