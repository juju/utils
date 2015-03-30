// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

// This package contains a mock implementation of the manager.PackageManager
// interface which always returns positive outcomes and a nil error.
package testing

import "github.com/juju/utils/proxy"

// MockPackageManager is a struct which always returns a positive outcome,
// constant ProxySettings and a nil error.
// It satisfies the PackageManager interface.
type MockPackageManager struct {
}

// InstallPrerequisite implements PackageManager.
func (pm *MockPackageManager) InstallPrerequisite() error {
	return nil
}

// Update implements PackageManager.
func (pm *MockPackageManager) Update() error {
	return nil
}

// Upgrade implements PackageManager.
func (pm *MockPackageManager) Upgrade() error {
	return nil
}

// Install implements PackageManager.
func (pm *MockPackageManager) Install(...string) error {
	return nil
}

// Remove implements PackageManager.
func (pm *MockPackageManager) Remove(...string) error {
	return nil
}

// Purge implements PackageManager.
func (pm *MockPackageManager) Purge(...string) error {
	return nil
}

// Search implements PackageManager.
func (pm *MockPackageManager) Search(string) (bool, error) {
	return true, nil
}

// IsInstalled implements PackageManager.
func (pm *MockPackageManager) IsInstalled(string) bool {
	return true
}

// AddRepository implements PackageManager.
func (pm *MockPackageManager) AddRepository(string) error {
	return nil
}

// RemoveRepository implements PackageManager.
func (pm *MockPackageManager) RemoveRepository(string) error {
	return nil
}

// Cleanup implements PackageManager.
func (pm *MockPackageManager) Cleanup() error {
	return nil
}

// GetProxySettings implements PackageManager.
func (pm *MockPackageManager) GetProxySettings() (proxy.Settings, error) {
	return proxy.Settings{"http proxy", "https proxy", "ftp proxy", "no proxy"}, nil
}

// SetProxy implements PackageManager.
func (pm *MockPackageManager) SetProxy(proxy.Settings) error {
	return nil
}
