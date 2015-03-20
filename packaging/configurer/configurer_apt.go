// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configurer

// aptConfigurer is the PackagingConfigurer implementation for apt-based systems.
type aptConfigurer struct {
	*baseConfigurer
}

// RenderSource implements PackagingConfigurer.
func (c *aptConfigurer) RenderSource(src PackageSource) string {
	return src.renderSourceFile(AptSourceTemplate[1:])
}

// RenderPreferences implements PackagingConfigurer.
func (c *aptConfigurer) RenderPreferences(prefs PackagePreferences) string {
	return prefs.renderPreferenceFile(AptPreferenceTemplate[1:])
}
