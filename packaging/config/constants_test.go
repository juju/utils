// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package config_test

import "github.com/juju/utils/packaging"

var (
	// testedSeriesUbuntu is simply the series to use in Ubuntu tests.
	testedSeriesUbuntu = "precise"

	// testedSeriesCentOS is simply the series we use in CentOS tests.
	testedSeriesCentOS = "centos7"

	// testedPackages is a slice of random package tests to run tests on.
	testedPackages = []string{
		"awesome-wm",
		"archey3",
		"arch-chroot",
		"ranger",
	}

	testedSource = packaging.PackageSource{
		Name: "Some Totally Official Source.",
		URL:  "some-source.com/packages",
		Key:  "some-key",
	}

	testedPrefs = packaging.PackagePreferences{
		Path:        "/etc/my-package-manager.d/prefs_file.conf",
		Explanation: "don't judge me",
		Package:     "some-package",
		Pin:         "releases/extra-special",
		Priority:    42,
	}
)
