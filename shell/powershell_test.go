// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package shell_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/shell"
)

var _ = gc.Suite(&powershellSuite{})

type powershellSuite struct {
	testing.IsolationSuite

	dirname  string
	filename string
	renderer *shell.PowershellRenderer
}

func (s *powershellSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.dirname = `C:\some\dir`
	s.filename = s.dirname + `\file`
	s.renderer = &shell.PowershellRenderer{}
}

func (s powershellSuite) TestExeSuffix(c *gc.C) {
	suffix := s.renderer.ExeSuffix()

	c.Check(suffix, gc.Equals, ".exe")
}

func (s powershellSuite) TestShQuote(c *gc.C) {
	quoted := s.renderer.Quote("abc")

	c.Check(quoted, gc.Equals, `'abc'`)
}

func (s powershellSuite) TestChmod(c *gc.C) {
	commands := s.renderer.Chmod(s.filename, 0644)

	c.Check(commands, gc.HasLen, 0)
}

func (s powershellSuite) TestWriteFile(c *gc.C) {
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

func (s powershellSuite) TestMkdir(c *gc.C) {
	commands := s.renderer.Mkdir(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir 'C:\some\dir'`,
	})
}

func (s powershellSuite) TestMkdirAll(c *gc.C) {
	commands := s.renderer.MkdirAll(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir 'C:\some\dir'`,
	})
}
