// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package utils_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
)

type randomStringSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&randomStringSuite{})

var (
	validChars = []rune("thisissorandom")
	length     = 7
)

func (randomStringSuite) TestLength(c *gc.C) {
	s := utils.RandomString(length, validChars)
	c.Assert(s, gc.HasLen, length)
}

func (randomStringSuite) TestContentInValidRunes(c *gc.C) {
	s := utils.RandomString(length, validChars)
	for _, char := range s {
		c.Assert(string(validChars), jc.Contains, string(char))
	}
}
