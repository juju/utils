// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

func doSomething() (int, error) { return 0, nil }

func shouldRetry(error) bool { return false }

func doSomethingWith(int) {}

func ExampleAttempt_HasNext() {
	// This example shows how Attempt.HasNext can be used to help
	// structure an attempt loop. If the godoc example code allowed
	// us to make the example return an error, we would uncomment
	// the commented return statements.
	attempts := utils.AttemptStrategy{
		Total: 1 * time.Second,
		Delay: 250 * time.Millisecond,
	}
	for attempt := attempts.Start(); attempt.Next(); {
		x, err := doSomething()
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		if err != nil {
			// return err
			return
		}
		doSomethingWith(x)
	}
	// return ErrTimedOut
	return
}

func (*utilsSuite) TestAttemptTiming(c *gc.C) {
	testAttempt := utils.AttemptStrategy{
		Total: 0.25e9,
		Delay: 0.1e9,
	}
	want := []time.Duration{0, 0.1e9, 0.2e9, 0.2e9}
	got := make([]time.Duration, 0, len(want)) // avoid allocation when testing timing
	t0 := time.Now()
	for a := testAttempt.Start(); a.Next(); {
		got = append(got, time.Now().Sub(t0))
	}
	got = append(got, time.Now().Sub(t0))
	c.Assert(got, gc.HasLen, len(want))
	const margin = 0.01e9
	for i, got := range want {
		lo := want[i] - margin
		hi := want[i] + margin
		if got < lo || got > hi {
			c.Errorf("attempt %d want %g got %g", i, want[i].Seconds(), got.Seconds())
		}
	}
}

func (s *utilsSuite) TestAttemptIncreasing(c *gc.C) {
	s.PatchValue(utils.SleepFunc, func(time.Duration) {})

	increase := 1 * time.Second
	max := 4 * time.Second
	testAttempt := utils.AttemptStrategy{
		Delay:     1 * time.Second,
		Min:       4,
		NextDelay: utils.DelayArithmeticMax(increase, max),
	}
	a := testAttempt.Start()
	before := utils.GetAttemptDelay(a)
	var delays []time.Duration
	for a.Next() {
		delays = append(delays, utils.GetAttemptDelay(a))
	}
	after := utils.GetAttemptDelay(a)

	c.Check(before, gc.Equals, time.Second)
	c.Check(after, gc.Equals, 4*time.Second)
	c.Check(delays, jc.DeepEquals, []time.Duration{
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
		4 * time.Second,
	})
}

func (*utilsSuite) TestAttemptNextHasNext(c *gc.C) {
	a := utils.AttemptStrategy{}.Start()
	c.Assert(a.Next(), gc.Equals, true)
	c.Assert(a.Next(), gc.Equals, false)

	a = utils.AttemptStrategy{}.Start()
	c.Assert(a.Next(), gc.Equals, true)
	c.Assert(a.HasNext(), gc.Equals, false)
	c.Assert(a.Next(), gc.Equals, false)

	a = utils.AttemptStrategy{Total: 2e8}.Start()
	c.Assert(a.Next(), gc.Equals, true)
	c.Assert(a.HasNext(), gc.Equals, true)
	time.Sleep(2e8)
	c.Assert(a.HasNext(), gc.Equals, true)
	c.Assert(a.Next(), gc.Equals, true)
	c.Assert(a.Next(), gc.Equals, false)

	a = utils.AttemptStrategy{Total: 1e8, Min: 2}.Start()
	time.Sleep(1e8)
	c.Assert(a.Next(), gc.Equals, true)
	c.Assert(a.HasNext(), gc.Equals, true)
	c.Assert(a.Next(), gc.Equals, true)
	c.Assert(a.HasNext(), gc.Equals, false)
	c.Assert(a.Next(), gc.Equals, false)
}
