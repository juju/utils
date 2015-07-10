// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package manager

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/juju/utils/proxy"
)

// proxyRe is a regexp which matches all proxy-related configuration options in
// the apt configuration file.
var proxyRE = regexp.MustCompile(`(?im)^\s*Acquire::(?P<protocol>[a-z]+)::Proxy\s+"(?P<proxy>[^"]+)";\s*$`)

// apt is the PackageManager implementation for deb-based systems.
type apt struct {
	basePackageManager
}

// Search is defined on the PackageManager interface.
func (apt *apt) Search(pack string) (bool, error) {
	out, _, err := RunCommandWithRetry(apt.cmder.SearchCmd(pack))
	if err != nil {
		return false, err
	}

	// apt-cache search --names-only package returns no output
	// if the search was unsuccesful
	if out == "" {
		return false, nil
	}
	return true, nil
}

// GetProxySettings is defined on the PackageManager interface.
func (apt *apt) GetProxySettings() (proxy.Settings, error) {
	var res proxy.Settings

	args := strings.Fields(apt.cmder.GetProxyCmd())
	if len(args) <= 1 {
		return proxy.Settings{}, fmt.Errorf("expected at least 2 arguments, got %d %v", len(args), args)
	}

	cmd := exec.Command(args[0], args[1:]...)
	out, err := CommandOutput(cmd)

	if err != nil {
		logger.Errorf("command failed: %v\nargs: %#v\n%s",
			err, args, string(out))
		return res, fmt.Errorf("command failed: %v", err)
	}

	output := string(bytes.Join(proxyRE.FindAll(out, -1), []byte("\n")))

	for _, match := range proxyRE.FindAllStringSubmatch(output, -1) {
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
