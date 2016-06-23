// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type yamlSuite struct {
}

var _ = gc.Suite(&yamlSuite{})

func (*yamlSuite) TestYamlRoundTrip(c *gc.C) {
	// test happy path of round tripping an object via yaml

	type T struct {
		A int    `yaml:"a"`
		B bool   `yaml:"deleted"`
		C string `yaml:"omitempty"`
		D string
	}

	v := T{A: 1, B: true, C: "", D: ""}

	f, err := ioutil.TempFile(c.MkDir(), "yaml")
	c.Assert(err, gc.IsNil)
	path := f.Name()
	f.Close()

	err = WriteYaml(path, v)
	c.Assert(err, gc.IsNil)

	var v2 T
	err = ReadYaml(path, &v2)
	c.Assert(err, gc.IsNil)

	c.Assert(v, gc.Equals, v2)
}

func (*yamlSuite) TestReadYamlReturnsNotFound(c *gc.C) {
	// The contract for ReadYaml requires it returns an error
	// that can be inspected by os.IsNotExist. Notably, we cannot
	// use juju/errors gift wrapping.
	f, err := ioutil.TempFile(c.MkDir(), "yaml")
	c.Assert(err, gc.IsNil)
	path := f.Name()
	err = os.Remove(path)
	c.Assert(err, gc.IsNil)
	err = ReadYaml(path, nil)

	// assert that the error is reported as NotExist
	c.Assert(os.IsNotExist(err), gc.Equals, true)
}

func (*yamlSuite) TestWriteYamlMissingDirectory(c *gc.C) {
	// WriteYaml tries to create a temporary file in the same
	// directory as the target. Test what happens if the path's
	// directory is missing

	root := c.MkDir()
	missing := filepath.Join(root, "missing", "filename")

	v := struct{ A, B int }{1, 2}
	err := WriteYaml(missing, v)
	c.Assert(err, gc.NotNil)
}

func (*yamlSuite) TestWriteYamlWriteGarbage(c *gc.C) {
	c.Skip("https://github.com/go-yaml/yaml/issues/144")
	// some things cannot be marshalled into yaml, check that
	// WriteYaml detects this.

	root := c.MkDir()
	path := filepath.Join(root, "f")

	v := struct{ A, B [10]bool }{}
	err := WriteYaml(path, v)
	c.Assert(err, gc.NotNil)
}

type ConformSuite struct{}

var _ = gc.Suite(&ConformSuite{})

func (s *ConformSuite) TestConformYAML(c *gc.C) {
	var goodInterfaceTests = []struct {
		description       string
		inputInterface    interface{}
		expectedInterface map[string]interface{}
		expectedError     string
	}{{
		description: "An interface requiring no changes.",
		inputInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
		expectedInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
	}, {
		description: "Substitute a single inner map[i]i.",
		inputInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[interface{}]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
		expectedInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
	}, {
		description: "Substitute nested inner map[i]i.",
		inputInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": "val2a",
			"key3a": map[interface{}]interface{}{
				"key1b": "val1b",
				"key2b": map[interface{}]interface{}{
					"key1c": "val1c"}}},
		expectedInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": "val2a",
			"key3a": map[string]interface{}{
				"key1b": "val1b",
				"key2b": map[string]interface{}{
					"key1c": "val1c"}}},
	}, {
		description: "Substitute nested map[i]i within []i.",
		inputInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": []interface{}{5, "foo", map[string]interface{}{
				"key1b": "val1b",
				"key2b": map[interface{}]interface{}{
					"key1c": "val1c"}}}},
		expectedInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": []interface{}{5, "foo", map[string]interface{}{
				"key1b": "val1b",
				"key2b": map[string]interface{}{
					"key1c": "val1c"}}}},
	}, {
		description: "An inner map[interface{}]interface{} with an int key.",
		inputInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[interface{}]interface{}{
				"foo1": "val1",
				5:      "val2"}},
		expectedError: "map keyed with non-string value",
	}, {
		description: "An inner []interface{} containing a map[i]i with an int key.",
		inputInterface: map[string]interface{}{
			"key1a": "val1b",
			"key2a": "val2b",
			"key3a": []interface{}{"foo1", 5, map[interface{}]interface{}{
				"key1b": "val1b",
				"key2b": map[interface{}]interface{}{
					"key1c": "val1c",
					5:       "val2c"}}}},
		expectedError: "map keyed with non-string value",
	}}

	for i, test := range goodInterfaceTests {
		c.Logf("test %d: %s", i, test.description)
		input := test.inputInterface
		cleansedInterfaceMap, err := ConformYAML(input)
		if test.expectedError == "" {
			if !c.Check(err, jc.ErrorIsNil) {
				continue
			}
			c.Check(cleansedInterfaceMap, jc.DeepEquals, test.expectedInterface)
		} else {
			c.Check(err, gc.ErrorMatches, test.expectedError)
		}
	}
}
