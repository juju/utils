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

	// yumSourceTemplate is the template specific to a yum source file.
	YumSourceTemplate = `
[{{.Name}}]
name={{.Name}} (added by Juju)
baseurl={{.Url}}
{{if .Key}}gpgcheck=1
gpgkey=%s{{end}}
enabled=1
`
)

const (
	// WgetRCFilePath is the default path of the wget config file.
	WgetRCFilePath = "/etc/wgetrc"

	// the basic command for all yum calls
	// 		--assumeyes to never prompt for confirmation
	//		--debuglevel=1 to limit output verbosity
	yum = "yum --assumeyes --debuglevel=1 "

	// the basic command for all yum repository configuration operations.
	yumconf = "yum-config-manager "

	// the basic format for specifying a proxy setting for wget
	// (which is used by yum in the background)
	wgetProxySettingFormat = "%s_proxy = %s"
)

// yumCmds is a map of available actions specific to a package manager
// and their direct equivalent command on a yum-based system.
var yumCmds map[string]string = map[string]string{
	"prereq":               yum + "install yum-utils",
	"update":               yum + "clean expire-cache",
	"upgrade":              yum + "update",
	"install":              yum + "install ",
	"remove":               yum + "remove ",
	"purge":                yum + "remove ", // purges by default
	"search":               yum + "list %s",
	"is-installed":         yum + " list installed %s",
	"list-available":       yum + "list all",
	"list-installed":       yum + "list installed",
	"list-repositories":    yum + "repolist all",
	"add-repository":       yumconf + "--add-repo %s",
	"remove-repository":    yumconf + "--disable %s",
	"cleanup":              yum + "clean all",
	"get-proxy":            "grep proxy " + WgetRCFilePath + " | grep -v ^#",
	"proxy-setting-format": wgetProxySettingFormat,
	"set-proxy":            "echo %s >> " + WgetRCFilePath,
}
