// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"time"
)

var (
	GOMAXPROCS        = &gomaxprocs
	NumCPU            = &numCPU
	Dial              = dial
	NetDial           = &netDial
	ResolveSudoByFunc = resolveSudo
)

func NewTestBackoffTimer(info BackoffTimerInfo, mockAfterFunc func(d time.Duration, f func()) StoppableTimer) *BackoffTimer {
	timer := NewBackoffTimer(info)
	timer.afterFunc = mockAfterFunc
	return timer
}

func ExposeBackoffTimerDuration(bot *BackoffTimer) time.Duration {
	return bot.currentDuration
}
