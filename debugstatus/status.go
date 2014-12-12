// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package debugstatus

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

// Check collects the status check results from the given checkers.
func Check(checkers map[string]StatusCheckerFunc) map[string]StatusCheckResult {
	results := make(map[string]StatusCheckResult, len(checkers))
	for key, c := range checkers {
		name, value, passed := c()
		results[key] = StatusCheckResult{
			Name:   name,
			Value:  value,
			Passed: passed,
		}
	}
	return results
}

// StatusCheckResult holds the result of a single status check.
type StatusCheckResult struct {
	Name   string
	Value  string
	Passed bool
}

// StatusCheckerFunc represents a function returning the name of the check,
// its value and reporting whether the check passed.
type StatusCheckerFunc func() (name, value string, passed bool)

// startTime holds the time that the code started running.
var startTime = time.Now()

// CheckStartTime reports the time when the application was started.
func CheckStartTime() (name, value string, passed bool) {
	return "Server started", startTime.UTC().String(), true
}

// CheckConnection returns a status checker reporting whether the given Pinger
// is connected.
func CheckConnection(p Pinger, n string) StatusCheckerFunc {
	return func() (name, value string, passed bool) {
		if err := p.Ping(); err != nil {
			return n, "Ping error: " + err.Error(), false
		}
		return n, "Connected", true
	}
}

// Pinger is an interface that wraps the Ping method.
type Pinger interface {
	Ping() error
}

// CheckMongoCollections returns a status checker checking that all the
// expected Mongo collections are present in the database.
func CheckMongoCollections(c Collector) StatusCheckerFunc {
	return func() (name, value string, passed bool) {
		name = "MongoDB collections"
		names, err := c.CollectionNames()
		if err != nil {
			return name, "Cannot get collections: " + err.Error(), false
		}
		var missing []string
		for _, coll := range c.Collections() {
			found := false
			for _, name := range names {
				if name == coll.Name {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, coll.Name)
			}
		}
		if len(missing) == 0 {
			return name, "All required collections exist", true
		}
		return name, fmt.Sprintf("Missing collections: %s", missing), false
	}
}

// Collector is an interface that groups the methods used to check that
// a Mongo database has the expected collections.
type Collector interface {
	// Collections returns the Mongo collections that we expect to exist in
	// the Mongo database.
	Collections() []*mgo.Collection

	// CollectionNames returns the names of the collections actually present in
	// the Mongo database.
	CollectionNames() ([]string, error)
}
