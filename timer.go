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
	// ElapsedChan returns a simple channel to which the timer
	// will send a signal once the duration runs down to 0.
	ElapsedChan() <-chan struct{}

	// Reset stops the timer and resets it's duration to the minimum one.
	// Signal must be called to start the timer again.
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
// The caller is responsible for using the returned cleanup function
// that will close the channel
func NewBackoffTimer(info BackoffTimerInfo) (t *BackoffTimer, closeFunc func()) {
	channel := make(chan struct{}, 1)
	closer := func() {
		close(channel)
	}
	if info.AfterFunc == nil {
		info.AfterFunc = func(d time.Duration, f func()) StoppableTimer {
			return time.AfterFunc(d, f)
		}
	}
	return &BackoffTimer{
		Info:            info,
		Chan:            channel,
		CurrentDuration: info.Min,
	}, closer
}

// BackoffTimer creates and initializer a new BackoffTimer
// A backoff timer starts at min and gets multiplied by factor
// until it reaches max. Jitter determines whether a small
// randomization is added to the duration.
// The struct is mainly exposed for testing purposes.
type BackoffTimer struct {
	Info BackoffTimerInfo

	Timer           StoppableTimer
	Chan            chan struct{}
	CurrentDuration time.Duration
}

// BackoffTimerInfo is a helper struct for backoff timer
// that encapsulates config information.
type BackoffTimerInfo struct {
	Min    time.Duration
	Max    time.Duration
	Jitter bool
	Factor int64

	// AfterFunc exists here for easier mocking
	// It is a function that will execute the function f
	// after duration d and return a timer object that will let
	// us stop the existing timer.
	AfterFunc func(d time.Duration, f func()) StoppableTimer
}

// Channel implements the Timer interface
func (t *BackoffTimer) ElapsedChan() <-chan struct{} {
	return t.Chan
}

// Signal implements the Timer interface
// Any existing timer execution is stopped before
// a new one is created.
func (t *BackoffTimer) Signal() {
	if t.Timer != nil {
		t.Timer.Stop()
	}
	t.Timer = t.Info.AfterFunc(t.CurrentDuration, func() {
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
