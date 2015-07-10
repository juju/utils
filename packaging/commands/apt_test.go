// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package commands_test

import (
	"github.com/juju/utils/packaging/commands"
	"github.com/juju/utils/proxy"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(&AptSuite{})

type AptSuite struct {
	paccmder commands.PackageCommander
}

func (s *AptSuite) SetUpSuite(c *gc.C) {
	s.paccmder = commands.NewAptPackageCommander()
}

func (s *AptSuite) TestProxyConfigContentsEmpty(c *gc.C) {
	out := s.paccmder.ProxyConfigContents(proxy.Settings{})
	c.Assert(out, gc.Equals, "")
}

func (s *AptSuite) TestProxyConfigContentsPartial(c *gc.C) {
	sets := proxy.Settings{
		Http: "dat-proxy.zone:8080",
	}

	output := s.paccmder.ProxyConfigContents(sets)
	c.Assert(output, gc.Equals, "Acquire::http::Proxy \"dat-proxy.zone:8080\";")
}

func (s *AptSuite) TestProxyConfigContentsFull(c *gc.C) {
	sets := proxy.Settings{
		Http:  "dat-proxy.zone:8080",
		Https: "https://much-security.com",
		Ftp:   "gimme-files.zone",
	}
	expected := `Acquire::http::Proxy "dat-proxy.zone:8080";
Acquire::https::Proxy "https://much-security.com";
Acquire::ftp::Proxy "gimme-files.zone";`

	output := s.paccmder.ProxyConfigContents(sets)
	c.Assert(output, gc.Equals, expected)
}
