// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"math"
	"time"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	"github.com/juju/utils"
)

type TestStdTimer struct {
	stdStub *testing.Stub
}

func (t *TestStdTimer) Stop() bool {
	t.stdStub.AddCall("Stop")
	return true
}

func (t *TestStdTimer) Reset(d time.Duration) bool {
	t.stdStub.AddCall("Reset", d)
	return true
}

type timerSuite struct {
	baseSuite      testing.CleanupSuite
	timer          utils.BackoffTimer
	afterFuncCalls int64
	stub           *testing.Stub

	min    time.Duration
	max    time.Duration
	factor int64
}

var _ = gc.Suite(&timerSuite{})

func (s *timerSuite) SetUpTest(c *gc.C) {
	s.baseSuite.SetUpTest(c)
	s.afterFuncCalls = 0
	s.stub = &testing.Stub{}
	s.baseSuite.PatchValue(utils.AfterFunc, func(d time.Duration, f func()) utils.StdTimer {
		s.afterFuncCalls++
		return &TestStdTimer{s.stub}
	})

	s.min = 2 * time.Second
	s.max = 16 * time.Second
	s.factor = 2
	s.timer = utils.BackoffTimer{
		Min:             s.min,
		Max:             s.max,
		Jitter:          false,
		Factor:          s.factor,
		Chan:            make(chan struct{}, 1),
		CurrentDuration: s.min,
	}
}

func (s *timerSuite) TestSignal(c *gc.C) {
	s.timer.Signal()
	s.testSignal(c, 1, 1)
}

func (s *timerSuite) TestMultipleSignals(c *gc.C) {
	s.timer.Signal()
	s.testSignal(c, 1, 1)

	s.timer.Signal()
	s.checkStopCalls(c, 1)
	s.testSignal(c, 2, 2)

	s.timer.Signal()
	s.checkStopCalls(c, 2)
	s.testSignal(c, 3, 3)
}

func (s *timerSuite) TestResetNoSignal(c *gc.C) {
	s.timer.Reset()
	c.Assert(s.timer.CurrentDuration, gc.Equals, s.min)
}

func (s *timerSuite) TestResetAndSignal(c *gc.C) {
	s.timer.Reset()
	c.Assert(s.timer.CurrentDuration, gc.Equals, s.min)

	// These variables are used to track the number
	// of afterFuncCalls(signalCallsNo) and the number
	// of Stop calls(resetStopCallsNo + signalCallsNo)
	resetStopCallsNo := 0
	signalCallsNo := 0

	signalCallsNo++
	s.timer.Signal()
	s.testSignal(c, 1, 1)

	resetStopCallsNo++
	s.timer.Reset()
	s.checkStopCalls(c, resetStopCallsNo+signalCallsNo-1)
	c.Assert(s.timer.CurrentDuration, gc.Equals, s.min)

	for i := 1; i < 200; i++ {
		signalCallsNo++
		s.timer.Signal()
		s.testSignal(c, int64(signalCallsNo), int64(i))
		s.checkStopCalls(c, resetStopCallsNo+signalCallsNo-1)
	}

	resetStopCallsNo++
	s.timer.Reset()
	s.checkStopCalls(c, signalCallsNo+resetStopCallsNo-1)

	for i := 1; i < 100; i++ {
		signalCallsNo++
		s.timer.Signal()
		s.testSignal(c, int64(signalCallsNo), int64(i))
		s.checkStopCalls(c, resetStopCallsNo+signalCallsNo-1)
	}

	resetStopCallsNo++
	s.timer.Reset()
	s.checkStopCalls(c, signalCallsNo+resetStopCallsNo-1)
}

func (s *timerSuite) testSignal(c *gc.C, afterFuncCalls int64, durationFactor int64) {
	c.Assert(s.afterFuncCalls, gc.Equals, afterFuncCalls)
	c.Logf("iteration %d", afterFuncCalls)
	expectedDuration := time.Duration(math.Pow(float64(s.factor), float64(durationFactor))) * s.min
	if expectedDuration > s.max || expectedDuration <= 0 {
		expectedDuration = s.max
	}
	c.Assert(s.timer.CurrentDuration, gc.Equals, expectedDuration)
}

func (s *timerSuite) checkStopCalls(c *gc.C, number int) {
	calls := make([]testing.StubCall, number)
	for i := 0; i < number; i++ {
		calls[i] = testing.StubCall{FuncName: "Stop"}
	}
	s.stub.CheckCalls(c, calls)
}
