// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package debugstatus_test

import (
	"errors"
	"time"

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

func makeCheckerFunc(key, name, value string, passed bool) debugstatus.CheckerFunc {
	return func() (string, debugstatus.CheckResult) {
		time.Sleep(time.Microsecond)
		return key, debugstatus.CheckResult{
			Name:   name,
			Value:  value,
			Passed: passed,
		}
	}
}

func (s *statusSuite) TestCheck(c *gc.C) {
	results := debugstatus.Check(
		makeCheckerFunc("check1", "check1 name", "value1", true),
		makeCheckerFunc("check2", "check2 name", "value2", false),
		makeCheckerFunc("check3", "check3 name", "value3", true),
	)
	for key, r := range results {
		if r.Duration < time.Microsecond {
			c.Errorf("got %v want >1µs", r.Duration)
		}
		r.Duration = 0
		results[key] = r
	}

	c.Assert(results, jc.DeepEquals, map[string]debugstatus.CheckResult{
		"check1": {
			Name:   "check1 name",
			Value:  "value1",
			Passed: true,
		},
		"check2": {
			Name:   "check2 name",
			Value:  "value2",
			Passed: false,
		},
		"check3": {
			Name:   "check3 name",
			Value:  "value3",
			Passed: true,
		},
	})
}

func (s *statusSuite) TestServerStartTime(c *gc.C) {
	startTime := time.Now()
	s.PatchValue(&debugstatus.StartTime, startTime)
	key, result := debugstatus.ServerStartTime()
	c.Assert(key, gc.Equals, "server_started")
	c.Assert(result, jc.DeepEquals, debugstatus.CheckResult{
		Name:   "Server started",
		Value:  startTime.String(),
		Passed: true,
	})
}

func (s *statusSuite) TestConnection(c *gc.C) {
	// Ensure a connection established is properly reported.
	check := debugstatus.Connection(pinger{nil})
	key, result := check()
	c.Assert(key, gc.Equals, "mongo_connected")
	c.Assert(result, jc.DeepEquals, debugstatus.CheckResult{
		Name:   "MongoDB is connected",
		Value:  "Connected",
		Passed: true,
	})

	// An error is reported if ping fails.
	check = debugstatus.Connection(pinger{errors.New("bad wolf")})
	key, result = check()
	c.Assert(key, gc.Equals, "mongo_connected")
	c.Assert(result, jc.DeepEquals, debugstatus.CheckResult{
		Name:   "MongoDB is connected",
		Value:  "Ping error: bad wolf",
		Passed: false,
	})
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
	expectValue: "Missing collections: [coll1 coll3]",
}, {
	about: "error retrieving collections",
	collector: collector{
		err: errors.New("bad wolf"),
	},
	expectValue: "Cannot get collections: bad wolf",
}}

func (s *statusSuite) TestMongoCollections(c *gc.C) {
	for i, test := range mongoCollectionsTests {
		c.Logf("test %d: %s", i, test.about)

		// Ensure a connection established is properly reported.
		check := debugstatus.MongoCollections(test.collector)
		key, result := check()
		c.Assert(key, gc.Equals, "mongo_collections")
		c.Assert(result, jc.DeepEquals, debugstatus.CheckResult{
			Name:   "MongoDB collections",
			Value:  test.expectValue,
			Passed: test.expectPassed,
		})
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

var renameTests = []struct {
	about string
	key   string
	name  string
}{{
	about: "rename key",
	key:   "new-key",
}, {
	about: "rename name",
	name:  "new name",
}, {
	about: "rename both",
	key:   "another-key",
	name:  "another name",
}, {
	about: "do not rename",
}}

func (s *statusSuite) TestRename(c *gc.C) {
	check := makeCheckerFunc("old-key", "old name", "value", true)
	for i, test := range renameTests {
		c.Logf("test %d: %s", i, test.about)

		// Rename and run the check.
		key, result := debugstatus.Rename(test.key, test.name, check)()

		// Ensure the results are successfully renamed.
		expectKey := test.key
		if expectKey == "" {
			expectKey = "old-key"
		}
		expectName := test.name
		if expectName == "" {
			expectName = "old name"
		}
		c.Assert(key, gc.Equals, expectKey)
		c.Assert(result, jc.DeepEquals, debugstatus.CheckResult{
			Name:   expectName,
			Value:  "value",
			Passed: true,
		})
	}
}
