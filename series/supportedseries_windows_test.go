// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package series_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/series"
)

type supportedSeriesWindowsSuite struct {
}

var _ = gc.Suite(&supportedSeriesWindowsSuite{})

func (s *supportedSeriesWindowsSuite) TestSeriesVersion(c *gc.C) {
	vers, err := series.SeriesVersion("win8")
	if err != nil {
		c.Assert(err, gc.Not(gc.ErrorMatches), `invalid series "win8"`, gc.Commentf(`unable to lookup series "win8"`))
	} else {
		c.Assert(err, jc.ErrorIsNil)
	}
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(vers, gc.Equals, "win8")
}

func (s *supportedSeriesWindowsSuite) TestSupportedSeries(c *gc.C) {
	expectedSeries := []string{
		"genericlinux",
		"centos7",
		"opensuseleap",

		"precise",
		"quantal",
		"raring",
		"saucy",
		"trusty",
		"utopic",
		"vivid",
		"wily",
		"xenial",

		"win10",
		"win2008r2",
		"win2012",
		"win2012hv",
		"win2012hvr2",
		"win2012r2",
		"win2016",
		"win2016nano",
		"win7",
		"win8",
		"win81",
	}
	series := series.SupportedSeries()
	c.Assert(series, jc.SameContents, expectedSeries)
}

func (s supportedSeriesWindowsSuite) TestWindowsVersions(c *gc.C) {
	sir := series.WindowsVersions()
	lsir := len(sir)
	wlen := len(WindowsVersionMap)
	nlen := len(WindowsNanoMap)
	verify := 0

	for i, ival := range sir {
		for j, jval := range WindowsVersionMap {
			if i == j && ival == jval {
				verify++
			}
		}
	}
	c.Assert(verify, c.Equals, wlen)
	verify = 0

	for i, ival := range sir {
		for j, jval := range WindowsNanoMap {
			if i == j && ival == jval {
				verify++
			}
		}
	}
	c.Assert(verify, c.Equals, nlen)
}
