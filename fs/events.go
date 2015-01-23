// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"os"

	"github.com/juju/errors"
)

const (
	OpExists     = "Exists"
	OpMkdirAll   = "MkdirAll"
	OpReadFile   = "ReadFile"
	OpCreateFile = "CreateFile"
	OpRemoveAll  = "RemoveAll"
	OpChmod      = "Chmod"
)

type OpEvent struct {
	Kind   string
	Target string
	Mode   os.FileMode
}

type OpEvents struct {
	events []OpEvent
}

func (oe *OpEvents) Add(name, target string) *OpEvent {
	event := OpEvent{
		Kind:   name,
		Target: target,
	}
	oe.events = append(oe.events, event)
	return &event
}

type commandRenderer interface {
	// TODO(ericsnow) Replace Render with event-specific methods?
	// Otherwise every rendered must know about OpEvent and make
	// the same decisions about how to handle each one.

	Render(event OpEvent) ([]string, error)
}

// TODO(ericsnow) Implement a renderer for Windows and one for
// Linux (under Bash).

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
