// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"
)

var (
	_ Operations         = (*Ops)(nil)
	_ Operations         = (*CachedOps)(nil)
	_ Operations         = (*SimpleOps)(nil)
	_ Operations         = (*StubOps)(nil)
	_ os.FileInfo        = (*fileInfo)(nil)
	_ Node               = (*NodeInfo)(nil)
	_ io.ReadWriteCloser = (*FileData)(nil)
	_ io.ReadWriteCloser = (*StubFile)(nil)
)
