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
	_ os.FileInfo        = (*File)(nil)
	_ io.ReadWriteCloser = (*FileData)(nil)
)
