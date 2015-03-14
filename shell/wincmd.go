// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package shell

import (
	"fmt"
	"os"

	"github.com/juju/utils"
)

// WinCmdRenderer is a shell renderer for Windows cmd.exe.
type WinCmdRenderer struct {
	windowsRenderer
}

// Quote implements Renderer.
func (wcr *WinCmdRenderer) Quote(str string) string {
	return utils.WinCmdQuote(str)
}

// Chmod implements Renderer.
func (wcr *WinCmdRenderer) Chmod(path string, perm os.FileMode) []string {
	path = wcr.Quote(path)
	// TODO(ericsnow) Use cacls?
	panic("not supported")
	return nil
}

// WriteFile implements Renderer.
func (wcr *WinCmdRenderer) WriteFile(filename string, data []byte) []string {
	filename = wcr.Quote(filename)
	// TODO(ericsnow) Use echo?
	panic("not supported")
	return nil
}

// MkDir implements Renderer.
func (wcr *WinCmdRenderer) Mkdir(dirname string) []string {
	dirname = wcr.Quote(dirname)
	return []string{
		fmt.Sprintf(`mkdir %s`, wcr.FromSlash(dirname)),
	}
}

// MkDirAll implements Renderer.
func (wcr *WinCmdRenderer) MkdirAll(dirname string) []string {
	dirname = wcr.Quote(dirname)
	// TODO(ericsnow) Wrap in "setlocal enableextensions...endlocal"?
	return []string{
		fmt.Sprintf(`mkdir %s`, wcr.FromSlash(dirname)),
	}
}

// ScriptFilename implements ScriptWriter.
func (wcr *WinCmdRenderer) ScriptFilename(name, dirname string) string {
	return wcr.Join(dirname, name+".bat")
}
