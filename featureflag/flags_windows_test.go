// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package featureflag_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"golang.org/x/sys/windows/registry"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/featureflag"
)

type flagWinSuite struct {
	testing.IsolationSuite
	k registry.Key
}

var _ = gc.Suite(&flagWinSuite{})

// We use a "random" key here for the tests
const regKey = `HKLM:\SOFTWARE\juju-9362394821442`

func (s *flagWinSuite) SetUpTest(c *gc.C) {
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, regKey[6:], registry.ALL_ACCESS)
	c.Assert(err, jc.ErrorIsNil)
	s.k = k
}

func (s *flagWinSuite) TearDownTest(c *gc.C) {
	s.k.DeleteValue("JUJU_TESTING_FEATURE")
	s.k.DeleteValue("JUJU_TESTING_FEATURE2")
	s.k.Close()
	registry.DeleteKey(registry.LOCAL_MACHINE, regKey[6:])
}

func (s *flagWinSuite) TestEmpty(c *gc.C) {
	s.k.SetStringValue("JUJU_TESTING_FEATURE", "")
	featureflag.SetFlagsFromRegistry(regKey, "JUJU_TESTING_FEATURE")
	c.Assert(featureflag.All(), gc.HasLen, 0)
	c.Assert(featureflag.AsEnvironmentValue(), gc.Equals, "")
	c.Assert(featureflag.String(), gc.Equals, "")
}

func (s *flagWinSuite) TestParsing(c *gc.C) {
	s.k.SetStringValue("JUJU_TESTING_FEATURE", "MAGIC, test, space ")
	s.k.SetStringValue("JUJU_TESTING_FEATURE2", "magic2")
	featureflag.SetFlagsFromRegistry(regKey, "JUJU_TESTING_FEATURE", "JUJU_TESTING_FEATURE2")
	c.Assert(featureflag.All(), jc.SameContents, []string{"magic", "space", "test", "magic2"})
	c.Assert(featureflag.AsEnvironmentValue(), gc.Equals, "magic,magic2,space,test")
	c.Assert(featureflag.String(), gc.Equals, `"magic", "magic2", "space", "test"`)
}

func (s *flagWinSuite) TestEnabled(c *gc.C) {
	c.Assert(featureflag.Enabled(""), jc.IsTrue)
	c.Assert(featureflag.Enabled(" "), jc.IsTrue)
	c.Assert(featureflag.Enabled("magic"), jc.IsFalse)
	c.Assert(featureflag.Enabled("magic2"), jc.IsFalse)

	s.k.SetStringValue("JUJU_TESTING_FEATURE", "MAGIC")
	s.k.SetStringValue("JUJU_TESTING_FEATURE2", "MAGIC2")
	featureflag.SetFlagsFromRegistry(regKey, "JUJU_TESTING_FEATURE", "JUJU_TESTING_FEATURE2")

	c.Assert(featureflag.Enabled("magic"), jc.IsTrue)
	c.Assert(featureflag.Enabled("Magic"), jc.IsTrue)
	c.Assert(featureflag.Enabled(" MAGIC "), jc.IsTrue)
	c.Assert(featureflag.Enabled("magic2"), jc.IsTrue)
	c.Assert(featureflag.Enabled("Magic2"), jc.IsTrue)
	c.Assert(featureflag.Enabled(" MAGIC2 "), jc.IsTrue)
}
