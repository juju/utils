// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package set_test

import (
	"sort"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/set"
)

type intSetSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(intSetSuite{})

// Helper methods for the tests.
func AssertIntValues(c *gc.C, s set.Ints, expected ...int) {
	values := s.Values()
	// Expect an empty slice, not a nil slice for values.
	if expected == nil {
		expected = []int{}
	}
	sort.Ints(expected)
	sort.Ints(values)
	c.Assert(values, gc.DeepEquals, expected)
	c.Assert(s.Size(), gc.Equals, len(expected))
	// Check the sorted values too.
	sorted := s.SortedValues()
	c.Assert(sorted, gc.DeepEquals, expected)
}

// Actual tests start here.

func (intSetSuite) TestEmpty(c *gc.C) {
	s := set.NewInts()
	AssertIntValues(c, s)
}

func (intSetSuite) TestInitialValues(c *gc.C) {
	values := []int{1, 2, 3}
	s := set.NewInts(values...)
	AssertIntValues(c, s, values...)
}

func (intSetSuite) TestSize(c *gc.C) {
	// Empty sets are empty.
	s := set.NewInts()
	c.Assert(s.Size(), gc.Equals, 0)

	// Size returns number of unique values.
	s = set.NewInts(1, 1, 2)
	c.Assert(s.Size(), gc.Equals, 2)
}

func (intSetSuite) TestIsEmpty(c *gc.C) {
	// Empty sets are empty.
	s := set.NewInts()
	c.Assert(s.IsEmpty(), jc.IsTrue)

	// Non-empty sets are not empty.
	s = set.NewInts(1)
	c.Assert(s.IsEmpty(), jc.IsFalse)
	// Newly empty sets work too.
	s.Remove(1)
	c.Assert(s.IsEmpty(), jc.IsTrue)
}

func (intSetSuite) TestAdd(c *gc.C) {
	s := set.NewInts()
	s.Add(1)
	s.Add(1)
	s.Add(2)
	AssertIntValues(c, s, 1, 2)
}

func (intSetSuite) TestRemove(c *gc.C) {
	s := set.NewInts(1, 2)
	s.Remove(1)
	AssertIntValues(c, s, 2)
}

func (intSetSuite) TestContains(c *gc.C) {
	s := set.NewInts(1, 2)
	c.Assert(s.Contains(1), jc.IsTrue)
	c.Assert(s.Contains(2), jc.IsTrue)
	c.Assert(s.Contains(3), jc.IsFalse)
}

func (intSetSuite) TestRemoveNonExistent(c *gc.C) {
	s := set.NewInts()
	s.Remove(1)
	AssertIntValues(c, s)
}

func (intSetSuite) TestUnion(c *gc.C) {
	s1 := set.NewInts(1, 2)
	s2 := set.NewInts(1, 3, 4)
	union1 := s1.Union(s2)
	union2 := s2.Union(s1)

	AssertIntValues(c, union1, 1, 2, 3, 4)
	AssertIntValues(c, union2, 1, 2, 3, 4)
}

func (intSetSuite) TestIntersection(c *gc.C) {
	s1 := set.NewInts(1, 2)
	s2 := set.NewInts(1, 3, 4)
	int1 := s1.Intersection(s2)
	int2 := s2.Intersection(s1)

	AssertIntValues(c, int1, 1)
	AssertIntValues(c, int2, 1)
}

func (intSetSuite) TestDifference(c *gc.C) {
	s1 := set.NewInts(1, 2)
	s2 := set.NewInts(1, 3, 4)
	diff1 := s1.Difference(s2)
	diff2 := s2.Difference(s1)

	AssertIntValues(c, diff1, 2)
	AssertIntValues(c, diff2, 3, 4)
}

func (intSetSuite) TestUninitializedPanics(c *gc.C) {
	f := func() {
		var s set.Ints
		s.Add(1)
	}
	c.Assert(f, gc.PanicMatches, "uninitalised set")
}
