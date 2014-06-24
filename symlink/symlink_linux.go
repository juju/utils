package symlink

import (
	"os"
)

func Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func Readlink(link string) (string, error) {
	return os.Readlink(link)
}
