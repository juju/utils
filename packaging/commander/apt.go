// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commander

const (
	// AptConfFilePath is the full file path for the proxy settings that are
	// written by cloud-init and the machine environ worker.
	AptConfFilePath = "/etc/apt/apt.conf.d/42-juju-proxy-settings"

	// the basic command for all dpkg calls:
	dpkg = "dpkg "

	// the basic command for all apt-get calls
	//		--assume-yes to never prompt for confirmation
	//		--force-confold is passed to dpkg to never overwrite config files
	aptget = "apt-get --assume-yes --option Dpkg::Options::=--force-confold "

	// the basic command for all apt-cache calls:
	aptcache = "apt-cache "

	// the basic command for all add-apt-repository calls
	//		--yes to never prompt for confirmation
	addaptrepo = "add-apt-repository --yes "

	// the basic command for all apt-config calls:
	aptconfig = "apt-config dump "

	// the basic format for specifying a proxy option for apt :
	aptProxySettingFormat = "Acquire::%s::Proxy %q;"
)

// aptCmds is a map of available actions specific to a package manager
// and their direct equivalent command on an apt-based system.
var aptCmds map[string]string = map[string]string{
	"prereq":               aptget + "install python-software-properties",
	"update":               aptget + "update",
	"upgrade":              aptget + "upgrade",
	"install":              aptget + "install ",
	"remove":               aptget + "remove ",
	"purge":                aptget + "purge ",
	"search":               aptcache + "search --names-only ^%s$",
	"is-installed":         dpkg + "-s %s",
	"list-available":       aptcache + "pkgnames",
	"list-installed":       dpkg + "--get-selections",
	"add-repository":       addaptrepo + "ppa:%s",
	"list-repositories":    `sed -r -n "s|^deb(-src)? (.*)|\1|p"`,
	"remove-repository":    addaptrepo + "--remove ppa:%s",
	"cleanup":              aptget + "autoremove",
	"get-proxy":            aptconfig + "Acquire::http::Proxy Acquire::https::Proxy Acquire::ftp::Proxy",
	"proxy-setting-format": aptProxySettingFormat,
	"set-proxy":            "echo %s >> " + AptConfFilePath,
}
