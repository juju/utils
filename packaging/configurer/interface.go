// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configurer

import (
	"github.com/juju/utils/packaging"
)

// PackagingConfigurer is an interface which handles various packaging-related configuration
// functions fot he specific distribution it represents.
type PackagingConfigurer interface {
	// DefaultPackages returns a list of default packages whcih should be
	// installed the vast majority of cases on any specific machine
	DefaultPackages() []string

	// GetPackageNameForSeries returns the equivalent package name of the
	// specified package for the given series.
	GetPackageNameForSeries(string, string) string

	// IsCloudArchivePackage signals whether the given package is a
	// cloud archive package and thus should be set as such.
	IsCloudArchivePackage(string) bool

	// ApplyCloudArchiveTarget returns the package with the required target
	// release bits preceding it.
	ApplyCloudArchiveTarget(string) []string

	// RenderSource returns the os-specific full file contents
	// of a given PackageSource.
	RenderSource(packaging.PackageSource) string

	// RenderPreferences returns the os-specific full file contents of a given
	// set of PackagePreferences.
	RenderPreferences(packaging.PackagePreferences) string
}

func NewPackagingConfigurer(series string) PackagingConfigurer {
	switch series {
	// TODO (aznashwan): find a more deterministic way of selection here
	// without importing version from core.
	case "centos7":
		return NewYumPackagingConfigurer(series)
	default:
		return NewAptPackagingConfigurer(series)
	}
}

// NewAptPackagingConfigurer returns a PackagingConfigurer for apt-based systems.
func NewAptPackagingConfigurer(series string) PackagingConfigurer {
	return &aptConfigurer{&baseConfigurer{
		series:               series,
		defaultPackages:      UbuntuDefaultPackages,
		cloudArchivePackages: cloudArchivePackagesUbuntu,
	}}
}

// NewYumPackagingConfigurer returns a PackagingConfigurer for yum-based systems.
func NewYumPackagingConfigurer(series string) PackagingConfigurer {
	return &yumConfigurer{&baseConfigurer{
		series:               series,
		defaultPackages:      CentOSDefaultPackages,
		cloudArchivePackages: cloudArchivePackagesCentOS,
	}}
}
