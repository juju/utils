// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

//go:build !windows
// +build !windows

package ssh

import (
	"io"
)

// WrapStdin returns the original stdin stream on nix platforms.
func WrapStdin(reader io.Reader) io.Reader {
	return reader
}
