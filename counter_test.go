// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
)

type counterSuite struct {
}

var _ = gc.Suite(&counterSuite{})

func (*counterSuite) TestConcurrentCounter_DefaultValueBeginsAtZero(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Count(), gc.Equals, int64(0))
}

func (*counterSuite) TestCount_ReturnsAccurateCount(c *gc.C) {
	var counter utils.ConcurrentCounter
	counter.Increment()
	c.Check(counter.Count(), gc.Equals, int64(1))
}

func (*counterSuite) TestIncrement_ReturnValueCorrect(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Increment(), gc.Equals, counter.Count())
}

func (*counterSuite) TestIncrement_AddsOne(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Increment(), gc.Equals, int64(1))
}

func (*counterSuite) TestDecrement_ReturnValueCorrect(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Decrement(), gc.Equals, counter.Count())
}

func (*counterSuite) TestDecrement_SubtractsOne(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Decrement(), gc.Equals, int64(-1))
}

func (*counterSuite) TestAdd_ReturnValueCorrect(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Add(1), gc.Equals, counter.Count())
}

func (*counterSuite) TestAdd_AddsAmount(c *gc.C) {
	var counter utils.ConcurrentCounter
	c.Check(counter.Add(2), gc.Equals, int64(2))
}
