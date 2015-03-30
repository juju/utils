// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commands

import (
	"fmt"
	"strings"

	"github.com/juju/utils/proxy"
)

// packageCommander is a struct which returns system-specific commands for all
// the operations that may be required of a package management system.
// It implements the PackageCommander interface.
type packageCommander struct {
	prereq                string // installs prerequisite repo management package
	update                string // updates the local package list
	upgrade               string // upgrades all packages
	install               string // installs the given packages
	remove                string // removes the given packages
	purge                 string // removes the given packages along with all data
	search                string // searches for the given package
	is_installed          string // checks if a given package is installed
	list_available        string // lists all packes available
	list_installed        string // lists all installed packages
	list_repositories     string // lists all currently configured repositories
	add_repository        string // adds the given repository
	remove_repository     string // removes the given repository
	cleanup               string // cleans up orhaned packages and the package cache
	get_proxy             string // command for getting the currently set packagemanager proxy
	proxy_settings_format string // format for proxy setting in package manager config file
	set_proxy             string // command for adding a proxy setting to the config file
}

// InstallPrerequisiteCmd implements PackageCommander.
func (p *packageCommander) InstallPrerequisiteCmd() string {
	return p.prereq
}

// UpdateCmd implements PackageCommander.
func (p *packageCommander) UpdateCmd() string {
	return p.update
}

// UpgradeCmd implements PackageCommander.
func (p *packageCommander) UpgradeCmd() string {
	return p.upgrade
}

// InstallCmd implements PackageCommander.
func (p *packageCommander) InstallCmd(packs ...string) string {
	cmd := p.install

	for _, pack := range packs {
		cmd = buildCommand(cmd, pack)
	}

	return cmd
}

// RemoveCmd implements PackageCommander.
func (p *packageCommander) RemoveCmd(packs ...string) string {
	cmd := p.remove

	for _, pack := range packs {
		cmd = buildCommand(cmd, pack)
	}

	return cmd
}

// PurgeCmd implements PackageCommander.
func (p *packageCommander) PurgeCmd(packs ...string) string {
	cmd := p.purge

	for _, pack := range packs {
		cmd = buildCommand(cmd, pack)
	}

	return cmd
}

// SearchCmd implements PackageCommander.
func (p *packageCommander) SearchCmd(pack string) string {
	return fmt.Sprintf(p.search, pack)
}

// IsInstalledCmd implements PackageCommander.
func (p *packageCommander) IsInstalledCmd(pack string) string {
	return fmt.Sprintf(p.is_installed, pack)
}

// ListAvailableCmd implements PackageCommander.
func (p *packageCommander) ListAvailableCmd() string {
	return p.list_available
}

// ListInstalledCmd implements PackageCommander.
func (p *packageCommander) ListInstalledCmd() string {
	return p.list_installed
}

// ListRepositoriesCmd implements PackageCommander.
func (p *packageCommander) ListRepositoriesCmd() string {
	return p.list_repositories
}

// AddRepositoryCmd implements PackageCommander.
func (p *packageCommander) AddRepositoryCmd(repo string) string {
	return fmt.Sprintf(p.add_repository, repo)
}

// RemoveRepositoryCmd implements PackageCommander.
func (p *packageCommander) RemoveRepositoryCmd(repo string) string {
	return fmt.Sprintf(p.remove_repository, repo)
}

// CleanupCmd implements PackageCommander.
func (p *packageCommander) CleanupCmd() string {
	return p.cleanup
}

// GetProxyCmd implements PackageCommander.
func (p *packageCommander) GetProxyCmd() string {
	return p.get_proxy
}

// giveProxyOptions is a helper function which takes a possible proxy setting
// and its value and returns the formatted option for it.
func (p *packageCommander) giveProxyOption(setting, proxy string) string {
	return fmt.Sprintf(p.proxy_settings_format, setting, proxy)
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
			cmds = append(cmds, fmt.Sprintf(p.set_proxy, p.giveProxyOption(setting, proxy)))
		}
	}

	addProxyCmd("http", settings.Http)
	addProxyCmd("https", settings.Https)
	addProxyCmd("ftp", settings.Ftp)

	return cmds
}
