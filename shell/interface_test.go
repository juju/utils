// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package shell

var _ Renderer = (*BashRenderer)(nil)
var _ Renderer = (*PowershellRenderer)(nil)
var _ Renderer = (*WinCmdRenderer)(nil)
