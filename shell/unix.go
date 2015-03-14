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
func (ur unixRenderer) Mkdir(dirname string) []string {
	dirname = ur.Quote(dirname)
	return []string{
		fmt.Sprintf("mkdir %s", dirname),
	}
}

// MkdirAll implements Renderer.
func (ur unixRenderer) MkdirAll(dirname string) []string {
	dirname = ur.Quote(dirname)
	return []string{
		fmt.Sprintf("mkdir -p %s", dirname),
	}
}

// Chmod implements Renderer.
func (ur unixRenderer) Chmod(path string, perm os.FileMode) []string {
	path = ur.Quote(path)
	return []string{
		fmt.Sprintf("chmod %04o %s", perm, path),
	}
}

// WriteFile implements Renderer.
func (ur unixRenderer) WriteFile(filename string, data []byte) []string {
	filename = ur.Quote(filename)
	return []string{
		// An alternate approach would be to use printf.
		fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", filename, data),
	}
}

// ScriptFilename implements ScriptWriter.
func (ur *unixRenderer) ScriptFilename(name, dirname string) string {
	return ur.Join(dirname, name+".sh")
}

// ScriptPermissions implements ScriptWriter.
func (ur *unixRenderer) ScriptPermissions() os.FileMode {
	return 0755
}
