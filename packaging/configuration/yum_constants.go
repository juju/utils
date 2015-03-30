// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configuration

const (
	// YumSourcesDir is the default directory in which yum sourcefiles are located.
	YumSourcesDir = "/etc/yum/repos.d"

	// YumKeyfileDir is the default directory for yum repository keys.
	YumKeyfileDir = "/etc/pki/rpm-gpg/"
)

// YumSourceTemplate is the template specific to a yum source file.
var YumSourceTemplate = `
[{{.Name}}]
name={{.Name}} (added by Juju)
baseurl={{.Url}}
{{if .Key}}gpgcheck=1
gpgkey=%s{{end}}
enabled=1
`[1:]
