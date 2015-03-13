// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/shell"
)

type windowsSuite struct {
	testing.IsolationSuite

	dirname  string
	filename string
	renderer *shell.WindowsRenderer
}

var _ = gc.Suite(&windowsSuite{})

func (s *windowsSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.dirname = `C:\some\dir`
	s.filename = s.dirname + `\file`
	s.renderer = &shell.WindowsRenderer{}
}

func (s windowsSuite) TestExeSuffix(c *gc.C) {
	suffix := s.renderer.ExeSuffix()

	c.Check(suffix, gc.Equals, ".exe")
}

func (s windowsSuite) TestShQuote(c *gc.C) {
	quoted := s.renderer.Quote("abc")

	c.Check(quoted, gc.Equals, `"abc"`)
}

func (s windowsSuite) TestChmod(c *gc.C) {
	commands := s.renderer.Chmod(s.filename, 0644)

	c.Check(commands, jc.DeepEquals, []string{
		"",
	})
}

func (s windowsSuite) TestWriteFile(c *gc.C) {
	data := []byte("something\nhere\n")
	commands := s.renderer.WriteFile(s.filename, data)

	expected := `
Set-Content 'C:\some\dir\file' @"
something
here

"@`[1:]
	c.Check(commands, jc.DeepEquals, []string{
		expected,
	})
}

func (s windowsSuite) TestMkdir(c *gc.C) {
	commands := s.renderer.Mkdir(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir C:\some\dir`,
	})
}

func (s windowsSuite) TestMkdirAll(c *gc.C) {
	commands := s.renderer.MkdirAll(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir C:\some\dir`,
	})
}
