// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package config_test

import (
	"fmt"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/packaging/config"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(&ZypperSuite{})

type ZypperSuite struct {
	pacconfer config.PackagingConfigurer
}

func (s *ZypperSuite) SetUpSuite(c *gc.C) {
	s.pacconfer = config.NewZypperPackagingConfigurer(testedSeriesOpenSUSE)
}

func (s *ZypperSuite) TestDefaultPackages(c *gc.C) {
	c.Assert(s.pacconfer.DefaultPackages(), gc.DeepEquals, config.OpenSUSEDefaultPackages)
}

func (s *ZypperSuite) TestGetPackageNameForSeriesSameSeries(c *gc.C) {
	for _, pack := range testedPackages {
		res, err := s.pacconfer.GetPackageNameForSeries(pack, testedSeriesOpenSUSE)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(res, gc.Equals, pack)
	}
}

func (s *ZypperSuite) TestGetPackageNameForSeriesErrors(c *gc.C) {
	for _, pack := range testedPackages {
		res, err := s.pacconfer.GetPackageNameForSeries(pack, "some-other-series")
		c.Assert(res, gc.Equals, "")
		c.Assert(err, gc.ErrorMatches, fmt.Sprintf("no equivalent package found for series %s: %s", "some-other-series", pack))
	}
}

func (s *ZypperSuite) TestIsCloudArchivePackage(c *gc.C) {
	testedPacks := []string{
		"random",
		"stuff",
		"mongodb",
		"cloud-utils",
		"more",
		"random stuff",
	}

	for i, pack := range testedPacks {
		c.Logf("Test %d: package %s:", i+1, pack)
		res := s.pacconfer.IsCloudArchivePackage(pack)
		_, there := config.CloudArchivePackagesOpenSUSE[pack]

		c.Assert(res, gc.Equals, there)
	}
}

func (s *ZypperSuite) TestRenderSource(c *gc.C) {
	expected, err := testedSource.RenderSourceFile(config.ZypperSourceTemplate)
	c.Assert(err, jc.ErrorIsNil)

	res, err := s.pacconfer.RenderSource(testedSource)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(res, gc.Equals, expected)
}
