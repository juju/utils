// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type SetenvSuite struct{}

var _ = gc.Suite(&SetenvSuite{})

var setenvTests = []struct {
	set    string
	expect []string
}{
	{"foo=1", []string{"foo=1", "arble="}},
	{"foo=", []string{"foo=", "arble="}},
	{"arble=23", []string{"foo=bar", "arble=23"}},
	{"zaphod=42", []string{"foo=bar", "arble=", "zaphod=42"}},
	{"bar", []string{"foo=bar", "arble="}},
}

func (*SetenvSuite) TestSetenv(c *gc.C) {
	env0 := []string{"foo=bar", "arble="}
	for i, t := range setenvTests {
		c.Logf("test %d", i)
		env := make([]string, len(env0))
		copy(env, env0)
		env = utils.Setenv(env, t.set)
		c.Check(env, gc.DeepEquals, t.expect)
	}
}
