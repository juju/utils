// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commander

const (
	// AptConfFilePath is the full file path for the proxy settings that are
	// written by cloud-init and the machine environ worker.
	AptConfFilePath = "/etc/apt/apt.conf.d/42-juju-proxy-settings"

	// the basic command for all dpkg calls:
	dpkg = "dpkg"

	// the basic command for all apt-get calls:
	//		--assume-yes to never prompt for confirmation
	//		--force-confold is passed to dpkg to never overwrite config files
	aptget = "apt-get --assume-yes --option Dpkg::Options::=--force-confold"

	// the basic command for all apt-cache calls:
	aptcache = "apt-cache"

	// the basic command for all add-apt-repository calls:
	//		--yes to never prompt for confirmation
	addaptrepo = "add-apt-repository --yes"

	// the basic command for all apt-config calls:
	aptconfig = "apt-config dump"

	// the basic format for specifying a proxy option for apt:
	aptProxySettingFormat = "Acquire::%s::Proxy %q;"
)

// aptCmder is the packageCommander instanciation for apt-based systems.
var aptCmder = packageCommander{
	prereq:                join(aptget, "install python-software-properties"),
	update:                join(aptget, "update"),
	upgrade:               join(aptget, "upgrade"),
	install:               join(aptget, "install"),
	remove:                join(aptget, "remove"),
	purge:                 join(aptget, "purge"),
	search:                join(aptcache, "search --names-only ^%s$"),
	is_installed:          join(dpkg, "-s %s"),
	list_available:        join(aptcache, "pkgnames"),
	list_installed:        join(dpkg, "--get-selections"),
	add_repository:        join(addaptrepo, "ppa:%s"),
	list_repositories:     `sed -r -n "s|^deb(-src)? (.*)|\1|p"`,
	remove_repository:     join(addaptrepo, "--remove ppa:%s"),
	cleanup:               join(aptget, "autoremove"),
	get_proxy:             join(aptconfig, "Acquire::http::Proxy Acquire::https::Proxy Acquire::ftp::Proxy"),
	proxy_settings_format: aptProxySettingFormat,
	set_proxy:             join("echo %s >> ", AptConfFilePath),
}
