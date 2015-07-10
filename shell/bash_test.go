// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package shell_test

import (
	"os"
	"time"

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

func (s bashSuite) TestChown(c *gc.C) {
	commands := s.renderer.Chown("/a/b/c", "x", "y")

	c.Check(commands, jc.DeepEquals, []string{
		"chown x:y '/a/b/c'",
	})
}

func (s bashSuite) TestTouchDefault(c *gc.C) {
	commands := s.renderer.Touch("/a/b/c", nil)

	c.Check(commands, jc.DeepEquals, []string{
		"touch '/a/b/c'",
	})
}

func (s bashSuite) TestTouchTimestamp(c *gc.C) {
	now := time.Date(2015, time.Month(3), 14, 12, 26, 38, 0, time.UTC)
	commands := s.renderer.Touch("/a/b/c", &now)

	c.Check(commands, jc.DeepEquals, []string{
		"touch -t 201503141226.38 '/a/b/c'",
	})
}

func (s bashSuite) TestRedirectFD(c *gc.C) {
	commands := s.renderer.RedirectFD("stdout", "stderr")

	c.Check(commands, jc.DeepEquals, []string{
		"exec 2>&1",
	})
}

func (s bashSuite) TestRedirectOutput(c *gc.C) {
	commands := s.renderer.RedirectOutput("/a/b/c")

	c.Check(commands, jc.DeepEquals, []string{
		"exec >> '/a/b/c'",
	})
}

func (s bashSuite) TestRedirectOutputReset(c *gc.C) {
	commands := s.renderer.RedirectOutputReset("/a/b/c")

	c.Check(commands, jc.DeepEquals, []string{
		"exec > '/a/b/c'",
	})
}

func (s bashSuite) TestScriptFilename(c *gc.C) {
	filename := s.renderer.ScriptFilename("spam", "/ham/eggs")

	c.Check(filename, gc.Equals, "/ham/eggs/spam.sh")
}

func (s bashSuite) TestScriptPermissions(c *gc.C) {
	perm := s.renderer.ScriptPermissions()

	c.Check(perm, gc.Equals, os.FileMode(0755))
}
