// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// chmod is intended to hold a wrapper around chmod which should
// be used instead of os.Chmod or File.Chmod.
// the intention is to provide an unified API for file permissions
// that works in both windows and linux.
// Use of Chmod form either os or File will provoke your code to
// panic or misvehave in windows.

// +build !windows

package fs

import (
	"os"
)

// Chmod is a straight alias to os.Chmod
var Chmod = os.Chmod
