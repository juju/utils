// Copyright 2016 Canonical Ltd.
// Copyright 2016 Cloudbase Solutions
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"fmt"
	"os"
	"time"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/clock"
	"github.com/juju/utils/exec"
)

type execSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&execSuite{})

func (*execSuite) TestWaitWithCancel(c *gc.C) {
	params := exec.RunParams{
		Commands: "sleep 100",
		Clock:    &mockClock{C: make(chan time.Time)},
	}

	err := params.Run()
	c.Assert(err, gc.IsNil)
	c.Assert(params.Process(), gc.Not(gc.IsNil))

	cancelChan := make(chan struct{}, 1)
	defer close(cancelChan)
	cancelChan <- struct{}{}
	result, err := params.WaitWithCancel(cancelChan)
	c.Assert(err, gc.Equals, exec.ErrCancelled)
	c.Assert(string(result.Stdout), gc.Equals, "")
	c.Assert(string(result.Stderr), gc.Equals, "")
	c.Assert(result.Code, gc.Equals, cancelErrCode)
}

func (s *execSuite) TestKillAbortedIfUnsuccessfull(c *gc.C) {
	killCalled := false

	mockChan := make(chan time.Time, 1)
	defer close(mockChan)
	params := exec.RunParams{
		Commands:    "sleep 100",
		WorkingDir:  "",
		Environment: []string{},
		Clock:       &mockClock{C: mockChan},
		KillProcess: func(*os.Process) error {
			killCalled = true
			return nil
		},
	}

	err := params.Run()
	c.Assert(err, gc.IsNil)
	c.Assert(params.Process(), gc.Not(gc.IsNil))

	cancelChan := make(chan struct{}, 1)
	defer close(cancelChan)
	cancelChan <- struct{}{}
	mockChan <- time.Now()
	res, err := params.WaitWithCancel(cancelChan)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("tried to kill process %d, but timed out", params.Process().Pid))
	c.Assert(res, gc.IsNil)
	c.Assert(killCalled, jc.IsTrue)
}

type mockClock struct {
	clock.Clock
	C <-chan time.Time
}

func (m *mockClock) After(t time.Duration) <-chan time.Time {
	return m.C
}
