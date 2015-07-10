// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package shell_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/shell"
)

var _ = gc.Suite(&winCmdSuite{})

type winCmdSuite struct {
	testing.IsolationSuite

	dirname  string
	filename string
	renderer *shell.WinCmdRenderer
}

func (s *winCmdSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.dirname = `C:\some\dir`
	s.filename = s.dirname + `\file`
	s.renderer = &shell.WinCmdRenderer{}
}

func (s winCmdSuite) TestExeSuffix(c *gc.C) {
	suffix := s.renderer.ExeSuffix()

	c.Check(suffix, gc.Equals, ".exe")
}

func (s winCmdSuite) TestShQuote(c *gc.C) {
	quoted := s.renderer.Quote("abc")

	c.Check(quoted, gc.Equals, `^"abc^"`)
}

func (s winCmdSuite) TestChmod(c *gc.C) {
	commands := s.renderer.Chmod(s.filename, 0644)

	c.Check(commands, gc.HasLen, 0)
}

func (s winCmdSuite) TestWriteFile(c *gc.C) {
	data := []byte("something\nhere\n")
	commands := s.renderer.WriteFile(s.filename, data)

	c.Check(commands, jc.DeepEquals, []string{
		`>>^"C:\\some\\dir\\file^" @echo something`,
		`>>^"C:\\some\\dir\\file^" @echo here`,
		`>>^"C:\\some\\dir\\file^" @echo `,
	})
}

func (s winCmdSuite) TestMkdir(c *gc.C) {
	commands := s.renderer.Mkdir(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir ^"C:\\some\\dir^"`,
	})
}

func (s winCmdSuite) TestMkdirAll(c *gc.C) {
	commands := s.renderer.MkdirAll(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir ^"C:\\some\\dir^"`,
	})
}
