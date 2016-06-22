// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package set_test

import (
	"testing"

	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v2"

	"github.com/juju/utils/set"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

type baseSet interface {
	Size() int
	IsEmpty() bool
}

type tagSet interface {
	Add(names.Tag)
	Remove(names.Tag)
	Contains(names.Tag) bool
	Values() []names.Tag
	SortedValues() []names.Tag

	Union(set.Tags) set.Tags
	Intersection(set.Tags) set.Tags
	Difference(set.Tags) set.Tags

	baseSet
}
