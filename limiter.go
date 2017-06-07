// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/juju/utils/clock"
)

type empty struct{}
type limiter struct {
	wait     chan empty
	maxPause time.Duration
	clock    clock.Clock
}

// Limiter represents a limited resource (eg a semaphore).
type Limiter interface {
	// Acquire another unit of the resource.
	// Acquire returns false to indicate there is no more availability,
	// until another entity calls Release.
	Acquire() bool
	// AcquireWait requests a unit of resource, but blocks until one is
	// available.
	AcquireWait()
	// Release returns a unit of the resource. Calling Release when there
	// are no units Acquired is an error.
	Release() error
}

// NewLimiter creates a limiter. If maxPause is > 0, there will be a random delay
// up to that duration before attempting an Acquire.
func NewLimiter(maxAllowed int, maxPause time.Duration, clk clock.Clock) Limiter {
	if clk == nil {
		clk = clock.WallClock
	}
	return limiter{
		wait:     make(chan empty, maxAllowed),
		maxPause: maxPause,
		clock:    clk,
	}
}

// Acquire requests some resources that you can return later
// It returns 'true' if there are resources available, but false if they are
// not. Callers are responsible for calling Release if this returns true, but
// should not release if this returns false.
func (l limiter) Acquire() bool {
	// Pause before attempting to grab a slot.
	// This is optional depending on what was used to
	// construct this limiter, and is used to throttle
	// incoming connections.
	l.pause()
	e := empty{}
	select {
	case l.wait <- e:
		return true
	default:
		return false
	}
}

// AcquireWait waits for the resource to become available before returning.
func (l limiter) AcquireWait() {
	e := empty{}
	l.wait <- e
}

// Release returns the resource to the available pool.
func (l limiter) Release() error {
	select {
	case <-l.wait:
		return nil
	default:
		return fmt.Errorf("Release without an associated Acquire")
	}
}

func (l limiter) pause() {
	if l.maxPause <= 0 {
		return
	}
	pauseTime := rand.Intn(int(l.maxPause / time.Millisecond))
	select {
	case <-l.clock.After(time.Duration(pauseTime) * time.Millisecond):
	}
}
