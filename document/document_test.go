// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package document_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/document"
)

var _ = gc.Suite(&DocSuite{})

type DocSuite struct {
	testing.IsolationSuite
}

func (s *DocSuite) TestID(c *gc.C) {
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	id := doc.ID()

	c.Check(id, gc.Equals, "some id")
}

func (s *DocSuite) TestSetIDInitial(c *gc.C) {
	doc := &document.Doc{}
	c.Assert(doc.ID(), gc.Equals, "")
	success := doc.SetID("some id")

	c.Check(success, gc.Equals, false)
	c.Check(doc.ID(), gc.Equals, "some id")
	c.Check(doc.Raw.ID, gc.Equals, "some id")
}

func (s *DocSuite) TestSetIDAlreadySetSame(c *gc.C) {
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	success := doc.SetID("some id")

	c.Check(success, gc.Equals, true)
	c.Check(doc.ID(), gc.Equals, "some id")
}

func (s *DocSuite) TestSetIDAlreadySetDifferent(c *gc.C) {
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	success := doc.SetID("another id")

	c.Check(success, gc.Equals, true)
	c.Check(doc.ID(), gc.Equals, "some id")
}

func (s *DocSuite) TestCopy(c *gc.C) {
	original := &document.Doc{}
	doc := original.Copy("")

	copied, ok := doc.(*document.Doc)
	c.Assert(ok, jc.IsTrue)

	c.Check(copied, gc.Not(gc.Equals), original)
	c.Check(copied, gc.DeepEquals, original)
}

func (s *DocSuite) TestCopyDifferent(c *gc.C) {
	original := &document.Doc{}
	original.Raw.ID = "some id"
	doc := original.Copy("another id")

	copied, ok := doc.(*document.Doc)
	c.Assert(ok, jc.IsTrue)

	c.Check(copied, gc.Not(gc.Equals), original)
	c.Check(copied, gc.Not(gc.DeepEquals), original)
}
