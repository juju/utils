// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/juju/errors"
)

// shellAndArgs returns the name of the shell command and arguments to run the
// specified script. shellAndArgs may write into the provided temporary
// directory, which will be maintained until the process exits.
func shellAndArgs(tempDir, script string) (string, []string, error) {
	var scriptFile string
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		scriptFile = filepath.Join(tempDir, "script.ps1")
		cmd = "powershell.exe"
		args = []string{
			"-NoProfile",
			"-NonInteractive",
			"-ExecutionPolicy", "RemoteSigned",
			"-File", scriptFile,
		}
		// Exceptions don't result in a non-zero exit code by default
		// when using -File. The exit code of an explicit "exit" when
		// using -Command is ignored and results in an exit code of 1.
		// We use -File and trap exceptions to cover both.
		script = "trap {Write-Error $_; exit 1}\n" + script
	default:
		scriptFile = filepath.Join(tempDir, "script.sh")
		cmd = "/bin/bash"
		args = []string{scriptFile}
	}
	err := ioutil.WriteFile(scriptFile, []byte(script), 0600)
	if err != nil {
		return "", nil, errors.Trace(err)
	}
	return cmd, args, nil
}
