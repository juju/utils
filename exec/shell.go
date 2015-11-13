// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/juju/errors"
)

// Run runs the provided script against the command and returns
// its standard out.
func RunScript(cmd Command, script string) ([]byte, error) {
	var stdout, stderr bytes.Buffer

	err := cmd.SetStdio(Stdio{
		In:  strings.NewReader(script),
		Out: &stdout,
		Err: &stderr,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	if _, err := Run(cmd); err != nil {
		// TODO(ericsnow) Fold this into Output()?
		if stderr.Len() == 0 {
			return nil, errors.Trace(err)
		}
		return nil, errors.Annotate(err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// RunBashScript runs the bash script within the provided execution system.
func RunBashScript(exec Exec, script string) ([]byte, error) {
	cmd, err := BashCommand(exec)
	if err != nil {
		return nil, errors.Trace(err)
	}

	output, err := RunScript(cmd, script)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return output, nil
}

// BashCommand returns a Command for the bash shell to run in
// the given system.
func BashCommand(exec Exec) (Command, error) {
	cmd, err := NewCommand(exec, "/bin/bash")
	if err != nil {
		return nil, errors.Trace(err)
	}
	return cmd, nil
}

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
