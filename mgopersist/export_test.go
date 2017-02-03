// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgopersist

var (
	PutInitialAtTime = (*Session).putInitialAtTime
	PutAtTime        = (*Session).putAtTime
	GetAtTime        = (*Session).getAtTime
)
