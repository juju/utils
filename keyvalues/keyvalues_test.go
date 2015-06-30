// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package keyvalues_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/keyvalues"
)

type keyValuesSuite struct{}

var _ = gc.Suite(&keyValuesSuite{})

var testCases = []struct {
	about         string
	input         []string
	allowEmptyVal bool
	output        map[string]string
	error         string
}{{
	about:         "simple test case",
	input:         []string{"key=value"},
	allowEmptyVal: false,
	output:        map[string]string{"key": "value"},
	error:         "",
}, {
	about:         "empty list",
	input:         []string{},
	allowEmptyVal: false,
	output:        map[string]string{},
	error:         "",
}, {
	about:         "nil list",
	input:         nil,
	allowEmptyVal: false,
	output:        map[string]string{},
	error:         "",
}, {
	about:         "invalid format - missing value",
	input:         []string{"key"},
	allowEmptyVal: false,
	output:        nil,
	error:         `expected "key=value", got "key"`,
}, {
	about:         "invalid format - missing value",
	input:         []string{"key="},
	allowEmptyVal: false,
	output:        nil,
	error:         `expected "key=value", got "key="`,
}, {
	about:         "invalid format - missing key",
	input:         []string{"=value"},
	allowEmptyVal: false,
	output:        nil,
	error:         `expected "key=value", got "=value"`,
}, {
	about:         "invalid format",
	input:         []string{"="},
	allowEmptyVal: false,
	output:        nil,
	error:         `expected "key=value", got "="`,
}, {
	about:         "invalid format, allowing empty",
	input:         []string{"="},
	allowEmptyVal: true,
	output:        nil,
	error:         `expected "key=value", got "="`,
}, {
	about:         "duplicate keys",
	input:         []string{"key=value", "key=value"},
	allowEmptyVal: true,
	output:        nil,
	error:         `key "key" specified more than once`,
}, {
	about:         "multiple keys",
	input:         []string{"key=value", "key2=value", "key3=value"},
	allowEmptyVal: true,
	output:        map[string]string{"key": "value", "key2": "value", "key3": "value"},
	error:         "",
}, {
	about:         "empty value",
	input:         []string{"key="},
	allowEmptyVal: true,
	output:        map[string]string{"key": ""},
	error:         "",
}, {
	about:         "whitespace trimmed",
	input:         []string{"key=value\n", "key2\t=\tvalue2"},
	allowEmptyVal: true,
	output:        map[string]string{"key": "value", "key2": "value2"},
	error:         "",
}, {
	about:         "whitespace trimming and duplicate keys",
	input:         []string{"key =value", "key\t=\tvalue2"},
	allowEmptyVal: true,
	output:        nil,
	error:         `key "key" specified more than once`,
}, {
	about:         "whitespace trimming and empty value not allowed",
	input:         []string{"key=    "},
	allowEmptyVal: false,
	output:        nil,
	error:         `expected "key=value", got "key="`,
}, {
	about:         "whitespace trimming and empty value",
	input:         []string{"key=    "},
	allowEmptyVal: true,
	output:        map[string]string{"key": ""},
	error:         "",
}, {
	about:         "whitespace trimming and missing key",
	input:         []string{"   =value"},
	allowEmptyVal: true,
	output:        nil,
	error:         `expected "key=value", got "=value"`,
}}

func (keyValuesSuite) TestMapParsing(c *gc.C) {
	for i, t := range testCases {
		c.Log("test %d: %s", i, t.about)
		result, err := keyvalues.Parse(t.input, t.allowEmptyVal)
		c.Check(result, gc.DeepEquals, t.output)
		if t.error == "" {
			c.Check(err, gc.IsNil)
		} else {
			c.Check(err, gc.ErrorMatches, t.error)
		}
	}
}
