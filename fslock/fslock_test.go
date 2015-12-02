// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fslock_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
	"launchpad.net/tomb"

	"github.com/juju/utils/clock"
	"github.com/juju/utils/fslock"
)

const (
	shortWait = 50 * time.Millisecond
	longWait  = 10 * time.Second
)

type fslockSuite struct {
	testing.CleanupSuite
	testing.IsolationSuite
	lockDelay  time.Duration
	lockConfig fslock.LockConfig
}

var _ = gc.Suite(&fslockSuite{})

type fastclock struct {
	c *gc.C
}

func (*fastclock) Now() time.Time {
	return time.Now()
}

func (f *fastclock) After(duration time.Duration) <-chan time.Time {
	f.c.Check(duration, gc.Equals, fslock.Defaults().WaitDelay)
	return time.After(time.Millisecond)
}

func (f *fastclock) AfterFunc(d time.Duration, af func()) clock.Timer {
	return time.AfterFunc(d, af)
}

func (s *fslockSuite) SetUpTest(c *gc.C) {
	s.lockConfig = fslock.Defaults()
	s.lockConfig.Clock = &fastclock{c}
}

// This test also happens to test that locks can get created when the parent
// lock directory doesn't exist.
func (s *fslockSuite) TestValidNamesLockDir(c *gc.C) {

	for _, name := range []string{
		"a",
		"longer",
		"longer-with.special-characters",
	} {
		dir := c.MkDir()
		_, err := fslock.NewLock(dir, name, s.lockConfig)
		c.Assert(err, gc.IsNil)
	}
}

func (s *fslockSuite) TestInvalidNames(c *gc.C) {

	for _, name := range []string{
		".start",
		"-start",
		"NoCapitals",
		"no+plus",
		"no/slash",
		"no\\backslash",
		"no$dollar",
		"no:colon",
	} {
		dir := c.MkDir()
		_, err := fslock.NewLock(dir, name, s.lockConfig)
		c.Assert(err, gc.ErrorMatches, "Invalid lock name .*")
	}
}

func (s *fslockSuite) TestNewLockWithExistingDir(c *gc.C) {
	dir := c.MkDir()
	err := os.MkdirAll(dir, 0755)
	c.Assert(err, gc.IsNil)
	_, err = fslock.NewLock(dir, "special", s.lockConfig)
	c.Assert(err, gc.IsNil)
}

func (s *fslockSuite) TestNewLockWithExistingFileInPlace(c *gc.C) {
	dir := c.MkDir()
	err := os.MkdirAll(dir, 0755)
	c.Assert(err, gc.IsNil)
	path := path.Join(dir, "locks")
	err = ioutil.WriteFile(path, []byte("foo"), 0644)
	c.Assert(err, gc.IsNil)
	_, err = fslock.NewLock(path, "special", s.lockConfig)
	c.Assert(err, gc.ErrorMatches, utils.MkdirFailErrRegexp)
}

func (s *fslockSuite) TestIsLockHeldBasics(c *gc.C) {
	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	c.Assert(lock.IsLockHeld(), gc.Equals, false)

	err = lock.Lock("")
	c.Assert(err, gc.IsNil)
	c.Assert(lock.IsLockHeld(), gc.Equals, true)

	err = lock.Unlock()
	c.Assert(err, gc.IsNil)
	c.Assert(lock.IsLockHeld(), gc.Equals, false)
}

func (s *fslockSuite) TestIsLockHeldTwoLocks(c *gc.C) {
	dir := c.MkDir()
	lock1, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	lock2, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock1.Lock("")
	c.Assert(err, gc.IsNil)
	c.Assert(lock2.IsLockHeld(), gc.Equals, false)
}

func (s *fslockSuite) TestLockBlocks(c *gc.C) {

	dir := c.MkDir()
	lock1, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	lock2, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	acquired := make(chan struct{})
	err = lock1.Lock("")
	c.Assert(err, gc.IsNil)

	go func() {
		lock2.Lock("")
		acquired <- struct{}{}
		close(acquired)
	}()

	// Waiting for something not to happen is inherently hard...
	select {
	case <-acquired:
		c.Fatalf("Unexpected lock acquisition")
	case <-time.After(shortWait):
		// all good
	}

	err = lock1.Unlock()
	c.Assert(err, gc.IsNil)

	select {
	case <-acquired:
		// all good
	case <-time.After(longWait):
		c.Fatalf("Expected lock acquisition")
	}

	c.Assert(lock2.IsLockHeld(), gc.Equals, true)
}

func (s *fslockSuite) TestLockWithTimeoutUnlocked(c *gc.C) {
	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock.LockWithTimeout(shortWait, "")
	c.Assert(err, gc.IsNil)
}

func (s *fslockSuite) TestLockWithTimeoutLocked(c *gc.C) {
	dir := c.MkDir()
	lock1, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	lock2, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock1.Lock("")
	c.Assert(err, gc.IsNil)

	err = lock2.LockWithTimeout(shortWait, "")
	c.Assert(err, gc.Equals, fslock.ErrTimeout)
}

func (s *fslockSuite) TestUnlock(c *gc.C) {
	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock.Unlock()
	c.Assert(err, gc.Equals, fslock.ErrLockNotHeld)
}

func (s *fslockSuite) TestIsLocked(c *gc.C) {
	dir := c.MkDir()
	lock1, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	lock2, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock1.Lock("")
	c.Assert(err, gc.IsNil)

	c.Assert(lock1.IsLocked(), gc.Equals, true)
	c.Assert(lock2.IsLocked(), gc.Equals, true)
}

func (s *fslockSuite) TestBreakLock(c *gc.C) {
	dir := c.MkDir()
	lock1, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	lock2, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock1.Lock("")
	c.Assert(err, gc.IsNil)

	err = lock2.BreakLock()
	c.Assert(err, gc.IsNil)
	c.Assert(lock2.IsLocked(), gc.Equals, false)

	// Normally locks are broken due to client crashes, not duration.
	err = lock1.Unlock()
	c.Assert(err, gc.Equals, fslock.ErrLockNotHeld)

	// Breaking a non-existant isn't an error
	err = lock2.BreakLock()
	c.Assert(err, gc.IsNil)
}

func (s *fslockSuite) TestMessage(c *gc.C) {
	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	c.Assert(lock.Message(), gc.Equals, "")

	err = lock.Lock("my message")
	c.Assert(err, gc.IsNil)
	c.Assert(lock.Message(), gc.Equals, "my message")

	// Unlocking removes the message.
	err = lock.Unlock()
	c.Assert(err, gc.IsNil)
	c.Assert(lock.Message(), gc.Equals, "")
}

func (s *fslockSuite) TestMessageAcrossLocks(c *gc.C) {
	dir := c.MkDir()
	lock1, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	lock2, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock1.Lock("very busy")
	c.Assert(err, gc.IsNil)
	c.Assert(lock2.Message(), gc.Equals, "very busy")
}

func (s *fslockSuite) TestInitialMessageWhenLocking(c *gc.C) {
	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)

	err = lock.Lock("initial message")
	c.Assert(err, gc.IsNil)
	c.Assert(lock.Message(), gc.Equals, "initial message")

	err = lock.Unlock()
	c.Assert(err, gc.IsNil)

	err = lock.LockWithTimeout(shortWait, "initial timeout message")
	c.Assert(err, gc.IsNil)
	c.Assert(lock.Message(), gc.Equals, "initial timeout message")
}

func (s *fslockSuite) TestStress(c *gc.C) {
	const lockAttempts = 200
	const concurrentLocks = 10

	var counter = new(int64)
	// Use atomics to update lockState to make sure the lock isn't held by
	// someone else. A value of 1 means locked, 0 means unlocked.
	var lockState = new(int32)
	var done = make(chan struct{})
	defer close(done)

	dir := c.MkDir()

	var stress = func(name string) {
		defer func() { done <- struct{}{} }()
		lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
		if err != nil {
			c.Errorf("Failed to create a new lock")
			return
		}
		for i := 0; i < lockAttempts; i++ {
			err = lock.Lock(name)
			c.Assert(err, gc.IsNil)
			state := atomic.AddInt32(lockState, 1)
			c.Assert(state, gc.Equals, int32(1))
			// Tell the go routine scheduler to give a slice to someone else
			// while we have this locked.
			runtime.Gosched()
			// need to decrement prior to unlock to avoid the race of someone
			// else grabbing the lock before we decrement the state.
			atomic.AddInt32(lockState, -1)
			err = lock.Unlock()
			c.Assert(err, gc.IsNil)
			// increment the general counter
			atomic.AddInt64(counter, 1)
		}
	}

	for i := 0; i < concurrentLocks; i++ {
		go stress(fmt.Sprintf("Lock %d", i))
	}
	for i := 0; i < concurrentLocks; i++ {
		<-done
	}
	c.Assert(*counter, gc.Equals, int64(lockAttempts*concurrentLocks))
}

func (s *fslockSuite) TestTomb(c *gc.C) {
	const timeToDie = 200 * time.Millisecond
	die := tomb.Tomb{}

	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", s.lockConfig)
	c.Assert(err, gc.IsNil)
	// Just use one lock, and try to lock it twice.
	err = lock.Lock("very busy")
	c.Assert(err, gc.IsNil)

	checkTomb := func() error {
		select {
		case <-die.Dying():
			return tomb.ErrDying
		default:
			// no-op to fall through to return.
		}
		return nil
	}

	go func() {
		time.Sleep(timeToDie)
		die.Killf("time to die")
	}()

	err = lock.LockWithFunc("won't happen", checkTomb)
	c.Assert(err, gc.Equals, tomb.ErrDying)
	c.Assert(lock.Message(), gc.Equals, "very busy")

}

func (s *fslockSuite) TestCleanStaleLock(c *gc.C) {
	lock, lockFile, dir := newLockedLock(c, s.lockConfig)
	c.Assert(fslock.IsAlive(lock, lock.PID), gc.Equals, true)
	c.Assert(fslock.IsAlive(lock, 1), gc.Equals, false)

	// Make a stale alive file, point the lock to it, then try to re-lock.
	PID := 1
	aliveFile := path.Join(dir, "testing", fmt.Sprintf("alive.%d", PID))
	ioutil.WriteFile(aliveFile, []byte{}, 644)
	oneHourAgo := time.Now().Add(-time.Hour)
	os.Chtimes(aliveFile, oneHourAgo, oneHourAgo)
	changeLockfilePID(c, lockFile, PID)
	assertCanLock(c, lock)
}

func (s *fslockSuite) TestCleanNoMatchingProcess(c *gc.C) {
	lock, lockFile, _ := newLockedLock(c, s.lockConfig)

	// Change the PID to a process that doesn't exist.
	changeLockfilePID(c, lockFile, 1)
	assertCanLock(c, lock)
}

// TestProofOfLife checks that the alive file doesn't get older than 500ms. Normally
// it can get older, but we crank up the refresh interval for testing.
func (s *fslockSuite) TestProofOfLife(c *gc.C) {
	s.lockConfig.WaitDelay = 20 * time.Millisecond
	lock, _, dir := newLockedLock(c, s.lockConfig)
	aliveFile := path.Join(dir, "testing", fmt.Sprintf("alive.%d", lock.PID))

	tests := 0
	for check := 0; check < 20; check++ {
		aliveInfo, err := os.Lstat(aliveFile)
		if err != nil {
			// Typically this is file not existing. Whatever the reason, just retry
			time.Sleep(50 * time.Millisecond)
			continue
		}

		c.Assert(time.Now().Sub(aliveInfo.ModTime()), jc.DurationLessThan, 500*time.Millisecond)
		tests++
		time.Sleep(50 * time.Millisecond)
	}

	// Make sure we actually spotted an alive file and checked its time.
	c.Assert(tests > 1, gc.Equals, true)
}
