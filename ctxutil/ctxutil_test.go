// Copyright 2016 Canonical Ltd.

package ctxutil_test

import (
	"time"

	"golang.org/x/net/context"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/ctxutil"
)

type ctxutilSuite struct {
}

var _ = gc.Suite(&ctxutilSuite{})

func (s *ctxutilSuite) TestJoinCancel1(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx := ctxutil.Join(ctx1, context.Background())
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestJoinCancel2(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx := ctxutil.Join(context.Background(), ctx1)
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestJoinCancelBoth1(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	ctx := ctxutil.Join(ctx1, ctx2)
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestErrNoErr(c *gc.C) {
	ctx := ctxutil.Join(context.Background(), context.Background())
	c.Assert(ctx.Err(), gc.Equals, nil)
}

func (s *ctxutilSuite) TestJoinCancelBoth2(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx := ctxutil.Join(ctx1, ctx2)
	cancel2()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestDeadline1(c *gc.C) {
	t := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t)
	defer cancel1()
	ctx := ctxutil.Join(ctx1, context.Background())
	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t)
}

func (s *ctxutilSuite) TestDeadline2(c *gc.C) {
	t := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t)
	defer cancel1()
	ctx := ctxutil.Join(context.Background(), ctx1)
	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t)
}

func (s *ctxutilSuite) TestDeadlineBoth1(c *gc.C) {
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx := ctxutil.Join(ctx1, ctx2)

	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t1)
}

func (s *ctxutilSuite) TestDeadlineBoth2(c *gc.C) {
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx := ctxutil.Join(ctx2, ctx1)

	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t1)
}

func (s *ctxutilSuite) TestValue1(c *gc.C) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx := ctxutil.Join(ctx1, context.Background())
	c.Assert(ctx.Value("foo"), gc.Equals, "bar")
}

func (s *ctxutilSuite) TestValue2(c *gc.C) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx := ctxutil.Join(context.Background(), ctx1)
	c.Assert(ctx.Value("foo"), gc.Equals, "bar")
}

func (s *ctxutilSuite) TestValueBoth(c *gc.C) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar1")
	ctx2 := context.WithValue(context.Background(), "foo", "bar2")
	ctx := ctxutil.Join(ctx1, ctx2)
	c.Assert(ctx.Value("foo"), gc.Equals, "bar1")
}

func (s *ctxutilSuite) TestDoneRace(c *gc.C) {
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()
	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()
	ctx := ctxutil.Join(ctx1, ctx2)
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		done <- struct{}{}
	}()
	go func() {
		<-ctx.Done()
		done <- struct{}{}
	}()
	waitFor(c, done)
	waitFor(c, done)
}

func (s *ctxutilSuite) TestErrRace(c *gc.C) {
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()
	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()
	ctx := ctxutil.Join(ctx1, ctx2)
	done := make(chan struct{})
	go func() {
		ctx.Err()
		done <- struct{}{}
	}()
	go func() {
		ctx.Err()
		done <- struct{}{}
	}()
	waitFor(c, done)
	waitFor(c, done)
}

func waitFor(c *gc.C, ch <-chan struct{}) {
	select {
	case <-ch:
		return
	case <-time.After(time.Second):
		c.Fatalf("timed out")
	}
}
