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
	// Delay is the interval between each try in the burst.
	Delay time.Duration
	// DelayIncrease is how much the delay increases between each try.
	// By default there will be no increase.
	DelayIncrease time.Duration
	// DelayMax is the maximum delay allowed. This helps keep increasing
	// delays under control.
	DelayMax time.Duration
	// Min is the minimum number of retries. It overrides Total.
	Min int
}

type Attempt struct {
	strategy AttemptStrategy
	last     time.Time
	end      time.Time
	delay    time.Duration
	force    bool
	count    int
	// updateDelay is the function used to set the next delay
	// value. If it is nil then the delay value stays the same.
	updateDelay func()
}

// Start begins a new sequence of attempts for the given strategy.
func (s AttemptStrategy) Start() *Attempt {
	now := time.Now()
	attempt := Attempt{
		strategy: s,
		last:     now,
		end:      now.Add(s.Total),
		delay:    s.Delay,
		force:    true,
	}

	if s.DelayIncrease > 0 {
		attempt.updateDelay = func() {
			// Follow an arithmetic progression.
			attempt.delay += s.DelayIncrease
			if s.DelayMax != 0 && attempt.delay > s.DelayMax {
				attempt.delay = s.DelayMax
				attempt.updateDelay = nil
			}
		}
	}

	return &attempt
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
	if a.updateDelay != nil {
		a.updateDelay()
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
