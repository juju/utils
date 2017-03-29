// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package series_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/os"
	"github.com/juju/utils/series"
)

type supportedSeriesSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&supportedSeriesSuite{})

func (s *supportedSeriesSuite) SetUpTest(c *gc.C) {
	s.CleanupSuite.SetUpTest(c)
	cleanup := series.SetSeriesVersions(make(map[string]string))
	s.AddCleanup(func(*gc.C) { cleanup() })
}

var getOSFromSeriesTests = []struct {
	series string
	want   os.OSType
	err    string
}{{
	series: "precise",
	want:   os.Ubuntu,
}, {
	series: "win2012r2",
	want:   os.Windows,
}, {
	series: "win2016nano",
	want:   os.Windows,
}, {
	series: "mountainlion",
	want:   os.OSX,
}, {
	series: "centos7",
	want:   os.CentOS,
}, {
	series: "opensuseleap",
	want:   os.OpenSUSE,
}, {
	series: "genericlinux",
	want:   os.GenericLinux,
}, {
	series: "",
	err:    "series \"\" not valid",
},
}

func (s *supportedSeriesSuite) TestGetOSFromSeries(c *gc.C) {
	for _, t := range getOSFromSeriesTests {
		got, err := series.GetOSFromSeries(t.series)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Check(err, jc.ErrorIsNil)
			c.Assert(got, gc.Equals, t.want)
		}
	}
}

func (s *supportedSeriesSuite) TestUnknownOSFromSeries(c *gc.C) {
	_, err := series.GetOSFromSeries("Xuanhuaceratops")
	c.Assert(err, jc.Satisfies, series.IsUnknownOSForSeriesError)
	c.Assert(err, gc.ErrorMatches, `unknown OS for series: "Xuanhuaceratops"`)
}

func setSeriesTestData() {
	series.SetSeriesVersions(map[string]string{
		"trusty":       "14.04",
		"utopic":       "14.10",
		"win7":         "win7",
		"win81":        "win81",
		"win2016nano":  "win2016nano",
		"centos7":      "centos7",
		"opensuseleap": "opensuse42",
		"genericlinux": "genericlinux",
	})
}

func (s *supportedSeriesSuite) TestOSSupportedSeries(c *gc.C) {
	setSeriesTestData()
	supported := series.OSSupportedSeries(os.Ubuntu)
	c.Assert(supported, jc.SameContents, []string{"trusty", "utopic"})
	supported = series.OSSupportedSeries(os.Windows)
	c.Assert(supported, jc.SameContents, []string{"win7", "win81", "win2016nano"})
	supported = series.OSSupportedSeries(os.CentOS)
	c.Assert(supported, jc.SameContents, []string{"centos7"})
	supported = series.OSSupportedSeries(os.OpenSUSE)
	c.Assert(supported, jc.SameContents, []string{"opensuseleap"})
	supported = series.OSSupportedSeries(os.GenericLinux)
	c.Assert(supported, jc.SameContents, []string{"genericlinux"})
}

func (s *supportedSeriesSuite) TestVersionSeriesValid(c *gc.C) {
	setSeriesTestData()
	seriesResult, err := series.VersionSeries("14.04")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert("trusty", gc.DeepEquals, seriesResult)
}

func (s *supportedSeriesSuite) TestVersionSeriesEmpty(c *gc.C) {
	setSeriesTestData()
	_, err := series.VersionSeries("")
	c.Assert(err, gc.ErrorMatches, `.*unknown series for version: "".*`)
}

func (s *supportedSeriesSuite) TestVersionSeriesInvalid(c *gc.C) {
	setSeriesTestData()
	_, err := series.VersionSeries("73655")
	c.Assert(err, gc.ErrorMatches, `.*unknown series for version: "73655".*`)
}

func (s *supportedSeriesSuite) TestSeriesVersionEmpty(c *gc.C) {
	setSeriesTestData()
	_, err := series.SeriesVersion("")
	c.Assert(err, gc.ErrorMatches, `.*unknown version for series: "".*`)
}

func (s *supportedSeriesSuite) TestIsWindowsNano(c *gc.C) {
	var isWindowsNanoTests = []struct {
		series   string
		expected bool
	}{
		{"win2016nano", true},
		{"win2016", false},
		{"win2012r2", false},
		{"trusty", false},
	}

	for _, t := range isWindowsNanoTests {
		c.Assert(series.IsWindowsNano(t.series), gc.Equals, t.expected)
	}
}

func (s *supportedSeriesSuite) TestLatestLts(c *gc.C) {
	table := []struct {
		latest, want string
	}{
		{"testseries", "testseries"},
		{"", "xenial"},
	}
	for _, test := range table {
		series.SetLatestLtsForTesting(test.latest)
		got := series.LatestLts()
		c.Assert(got, gc.Equals, test.want)
	}
}

func (s *supportedSeriesSuite) TestSetLatestLtsForTesting(c *gc.C) {
	table := []struct {
		value, want string
	}{
		{"1", "xenial"}, {"2", "1"}, {"3", "2"}, {"4", "3"},
	}
	for _, test := range table {
		got := series.SetLatestLtsForTesting(test.value)
		c.Assert(got, gc.Equals, test.want)
	}
}

func (s *supportedSeriesSuite) TestSupportedLts(c *gc.C) {
	got := series.SupportedLts()
	want := []string{"precise", "trusty", "xenial"}
	c.Assert(got, gc.DeepEquals, want)
}
