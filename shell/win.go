// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
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
