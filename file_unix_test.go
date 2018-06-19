// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build !windows

package utils_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	gc "gopkg.in/check.v1"

	"github.com/juju/errors"
	"github.com/juju/utils"
)

type unixFileSuite struct {
}

var _ = gc.Suite(&unixFileSuite{})

func (s *unixFileSuite) TestEnsureBaseDir(c *gc.C) {
	c.Assert(utils.EnsureBaseDir(`/a`, `/b/c`), gc.Equals, `/a/b/c`)
	c.Assert(utils.EnsureBaseDir(`/`, `/b/c`), gc.Equals, `/b/c`)
	c.Assert(utils.EnsureBaseDir(``, `/b/c`), gc.Equals, `/b/c`)
}

func (s *unixFileSuite) TestFileOwner(c *gc.C) {
	username, err := utils.LocalUsername()
	c.Assert(err, gc.IsNil)

	path := filepath.Join(os.TempDir(), fmt.Sprintf("file-%d", time.Now().UnixNano()))
	_, err = os.Create(path)
	c.Assert(err, gc.IsNil)

	ok, err := utils.IsFileOwner(path, username)
	c.Assert(err, gc.IsNil)
	c.Assert(ok, gc.Equals, true)
}

func (s *unixFileSuite) TestFileOwnerUsingRoot(c *gc.C) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file-%d", time.Now().UnixNano()))
	_, err := os.Create(path)
	c.Assert(err, gc.IsNil)

	ok, err := utils.IsFileOwner(path, "root")
	c.Assert(err, gc.IsNil)
	c.Assert(ok, gc.Equals, false)
}

func (s *unixFileSuite) TestFileOwnerWithInvalidPath(c *gc.C) {
	username, err := utils.LocalUsername()
	c.Assert(err, gc.IsNil)

	path := filepath.Join(os.TempDir(), "file-bad")
	ok, err := utils.IsFileOwner(path, username)
	c.Assert(errors.Cause(err), gc.ErrorMatches, "stat .*: no such file or directory")
	c.Assert(ok, gc.Equals, false)
}

func (s *unixFileSuite) TestFileOwnerWithInvalidUsername(c *gc.C) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file-%d", time.Now().UnixNano()))
	_, err := os.Create(path)
	c.Assert(err, gc.IsNil)

	ok, err := utils.IsFileOwner(path, "invalid")
	c.Assert(errors.Cause(err), gc.ErrorMatches, "user: unknown user invalid")
	c.Assert(ok, gc.Equals, false)
}
