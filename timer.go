// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package utils

import (
	"math/rand"
	"time"
)

// Countdown implements a timer that will signal on the provided channel
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
// The caller is responsible for using the returned cleanup function
// that will close the channel
func NewBackoffTimer(info BackoffTimerInfo) *BackoffTimer {
	if info.AfterFunc == nil {
		info.AfterFunc = func(d time.Duration, f func()) StoppableTimer {
			return time.AfterFunc(d, f)
		}
	}
	return &BackoffTimer{
		Info:            info,
		CurrentDuration: info.Min,
	}
}

// BackoffTimer creates and initializer a new BackoffTimer
// A backoff timer starts at min and gets multiplied by factor
// until it reaches max. Jitter determines whether a small
// randomization is added to the duration.
// The struct is mainly exposed for testing purposes.
type BackoffTimer struct {
	Info BackoffTimerInfo

	Timer           StoppableTimer
	CurrentDuration time.Duration
}

// BackoffTimerInfo is a helper struct for backoff timer
// that encapsulates config information.
type BackoffTimerInfo struct {
	Min    time.Duration
	Max    time.Duration
	Jitter bool
	Factor int64

	// Func is the function that will be called when the countdown reaches 0.
	Func func()

	// AfterFunc exists here for easier mocking
	// It is a function that will execute the function f
	// after duration d and return a timer object that will let
	// us stop the existing timer.
	AfterFunc func(d time.Duration, f func()) StoppableTimer
}

// Signal implements the Timer interface
// Any existing timer execution is stopped before
// a new one is created.
func (t *BackoffTimer) Start() {
	if t.Timer != nil {
		t.Timer.Stop()
	}
	t.Timer = t.Info.AfterFunc(t.CurrentDuration, t.Info.Func)

	// Since it's a backoff timer we will increase
	// the duration after each signal.
	t.increaseDuration()
}

// increaseDuration will increase the duration based on
// the current value and the factor. If jitter is true
// it will add a 0.3% jitter to the final value.
func (t *BackoffTimer) increaseDuration() {
	current := int64(t.CurrentDuration)
	nextDuration := time.Duration(current * t.Info.Factor)
	if t.Info.Jitter {
		// Get a factor in [-1; 1]
		randFactor := (rand.Float64() * 2) - 1
		jitter := float64(nextDuration) * randFactor * 0.03
		nextDuration = nextDuration + time.Duration(jitter)
	}
	if nextDuration > t.Info.Max {
		nextDuration = t.Info.Max
	}
	t.CurrentDuration = nextDuration
}

// Reset implements the Timer interface
func (t *BackoffTimer) Reset() {
	if t.Timer != nil {
		t.Timer.Stop()
	}
	if t.CurrentDuration > t.Info.Min {
		t.CurrentDuration = t.Info.Min
	}
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
