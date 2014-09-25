// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs_test

import (
	"path/filepath"
	"testing"

	ft "github.com/juju/testing/filetesting"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/fs"
)

type copySuite struct{}

var _ = gc.Suite(&copySuite{})

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

var copyTests = []struct {
	about string
	src   ft.Entries
	dst   ft.Entries
	err   string
}{{
	about: "one file",
	src: []ft.Entry{
		ft.File{"file", "data", 0756},
	},
}, {
	about: "one directory",
	src: []ft.Entry{
		ft.Dir{"dir", 0777},
	},
}, {
	about: "one symlink",
	src: []ft.Entry{
		ft.Symlink{"link", "/foo"},
	},
}, {
	about: "several entries",
	src: []ft.Entry{
		ft.Dir{"top", 0755},
		ft.File{"top/foo", "foodata", 0644},
		ft.File{"top/bar", "bardata", 0633},
		ft.Dir{"top/next", 0721},
		ft.Symlink{"top/next/link", "../foo"},
		ft.File{"top/next/another", "anotherdata", 0644},
	},
}, {
	about: "destination already exists",
	src: []ft.Entry{
		ft.Dir{"dir", 0777},
	},
	dst: []ft.Entry{
		ft.Dir{"dir", 0777},
	},
	err: `will not overwrite ".+dir"`,
}, {
	about: "source with unwritable directory",
	src: []ft.Entry{
		ft.Dir{"dir", 0555},
	},
}}

func (*copySuite) TestCopy(c *gc.C) {
	for i, test := range copyTests {
		c.Logf("test %d: %v", i, test.about)
		src, dst := c.MkDir(), c.MkDir()
		test.src.Create(c, src)
		test.dst.Create(c, dst)
		path := test.src[0].GetPath()
		err := fs.Copy(
			filepath.Join(src, path),
			filepath.Join(dst, path),
		)
		if test.err != "" {
			c.Check(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, gc.IsNil)
			test.src.Check(c, dst)
		}
	}
}
