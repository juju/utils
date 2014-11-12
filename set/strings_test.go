// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package set_test

import (
	"sort"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/set"
)

type stringSetSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(stringSetSuite{})

// Helper methods for the tests.
func AssertValues(c *gc.C, s set.Strings, expected ...string) {
	values := s.Values()
	// Expect an empty slice, not a nil slice for values.
	if expected == nil {
		expected = []string{}
	}
	sort.Strings(expected)
	sort.Strings(values)
	c.Assert(values, gc.DeepEquals, expected)
	c.Assert(s.Size(), gc.Equals, len(expected))
	// Check the sorted values too.
	sorted := s.SortedValues()
	c.Assert(sorted, gc.DeepEquals, expected)
}

// Actual tests start here.

func (stringSetSuite) TestEmpty(c *gc.C) {
	s := set.NewStrings()
	AssertValues(c, s)
}

func (stringSetSuite) TestInitialValues(c *gc.C) {
	values := []string{"foo", "bar", "baz"}
	s := set.NewStrings(values...)
	AssertValues(c, s, values...)
}

func (stringSetSuite) TestSize(c *gc.C) {
	// Empty sets are empty.
	s := set.NewStrings()
	c.Assert(s.Size(), gc.Equals, 0)

	// Size returns number of unique values.
	s = set.NewStrings("foo", "foo", "bar")
	c.Assert(s.Size(), gc.Equals, 2)
}

func (stringSetSuite) TestIsEmpty(c *gc.C) {
	// Empty sets are empty.
	s := set.NewStrings()
	c.Assert(s.IsEmpty(), gc.Equals, true)

	// Non-empty sets are not empty.
	s = set.NewStrings("foo")
	c.Assert(s.IsEmpty(), gc.Equals, false)
	// Newly empty sets work too.
	s.Remove("foo")
	c.Assert(s.IsEmpty(), gc.Equals, true)
}

func (stringSetSuite) TestAdd(c *gc.C) {
	s := set.NewStrings()
	s.Add("foo")
	s.Add("foo")
	s.Add("bar")
	AssertValues(c, s, "foo", "bar")
}

func (stringSetSuite) TestRemove(c *gc.C) {
	s := set.NewStrings("foo", "bar")
	s.Remove("foo")
	AssertValues(c, s, "bar")
}

func (stringSetSuite) TestContains(c *gc.C) {
	s := set.NewStrings("foo", "bar")
	c.Assert(s.Contains("foo"), gc.Equals, true)
	c.Assert(s.Contains("bar"), gc.Equals, true)
	c.Assert(s.Contains("baz"), gc.Equals, false)
}

func (stringSetSuite) TestRemoveNonExistent(c *gc.C) {
	s := set.NewStrings()
	s.Remove("foo")
	AssertValues(c, s)
}

func (stringSetSuite) TestUnion(c *gc.C) {
	s1 := set.NewStrings("foo", "bar")
	s2 := set.NewStrings("foo", "baz", "bang")
	union1 := s1.Union(s2)
	union2 := s2.Union(s1)

	AssertValues(c, union1, "foo", "bar", "baz", "bang")
	AssertValues(c, union2, "foo", "bar", "baz", "bang")
}

func (stringSetSuite) TestIntersection(c *gc.C) {
	s1 := set.NewStrings("foo", "bar")
	s2 := set.NewStrings("foo", "baz", "bang")
	int1 := s1.Intersection(s2)
	int2 := s2.Intersection(s1)

	AssertValues(c, int1, "foo")
	AssertValues(c, int2, "foo")
}

func (stringSetSuite) TestDifference(c *gc.C) {
	s1 := set.NewStrings("foo", "bar")
	s2 := set.NewStrings("foo", "baz", "bang")
	diff1 := s1.Difference(s2)
	diff2 := s2.Difference(s1)

	AssertValues(c, diff1, "bar")
	AssertValues(c, diff2, "baz", "bang")
}

func (stringSetSuite) TestUninitializedPanics(c *gc.C) {
	f := func() {
		var s set.Strings
		s.Add("foo")
	}
	c.Assert(f, gc.PanicMatches, "uninitalised set")
}
