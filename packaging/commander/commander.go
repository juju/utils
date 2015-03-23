// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commander

import (
	"fmt"
	"strings"

	"github.com/juju/utils/proxy"
)

// packageCommander is a struct which returns system-specific commands for all
// the operations that may be required of a package management system.
// It implements the PackageCommander interface.
type packageCommander struct {
	// The following are the options expected of the cmds map of a
	// packageCommander. They must all be present for the Commander to be
	// fully-featured:
	//
	// prereq				installs prerequisite repo management package
	// update				updates the local package list
	// upgrade				upgrades all packages
	// install packages...	installs the given packages
	// remove packages...	removes the given packages
	// purge packages...	removes the given packages along with all data
	// search package		searches for the given package
	// is-installed package	checks if a given package is installed
	// list-available		lists all packes available
	// list-installed		lists all installed packages
	// list-repositories	lists all currently configured repositories
	// add-repository repo	adds the given repository
	// remove-repository	removes the given repository
	// cleanup				cleans up orhaned packages and the package cache
	// get-proxy			command for getting the currently set packagemanager proxy
	// proxy-setting-format	format for proxy setting in package manager config file
	// set-proxy			command for adding a proxy setting to the config file
	cmds map[string]string
}

// InstallPrerequisiteCmd implements PackageCommander.
func (p *packageCommander) InstallPrerequisiteCmd() string {
	return p.cmds["prereq"]
}

// UpdateCmd implements PackageCommander.
func (p *packageCommander) UpdateCmd() string {
	return p.cmds["update"]
}

// UpgradeCmd implements PackageCommander.
func (p *packageCommander) UpgradeCmd() string {
	return p.cmds["upgrade"]
}

// InstallCmd implements PackageCommander.
func (p *packageCommander) InstallCmd(packs ...string) string {
	cmd := p.cmds["install"]

	for _, pack := range packs {
		cmd = cmd + pack + " "
	}

	return cmd[:len(cmd)-1]
}

// RemoveCmd implements PackageCommander.
func (p *packageCommander) RemoveCmd(packs ...string) string {
	cmd := p.cmds["remove"]

	for _, pack := range packs {
		cmd = cmd + pack + " "
	}

	return cmd[:len(cmd)-1]
}

// PurgeCmd implements PackageCommander.
func (p *packageCommander) PurgeCmd(packs ...string) string {
	cmd := p.cmds["purge"]

	for _, pack := range packs {
		cmd = cmd + pack + " "
	}

	return cmd[:len(cmd)-1]
}

// SearchCmd implements PackageCommander.
func (p *packageCommander) SearchCmd(pack string) string {
	return fmt.Sprintf(p.cmds["search"], pack)
}

// IsInstalledCmd implements PackageCommander.
func (p *packageCommander) IsInstalledCmd(pack string) string {
	return fmt.Sprintf(p.cmds["is-installed"], pack)
}

// ListAvailableCmd implements PackageCommander.
func (p *packageCommander) ListAvailableCmd() string {
	return p.cmds["list-available"]
}

// ListInstalledCmd implements PackageCommander.
func (p *packageCommander) ListInstalledCmd() string {
	return p.cmds["list-installed"]
}

// ListRepositoriesCmd implements PackageCommander.
func (p *packageCommander) ListRepositoriesCmd() string {
	return p.cmds["list-repositories"]
}

// AddRepositoryCmd implements PackageCommander.
func (p *packageCommander) AddRepositoryCmd(repo string) string {
	return fmt.Sprintf(p.cmds["add-repository"], repo)
}

// RemoveRepositoryCmd implements PackageCommander.
func (p *packageCommander) RemoveRepositoryCmd(repo string) string {
	return fmt.Sprintf(p.cmds["remove-repository"], repo)
}

// CleanupCmd implements PackageCommander.
func (p *packageCommander) CleanupCmd() string {
	return p.cmds["cleanup"]
}

// GetProxyCmd implements PackageCommander.
func (p *packageCommander) GetProxyCmd() string {
	return p.cmds["get-proxy"]
}

// giveProxyOptions is a helper function which takes a possible proxy setting
// and its value and returns the formatted option for it.
func (p *packageCommander) giveProxyOption(setting, proxy string) string {
	return fmt.Sprintf(p.cmds["proxy-setting-format"], setting, proxy)
}

// ProxyConfigContents implements PackageCommander.
func (p *packageCommander) ProxyConfigContents(settings proxy.Settings) string {
	options := []string{}

	addOption := func(setting, proxy string) {
		if proxy != "" {
			options = append(options, p.giveProxyOption(setting, proxy))
		}
	}

	addOption("http", settings.Http)
	addOption("https", settings.Https)
	addOption("ftp", settings.Ftp)

	return strings.Join(options, "\n")
}

// SetProxyCmds implements PackageCommander.
func (p *packageCommander) SetProxyCmds(settings proxy.Settings) []string {
	cmds := []string{}

	addProxyCmd := func(setting, proxy string) {
		if proxy != "" {
			cmds = append(cmds, fmt.Sprintf(p.cmds["set-proxy"], p.giveProxyOption(setting, proxy)))
		}
	}

	addProxyCmd("http", settings.Http)
	addProxyCmd("https", settings.Https)
	addProxyCmd("ftp", settings.Ftp)

	return cmds
}
