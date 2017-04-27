// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh

import (
	"sync/atomic"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
)

var (
	ReadAuthorisedKeys  = readAuthorisedKeys
	WriteAuthorisedKeys = writeAuthorisedKeys
	InitDefaultClient   = initDefaultClient
	DefaultIdentities   = &defaultIdentities
	SSHDial             = &sshDial
	RSAGenerateKey      = &rsaGenerateKey
	TestCopyReader      = copyReader
	TestNewCmd          = newCmd
)

type ReadLineWriter readLineWriter

func PatchTerminal(s *testing.CleanupSuite, rlw ReadLineWriter) {
	var balance int64
	s.PatchValue(&getTerminal, func() (readLineWriter, func(), error) {
		atomic.AddInt64(&balance, 1)
		cleanup := func() {
			atomic.AddInt64(&balance, -1)
		}
		return rlw, cleanup, nil
	})
	s.AddCleanup(func(c *gc.C) {
		c.Assert(atomic.LoadInt64(&balance), gc.Equals, int64(0))
	})
}
