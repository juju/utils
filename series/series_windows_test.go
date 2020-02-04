// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package series_test

import (
	"fmt"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"golang.org/x/sys/windows/registry"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils.v2"
	"github.com/juju/utils.v2/series"
)

type windowsSeriesSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&windowsSeriesSuite{})

var versionTests = []struct {
	version string
	want    string
}{
	{
		"Hyper-V Server 2012 R2",
		"win2012hvr2",
	},
	{
		"Hyper-V Server 2012",
		"win2012hv",
	},
	{
		"Windows Server 2008 R2",
		"win2008r2",
	},
	{
		"Windows Server 2012 R2",
		"win2012r2",
	},
	{
		"Windows Server 2012",
		"win2012",
	},
	{
		"Windows Server 2012 R2 Datacenter",
		"win2012r2",
	},
	{
		"Windows Server 2012 Standard",
		"win2012",
	},
	{
		"Windows Storage Server 2012 R2",
		"win2012r2",
	},
	{
		"Windows Storage Server 2012 Standard",
		"win2012",
	},
	{
		"Windows Storage Server 2012 R2 Standard",
		"win2012r2",
	},
	{
		"Windows 7 Home",
		"win7",
	},
	{
		"Windows 8 Pro",
		"win8",
	},
	{
		"Windows 8.1 Pro",
		"win81",
	},
}

func (s *windowsSeriesSuite) SetUpTest(c *gc.C) {
	s.CleanupSuite.SetUpTest(c)
	s.createRegKey(c, series.CurrentVersionKey)
}

func (s *windowsSeriesSuite) createRegKey(c *gc.C, key *string) {
	salt, err := utils.RandomPassword()
	c.Assert(err, jc.ErrorIsNil)
	regKey := fmt.Sprintf(`SOFTWARE\JUJU\%s`, salt)
	s.PatchValue(key, regKey)

	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, *key, registry.ALL_ACCESS)
	c.Assert(err, jc.ErrorIsNil)

	err = k.Close()
	c.Assert(err, jc.ErrorIsNil)

	s.AddCleanup(func(*gc.C) {
		registry.DeleteKey(registry.LOCAL_MACHINE, *series.CurrentVersionKey)
	})
}

func (s *windowsSeriesSuite) TestReadSeries(c *gc.C) {
	for _, value := range versionTests {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, *series.CurrentVersionKey, registry.ALL_ACCESS)
		c.Assert(err, jc.ErrorIsNil)

		err = k.SetStringValue("ProductName", value.version)
		c.Assert(err, jc.ErrorIsNil)

		err = k.Close()
		c.Assert(err, jc.ErrorIsNil)

		ver, err := series.ReadSeries()
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(ver, gc.Equals, value.want)
	}
}

type windowsNanoSeriesSuite struct {
	windowsSeriesSuite
}

var _ = gc.Suite(&windowsNanoSeriesSuite{})

func (s *windowsNanoSeriesSuite) SetUpTest(c *gc.C) {
	s.windowsSeriesSuite.SetUpTest(c)
	s.createRegKey(c, series.IsNanoKey)

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, *series.IsNanoKey, registry.ALL_ACCESS)
	c.Assert(err, jc.ErrorIsNil)

	err = k.SetDWordValue("NanoServer", 1)
	c.Assert(err, jc.ErrorIsNil)

	err = k.Close()
	c.Assert(err, jc.ErrorIsNil)
}

var nanoVersionTests = []struct {
	version string
	want    string
}{{
	"Windows Server 2016",
	"win2016nano",
}}

func (s *windowsNanoSeriesSuite) TestReadSeries(c *gc.C) {
	for _, value := range nanoVersionTests {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, *series.CurrentVersionKey, registry.ALL_ACCESS)
		c.Assert(err, jc.ErrorIsNil)

		err = k.SetStringValue("ProductName", value.version)
		c.Assert(err, jc.ErrorIsNil)

		err = k.Close()
		c.Assert(err, jc.ErrorIsNil)

		ver, err := series.ReadSeries()
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(ver, gc.Equals, value.want)
	}
}
