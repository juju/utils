// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"fmt"
	"runtime"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type IsUbuntuSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&IsUbuntuSuite{})

func (s *IsUbuntuSuite) patchLsbRelease(c *gc.C, name string) {
	var content string
	var execName string
	if runtime.GOOS != "windows" {
		content = fmt.Sprintf("#!/bin/bash --norc\n%s", name)
		execName = "lsb_release"
	} else {
		execName = "lsb_release.bat"
		content = fmt.Sprintf("@echo off\r\n%s", name)
	}
	patchExecutable(s, c.MkDir(), execName, content)
}

func (s *IsUbuntuSuite) TestIsUbuntu(c *gc.C) {
	s.patchLsbRelease(c, "echo Ubuntu")
	c.Assert(utils.IsUbuntu(), jc.IsTrue)
}

func (s *IsUbuntuSuite) TestIsNotUbuntu(c *gc.C) {
	s.patchLsbRelease(c, "echo Windows NT")
	c.Assert(utils.IsUbuntu(), jc.IsFalse)
}

func (s *IsUbuntuSuite) TestIsNotUbuntuLsbReleaseNotFound(c *gc.C) {
	if runtime.GOOS != "windows" {
		s.patchLsbRelease(c, "exit 127")
	}
	c.Assert(utils.IsUbuntu(), jc.IsFalse)
}
