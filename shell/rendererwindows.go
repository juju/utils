// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
	"fmt"
	"os"

	"github.com/juju/utils/filepath"
)

// WindowsRenderer represents a Windows specific script render
// type that is responsible for this particular OS. It implements
// the Renderer interface
type WindowsRenderer struct {
	filepath.WindowsRenderer
}

// ExeSuffix implements Renderer.
func (w *WindowsRenderer) ExeSuffix() string {
	return ".exe"
}

// Quote implements Renderer.
func (w *WindowsRenderer) Quote(str string) string {
	return `"` + str + `"`
}

// Chmod implements Renderer.
func (w *WindowsRenderer) Chmod(path string, perm os.FileMode) []string {
	// TODO(ericsnow) Use cacls?
	panic("not supported")
	return nil
}

// WriteFile implements Renderer.
func (w *WindowsRenderer) WriteFile(filename string, data []byte) []string {
	return []string{
		fmt.Sprintf("Set-Content '%s' @\"\n%s\n\"@", filename, data),
	}
}

// MkDir implements Renderer.
func (w *WindowsRenderer) Mkdir(dirname string) []string {
	return []string{fmt.Sprintf(`mkdir %s`, w.FromSlash(dirname))}
}

// MkDirAll implements Renderer.
func (w *WindowsRenderer) MkdirAll(dirname string) []string {
	// TODO(ericsnow) Wrap in "setlocal enableextensions...endlocal"?
	return []string{fmt.Sprintf(`mkdir %s`, w.FromSlash(dirname))}
}
