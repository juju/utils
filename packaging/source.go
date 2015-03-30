// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

// The packaging package contains definitions for general packaging-related
// types as well as implementations of objects which either return or configure
// packaging-related parameters in its three subpackages.
package packaging

// Source contains all the data required for a package source.
type PackageSource struct {
	Name string `yaml:"-"`
	Url  string `yaml:"source"`
	Key  string `yaml:"key,omitempty"`
}

// KeyFileName returns the name of this source's keyfile.
func (s *PackageSource) KeyFileName() string {
	return s.Name + ".key"
}

// RenderSourceFile renders the current source based on a template it recieves.
func (s *PackageSource) RenderSourceFile(fileTemplate string) string {
	return renderTemplate(fileTemplate, s)
}
