// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filepath_test

import (
	"runtime"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filepath"
)

type filepathSuite struct {
	testing.IsolationSuite

	unix    *filepath.UnixRenderer
	windows *filepath.WindowsRenderer
}

var _ = gc.Suite(&filepathSuite{})

func (s *filepathSuite) SetupTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.unix = &filepath.UnixRenderer{}
	s.windows = &filepath.WindowsRenderer{}
}

func (s filepathSuite) checkRenderer(c *gc.C, renderer filepath.Renderer, expected string) {
	switch expected {
	case "windows":
		c.Check(renderer, gc.FitsTypeOf, s.windows)
	case "unix":
		c.Check(renderer, gc.FitsTypeOf, s.unix)
	default:
		c.Errorf("unknown kind %q", expected)
	}
}

func (s filepathSuite) TestNewRendererDefault(c *gc.C) {
	// All possible values of runtime.GOOS should be supported.
	renderer, err := filepath.NewRenderer("")
	c.Assert(err, jc.ErrorIsNil)

	switch runtime.GOOS {
	case "windows":
		s.checkRenderer(c, renderer, "windows")
	default:
		s.checkRenderer(c, renderer, "unix")
	}
}

func (s filepathSuite) TestNewRendererGOOS(c *gc.C) {
	// All possible values of runtime.GOOS should be supported.
	renderer, err := filepath.NewRenderer(runtime.GOOS)
	c.Assert(err, jc.ErrorIsNil)

	switch runtime.GOOS {
	case "windows":
		s.checkRenderer(c, renderer, "windows")
	default:
		s.checkRenderer(c, renderer, "unix")
	}
}

func (s filepathSuite) TestNewRendererWindows(c *gc.C) {
	renderer, err := filepath.NewRenderer("windows")
	c.Assert(err, jc.ErrorIsNil)

	s.checkRenderer(c, renderer, "windows")
}

func (s filepathSuite) TestNewRendererUnix(c *gc.C) {
	for _, os := range utils.OSUnix {
		c.Logf("trying %q", os)
		renderer, err := filepath.NewRenderer(os)
		c.Assert(err, jc.ErrorIsNil)

		s.checkRenderer(c, renderer, "unix")
	}
}

func (s filepathSuite) TestNewRendererDistros(c *gc.C) {
	distros := []string{"ubuntu"}
	for _, distro := range distros {
		c.Logf("trying %q", distro)
		renderer, err := filepath.NewRenderer(distro)
		c.Assert(err, jc.ErrorIsNil)

		s.checkRenderer(c, renderer, "unix")
	}
}

func (s filepathSuite) TestNewRendererUnknown(c *gc.C) {
	_, err := filepath.NewRenderer("<unknown OS>")

	c.Check(err, jc.Satisfies, errors.IsNotFound)
}
