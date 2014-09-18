// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package set_test

import (
	"github.com/juju/names"
	"github.com/juju/testing"
	gc "launchpad.net/gocheck"

	"github.com/juju/utils/set"
)

type tagSetSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(tagSetSuite{})

func (tagSetSuite) TestEmpty(c *gc.C) {
	t := set.NewTags()
	c.Assert(t.Size(), gc.Equals, 0)
}

func (tagSetSuite) TestInitialValues(c *gc.C) {
	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-rabbitmq-server-0")

	t := set.NewTags(foo, bar)
	c.Assert(t.Size(), gc.Equals, 2)
}

func (tagSetSuite) TestInitialStringValues(c *gc.C) {
	t, err := set.NewTagsFromStrings("unit-wordpress-0", "unit-rabbitmq-server-0")
	c.Assert(err, gc.IsNil)
	c.Assert(t.Size(), gc.Equals, 2)
}

func (tagSetSuite) TestSize(c *gc.C) {
	// Empty sets are empty.
	s := set.NewTags()
	c.Assert(s.Size(), gc.Equals, 0)

	s, err := set.NewTagsFromStrings(
		"unit-wordpress-0",
		"unit-rabbitmq-server-0",
	)
	c.Assert(err, gc.IsNil)
	c.Assert(s.Size(), gc.Equals, 2)
}

func (tagSetSuite) TestSizeDuplicate(c *gc.C) {
	// Empty sets are empty.
	s := set.NewTags()
	c.Assert(s.Size(), gc.Equals, 0)

	// Size returns number of unique values.
	s, err := set.NewTagsFromStrings(
		"unit-wordpress-0",
		"unit-rabbitmq-server-0",
		"unit-wordpress-0",
	)
	c.Assert(err, gc.IsNil)
	c.Assert(s.Size(), gc.Equals, 2)
}

func (tagSetSuite) TestIsEmpty(c *gc.C) {
	// Empty sets are empty.
	s := set.NewTags()
	c.Assert(s.IsEmpty(), gc.Equals, true)

	// Non-empty sets are not empty.
	tag, _ := names.ParseTag("unit-wordpress-0")
	s = set.NewTags(tag)
	c.Assert(s.IsEmpty(), gc.Equals, false)

	// Newly empty sets work too.
	s.Remove(tag)
	c.Assert(s.IsEmpty(), gc.Equals, true)
}

func (tagSetSuite) TestAdd(c *gc.C) {
	t := set.NewTags()
	foo, _ := names.ParseTag("unit-wordpress-0")
	t.Add(foo)
	c.Assert(t.Size(), gc.Equals, 1)
	c.Assert(t.Contains(foo), gc.Equals, true)
}

func (tagSetSuite) TestAddDuplicate(c *gc.C) {
	t := set.NewTags()
	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-rabbitmq-server-0")

	t.Add(foo)
	t.Add(bar)

	bar, _ = names.ParseTag("unit-wordpress-0")
	t.Add(bar)

	c.Assert(t.Size(), gc.Equals, 2)
}

func (tagSetSuite) TestRemove(c *gc.C) {
	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-rabbitmq-server-0")

	t := set.NewTags(foo, bar)
	t.Remove(foo)

	c.Assert(t.Contains(foo), gc.Equals, false)
	c.Assert(t.Contains(bar), gc.Equals, true)
}

func (tagSetSuite) TestContains(c *gc.C) {
	t, err := set.NewTagsFromStrings("unit-wordpress-0", "unit-rabbitmq-server-0")
	c.Assert(err, gc.IsNil)

	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-rabbitmq-server-0")
	baz, _ := names.ParseTag("unit-mongodb-0")

	c.Assert(t.Contains(foo), gc.Equals, true)
	c.Assert(t.Contains(bar), gc.Equals, true)
	c.Assert(t.Contains(baz), gc.Equals, false)
}

func (tagSetSuite) TestSortedValues(c *gc.C) {
	m1, _ := names.ParseTag("machine-0")
	z1, _ := names.ParseTag("unit-z-server-0")
	z2, _ := names.ParseTag("unit-z-server-1")
	a1, _ := names.ParseTag("unit-a-server-0")

	t := set.NewTags(z2, a1, z1, m1)
	values := t.SortedValues()

	c.Assert(values, gc.DeepEquals, []names.Tag{m1, a1, z1, z2})
}

func (tagSetSuite) TestRemoveNonExistent(c *gc.C) {
	t := set.NewTags()
	foo, _ := names.ParseTag("unit-wordpress-0")
	t.Remove(foo)
	c.Assert(t.Size(), gc.Equals, 0)
}

func (tagSetSuite) TestUnion(c *gc.C) {
	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-mongodb-0")
	baz, _ := names.ParseTag("unit-rabbitmq-server-0")
	bang, _ := names.ParseTag("unit-mysql-server-0")

	t1 := set.NewTags(foo, bar)
	t2 := set.NewTags(foo, baz, bang)
	union1 := t1.Union(t2)
	union2 := t2.Union(t1)

	c.Assert(union1.Size(), gc.Equals, 4)
	c.Assert(union2.Size(), gc.Equals, 4)

	c.Assert(union1, gc.DeepEquals, union2)
	c.Assert(union1, gc.DeepEquals, set.NewTags(foo, bar, baz, bang))
}

func (tagSetSuite) TestIntersection(c *gc.C) {
	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-mongodb-0")
	baz, _ := names.ParseTag("unit-rabbitmq-server-0")
	bang, _ := names.ParseTag("unit-mysql-server-0")

	t1 := set.NewTags(foo, bar)
	t2 := set.NewTags(foo, baz, bang)

	int1 := t1.Intersection(t2)
	int2 := t2.Intersection(t1)

	c.Assert(int1.Size(), gc.Equals, 1)
	c.Assert(int2.Size(), gc.Equals, 1)

	c.Assert(int1, gc.DeepEquals, int2)
	c.Assert(int1, gc.DeepEquals, set.NewTags(foo))
}

func (tagSetSuite) TestDifference(c *gc.C) {
	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-mongodb-0")
	baz, _ := names.ParseTag("unit-rabbitmq-server-0")
	bang, _ := names.ParseTag("unit-mysql-server-0")

	t1 := set.NewTags(foo, bar)
	t2 := set.NewTags(foo, baz, bang)

	diff1 := t1.Difference(t2)
	diff2 := t2.Difference(t1)

	c.Assert(diff1, gc.DeepEquals, set.NewTags(bar))
	c.Assert(diff2, gc.DeepEquals, set.NewTags(baz, bang))
}

func (tagSetSuite) TestUninitialized(c *gc.C) {
	var uninitialized set.Tags

	foo, _ := names.ParseTag("unit-wordpress-0")
	bar, _ := names.ParseTag("unit-mongodb-0")

	c.Assert(uninitialized.Size(), gc.Equals, 0)
	c.Assert(uninitialized.IsEmpty(), gc.Equals, true)
	// You can get values and sorted values from an unitialized set.
	c.Assert(uninitialized.Values(), gc.DeepEquals, []names.Tag{})
	// All contains checks are false
	c.Assert(uninitialized.Contains(foo), gc.Equals, false)
	// Remove works on an uninitialized Strings
	uninitialized.Remove(foo)

	var other set.Tags
	// Union returns a new set that is empty but initialized.
	c.Assert(uninitialized.Union(other), gc.DeepEquals, set.NewTags())
	c.Assert(uninitialized.Intersection(other), gc.DeepEquals, set.NewTags())
	c.Assert(uninitialized.Difference(other), gc.DeepEquals, set.NewTags())

	other = set.NewTags(foo, bar)
	c.Assert(uninitialized.Union(other), gc.DeepEquals, other)
	c.Assert(uninitialized.Intersection(other), gc.DeepEquals, set.NewTags())
	c.Assert(uninitialized.Difference(other), gc.DeepEquals, set.NewTags())
	c.Assert(other.Union(uninitialized), gc.DeepEquals, other)
	c.Assert(other.Intersection(uninitialized), gc.DeepEquals, set.NewTags())
	c.Assert(other.Difference(uninitialized), gc.DeepEquals, other)

	// Once something is added, the set becomes initialized.
	uninitialized.Add(foo)
	c.Assert(uninitialized.Contains(foo), gc.Equals, true)
}
