// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package clock

import "time"

// Clock provides an interface for dealing with clocks.
type Clock interface {

	// Now returns the current clock time.
	Now() time.Time

	// After waits for the duration to elapse and then sends the
	// current time on the returned channel.
	After(time.Duration) <-chan time.Time
}

// Alarm returns a channel that will have the time sent on it at some point
// after the supplied time occurs.
//
// This is short for c.After(t.Sub(c.Now())).
func Alarm(c Clock, t time.Time) <-chan time.Time {
	return c.After(t.Sub(c.Now()))
}
