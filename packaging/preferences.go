// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package packaging

import (
	"bytes"
	"text/template"
)

// AptPreferences is a set of apt_preferences(5) compatible preferences for an
// apt source. It can be used to override the default priority for the source.
// Path where the file will be created (usually in /etc/apt/preferences.d/).
type PackagePreferences struct {
	Path        string // the file the prefs will be written at
	Explanation string // a short explanation for the preference
	Package     string // the name of the package the preference applies to
	Pin         string // a pin on a certain source
	Priority    int    // the priority of that source
}

// RenderPreferenceFile returns contents of the package-manager specific config file
// of this paritcular package source.
func (s *PackagePreferences) RenderPreferenceFile(temp string) string {
	var buf bytes.Buffer

	t := template.Must(template.New("").Parse(temp))
	err := t.Execute(&buf, s)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
