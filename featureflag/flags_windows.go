// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

// +build windows

package featureflag

import (
	"github.com/gabriel-samfira/sys/windows/registry"
)

// getFlagsFromRegistry returns the string value from a registry key
func getFlagsFromRegistry(envVarKey, envVarName string) string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, envVarKey[6:], registry.QUERY_VALUE)
	if err != nil {
		// Since this is called during init, we can't fail here. We just log
		// the failure and move on.
		logger.Infof("Failed to open juju registry key %s; feature flags not enabled", envVarKey)
		return ""
	}
	defer k.Close()

	f, _, err := k.GetStringValue(envVarName)
	if err != nil {
		// Since this is called during init, we can't fail here. We just log
		// the failure and move on.
		logger.Infof("Failed to read juju registry value %s; feature flags not enabled", envVarName)
		return ""
	}

	return f
}
