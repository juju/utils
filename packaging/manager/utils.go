// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package manager

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/juju/loggo"

	"github.com/juju/utils"
)

var (
	logger = loggo.GetLogger("juju.utils.packaging.manager")

	attemptStrategy = utils.AttemptStrategy{
		Delay: 10 * time.Second,
		Min:   30,
	}
)

// CommandOutput is cmd.Output. It was aliased for testing purposes.
var commandOutput = (*exec.Cmd).CombinedOutput

// processStateSys is ps.Sys. It was aliased for testing purposes.
var processStateSys = (*os.ProcessState).Sys

// runCommand is utils.RunCommand. It was aliased for testing purposes.
var runCommand = utils.RunCommand

// exitStatuser is a mini-interface for the ExitStatus() method.
type exitStatuser interface {
	ExitStatus() int
}

// runCommandWithRetry is a helper function which tries to execute the given command.
// It tries to do so for 30 times with a 10 second sleep between commands.
// It returns the output of the command, the exit code, and an error, if one occurs,
// logging along the way.
func runCommandWithRetry(cmd string) (string, int, error) {
	var code int
	var err error
	var out []byte

	// split the command for use with exec
	args := strings.Fields(cmd)

	logger.Infof("Running: %s", cmd)

	// Retry oeration 30 times, sleeping every 10 seconds between attempts.
	// This avoids failure in the case of
	// something else having the dpkg lock (e.g. a charm on the
	// machine we're deploying containers to).
	for a := attemptStrategy.Start(); a.Next(); {
		// Create the command for each attempt, because we need to
		// call cmd.CombinedOutput only once. See http://pad.lv/1394524.
		cmd := exec.Command(args[0], args[1:]...)

		out, err = commandOutput(cmd)

		if err == nil {
			return string(out), 0, nil
		}

		exitError, ok := err.(*exec.ExitError)
		if !ok {
			err = errors.Annotatef(err, "unexpected error type %T", err)
			break
		}
		waitStatus, ok := processStateSys(exitError.ProcessState).(exitStatuser)
		if !ok {
			err = errors.Annotatef(err, "unexpected process state type %T", exitError.ProcessState.Sys())
			break
		}

		// Both apt-get and yum return 100 on abnormal execution
		code = waitStatus.ExitStatus()
		if code != 100 {
			break
		}

		logger.Infof("Retrying: %s", cmd)
	}

	if err != nil {
		logger.Errorf("packaging command failed: %v; cmd: %q; output: %s",
			err, cmd, string(out))
		return "", code, errors.Errorf("packaging command failed: %v", err)
	}

	return string(out), 0, nil
}
