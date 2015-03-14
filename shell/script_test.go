// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package shell_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/shell"
)

type scriptSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&scriptSuite{})

func (*scriptSuite) TestDumpFileOnErrorScriptOutput(c *gc.C) {
	script := shell.DumpFileOnErrorScript("a b c")
	c.Assert(script, gc.Equals, `
dump_file() {
    code=$?
    if [ $code -ne 0 -a -e 'a b c' ]; then
        cat 'a b c' >&2
    fi
    exit $code
}
trap dump_file EXIT
`[1:])
}

func (*scriptSuite) TestDumpFileOnErrorScript(c *gc.C) {
	tempdir := c.MkDir()
	filename := filepath.Join(tempdir, "log.txt")
	err := ioutil.WriteFile(filename, []byte("abc"), 0644)
	c.Assert(err, gc.IsNil)

	dumpScript := shell.DumpFileOnErrorScript(filename)
	c.Logf("%s", dumpScript)
	run := func(command string) (stdout, stderr string) {
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd := exec.Command("/bin/bash", "-s")
		cmd.Stdin = strings.NewReader(dumpScript + command)
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		cmd.Run()
		return stdoutBuf.String(), stderrBuf.String()
	}

	stdout, stderr := run("exit 0")
	c.Assert(stdout, gc.Equals, "")
	c.Assert(stderr, gc.Equals, "")

	stdout, stderr = run("exit 1")
	c.Assert(stdout, gc.Equals, "")
	c.Assert(stderr, gc.Equals, "abc")

	err = os.Remove(filename)
	c.Assert(err, gc.IsNil)
	stdout, stderr = run("exit 1")
	c.Assert(stdout, gc.Equals, "")
	c.Assert(stderr, gc.Equals, "")
}

func (*scriptSuite) TestWriteScriptUnix(c *gc.C) {
	renderer := &shell.BashRenderer{}
	script := `
exec a-command
exec another-command
`
	commands := shell.WriteScript(renderer, "spam", "/ham/eggs", strings.Split(script, "\n"))

	cmd := `
cat > '/ham/eggs/spam.sh' << 'EOF'
#!/usr/bin/env bash


exec a-command
exec another-command

EOF`[1:]
	c.Check(commands, jc.DeepEquals, []string{
		cmd,
		"chmod 0755 '/ham/eggs/spam.sh'",
	})
}

func (*scriptSuite) TestWriteScriptWindows(c *gc.C) {
	renderer := &shell.PowershellRenderer{}
	script := `
exec a-command
exec another-command
`
	commands := shell.WriteScript(renderer, "spam", `C:\ham\eggs`, strings.Split(script, "\n"))

	c.Check(commands, jc.DeepEquals, []string{
		`Set-Content 'C:\ham\eggs\spam.ps1' @"

exec a-command
exec another-command

"@`,
	})
}
