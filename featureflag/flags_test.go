// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package featureflag_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/featureflag"
)

type flagSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&flagSuite{})

func (s *flagSuite) TestEmpty(c *gc.C) {
	s.PatchEnvironment("JUJU_TESTING_FEATURE", "")
	featureflag.SetFlagsFromEnvironment("JUJU_TESTING_FEATURE")
	c.Assert(featureflag.All(), gc.HasLen, 0)
	c.Assert(featureflag.AsEnvironmentValue(), gc.Equals, "")
	c.Assert(featureflag.String(), gc.Equals, "")
}

func (s *flagSuite) TestParsing(c *gc.C) {
	s.PatchEnvironment("JUJU_TESTING_FEATURE", "MAGIC, test, space ")
	featureflag.SetFlagsFromEnvironment("JUJU_TESTING_FEATURE")
	c.Assert(featureflag.All(), jc.SameContents, []string{"magic", "space", "test"})
	c.Assert(featureflag.AsEnvironmentValue(), gc.Equals, "magic,space,test")
	c.Assert(featureflag.String(), gc.Equals, `"magic", "space", "test"`)
}

func (s *flagSuite) TestEnabled(c *gc.C) {
	c.Assert(featureflag.Enabled(""), jc.IsTrue)
	c.Assert(featureflag.Enabled(" "), jc.IsTrue)
	c.Assert(featureflag.Enabled("magic"), jc.IsFalse)

	s.PatchEnvironment("JUJU_TESTING_FEATURE", "MAGIC")
	featureflag.SetFlagsFromEnvironment("JUJU_TESTING_FEATURE")

	c.Assert(featureflag.Enabled("magic"), jc.IsTrue)
	c.Assert(featureflag.Enabled("Magic"), jc.IsTrue)
	c.Assert(featureflag.Enabled(" MAGIC "), jc.IsTrue)
}
