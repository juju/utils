// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filepath

import (
	"runtime"
	"strings"

	"github.com/juju/errors"
)

// Renderer provides methods for the different functions in
// the stdlib path/filepath package that don't relate to a concrete
// filesystem. So Abs, EvalSymlinks, Glob, Rel, and Walk are not
// included. Also, while the functions in path/filepath relate to the
// current host, the PathRenderer methods relate to the renderer's
// target platform. So for example, a windows-oriented implementation
// will give windows-specific results even when used on linux.
type Renderer interface {
	// Base mimics path/filepath.
	Base(path string) string

	// Clean mimics path/filepath.
	Clean(path string) string

	// Dir mimics path/filepath.
	Dir(path string) string

	// Ext mimics path/filepath.
	Ext(path string) string

	// FromSlash mimics path/filepath.
	FromSlash(path string) string

	// IsAbs mimics path/filepath.
	IsAbs(path string) bool

	// Join mimics path/filepath.
	Join(path ...string) string

	// Match mimics path/filepath.
	Match(pattern, name string) (matched bool, err error)

	// Split mimics path/filepath.
	Split(path string) (dir, file string)

	// SplitList mimics path/filepath.
	SplitList(path string) []string

	// ToSlash mimics path/filepath.
	ToSlash(path string) string

	// VolumeName mimics path/filepath.
	VolumeName(path string) string
}

// NewRenderer returns a Renderer for the given os.
func NewRenderer(os string) (Renderer, error) {
	if os == "" {
		os = runtime.GOOS
	}

	switch strings.ToLower(os) {
	case "windows":
		return &WindowsRenderer{}, nil
	case "ubuntu":
		return &UnixRenderer{}, nil
	case "darwin", "dragonfly", "freebsd", "linux", "nacl", "netbsd", "openbsd", "solaris":
		// These match the the OS names from
		// http://golang.org/src/path/filepath/path_unix.go.
		return &UnixRenderer{}, nil
	default:
		return nil, errors.NotFoundf("renderer for %q", os)
	}
}
