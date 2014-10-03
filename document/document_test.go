// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package document_test

import (
	"bytes"
	"time"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/document"
)

var _ = gc.Suite(&DocSuite{})

type DocSuite struct {
	testing.IsolationSuite
}

func (s *DocSuite) TestNewDocument(c *gc.C) {
	timestamp := time.Now().UTC()
	doc := document.NewDocument(&timestamp)

	c.Check(doc.Raw.ID, gc.Equals, "")
	c.Check(doc.Raw.Created, gc.Equals, timestamp)
}

func (s *DocSuite) TestNewDocumentDefaultTimestamp(c *gc.C) {
	doc := document.NewDocument(nil)

	c.Check(doc.Raw.Created, gc.NotNil)
}

func (s *DocSuite) TestID(c *gc.C) {
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	id := doc.ID()

	c.Check(id, gc.Equals, "some id")
	c.Check(id, gc.Equals, doc.Raw.ID)
}

func (s *DocSuite) TestCreated(c *gc.C) {
	timestamp := time.Now().UTC()
	doc := document.NewDocument(&timestamp)
	created := doc.Created()

	c.Check(created, gc.Equals, timestamp)
}

func (s *DocSuite) TestSetIDInitial(c *gc.C) {
	doc := &document.Doc{}
	c.Assert(doc.ID(), gc.Equals, "")
	already := doc.SetID("some id")

	c.Check(already, gc.Equals, false)
	c.Check(doc.Raw.ID, gc.Equals, "some id")
}

func (s *DocSuite) TestSetIDAlreadySetSame(c *gc.C) {
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	already := doc.SetID("some id")

	c.Check(already, gc.Equals, true)
	c.Check(doc.ID(), gc.Equals, "some id")
}

func (s *DocSuite) TestSetIDAlreadySetDifferent(c *gc.C) {
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	already := doc.SetID("another id")

	c.Check(already, gc.Equals, true)
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

func (s *DocSuite) TestDumpJSON(c *gc.C) {
	timestamp := time.Now().UTC()
	doc := &document.Doc{}
	doc.Raw.ID = "some id"
	doc.Raw.Created = timestamp
	var file bytes.Buffer
	err := doc.Dump(&file, "json")
	c.Assert(err, gc.IsNil)

	ts := timestamp.Format(time.RFC3339Nano)
	expected := `{"Raw":{"ID":"some id","Created":"` + ts + `"}}` + "\n"
	c.Check(file.String(), gc.Equals, expected)
}

func (s *DocSuite) TestDumpNotSupported(c *gc.C) {
	doc := &document.Doc{}
	err := doc.Dump(nil, "bogus")

	c.Check(err, jc.Satisfies, errors.IsNotSupported)
}

func (s *DocSuite) TestLoadJSON(c *gc.C) {
	timestamp := time.Now().UTC()
	ts := timestamp.Format(time.RFC3339Nano)
	data := `{"Raw":{"ID":"some id","Created":"` + ts + `"}}` + "\n"
	file := bytes.NewBufferString(data)
	doc := &document.Doc{}
	err := doc.Load(file, "json")
	c.Assert(err, gc.IsNil)

	c.Check(doc.ID(), gc.Equals, "some id")
	c.Check(doc.Created(), gc.Equals, timestamp)
}

func (s *DocSuite) TestLoadNotSupported(c *gc.C) {
	doc := &document.Doc{}
	err := doc.Load(nil, "bogus")

	c.Check(err, jc.Satisfies, errors.IsNotSupported)
}

func (s *DocSuite) TestJSONRoundTrip(c *gc.C) {
	var doc document.Doc
	doc.Raw.ID = "some id"
	doc.Raw.Created = time.Now().UTC()
	var file bytes.Buffer
	err := doc.Dump(&file, "json")
	c.Assert(err, gc.IsNil)

	var copied document.Doc
	err = copied.Load(&file, "json")
	c.Assert(err, gc.IsNil)

	c.Check(doc.Raw.ID, gc.Equals, copied.Raw.ID)
	c.Check(doc.Raw.Created, gc.Equals, copied.Raw.Created)
}

func (s *DocSuite) TestDefaultID(c *gc.C) {
	var doc document.Doc
	_, err := doc.DefaultID()

	c.Check(err, jc.Satisfies, errors.IsNotSupported)
}

func (s *DocSuite) TestValidateFull(c *gc.C) {
	var doc document.Doc
	doc.Raw.ID = "some id"
	doc.Raw.Created = time.Now().UTC()
	err := doc.Validate("full")

	c.Check(err, gc.IsNil)
}

func (s *DocSuite) TestValidateIDValid(c *gc.C) {
	var doc document.Doc
	doc.Raw.ID = "some id"
	doc.Raw.Created = time.Now().UTC()
	err := doc.Validate("id")

	c.Check(err, gc.IsNil)
}

func (s *DocSuite) TestValidateIDInvalid(c *gc.C) {
	var doc document.Doc
	doc.Raw.Created = time.Now().UTC()
	err := doc.Validate("id")

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *DocSuite) TestValidateIniitialized(c *gc.C) {
	var doc document.Doc
	doc.Raw.Created = time.Now().UTC()
	err := doc.Validate("initialized")

	c.Check(err, gc.IsNil)
}

func (s *DocSuite) TestValidateIniitializedInvalid(c *gc.C) {
	var doc document.Doc
	err := doc.Validate("initialized")

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *DocSuite) TestValidateNotSupported(c *gc.C) {
	var doc document.Doc
	err := doc.Validate("bogus")

	c.Check(err, jc.Satisfies, errors.IsNotSupported)
}
