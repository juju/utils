// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commander

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

// yumCmder is the packageCommander instanciation for yum-based systems.
var yumCmder = packageCommander{
	prereq:                join(yum, "install yum-utils"),
	update:                join(yum, "clean expire-cache"),
	upgrade:               join(yum, "update"),
	install:               join(yum, "install"),
	remove:                join(yum, "remove"),
	purge:                 join(yum, "remove"), // purges by default
	search:                join(yum, "list %s"),
	is_installed:          join(yum, " list installed %s"),
	list_available:        join(yum, "list all"),
	list_installed:        join(yum, "list installed"),
	list_repositories:     join(yum, "repolist all"),
	add_repository:        join(yumconf, "--add-repo %s"),
	remove_repository:     join(yumconf, "--disable %s"),
	cleanup:               join(yum, "clean all"),
	get_proxy:             join("grep proxy ", WgetRCFilePath, " | grep -v ^#"),
	proxy_settings_format: wgetProxySettingFormat,
	set_proxy:             join("echo %s >> ", WgetRCFilePath),
}
