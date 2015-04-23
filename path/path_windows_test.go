// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package path_test

import (
	"testing"

	gc "launchpad.net/gocheck"

	"github.com/juju/utils/path"
)

type PathUtilsSuite struct{}

var _ = gc.Suite(&PathUtilsSuite{})

func Test(t *testing.T) {
	gc.TestingT(t)
}

func (*PathUtilsSuite) TestLongPath(c *gc.C) {
	programFiles := `C:\PROGRA~1`
	longProg := `C:\Program Files`
	target, err := path.GetLongPathAsString(programFiles)
	c.Assert(err, gc.IsNil)
	c.Assert(target, gc.Equals, longProg)
}
