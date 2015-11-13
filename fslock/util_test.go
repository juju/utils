// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fslock_test

import (
	"io/ioutil"
	"path"

	gc "gopkg.in/check.v1"
	goyaml "gopkg.in/yaml.v2"

	"github.com/juju/utils/fslock"
)

func changeLockfilePID(c *gc.C, lockFile string, PID int) {
	var l fslock.OnDisk
	heldLock, err := ioutil.ReadFile(lockFile)
	c.Assert(err, gc.IsNil)
	err = goyaml.Unmarshal(heldLock, &l)
	c.Assert(err, gc.IsNil)
	l.PID = PID
	heldLock, err = goyaml.Marshal(l)
	c.Assert(err, gc.IsNil)
	err = ioutil.WriteFile(lockFile, heldLock, 644)
	c.Assert(err, gc.IsNil)
}

func assertCanLock(c *gc.C, lock *fslock.Lock) {
	err := lock.Lock("")
	c.Assert(err, gc.IsNil)
	c.Assert(lock.IsLocked(), gc.Equals, true)
}

func newLockedLock(c *gc.C, cfg fslock.LockConfig) (lock *fslock.Lock, lockFile, aliveFile string) {
	dir := c.MkDir()
	lock, err := fslock.NewLock(dir, "testing", cfg)
	c.Assert(err, gc.IsNil)
	assertCanLock(c, lock)
	lockFile = path.Join(dir, "testing", "held")
	return lock, lockFile, dir
}
