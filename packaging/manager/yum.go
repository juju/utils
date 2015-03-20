// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package manager

import (
	"fmt"
	"strings"

	"github.com/juju/utils/proxy"
)

// yum is the PackageManager implementations for rpm-based systems.
type yum struct {
	basePackageManager
}

// Search implements PackageManager.
func (yum *yum) Search(pack string) (bool, error) {
	_, code, err := runCommandWithRetry(yum.cmder.SearchCmd(pack))

	// yum list package returns 1 when it cannot find it
	if code == 1 {
		return false, nil
	}

	return false, err
}

// GetProxySettings implements PackageManager.
func (yum *yum) GetProxySettings() (proxy.Settings, error) {
	var res proxy.Settings
	args := []string{"bash", "-c", fmt.Sprintf("%q", yum.cmder.GetProxyCmd())}

	out, err := runCommand(args[0], args[1:]...)
	if err != nil {
		logger.Errorf("command failed: %v\nargs: %#v\n%s",
			err, args, string(out))
		return res, fmt.Errorf("command failed: %v", err)
	}

	for _, match := range strings.Split(out, "\n") {
		fields := strings.Split(match, "=")
		if strings.HasPrefix(fields[0], "https") {
			res.Https = strings.TrimSpace(fields[1])
		} else if strings.HasPrefix(fields[0], "http") {
			res.Http = strings.TrimSpace(fields[1])
		} else if strings.HasPrefix(fields[0], "ftp") {
			res.Ftp = strings.TrimSpace(fields[1])
		}
	}

	return res, nil
}
