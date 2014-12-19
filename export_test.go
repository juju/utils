// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"time"
)

var (
	GOMAXPROCS = &gomaxprocs
	NumCPU     = &numCPU
	Dial       = dial
	NetDial    = &netDial
	SleepFunc  = &sleepFunc
)

func GetAttemptDelay(attempt *Attempt) time.Duration {
	return attempt.delay
}
