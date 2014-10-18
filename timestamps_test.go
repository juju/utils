// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"time"

	"github.com/juju/testing"
	gc "launchpad.net/gocheck"

	"github.com/juju/utils"
)

type tsSuite struct {
	testing.IsolationSuite
	timestamp *time.Time
}

var _ = gc.Suite(&tsSuite{})

func (s *tsSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	ts := time.Date(2014, 7, 31, 4, 59, 59, 0, time.UTC)
	s.timestamp = &ts
}

//---------------------------
// FormatTimestamp()

func (s *tsSuite) TestTimestampFormatTimestampSuccess(c *gc.C) {
	ts := utils.FormatTimestamp("%Y%M%D-%h%m%s", s.timestamp)

	c.Check(ts, gc.Equals, "20140731-045959")
}

func (s *tsSuite) TestTimestampFormatTimestampExtras(c *gc.C) {
	ts := utils.FormatTimestamp("juju-%Y%M%D-%h%m%s-juju", s.timestamp)

	c.Check(ts, gc.Equals, "juju-20140731-045959-juju")
}

func (s *tsSuite) TestTimestampFormatTimestampEmpty(c *gc.C) {
	ts := utils.FormatTimestamp("", nil)

	c.Check(ts, gc.Equals, "")
}

func (s *tsSuite) TestTimestampFormatTimestampBarePercents(c *gc.C) {
	ts := utils.FormatTimestamp("spam%-%%%Y%M%D-%h%m%s%% %", s.timestamp)

	c.Check(ts, gc.Equals, "spam%-%%20140731-045959%% %")
}

//---------------------------
// ParseTimestamp()

func (s *tsSuite) TestTimestampParseTimestampSuccess(c *gc.C) {
	ts := utils.ParseTimestamp("%Y%M%D-%h%m%s", "20140731-045959")

	c.Check(ts, gc.DeepEquals, s.timestamp)
}

func (s *tsSuite) TestTimestampParseTimestampExtras(c *gc.C) {
	ts := utils.ParseTimestamp("%Y%M%D-%h%m%s", "20140731-045959-juju")

	c.Check(ts, gc.DeepEquals, s.timestamp)
}

func (s *tsSuite) TestTimestampParseTimestampNoMatch(c *gc.C) {
	ts := utils.ParseTimestamp("%Y%M%D-%h%m%s", "20140731---045959")

	c.Check(ts, gc.IsNil)
}
