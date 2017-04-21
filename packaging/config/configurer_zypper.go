// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.
// Copied from configurer_yum.go (with all pending TODOs)

package config

import (
	"github.com/juju/utils/packaging"
)

// yumConfigurer is the PackagingConfigurer implementation for apt-based systems.
type zypperConfigurer struct {
	*baseConfigurer
}

// RenderSource is defined on the PackagingConfigurer interface.
func (c *zypperConfigurer) RenderSource(src packaging.PackageSource) (string, error) {
	return src.RenderSourceFile(ZypperSourceTemplate)
}

// RenderPreferences is defined on the PackagingConfigurer interface.
func (c *zypperConfigurer) RenderPreferences(src packaging.PackagePreferences) (string, error) {
	// TODO (aznashwan): research a way of using zypper-priorities in the context
	// of single/multiple package pinning and implement it.
	return "", nil
}

// ApplyCloudArchiveTarget is defined on the PackagingConfigurer interface.
func (c *zypperConfigurer) ApplyCloudArchiveTarget(pack string) []string {
	// TODO (aznashwan): implement target application when archive is available.
	return []string{pack}
}
