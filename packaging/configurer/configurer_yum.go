// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configurer

import (
	"github.com/juju/utils/packaging"
)

// yumConfigurer is the PackagingConfigurer implementation for apt-based systems.
type yumConfigurer struct {
	*baseConfigurer
}

// RenderSource implements PackagingConfigurer.
func (c *yumConfigurer) RenderSource(src packaging.PackageSource) string {
	return src.RenderSourceFile(YumSourceTemplate)
}

// RenderPreferences implements PackagingConfigurer.
func (c *yumConfigurer) RenderPreferences(src packaging.PackagePreferences) string {
	// TODO (aznashwan): research a way of using yum-priorities in the context
	// of single/multiple package pinning and implement it.
	return ""
}

// ApplyCloudArchiveTarget implements PackagingConfigurer.
func (c *yumConfigurer) ApplyCloudArchiveTarget(pack string) []string {
	// TODO (aznashwan): implement target application when archive is available.
	return []string{pack}
}
