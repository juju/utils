// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package manager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/utils/proxy"
)

// apt is the PackageManager implementation for deb-based systems.
type apt struct {
	basePackageManager
}

// Search implements PackageManager.
func (apt *apt) Search(pack string) (bool, error) {
	out, _, err := runCommandWithRetry(apt.cmder.SearchCmd(pack))
	if err != nil {
		return false, err
	}

	// apt-cache search --names-only package returns no output
	// if the search was unsuccesfull
	if out == "" {
		return false, nil
	}
	return true, nil
}

// GetProxySettings implements PackageManager.
func (apt *apt) GetProxySettings() (proxy.Settings, error) {
	var res proxy.Settings
	proxyRE := regexp.MustCompile(`(?im)^\s*Acquire::(?P<protocol>[a-z]+)::Proxy\s+"(?P<proxy>[^"]+)";\s*$`)
	args := strings.Fields(apt.cmder.GetProxyCmd())

	out, err := runCommand(args[0], args[1:]...)
	if err != nil {
		logger.Errorf("command failed: %v\nargs: %#v\n%s",
			err, args, string(out))
		return res, fmt.Errorf("command failed: %v", err)
	}

	for _, match := range proxyRE.FindAllStringSubmatch(out, -1) {
		switch match[1] {
		case "http":
			res.Http = match[2]
		case "https":
			res.Https = match[2]
		case "ftp":
			res.Ftp = match[2]
		}
	}

	return res, nil
}
