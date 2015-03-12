// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
	"fmt"
	"os"

	"github.com/juju/utils"
	"github.com/juju/utils/filepath"
)

// UnixRenderer represents an Ubuntu specific script render
// type that is responsible for this particular OS. It implements
// the Renderer interface
type UnixRenderer struct {
	filepath.UnixRenderer
}

// ShQuote implements Renderer.
func (UnixRenderer) ShQuote(str string) string {
	return utils.ShQuote(str)
}

// ExeSuffix implements Renderer.
func (UnixRenderer) ExeSuffix() string {
	return ""
}

// Mkdir implements Renderer.
func (lr UnixRenderer) Mkdir(dirname string) []string {
	dirname = lr.ShQuote(dirname)
	return []string{
		fmt.Sprintf("mkdir %s", dirname),
	}
}

// MkdirAll implements Renderer.
func (lr UnixRenderer) MkdirAll(dirname string) []string {
	dirname = lr.ShQuote(dirname)
	return []string{
		fmt.Sprintf("mkdir -p %s", dirname),
	}
}

// Chmod implements Renderer.
func (lr UnixRenderer) Chmod(path string, perm os.FileMode) []string {
	path = lr.ShQuote(path)
	return []string{
		fmt.Sprintf("chmod %04s %s", path, perm),
	}
}

// WriteFile implements Renderer.
func (lr UnixRenderer) WriteFile(filename string, data []byte) []string {
	filename = lr.ShQuote(filename)
	return []string{
		// An alternate approach would be to use printf.
		fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", filename, data),
	}
}
