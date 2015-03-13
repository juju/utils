// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/shell"
)

type bashSuite struct {
	testing.IsolationSuite

	dirname  string
	filename string
	renderer *shell.BashRenderer
}

var _ = gc.Suite(&bashSuite{})

func (s *bashSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.dirname = `/some/dir`
	s.filename = s.dirname + `/file`
	s.renderer = &shell.BashRenderer{}
}

func (s bashSuite) TestExeSuffix(c *gc.C) {
	suffix := s.renderer.ExeSuffix()

	c.Check(suffix, gc.Equals, "")
}

func (s bashSuite) TestShQuote(c *gc.C) {
	quoted := s.renderer.Quote("abc")

	c.Check(quoted, gc.Equals, `'abc'`)
}

func (s bashSuite) TestChmod(c *gc.C) {
	commands := s.renderer.Chmod(s.filename, 0644)

	c.Check(commands, jc.DeepEquals, []string{
		"chmod 0644 '/some/dir/file'",
	})
}

func (s bashSuite) TestWriteFile(c *gc.C) {
	data := []byte("something\nhere\n")
	commands := s.renderer.WriteFile(s.filename, data)

	expected := `cat > '/some/dir/file' << 'EOF'
something
here

EOF`
	c.Check(commands, jc.DeepEquals, []string{
		expected,
	})
}

func (s bashSuite) TestMkdir(c *gc.C) {
	commands := s.renderer.Mkdir(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir '/some/dir'`,
	})
}

func (s bashSuite) TestMkdirAll(c *gc.C) {
	commands := s.renderer.MkdirAll(s.dirname)

	c.Check(commands, jc.DeepEquals, []string{
		`mkdir -p '/some/dir'`,
	})
}
