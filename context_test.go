// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/juju/clock/testclock"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type contextSuite struct{}

var _ = gc.Suite(&contextSuite{})

// Note: the logic in these tests was copied from the tests
// in the Go standard library.

func (*contextSuite) TestDeadline(c *gc.C) {
	clk := testclock.NewClock(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))
	ctx, cancel := utils.ContextWithDeadline(context.Background(), clk, clk.Now().Add(50*time.Millisecond))
	defer cancel()
	c.Assert(fmt.Sprint(ctx), gc.Equals, `context.Background.WithDeadline(2000-01-01 00:00:00.05 +0000 UTC [50ms])`)
	testContextDeadline(c, ctx, "WithDeadline", clk, 1, 50*time.Millisecond)

	ctx, cancel = utils.ContextWithDeadline(context.Background(), clk, clk.Now().Add(50*time.Millisecond))
	defer cancel()
	o := otherContext{ctx}
	testContextDeadline(c, o, "WithDeadline+otherContext", clk, 1, 50*time.Millisecond)

	ctx, cancel = utils.ContextWithDeadline(context.Background(), clk, clk.Now().Add(50*time.Millisecond))
	defer cancel()
	o = otherContext{ctx}
	ctx, _ = utils.ContextWithDeadline(o, clk, clk.Now().Add(4*time.Second))
	testContextDeadline(c, ctx, "WithDeadline+otherContext+WithDeadline", clk, 2, 50*time.Millisecond)

	ctx, cancel = utils.ContextWithDeadline(context.Background(), clk, clk.Now().Add(-time.Millisecond))
	defer cancel()
	testContextDeadline(c, ctx, "WithDeadline+inthepast", clk, 0, 0)

	ctx, cancel = utils.ContextWithDeadline(context.Background(), clk, clk.Now())
	testContextDeadline(c, ctx, "WithDeadline+now", clk, 0, 0)
}

func (*contextSuite) TestTimeout(c *gc.C) {
	clk := testclock.NewClock(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))
	ctx, _ := utils.ContextWithTimeout(context.Background(), clk, 50*time.Millisecond)
	c.Assert(fmt.Sprint(ctx), gc.Equals, `context.Background.WithDeadline(2000-01-01 00:00:00.05 +0000 UTC [50ms])`)
	testContextDeadline(c, ctx, "WithTimeout", clk, 1, 50*time.Millisecond)

	ctx, _ = utils.ContextWithTimeout(context.Background(), clk, 50*time.Millisecond)
	o := otherContext{ctx}
	testContextDeadline(c, o, "WithTimeout+otherContext", clk, 1, 50*time.Millisecond)

	ctx, _ = utils.ContextWithTimeout(context.Background(), clk, 50*time.Millisecond)
	o = otherContext{ctx}
	ctx, _ = utils.ContextWithTimeout(o, clk, 3*time.Second)
	testContextDeadline(c, ctx, "WithTimeout+otherContext+WithTimeout", clk, 2, 50*time.Millisecond)
}

func (*contextSuite) TestCanceledTimeout(c *gc.C) {
	clk := testclock.NewClock(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))
	ctx, _ := utils.ContextWithTimeout(context.Background(), clk, time.Second)
	o := otherContext{ctx}
	ctx, cancel := utils.ContextWithTimeout(o, clk, 2*time.Second)
	cancel()
	time.Sleep(100 * time.Millisecond) // let cancelation propagate
	select {
	case <-ctx.Done():
	default:
		c.Errorf("<-ctx.Done() blocked, but shouldn't have")
	}
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func testContextDeadline(c *gc.C, ctx context.Context, name string, clk *testclock.Clock, waiters int, failAfter time.Duration) {
	err := clk.WaitAdvance(failAfter, 0, waiters)
	c.Assert(err, jc.ErrorIsNil)
	select {
	case <-time.After(time.Second):
		c.Fatalf("%s: context should have timed out", name)
	case <-ctx.Done():
	}
	c.Assert(ctx.Err(), gc.Equals, context.DeadlineExceeded)
}

// otherContext is a Context that's not one of the types defined in context.go.
// This lets us test code paths that differ based on the underlying type of the
// Context.
type otherContext struct {
	context.Context
}
