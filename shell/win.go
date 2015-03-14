// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
	"os"
	"strings"

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
