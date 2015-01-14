// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"time"
)

// The Attempt and AttemptStrategy types are copied from those in launchpad.net/goamz/aws.

// AttemptStrategy represents a strategy for waiting for an action
// to complete successfully.
type AttemptStrategy struct {
	// Total is the total duration of the attempt.
	Total time.Duration

	// Delay is the interval between each try in the burst. When the
	// delay is dynamic (as defined by NextDelay), Delay is the initial
	// delay value.
	Delay time.Duration

	// Min is the minimum number of retries. It overrides Total.
	Min int

	// NextDelay is used to determine a new delay value given the old
	// one. If it is not set then the delay will not change.
	NextDelay func(delay time.Duration) time.Duration
}

// Attempt is the realization of an attempt strategy. It provides an
// iteration mechanism for use in for loops that satisfies the strategy.
// As such it holds the state that will be used to decide the outcome of
// the next iteration (i.e. call to Next).
type Attempt struct {
	// strategy is the attempt strategy that this attempt satisfies.
	strategy AttemptStrategy
	// last identifies when the last iteration happened.
	last time.Time
	// end indicates the timeout time for the attempt.
	end time.Time
	// delay is (roughly) how long the attempt will sleep at the next
	// iteration. This is initialized to the strategy's Delay, and may
	// change if the NextDelay func is set on the strategy.
	delay time.Duration
	// force is used to ensure at least one iteration takes place.
	force bool
	// count keeps track of the number of completed iterations.
	count int
}

// Start begins a new sequence of attempts for the given strategy.
func (s AttemptStrategy) Start() *Attempt {
	now := time.Now()
	return &Attempt{
		strategy: s,
		last:     now,
		end:      now.Add(s.Total),
		delay:    s.Delay,
		force:    true,
	}
}

var sleepFunc = time.Sleep

// Next waits until it is time to perform the next attempt or returns
// false if it is time to stop trying.
// It always returns true the first time it is called - we are guaranteed to
// make at least one attempt.
func (a *Attempt) Next() bool {
	now := time.Now()
	sleep := a.nextSleep(now)
	if !a.force && !now.Add(sleep).Before(a.end) && a.strategy.Min <= a.count {
		return false
	}
	a.force = false
	if sleep > 0 && a.count > 0 {
		sleepFunc(sleep)
		now = time.Now()
	}
	a.count++
	a.last = now
	if a.strategy.NextDelay != nil {
		a.delay = a.strategy.NextDelay(a.delay)
	}
	return true
}

func (a *Attempt) nextSleep(now time.Time) time.Duration {
	sleep := a.delay - now.Sub(a.last)
	if sleep < 0 {
		return 0
	}
	return sleep
}

// HasNext returns whether another attempt will be made if the current
// one fails. If it returns true, the following call to Next is
// guaranteed to return true.
func (a *Attempt) HasNext() bool {
	if a.force || a.strategy.Min > a.count {
		return true
	}
	now := time.Now()
	if now.Add(a.nextSleep(now)).Before(a.end) {
		a.force = true
		return true
	}
	return false
}

// MaxDelay returns a delay func that sets a hard limit on the delay
// provided by the wrapped func.
func MaxDelay(max time.Duration, nextDelay func(time.Duration) time.Duration) func(time.Duration) time.Duration {
	return func(delay time.Duration) time.Duration {
		if delay == max {
			return delay
		}
		delay = nextDelay(delay)
		if delay > max {
			return max
		}
		return delay
	}
}

// DelayArithmetic returns a "next delay" function that increases a
// delay by a fixed amount. To limit how big the delay might get, use
// MaxDelay.
func DelayArithmetic(increase time.Duration) func(time.Duration) time.Duration {
	return func(delay time.Duration) time.Duration {
		// Follow an arithmetic progression.
		return delay + increase
	}
}

// DelayGeometric returns a "next delay" function that multiplies a
// delay by fixed amount. To limit how big the delay might get, use
// MaxDelay.
func DelayGeometric(scale int) func(time.Duration) time.Duration {
	next := time.Duration(1)
	return func(delay time.Duration) time.Duration {
		// Follow a geometric progression.
		delay *= next
		next *= time.Duration(scale)
		return delay
	}
}
