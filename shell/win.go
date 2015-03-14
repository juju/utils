// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
	"os"
	"strings"
	"time"

	"github.com/juju/utils/filepath"
)

// windowsRenderer is the base implementation for Windows shells.
type windowsRenderer struct {
	filepath.WindowsRenderer
}

// ExeSuffix implements Renderer.
func (w *windowsRenderer) ExeSuffix() string {
	return ".exe"
}

// ScriptPermissions implements ScriptWriter.
func (w *windowsRenderer) ScriptPermissions() os.FileMode {
	return 0755
}

// Render implements ScriptWriter.
func (w *windowsRenderer) RenderScript(commands []string) []byte {
	return []byte(strings.Join(commands, "\n"))
}

// Chown implements Renderer.
func (w windowsRenderer) Chown(path, owner, group string) []string {
	// TODO(ericsnow) Use ???
	panic("not supported")
	return nil
}

// Touch implements Renderer.
func (w windowsRenderer) Touch(path string, timestamp *time.Time) []string {
	// TODO(ericsnow) Use ???
	panic("not supported")
	return nil
}

// RedirectFD implements Renderer.
func (w windowsRenderer) RedirectFD(dst, src string) []string {
	// TODO(ericsnow) Use ???
	panic("not supported")
	return nil
}

// RedirectOutput implements Renderer.
func (w windowsRenderer) RedirectOutput(filename string) []string {
	// TODO(ericsnow) Use ???
	panic("not supported")
	return nil
}
