// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// On-disk mutex protecting a resource
//
// A lock is represented on disk by a directory of a particular name,
// containing an information file.  Taking a lock is done by renaming a
// temporary directory into place.  We use temporary directories because for
// all filesystems we believe that exactly one attempt to claim the lock will
// succeed and the others will fail.
package fslock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"time"

	"github.com/juju/errors"
	"github.com/juju/loggo"

	"github.com/juju/utils"
	"github.com/juju/utils/clock"
)

const (
	// NameRegexp specifies the regular expression used to identify valid lock names.
	NameRegexp      = "^[a-z]+[a-z0-9.-]*$"
	heldFilename    = "held"
	messageFilename = "message"
)

var (
	logger = loggo.GetLogger("juju.utils.fslock")

	ErrLockNotHeld = errors.New("lock not held")
	ErrTimeout     = errors.New("lock timeout exceeded")

	validName = regexp.MustCompile(NameRegexp)

	LockWaitDelay = 1 * time.Second
)

type Lock struct {
	name   string
	parent string
	nonce  []byte
	clock  clock.Clock
}

// NewLockWithClock is like NewLock but uses the given clock
// to enable faking time passing.
func NewLockWithClock(lockDir, name string, clock clock.Clock) (*Lock, error) {
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("Invalid lock name %q.  Names must match %q", name, NameRegexp)
	}
	nonce, err := utils.NewUUID()
	if err != nil {
		return nil, err
	}
	lock := &Lock{
		name:   name,
		parent: lockDir,
		nonce:  nonce[:],
		clock:  clock,
	}
	// Ensure the parent exists.
	if err := os.MkdirAll(lock.parent, 0755); err != nil {
		return nil, err
	}
	return lock, nil
}

// NewLock returns a new lock with the given name within the given lock
// directory, without acquiring it. The lock name must match the regular
// expression defined by NameRegexp.
func NewLock(lockDir, name string) (*Lock, error) {
	return NewLockWithClock(lockDir, name, clock.WallClock)
}

func (lock *Lock) lockDir() string {
	return path.Join(lock.parent, lock.name)
}

func (lock *Lock) heldFile() string {
	return path.Join(lock.lockDir(), "held")
}

func (lock *Lock) messageFile() string {
	return path.Join(lock.lockDir(), "message")
}

// If message is set, it will write the message to the lock directory as the
// lock is taken.
func (lock *Lock) acquire(message string) (bool, error) {
	// If the lockDir exists, then the lock is held by someone else.
	_, err := os.Stat(lock.lockDir())
	if err == nil {
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	// Create a temporary directory (in the parent dir), and then move it to
	// the right name.  Using the same directory to make sure the directories
	// are on the same filesystem.  Use a directory name starting with "." as
	// it isn't a valid lock name.
	tempLockName := fmt.Sprintf(".%x", lock.nonce)
	tempDirName, err := ioutil.TempDir(lock.parent, tempLockName)
	if err != nil {
		return false, err // this shouldn't really fail...
	}
	// write nonce into the temp dir
	err = ioutil.WriteFile(path.Join(tempDirName, heldFilename), lock.nonce, 0755)
	if err != nil {
		return false, err
	}
	if message != "" {
		err = ioutil.WriteFile(path.Join(tempDirName, messageFilename), []byte(message), 0755)
		if err != nil {
			return false, err
		}
	}
	// Now move the temp directory to the lock directory.
	err = utils.ReplaceFile(tempDirName, lock.lockDir())
	if err != nil {
		// Any error on rename means we failed.
		// Beaten to it, clean up temporary directory.
		os.RemoveAll(tempDirName)
		return false, nil
	}
	// We now have the lock.
	return true, nil
}

// lockLoop tries to acquire the lock. If the acquisition fails, the
// continueFunc is run to see if the function should continue waiting.
func (lock *Lock) lockLoop(message string, continueFunc func() error) error {
	var heldMessage = ""
	for {
		acquired, err := lock.acquire(message)
		if err != nil {
			return err
		}
		if acquired {
			return nil
		}
		if err = continueFunc(); err != nil {
			return err
		}
		currMessage := lock.Message()
		if currMessage != heldMessage {
			logger.Infof("attempted lock failed %q, %s, currently held: %s", lock.name, message, currMessage)
			heldMessage = currMessage
		}
		<-lock.clock.After(LockWaitDelay)
	}
}

// clean reads the lock and checks that it is valid. If the lock points to a running
// juju process that is older than the lock file, the lock is left in place, else
// the lock is removed.
func (lock *Lock) clean() error {
	// If a lock exists, see if it is stale
	heldNonce, err := ioutil.ReadFile(lock.heldFile())
	if err != nil {
		// No lock or we can't read it, so nothing to do/that we can do
		logger.Tracef("No lock to clean")
		return nil
	}

	// There is a lock...
	var nonce Nonce
	err = bson.Unmarshal(heldNonce, &nonce)
	if err != nil {
		// The lock should contain a BSON encoded Nonce object. If we can't decode
		// it then we consider it garbage and just delete the lock.
		logger.Debugf("Can't decode lock %s (%s): %s", lock.name, lock.Message(), err)
		return lock.BreakLock()
	}

	var processStartTime time.Time

	if runtime.GOOS == "windows" {
		cmd := fmt.Sprintf("powershell \"'{0:O}' -f (Get-Process -Id %d).StartTime\"", nonce.PID)
		out, err := exec.Command(cmd).Output()
		if err != nil {
			logger.Debugf("Lock is stale (can't find process) %s (%s): %s", lock.name, lock.Message(), err)
			return lock.BreakLock()
		}
		processStartTime, err = time.Parse(time.RFC3339Nano, string(out))
		if err != nil {
			logger.Errorf("Unable to parse time string: %s", out)
		}
	} else {
		// Find if the lock points to a running process...
		procExeLink := fmt.Sprintf("/proc/%d/exe", nonce.PID)
		_, err = filepath.EvalSymlinks(procExeLink)
		if err != nil {
			// If we can't read the symlink, it can't be a Juju process started by
			// the same user (or something really bad is going on)
			logger.Debugf("Lock is stale (can't read exe symlink) %s (%s): %s", lock.name, lock.Message(), err)
			return lock.BreakLock()
		}

		// Lock is current and points to a running process
		procFileInfo, err := os.Lstat(procExeLink)
		if err != nil {
			logger.Debugf("Lock cleaner error -- can't os.Lstat(procExeLink) %s (%s): %s", lock.name, lock.Message(), err)
			return err
		}
		processStartTime = procFileInfo.ModTime()
	}

	lockFileInfo, err := os.Lstat(lock.heldFile())
	if err != nil {
		logger.Debugf("Lock cleaner error -- can't os.Lstat(lock.heldFile()) %s (%s): %s", lock.name, lock.Message(), err)
		return err
	}

	if processStartTime.After(lockFileInfo.ModTime().Add(time.Second)) {
		// If the process is newer than the lock, the lock is stale. The 1s fiddle is much more than is needed
		// to prevent errant test failures (on dooferlad's dev box 50ms is plenty). It is fine to have this much
		// margin for error though because this branch should only be taken when a PID has been recycled and that
		// only happens when all 32k (/proc/sys/kernel/pid_max) have been used or the machine reboots.
		logger.Debugf("Lock is stale (older then juju process) %s (%s)", lock.name, lock.Message())
		return lock.BreakLock()
	}

	logger.Tracef("Lock is current %s (%s)", lock.name, lock.Message())
	// lock is current. Do nothing.
	return nil
}

// Lock blocks until it is able to acquire the lock.  Since we are dealing
// with sharing and locking using the filesystem, it is good behaviour to
// provide a message that is saved with the lock.  This is output in debugging
// information, and can be queried by any other Lock dealing with the same
// lock name and lock directory.
func (lock *Lock) Lock(message string) error {
	// The continueFunc is effectively a no-op, causing continual looping
	// until the lock is acquired.
	continueFunc := func() error { return nil }
	return lock.lockLoop(message, continueFunc)
}

// LockWithTimeout tries to acquire the lock. If it cannot acquire the lock
// within the given duration, it returns ErrTimeout.  See `Lock` for
// information about the message.
func (lock *Lock) LockWithTimeout(duration time.Duration, message string) error {
	deadline := lock.clock.Now().Add(duration)
	continueFunc := func() error {
		if lock.clock.Now().After(deadline) {
			return ErrTimeout
		}
		return nil
	}
	return lock.lockLoop(message, continueFunc)
}

// LockWithFunc blocks until it is able to acquire the lock.  If the lock is failed to
// be acquired, the continueFunc is called prior to the sleeping.  If the
// continueFunc returns an error, that error is returned from LockWithFunc.
func (lock *Lock) LockWithFunc(message string, continueFunc func() error) error {
	return lock.lockLoop(message, continueFunc)
}

// IsLockHeld returns whether the lock is currently held by the receiver.
func (lock *Lock) IsLockHeld() bool {
	heldNonce, err := ioutil.ReadFile(lock.heldFile())
	if err != nil {
		return false
	}
	return bytes.Equal(heldNonce, lock.nonce)
}

// Unlock releases a held lock.  If the lock is not held ErrLockNotHeld is
// returned.
func (lock *Lock) Unlock() error {
	if !lock.IsLockHeld() {
		return ErrLockNotHeld
	}
	// To ensure reasonable unlocking, we should rename to a temp name, and delete that.
	tempLockName := fmt.Sprintf(".%s.%x", lock.name, lock.nonce)
	tempDirName := path.Join(lock.parent, tempLockName)
	// Now move the lock directory to the temp directory to release the lock.
	if err := utils.ReplaceFile(lock.lockDir(), tempDirName); err != nil {
		return err
	}
	// And now cleanup.
	return os.RemoveAll(tempDirName)
}

// IsLocked returns true if the lock is currently held by anyone.
func (lock *Lock) IsLocked() bool {
	_, err := os.Stat(lock.heldFile())
	return err == nil
}

// BreakLock forcably breaks the lock that is currently being held.
func (lock *Lock) BreakLock() error {
	return os.RemoveAll(lock.lockDir())
}

// Message returns the saved message, or the empty string if there is no
// saved message.
func (lock *Lock) Message() string {
	message, err := ioutil.ReadFile(lock.messageFile())
	if err != nil {
		return ""
	}
	return string(message)
}
