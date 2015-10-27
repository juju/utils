// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package utils

import (
	"math/rand"
	"time"
)

// Timer implements a timer that will signal after a
// internally stored duration. The steps as well as min and max
// durations are declared upon initialization and depend on
// the particular implementation
type Timer interface {
	// Channel returns the channel that can be listened for events.
	Channel() <-chan struct{}

	// Reset will reset the timer to it's minimum value
	// and stop any existing timer.
	Reset()

	// Signal will send a signal on the channel returned by Channel
	// after a internally stored duration and increase the duration
	// for the next call. Only one signal can be pending to be sent
	// at a time.
	Signal()
}

// NewBackoffTimer creates and initializer a new BackoffTimer
// A backoff timer starts at min and gets multiplied by factor
// until it reaches max. Jitter determines whether a small
// randomization is added to the duration.
func NewBackoffTimer(min, max time.Duration, jitter bool, factor int64) Timer {
	return &BackoffTimer{
		Min:             min,
		Max:             max,
		Jitter:          jitter,
		Factor:          factor,
		Chan:            make(chan struct{}, 1),
		CurrentDuration: min,
	}
}

// BackoffTimer creates and initializer a new BackoffTimer
// A backoff timer starts at min and gets multiplied by factor
// until it reaches max. Jitter determines whether a small
// randomization is added to the duration.
// The struct is mainly exposed for testing purposes.
type BackoffTimer struct {
	Timer StdTimer

	Min    time.Duration
	Max    time.Duration
	Jitter bool
	Factor int64

	Chan chan struct{}

	CurrentDuration time.Duration
}

// Channel implements the Timer interface
func (t *BackoffTimer) Channel() <-chan struct{} {
	return t.Chan
}

// Signal implements the Timer interface
// Any existing timer execution is stopped before
// a new one is created.
func (t *BackoffTimer) Signal() {
	if t.Timer != nil {
		t.Timer.Stop()
	}
	t.Timer = afterFunc(t.CurrentDuration, func() {
		t.Chan <- struct{}{}
	})
	// Since it's a backoff timer we will increase
	// the duration after each signal.
	t.increaseDuration()
}

// increaseDuration will increase the duration based on
// the current value and the factor. If jitter is true
// it will add a 0.3% jitter to the final value.
func (t *BackoffTimer) increaseDuration() {
	current := int64(t.CurrentDuration)
	nextDuration := time.Duration(current * t.Factor)
	if t.Jitter {
		// Get a factor in [-1; 1]
		randFactor := (rand.Float64() * 2) - 1
		jitter := float64(nextDuration) * randFactor * 0.03
		nextDuration = nextDuration + time.Duration(jitter)
	}
	if nextDuration > t.Max {
		nextDuration = t.Max
	}
	t.CurrentDuration = nextDuration
}

// Reset implements the Timer interface
func (t *BackoffTimer) Reset() {
	if t.Timer != nil {
		t.Timer.Stop()
	}
	if t.CurrentDuration > t.Min {
		t.CurrentDuration = t.Min
	}
}

// StdTimer defines a interface for time.Timer
// so it's easier to mock it in tests
type StdTimer interface {
	// Reset changes the timer to expire after duration d. It returns true
	// if the timer had been active, false if the timer had expired or been stopped.
	Reset(time.Duration) bool

	// Stop prevents the Timer from firing. It returns true if the call stops the timer,
	// false if the timer has already expired or been stopped. Stop does not close the
	// channel, to prevent a read from the channel succeeding incorrectly.
	Stop() bool
}

var afterFunc = func(d time.Duration, f func()) StdTimer {
	return time.AfterFunc(d, f)
}
