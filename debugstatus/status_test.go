// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package debugstatus_test

import (
	"errors"

	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"

	"github.com/juju/utils/debugstatus"
)

type statusSuite struct {
	jujutesting.IsolationSuite
}

var _ = gc.Suite(&statusSuite{})

func (s *statusSuite) TestCheck(c *gc.C) {
	checkers := map[string]debugstatus.CheckerFunc{
		"check1": func() (name, value string, passed bool) {
			return "check1 results", "value1", true
		},
		"check2": func() (name, value string, passed bool) {
			return "check2 results", "value2", false
		},
		"check3": func() (name, value string, passed bool) {
			return "check3 results", "value3", true
		},
	}
	results := debugstatus.Check(checkers)
	c.Assert(results, jc.DeepEquals, map[string]debugstatus.CheckResult{
		"check1": {
			Name:   "check1 results",
			Value:  "value1",
			Passed: true,
		},
		"check2": {
			Name:   "check2 results",
			Value:  "value2",
			Passed: false,
		},
		"check3": {
			Name:   "check3 results",
			Value:  "value3",
			Passed: true,
		},
	})
}

func (s *statusSuite) TestConnection(c *gc.C) {
	// Ensure a connection established is properly reported.
	check := debugstatus.Connection(pinger{nil}, "valid connection")
	name, value, passed := check()
	c.Assert(name, gc.Equals, "valid connection")
	c.Assert(value, gc.Equals, "Connected")
	c.Assert(passed, jc.IsTrue)

	// An error is reported if ping fails.
	check = debugstatus.Connection(pinger{errors.New("bad wolf")}, "connection error")
	name, value, passed = check()
	c.Assert(name, gc.Equals, "connection error")
	c.Assert(value, gc.Equals, "Ping error: bad wolf")
	c.Assert(passed, jc.IsFalse)
}

// pinger implements a debugstatus.Pinger used for tests.
type pinger struct {
	err error
}

func (p pinger) Ping() error {
	return p.err
}

var mongoCollectionsTests = []struct {
	about        string
	collector    collector
	expectValue  string
	expectPassed bool
}{{
	about: "all collection exist",
	collector: collector{
		expected: []string{"coll1", "coll2"},
		obtained: []string{"coll1", "coll2"},
	},
	expectValue:  "All required collections exist",
	expectPassed: true,
}, {
	about:        "no collections",
	expectValue:  "All required collections exist",
	expectPassed: true,
}, {
	about: "missing collections",
	collector: collector{
		expected: []string{"coll1", "coll2", "coll3"},
		obtained: []string{"coll2"},
	},
	expectValue:  "Missing collections: [coll1 coll3]",
	expectPassed: false,
}, {
	about: "error retrieving collections",
	collector: collector{
		err: errors.New("bad wolf"),
	},
	expectValue:  "Cannot get collections: bad wolf",
	expectPassed: false,
}}

func (s *statusSuite) TestMongoCollections(c *gc.C) {
	for i, test := range mongoCollectionsTests {
		c.Logf("test %d: %s", i, test.about)

		// Ensure a connection established is properly reported.
		check := debugstatus.MongoCollections(test.collector)
		name, value, passed := check()
		c.Assert(name, gc.Equals, "MongoDB collections")
		c.Assert(value, gc.Equals, test.expectValue)
		c.Assert(passed, gc.Equals, test.expectPassed)
	}
}

// collector implements a debugstatus.Collector used for tests.
type collector struct {
	expected []string
	obtained []string
	err      error
}

func (c collector) CollectionNames() ([]string, error) {
	return c.obtained, c.err
}

func (c collector) Collections() []*mgo.Collection {
	collections := make([]*mgo.Collection, len(c.expected))
	for i, name := range c.expected {
		collections[i] = &mgo.Collection{Name: name}
	}
	return collections
}
