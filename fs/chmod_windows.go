// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// chmod is intended to hold a wrapper around chmod which should
// be used instead of os.Chmod or File.Chmod.
// the intention is to provide an unified API for file permissions
// that works in both windows and linux.
// Use of Chmod form either os or File will provoke your code to
// panic or misvehave in windows.


package fs

import (
	"os"

	"github.com/juju/loggo"
	"github.com/juju/utils/exec"
)

var logger = loggo.GetLogger("juju.utils.fs")

const (
	// This is a list of permissions masks and bits for reference:
	// IRWXU 0700  mask for file owner permissions
	// IRUSR 0400  owner has read permission
	// IWUSR 0200  owner has write permission
	// IXUSR 0100  owner has execute permission
	// IRWXG 0070  mask for group permissions
	// IRGRP 0040  group has read permission
	// IWGRP 0020  group has write permission
	// IXGRP 0010  group has execute permission
	// IRWXO 0007  mask for permissions for others (not in group)
	// IROTH 0004  others have read permission
	// IWOTH 0002  others have write permission
	// IXOTH 0001  others have execute permission

	// Owner level permissions
	OwnerR			= 0400
	OwnerW			= 0200
	OwnerRW			= 0600

	// mixed permissions
	OwnerRWXGroupOthersRX	= 0755

	// all permissions
	AllRX			= 0555
	AllRW			= 0666
	AllRWX			= 0777
)

type acl []string

// applyACL applies all the ACL rules to the named file
func applyACL(name string, acls acl) error {
	for _, cacl := range acls {
		icaclsCommand := fmt.Sprintf("%s %s %s","icacls", name, cacl)
		icaclsRun := exec.RunParams{Command: icaclsCommand}
		result, err := exec.RunCommands(icaclsRun)
		if err != nil{
			logger.Errorf("failed to execute: %q", icaclsCommand)
			return errors.Trace(err)
		}
		logger.Debugf("changed permissions of %q to: %v", name, cacl)
	}

}

// Chmod for windows will convert, as best as possible, the mode passed
// to acls using icacls.exe in windows.
// os.Chmod does settings of file attributes that do not mean what we
// would expect when invoking unix chmod and File.Chmod is not implemented
// and it panics.
// This is explicitly not on general form because we want ot make acls
// for each concrete necessity to avoid having unnoticed mis assignements
// About ACLs (ref http://technet.microsoft.com/en-us/library/cc753525.aspx) :
// :r after first grant will make sure that the permissions
// replace old ones instead of appending.
// Permissions:
// F (full access)
// M (modify access)
// RX (read and execute access)
// R (read-only access)
// W (write-only access)
func Chmod(name string, mode os.FileMode) error {
	var acls acl
	switch mode {
	case OwnerR:
		acls = []string{`/grant:r "jujud":R`, }
	case OwnerW:
		acls = []string{`/grant:r "jujud":W`, }
	case OwnerRW:
		acls = []string{`/grant:r "jujud":M`, }
	case OwnerRWXGroupOthersRX:
		acls = []string{`/grant:r "jujud":F`, `/grant "everyone":RX`}
	case AllRX:
		acls = []string{`/grant:r "everyone":RX`, }
	case AllRW:
		acls = []string{`/grant:r "everyone":(R)`, }
	case AllRWX:
		acls = []string{`/grant:r "everyone":(R)`, }
	default:
		return errors.Errorf("permission %q is not supported in windows", mode)
	}
	return errors.Trace(applyACL(name, acls))
}
