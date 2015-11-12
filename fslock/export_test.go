// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fslock

type OnDisk onDisk

func IsAlive(lock *Lock, PID int) bool {
	return lock.isAlive(PID)
}

func DeclareDead(lock *Lock) {
	lock.declareDead()
}
