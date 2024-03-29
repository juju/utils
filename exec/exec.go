// Copyright 2016 Canonical Ltd.
// Copyright 2016 Cloudbase Solutions
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/juju/clock"
	"github.com/juju/errors"
	"github.com/juju/loggo/v2"
)

var logger = loggo.GetLogger("juju.util.exec")

// Parameters for RunCommands.  Commands contains one or more commands to be
// executed using bash or PowerShell.  If WorkingDir is set, this is passed
// through.  Similarly if the Environment is specified, this is used
// for executing the command.
// TODO: refactor this to use a config struct and a constructor. Remove todo
// and extra code from WaitWithCancel once this is done.
type RunParams struct {
	Commands    string
	WorkingDir  string
	Environment []string
	Clock       clock.Clock
	KillProcess func(*os.Process) error
	User        string

	tempDir string
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	ps      *exec.Cmd
}

// ExecResponse contains the return code and output generated by executing a
// command.
type ExecResponse struct {
	Code   int
	Stdout []byte
	Stderr []byte
}

// mergeEnvironment takes in a string array representing the desired environment
// and merges it with the current environment. On Windows, clearing the environment,
// or having missing environment variables, may lead to standard go packages not working
// (os.TempDir relies on $env:TEMP), and powershell erroring out
// Currently this function is only used for windows
func mergeEnvironment(env []string) []string {
	if env == nil {
		return nil
	}
	m := make(map[string]string)
	var tmpEnv []string
	for _, val := range os.Environ() {
		varSplit := strings.SplitN(val, "=", 2)
		m[varSplit[0]] = varSplit[1]
	}

	for _, val := range env {
		varSplit := strings.SplitN(val, "=", 2)
		m[varSplit[0]] = varSplit[1]
	}

	for key, val := range m {
		tmpEnv = append(tmpEnv, key+"="+val)
	}

	return tmpEnv
}

// shellAndArgs returns the name of the shell command and arguments to run the
// specified script. shellAndArgs may write into the provided temporary
// directory, which will be maintained until the process exits.
func shellAndArgs(tempDir, script, user string) (string, []string, error) {
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
		if user == "" {
			cmd = "/bin/bash"
			args = []string{scriptFile}
		} else {
			// Need to make the tempDir readable by all so the user can see it.
			err := os.Chmod(tempDir, 0755)
			if err != nil {
				return "", nil, errors.Annotatef(err, "making tempdir readable by %q", user)
			}
			cmd = "/bin/su"
			args = []string{user, "--login", "--command", fmt.Sprintf("/bin/bash %s", scriptFile)}
		}
	}
	err := ioutil.WriteFile(scriptFile, []byte(script), 0644)
	if err != nil {
		return "", nil, err
	}
	return cmd, args, nil
}

// Run sets up the command environment (environment variables, working dir)
// and starts the process. The commands are passed into bash on Linux machines
// and to powershell on Windows machines.
func (r *RunParams) Run() error {
	if runtime.GOOS == "windows" {
		r.Environment = mergeEnvironment(r.Environment)
	}

	tempDir, err := ioutil.TempDir("", "juju-exec")
	if err != nil {
		return err
	}

	shell, args, err := shellAndArgs(tempDir, r.Commands, r.User)
	if err != nil {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Warningf("failed to remove temporary directory: %v", err)
		}
		return err
	}

	r.ps = exec.Command(shell, args...)
	if r.Environment != nil {
		r.ps.Env = r.Environment
	}
	if r.WorkingDir != "" {
		r.ps.Dir = r.WorkingDir
	}

	r.populateSysProcAttr()

	// If there is no user provided KillProcess function we
	// use the default one.
	if r.KillProcess == nil {
		r.KillProcess = KillProcess
	}

	r.tempDir = tempDir
	r.stdout = &bytes.Buffer{}
	r.stderr = &bytes.Buffer{}

	r.ps.Stdout = r.stdout
	r.ps.Stderr = r.stderr

	return r.ps.Start()
}

// Process returns the *os.Process instance of the current running process
// This will allow us to kill the process if needed, or get more information
// on the process
func (r *RunParams) Process() *os.Process {
	if r.ps != nil && r.ps.Process != nil {
		return r.ps.Process
	}
	return nil
}

// Wait blocks until the process exits, and returns an ExecResponse type
// containing stdout, stderr and the return code of the process. If a non-zero
// return code is returned, this is collected as the code for the response and
// this does not classify as an error.
func (r *RunParams) Wait() (*ExecResponse, error) {
	var err error
	if r.ps == nil {
		return nil, errors.New("No process has been started yet")
	}
	err = r.ps.Wait()
	if err := os.RemoveAll(r.tempDir); err != nil {
		logger.Warningf("failed to remove temporary directory: %v", err)
	}

	result := &ExecResponse{
		Stdout: r.stdout.Bytes(),
		Stderr: r.stderr.Bytes(),
	}

	if ee, ok := err.(*exec.ExitError); ok && err != nil {
		status := ee.ProcessState.Sys().(syscall.WaitStatus)
		if status.Exited() {
			// A non-zero return code isn't considered an error here.
			result.Code = status.ExitStatus()
			err = nil
		}
		logger.Infof("run result: %v", ee)
	}
	return result, err
}

// ErrCancelled is returned by WaitWithCancel in case it successfully manages to kill
// the running process.
var ErrCancelled = errors.New("command cancelled")

// timeWaitForKill reperesent the time we wait after attempting to kill a
// process before bailing out and returning.
const timeWaitForKill = 30 * time.Second

type resultWithError struct {
	execResult *ExecResponse
	err        error
}

// WaitWithCancel waits until the process exits or until a signal is sent on the
// cancel channel. In case a signal is sent it first tries to kill the process and
// return ErrCancelled. If it fails at killing the process it will return anyway
// and report the problematic PID.
func (r *RunParams) WaitWithCancel(cancel <-chan struct{}) (*ExecResponse, error) {
	// TODO: Remove this once we make Clock a required field
	_clock := r.Clock
	if _clock == nil {
		_clock = clock.WallClock
	}

	done := make(chan resultWithError, 1)
	go func() {
		defer close(done)
		waitResult, err := r.Wait()
		done <- resultWithError{waitResult, err}
	}()

	select {
	case resWithError := <-done:
		return resWithError.execResult, errors.Trace(resWithError.err)
	case <-cancel:
		logger.Debugf("attempting to kill process")
		err := r.KillProcess(r.ps.Process)
		if err != nil {
			logger.Debugf("kill returned: %s", err)
		}

		// After we issue a kill we expect the wait above to return within timeWaitForKill.
		// In case it doesn't we just go on and assume the process is stuck, but we don't block
		select {
		case resWithError := <-done:
			return resWithError.execResult, ErrCancelled
		case <-_clock.After(timeWaitForKill):
			return nil, errors.Errorf("tried to kill process %v, but timed out", r.ps.Process.Pid)
		}
	}
}

// RunCommands executes the Commands specified in the RunParams using
// powershell on windows, and '/bin/bash -s' on everything else,
// passing the commands through as stdin, and collecting
// stdout and stderr.  If a non-zero return code is returned, this is
// collected as the code for the response and this does not classify as an
// error.
func RunCommands(run RunParams) (*ExecResponse, error) {
	err := run.Run()
	if err != nil {
		return nil, err
	}
	return run.Wait()
}
