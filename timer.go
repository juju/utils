// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package utils

import (
	"math/rand"
	"time"
)

// Countdown implements a timer that will call a provided function.
// after a internally stored duration. The steps as well as min and max
// durations are declared upon initialization and depend on
// the particular implementation.
type Countdown interface {
	// Reset stops the timer and resets it's duration to the minimum one.
	// Signal must be called to start the timer again.
	Reset()

	// Start starts the internal timer.
	// At the end of the timer, if Reset hasn't been called in the mean time
	// Func will be called and the duration is increased for the next call.
	Start()
}

// NewBackoffTimer creates and initializer a new BackoffTimer
// A backoff timer starts at min and gets multiplied by factor
// until it reaches max. Jitter determines whether a small
// randomization is added to the duration.
func NewBackoffTimer(info BackoffTimerInfo) *BackoffTimer {
	return &BackoffTimer{
		info:            info,
		currentDuration: info.Min,
		afterFunc: func(d time.Duration, f func()) StoppableTimer {
			return time.AfterFunc(d, f)
		},
	}
}

// BackoffTimer creates and initializer a new BackoffTimer
// A backoff timer starts at min and gets multiplied by factor
// until it reaches max. Jitter determines whether a small
// randomization is added to the duration.
type BackoffTimer struct {
	info BackoffTimerInfo

	timer           StoppableTimer
	currentDuration time.Duration
	afterFunc       func(d time.Duration, f func()) StoppableTimer
}

// BackoffTimerInfo is a helper struct for backoff timer
// that encapsulates config information.
type BackoffTimerInfo struct {
	// The minimum duration after which Func is called.
	Min time.Duration

	// The maximum duration after which Func is called.
	Max time.Duration

	// Determines whether a small randomization is applied to
	// the duration.
	Jitter bool

	// The factor by which you want the duration to increase
	// every time.
	Factor int64

	// Func is the function that will be called when the countdown reaches 0.
	Func func()
}

// Signal implements the Timer interface
// Any existing timer execution is stopped before
// a new one is created.
func (t *BackoffTimer) Start() {
	if t.timer != nil {
		t.timer.Stop()
	}
	t.timer = t.afterFunc(t.currentDuration, t.info.Func)

	// Since it's a backoff timer we will increase
	// the duration after each signal.
	t.increaseDuration()
}

// Reset implements the Timer interface
func (t *BackoffTimer) Reset() {
	if t.timer != nil {
		t.timer.Stop()
	}
	if t.currentDuration > t.info.Min {
		t.currentDuration = t.info.Min
	}
}

// increaseDuration will increase the duration based on
// the current value and the factor. If jitter is true
// it will add a 0.3% jitter to the final value.
func (t *BackoffTimer) increaseDuration() {
	current := int64(t.currentDuration)
	nextDuration := time.Duration(current * t.info.Factor)
	if t.info.Jitter {
		// Get a factor in [-1; 1]
		randFactor := (rand.Float64() * 2) - 1
		jitter := float64(nextDuration) * randFactor * 0.03
		nextDuration = nextDuration + time.Duration(jitter)
	}
	if nextDuration > t.info.Max {
		nextDuration = t.info.Max
	}
	t.currentDuration = nextDuration
}

// StoppableTimer defines a interface for a time.Timer
// usually returned by AfterFunc so it's easier to mock it
// in tests. We only use Stop from that interface.
type StoppableTimer interface {
	// Stop prevents the Timer from firing. It returns true if the call stops the timer,
	// false if the timer has already expired or been stopped. Stop does not close the
	// channel, to prevent a read from the channel succeeding incorrectly.
	Stop() bool
}
