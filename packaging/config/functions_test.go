// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package config_test

import (
	"fmt"

	"github.com/juju/utils/packaging"
	"github.com/juju/utils/packaging/config"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(&FunctionsSuite{})

type FunctionsSuite struct {
}

func (s *FunctionsSuite) TestSeriesRequiresCloudArchiveTools(c *gc.C) {
	testedSeries := []string{
		"precise",
		"centos7",
		"rogentos",
		"debian",
		"mint",
		"makulu",
		"rhel lol",
	}

	for i, series := range testedSeries {
		c.Logf("Test %d: series %s:", i+1, series)

		res := config.SeriesRequiresCloudArchiveTools(series)
		_, req := config.SeriesRequiringCloudTools[series]

		c.Assert(res, gc.Equals, req)
	}
}

func (s *FunctionsSuite) TestGetCloudArchiveSourceCentOS(c *gc.C) {
	src, prefs := config.GetCloudArchiveSource("centos7")

	c.Assert(src, gc.Equals, packaging.PackageSource{})
	c.Assert(prefs, gc.Equals, packaging.PackagePreferences{})
}

func (s *FunctionsSuite) TestGetCloudArchiveSourceUbuntu(c *gc.C) {
	expectedSrc := packaging.PackageSource{
		URL: fmt.Sprintf("deb %s %s-updates/cloud-tools main", config.UbuntuCloudArchiveUrl, "precise"),
		Key: config.UbuntuCloudArchiveSigningKey,
	}

	expectedPrefs := packaging.PackagePreferences{
		Path:        config.UbuntuCloudToolsPrefsPath,
		Explanation: "Pin with lower priority, not to interfere with charms.",
		Package:     "*",
		Pin:         fmt.Sprintf("release n=%s-updates/cloud-tools", "precise"),
		Priority:    400,
	}

	src, prefs := config.GetCloudArchiveSource("precise")

	c.Assert(src, gc.Equals, expectedSrc)
	c.Assert(prefs, gc.Equals, expectedPrefs)
}

func (s *FunctionsSuite) TestRequiresBackportsTrustyLXD(c *gc.C) {
	requiresBackports := config.RequiresBackports("trusty", "lxd")
	c.Assert(requiresBackports, jc.IsTrue)
}
