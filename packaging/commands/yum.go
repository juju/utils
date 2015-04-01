// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commands

const (
	// CentOSSourcesDir is the default directory in which yum sourcefiles
	// may be found.
	CentOSSourcesDir = "/etc/yum/repos.d"

	// CentOSYumKeyfileDir is the default directory for yum repository keys.
	CentOSYumKeyfileDir = "/etc/pki/rpm-gpg/"

	// CentOSSourcesFile is the default file which lists all core sources
	// for yum packages on CentOS.
	CentOSSourcesFile = "/etc/yum/repos.d/CentOS-Base.repo"
)

const (
	// WgetRCFilePath is the default path of the wget config file.
	WgetRCFilePath = "/etc/wgetrc"

	// the basic command for all yum calls
	// 		--assumeyes to never prompt for confirmation
	//		--debuglevel=1 to limit output verbosity
	yum = "yum --assumeyes --debuglevel=1"

	// the basic command for all yum repository configuration operations.
	yumconf = "yum-config-manager"

	// the basic format for specifying a proxy setting for wget
	// (which is used by yum in the background)
	wgetProxySettingFormat = "%s_proxy = %s"
)

// yumCmder is the packageCommander instantiation for yum-based systems.
var yumCmder = packageCommander{
	prereq:              buildCommand(yum, "install yum-utils"),
	update:              buildCommand(yum, "clean expire-cache"),
	upgrade:             buildCommand(yum, "update"),
	install:             buildCommand(yum, "install"),
	remove:              buildCommand(yum, "remove"),
	purge:               buildCommand(yum, "remove"), // purges by default
	search:              buildCommand(yum, "list %s"),
	isInstalled:         buildCommand(yum, " list installed %s"),
	listAvailable:       buildCommand(yum, "list all"),
	listInstalled:       buildCommand(yum, "list installed"),
	listRepositories:    buildCommand(yum, "repolist all"),
	addRepository:       buildCommand(yumconf, "--add-repo %s"),
	removeRepository:    buildCommand(yumconf, "--disable %s"),
	cleanup:             buildCommand(yum, "clean all"),
	getProxy:            buildCommand("grep proxy ", WgetRCFilePath, " | grep -v ^#"),
	proxySettingsFormat: wgetProxySettingFormat,
	setProxy:            buildCommand("echo %s >> ", WgetRCFilePath),
}
