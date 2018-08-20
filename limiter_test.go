// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"fmt"
	"time"

	"github.com/juju/clock/testclock"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

const longWait = 10 * time.Second

type limiterSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&limiterSuite{})

func (*limiterSuite) TestAcquireUntilFull(c *gc.C) {
	l := utils.NewLimiter(2)
	c.Check(l.Acquire(), jc.IsTrue)
	c.Check(l.Acquire(), jc.IsTrue)
	c.Check(l.Acquire(), jc.IsFalse)
}

func (*limiterSuite) TestBadRelease(c *gc.C) {
	l := utils.NewLimiter(2)
	c.Check(l.Release(), gc.ErrorMatches, "Release without an associated Acquire")
}

func (*limiterSuite) TestAcquireAndRelease(c *gc.C) {
	l := utils.NewLimiter(2)
	c.Check(l.Acquire(), jc.IsTrue)
	c.Check(l.Acquire(), jc.IsTrue)
	c.Check(l.Acquire(), jc.IsFalse)
	c.Check(l.Release(), gc.IsNil)
	c.Check(l.Acquire(), jc.IsTrue)
	c.Check(l.Release(), gc.IsNil)
	c.Check(l.Release(), gc.IsNil)
	c.Check(l.Release(), gc.ErrorMatches, "Release without an associated Acquire")
}

func (*limiterSuite) TestAcquireWaitBlocksUntilRelease(c *gc.C) {
	l := utils.NewLimiter(2)
	calls := make([]string, 0, 10)
	start := make(chan bool, 0)
	waiting := make(chan bool, 0)
	done := make(chan bool, 0)
	go func() {
		<-start
		calls = append(calls, fmt.Sprintf("%v", l.Acquire()))
		calls = append(calls, fmt.Sprintf("%v", l.Acquire()))
		calls = append(calls, fmt.Sprintf("%v", l.Acquire()))
		waiting <- true
		l.AcquireWait()
		calls = append(calls, "waited")
		calls = append(calls, fmt.Sprintf("%v", l.Acquire()))
		done <- true
	}()
	// Start the routine, and wait for it to get to the first checkpoint
	start <- true
	select {
	case <-waiting:
	case <-time.After(longWait):
		c.Fatalf("timed out waiting for 'waiting' to trigger")
	}
	c.Check(l.Acquire(), jc.IsFalse)
	l.Release()
	select {
	case <-done:
	case <-time.After(longWait):
		c.Fatalf("timed out waiting for 'done' to trigger")
	}
	c.Check(calls, gc.DeepEquals, []string{"true", "true", "false", "waited", "false"})
}

func (*limiterSuite) TestAcquirePauses(c *gc.C) {
	clk := testclock.NewClock(time.Now())
	l := utils.NewLimiterWithPause(2, 10*time.Millisecond, 20*time.Millisecond, clk)
	acquired := make(chan bool, 1)
	start := make(chan bool, 0)
	go func() {
		<-start
		defer l.Release()
		acquired <- l.Acquire()
	}()

	start <- true
	// Minimum pause time not exceeded, acquire should not happen.
	clk.Advance(9 * time.Millisecond)
	select {
	case <-acquired:
		c.Fail()
	case <-time.After(50 * time.Millisecond):
	}

	clk.Advance(11 * time.Millisecond)
	select {
	case <-acquired:
	case <-time.After(50 * time.Millisecond):
		c.Fatal("acquire failed")
	}
}
