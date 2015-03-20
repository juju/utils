// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configurer

// yumConfigurer is the PackagingConfigurer implementation for apt-based systems.
type yumConfigurer struct {
	*baseConfigurer
}

// RenderSource implements PackagingConfigurer.
func (c *yumConfigurer) RenderSource(src PackageSource) string {
	return src.renderSourceFile(YumSourceTemplate[1:])
}

// RenderPreferences implements PackagingConfigurer.
func (c *yumConfigurer) RenderPreferences(src PackagePreferences) string {
	// TODO (aznashwan): research a way of using yum-priorities in the context
	// of single/multiple package pinning and implement it.
	return ""
}
