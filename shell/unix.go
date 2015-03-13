// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
	"fmt"
	"os"

	"github.com/juju/utils"
	"github.com/juju/utils/filepath"
)

// unixRenderer is the base shell renderer for "unix" shells.
type unixRenderer struct {
	filepath.UnixRenderer
}

// Quote implements Renderer.
func (unixRenderer) Quote(str string) string {
	// This *may* not be correct for *all* unix shells...
	return utils.ShQuote(str)
}

// ExeSuffix implements Renderer.
func (unixRenderer) ExeSuffix() string {
	return ""
}

// Mkdir implements Renderer.
func (lr unixRenderer) Mkdir(dirname string) []string {
	dirname = lr.Quote(dirname)
	return []string{
		fmt.Sprintf("mkdir %s", dirname),
	}
}

// MkdirAll implements Renderer.
func (lr unixRenderer) MkdirAll(dirname string) []string {
	dirname = lr.Quote(dirname)
	return []string{
		fmt.Sprintf("mkdir -p %s", dirname),
	}
}

// Chmod implements Renderer.
func (lr unixRenderer) Chmod(path string, perm os.FileMode) []string {
	path = lr.Quote(path)
	return []string{
		fmt.Sprintf("chmod %04o %s", perm, path),
	}
}

// WriteFile implements Renderer.
func (lr unixRenderer) WriteFile(filename string, data []byte) []string {
	filename = lr.Quote(filename)
	return []string{
		// An alternate approach would be to use printf.
		fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", filename, data),
	}
}
