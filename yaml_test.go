// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
