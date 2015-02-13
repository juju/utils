// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"os"
	"time"

	"github.com/juju/errors"
)

// These are the recognized OpEvent kinds. They essentially match the
// methods of the Operations interface.
const (
	OpExists     = "Exists"
	OpInfo       = "Info"
	OpMkdirAll   = "MkdirAll"
	OpListDir    = "ListDir"
	OpReadFile   = "ReadFile"
	OpCreateFile = "CreateFile"
	OpWriteFile  = "WriteFile"
	OpRemoveAll  = "RemoveAll"
	OpChmod      = "Chmod"
	OpSymlink    = "Symlink"
	OpReadlink   = "Readlink"
)

// OpEvent describes a filesystem operation that took place.
type OpEvent struct {
	// Kind is name of the kind of event.
	Kind string

	// Target is the path to the affected file or directory.
	Target string

	// Source is the path to the file or directory from which the
	// operation acted. This is only set for operations that involve
	// a source and target (e.g. copy).
	Source string

	// Permissions is the file permissions of the affected file.
	Permissions os.FileMode

	// Timestamp is when the event happened.
	Timestamp time.Time

	// TODO(ericsnow) Record the User? Notes? Other info?
}

// OpEvents is a record of FS operations that have happened, in the
// order that they happened.
type OpEvents struct {
	events []OpEvent
}

// Add adds a new event to the recorded history, copying the info from
// the provided event. If the timestamp is not set then the current
// time is used. A pointer to the added event is returned so it can be
// further modified after the fact.
func (oe *OpEvents) Add(event OpEvent) *OpEvent {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	// TODO(ericsnow) Fail if the timestamp is not later than that of
	// the last event in the history?
	oe.events = append(oe.events, event)
	return &event
}

// Record builds a new event from the provided information and adds it
// to the event history. The new event is returned.
func (oe *OpEvents) Record(kind, target string) *OpEvent {
	event := OpEvent{
		Kind:   kind,
		Target: target,
	}
	return oe.Add(event)
}

// Reset clears the event history.
func (oe *OpEvents) Reset() {
	oe.events = nil
}

type commandRenderer interface {
	// TODO(ericsnow) Replace Render with event-specific methods?
	// Otherwise every rendered must know about OpEvent and make
	// the same decisions about how to handle each one.

	Render(event OpEvent) ([]string, error)
}

// TODO(ericsnow) Implement a renderer for Windows and one for
// Linux (under Bash).

// ReplayCommands builds the list of commands that would reproduce they
// recorded event history. The provided renderer is used to convert the
// events into commands.
func (oe *OpEvents) ReplayCommands(renderer commandRenderer) ([]string, error) {
	// TODO(ericsnow) This approach does not accommodate the ability
	// of a renderer to look ahead at subsequent events when deciding
	// how to render an event.
	var allCommands []string
	for _, event := range oe.events {
		commands, err := renderer.Render(event)
		if err != nil {
			return nil, errors.Trace(err)
		}
		allCommands = append(allCommands, commands...)
	}
	return allCommands, nil
}
