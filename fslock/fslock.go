// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// Package fslock provides an on-disk mutex protecting a resource
//
// A lock is represented on disk by a directory of a particular name,
// containing an information file.  Taking a lock is done by renaming a
// temporary directory into place.  We use temporary directories because for
// all filesystems we believe that exactly one attempt to claim the lock will
// succeed and the others will fail.
package fslock

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/juju/loggo"

	"github.com/juju/utils"
	"github.com/juju/utils/clock"
	goyaml "gopkg.in/yaml.v2"
)

const (
	// NameRegexp specifies the regular expression used to identify valid lock names.
	NameRegexp   = "^[a-z]+[a-z0-9.-]*$"
	heldFilename = "held"
)

var (
	logger = loggo.GetLogger("juju.utils.fslock")

	// ErrLockNotHeld is returned by Unlock if the lock file is not held by this lock
	ErrLockNotHeld = errors.New("lock not held")
	// ErrTimeout is returned by LockWithTimeout if the lock could not be obtained before the given deadline
	ErrTimeout = errors.New("lock timeout exceeded")

	validName = regexp.MustCompile(NameRegexp)
)

// LockConfig defines the configuration of the new lock. Sensible defaults can be
// obtained from Defaults().
type LockConfig struct {
	// Clock is used to generate delays
	Clock clock.Clock
	// WaitDelay is how long to wait after trying to aquire a lock before trying again
	WaitDelay time.Duration
	// LividityTimeout is how old a lock can be without us considering its
	// parent process dead.
	LividityTimeout time.Duration
	// ReadRetryTimeout is how long to wait after trying to examine a lock
	// and not finding it before trying again.
	ReadRetryTimeout time.Duration
}

// Defaults generates a LockConfig pre-filled with sensible defaults.
func Defaults() LockConfig {
	return LockConfig{
		Clock:            clock.WallClock,
		WaitDelay:        1 * time.Second,
		LividityTimeout:  30 * time.Second,
		ReadRetryTimeout: time.Millisecond * 10,
	}
}

// Lock is a file system lock
type Lock struct {
	name                   string
	parent                 string
	clock                  clock.Clock
	nonce                  string
	PID                    int
	stopWritingAliveFile   chan struct{}
	createAliveFileRunning sync.WaitGroup
	waitDelay              time.Duration
	lividityTimeout        time.Duration
	readRetryTimeout       time.Duration
	sanityCheck            chan struct{}
}

type onDisk struct {
	Nonce   string
	PID     int
	Message string
}

// NewLock returns a new lock with the given name within the given lock
// directory, without acquiring it. The lock name must match the regular
// expression defined by NameRegexp.
func NewLock(lockDir, name string, cfg LockConfig) (*Lock, error) {
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("Invalid lock name %q.  Names must match %q", name, NameRegexp)
	}
	uuid, err := utils.NewUUID()
	if err != nil {
		return nil, err
	}
	lock := &Lock{
		name:                 name,
		parent:               lockDir,
		clock:                cfg.Clock,
		nonce:                uuid.String(),
		PID:                  os.Getpid(),
		stopWritingAliveFile: make(chan struct{}, 1),
		waitDelay:            cfg.WaitDelay,
		lividityTimeout:      cfg.LividityTimeout,
		readRetryTimeout:     cfg.ReadRetryTimeout,
		sanityCheck:          make(chan struct{}),
	}
	// Ensure the parent exists.
	if err := os.MkdirAll(lock.parent, 0755); err != nil {
		return nil, err
	}
	// Ensure that an old alive file doesn't exist. RemoveAll doesn't raise
	// an error if the target doesn't exist, so we don't expect any errors.
	if err := os.RemoveAll(lock.aliveFile(lock.PID)); err != nil {
		return nil, err
	}

	return lock, nil
}

func (lock *Lock) lockDir() string {
	return path.Join(lock.parent, lock.name)
}

func (lock *Lock) heldFile() string {
	return path.Join(lock.lockDir(), "held")
}

func (lock *Lock) aliveFile(PID int) string {
	return path.Join(lock.lockDir(), fmt.Sprintf("alive.%d", PID))
}

// isAlive checks that the PID given is alive by looking to see if it is the
// current process's PID or, if it isn't, for a file named alive.<PID>, which
// has been updated in the last 30 seconds.
func (lock *Lock) isAlive(PID int) bool {
	if PID == lock.PID {
		return true
	}
	for i := 0; i < 10; i++ {
		aliveInfo, err := os.Lstat(lock.aliveFile(PID))
		if err == nil {
			return time.Now().Before(aliveInfo.ModTime().Add(lock.lividityTimeout))
		}
		time.Sleep(lock.readRetryTimeout)
	}
	return false
}

// createAliveFile kicks off a gorouteine that creates a proof of life file
// and keeps its timestamp current.
func (lock *Lock) createAliveFile() {
	lock.createAliveFileRunning.Add(1)
	close(lock.sanityCheck)
	go func() {
		defer lock.createAliveFileRunning.Done()

		aliveFile := lock.aliveFile(lock.PID)
		if err := ioutil.WriteFile(aliveFile, []byte{}, 644); err != nil {
			return
		}

		for {
			select {
			case <-time.After(5 * lock.waitDelay):
				now := time.Now()
				if err := os.Chtimes(aliveFile, now, now); err != nil {
					return
				}
			case <-lock.stopWritingAliveFile:
				return
			}
		}
	}()
}

func (lock *Lock) declareDead() {
	select {
	case lock.stopWritingAliveFile <- struct{}{}:
	default:
	}
	lock.createAliveFileRunning.Wait()
	lock.sanityCheck = make(chan struct{}) // refresh sanity check
}

// clean reads the lock and checks that it is valid. If the lock points to a running
// juju process that is older than the lock file, the lock is left in place, else
// the lock is removed.
func (lock *Lock) clean() error {
	// If a lock exists, see if it is stale
	lockInfo, err := lock.readLock()
	if err != nil {
		return nil
	}

	if lock.isAlive(lockInfo.PID) {
		// lock is current. Do nothing.
		logger.Debugf("Lock alive")
		return nil
	}

	logger.Debugf("Lock dead")
	return lock.BreakLock()
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
	tempLockName := fmt.Sprintf(".%s", lock.nonce)
	tempDirName, err := ioutil.TempDir(lock.parent, tempLockName)
	if err != nil {
		return false, err // this shouldn't really fail...
	}

	// write lock into the temp dir
	l := onDisk{
		PID:     lock.PID,
		Nonce:   lock.nonce,
		Message: message,
	}
	lockInfo, err := goyaml.Marshal(&l)
	if err != nil {
		return false, err // this shouldn't fail either...
	}
	err = ioutil.WriteFile(path.Join(tempDirName, heldFilename), lockInfo, 0664)
	if err != nil {
		return false, err
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
	lock.createAliveFile()
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
		<-lock.clock.After(lock.waitDelay)
	}
}

// Lock blocks until it is able to acquire the lock.  Since we are dealing
// with sharing and locking using the filesystem, it is good behaviour to
// provide a message that is saved with the lock.  This is output in debugging
// information, and can be queried by any other Lock dealing with the same
// lock name and lock directory.
func (lock *Lock) Lock(message string) error {
	lock.clean()
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

func (lock *Lock) readLock() (lockInfo onDisk, err error) {
	lockFile, err := ioutil.ReadFile(lock.heldFile())
	if err != nil {
		return lockInfo, err
	}

	err = goyaml.Unmarshal(lockFile, &lockInfo)
	return lockInfo, err
}

// IsLockHeld returns whether the lock is currently held by the receiver.
func (lock *Lock) IsLockHeld() bool {
	lockInfo, err := lock.readLock()
	if err != nil {
		return false
	}
	return lockInfo.Nonce == lock.nonce
}

// Unlock releases a held lock.  If the lock is not held ErrLockNotHeld is
// returned.
func (lock *Lock) Unlock() error {
	if !lock.IsLockHeld() {
		return ErrLockNotHeld
	}
	// To ensure reasonable unlocking, we should rename to a temp name, and delete that.
	lock.declareDead()
	tempLockName := fmt.Sprintf(".%s.%s", lock.name, lock.nonce)
	tempDirName := path.Join(lock.parent, tempLockName)
	// Now move the lock directory to the temp directory to release the lock.
	for i := 0; ; i++ {
		err := utils.ReplaceFile(lock.lockDir(), tempDirName)
		if err == nil {
			break
		}
		if i == 100 {
			logger.Debugf("Failed to replace lock, giving up: (%s)", err)
			return err
		}
		logger.Debugf("Failed to replace lock, retrying: (%s)", err)
		runtime.Gosched()
	}
	// And now cleanup.
	if err := os.RemoveAll(tempDirName); err != nil {
		logger.Debugf("Failed to remove lock: %s", err)
		return err
	}
	return nil
}

// IsLocked returns true if the lock is currently held by anyone.
func (lock *Lock) IsLocked() bool {
	_, err := os.Stat(lock.heldFile())
	return err == nil
}

// BreakLock forcibly breaks the lock that is currently being held.
func (lock *Lock) BreakLock() error {
	lock.declareDead()
	return os.RemoveAll(lock.lockDir())
}

// Message returns the saved message, or the empty string if there is no
// saved message.
func (lock *Lock) Message() string {
	lockInfo, err := lock.readLock()
	if err != nil {
		return ""
	}
	return lockInfo.Message
}
