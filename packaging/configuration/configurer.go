// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configuration

// baseConfigurer is the base type of a Configurer object.
type baseConfigurer struct {
	series               string
	defaultPackages      []string
	cloudArchivePackages map[string]bool
}

// DefaultPackages implements PackagingConfigurer.
func (c *baseConfigurer) DefaultPackages() []string {
	return c.defaultPackages
}

// GetPackageNameForSeries implements PackagingConfigurer.
func (c *baseConfigurer) GetPackageNameForSeries(pack, series string) string {
	if c.series == series {
		return pack
	}

	// TODO(aznashwan): find a more deterministic way of filtering series that
	// does not imply importing version from core.
	switch series {
	case "centos7":
		res, ok := centOSToUbuntuPackageNameMap[pack]
		if !ok {
			// seems harsh, but this is to encourage all further additions of
			// packages to be made for all distributions...
			panic("Cannot find equivalent Ubuntu package: " + pack)
		}
		return res
	default:
		res, ok := ubuntuToCentOSPackageNameMap[pack]
		if !ok {
			panic("Cannot find equivalent CentOS package: " + pack)
		}
		return res
	}

	return pack
}

// IsCloudArchivePackage implements PackagingConfigurer.
func (c *baseConfigurer) IsCloudArchivePackage(pack string) bool {
	return c.cloudArchivePackages[pack]
}
