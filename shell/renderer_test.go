// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package shell_test

import (
	"runtime"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/shell"
)

type rendererSuite struct {
	testing.IsolationSuite

	unix    *shell.BashRenderer
	windows *shell.PowershellRenderer
}

var _ = gc.Suite(&rendererSuite{})

func (s *rendererSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.unix = &shell.BashRenderer{}
	s.windows = &shell.PowershellRenderer{}
}

func (s rendererSuite) checkRenderer(c *gc.C, renderer shell.Renderer, expected string) {
	switch expected {
	case "powershell":
		c.Check(renderer, gc.FitsTypeOf, s.windows)
	case "bash":
		c.Check(renderer, gc.FitsTypeOf, s.unix)
	default:
		c.Errorf("unknown kind %q", expected)
	}
}

func (s rendererSuite) TestNewRendererDefault(c *gc.C) {
	// All possible values of runtime.GOOS should be supported.
	renderer, err := shell.NewRenderer("")
	c.Assert(err, jc.ErrorIsNil)

	switch runtime.GOOS {
	case "windows":
		s.checkRenderer(c, renderer, "powershell")
	default:
		s.checkRenderer(c, renderer, "bash")
	}
}

func (s rendererSuite) TestNewRendererGOOS(c *gc.C) {
	// All possible values of runtime.GOOS should be supported.
	renderer, err := shell.NewRenderer(runtime.GOOS)
	c.Assert(err, jc.ErrorIsNil)

	switch runtime.GOOS {
	case "windows":
		s.checkRenderer(c, renderer, "powershell")
	default:
		s.checkRenderer(c, renderer, "bash")
	}
}

func (s rendererSuite) TestNewRendererWindows(c *gc.C) {
	renderer, err := shell.NewRenderer("windows")
	c.Assert(err, jc.ErrorIsNil)

	s.checkRenderer(c, renderer, "powershell")
}

func (s rendererSuite) TestNewRendererUnix(c *gc.C) {
	for _, os := range utils.OSUnix {
		c.Logf("trying %q", os)
		renderer, err := shell.NewRenderer(os)
		c.Assert(err, jc.ErrorIsNil)

		s.checkRenderer(c, renderer, "bash")
	}
}

func (s rendererSuite) TestNewRendererDistros(c *gc.C) {
	distros := []string{"ubuntu"}
	for _, distro := range distros {
		c.Logf("trying %q", distro)
		renderer, err := shell.NewRenderer(distro)
		c.Assert(err, jc.ErrorIsNil)

		s.checkRenderer(c, renderer, "bash")
	}
}

func (s rendererSuite) TestNewRendererUnknown(c *gc.C) {
	_, err := shell.NewRenderer("<unknown OS>")

	c.Check(err, jc.Satisfies, errors.IsNotFound)
}
