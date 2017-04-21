// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package config_test

import (
	"github.com/juju/utils/packaging/config"
)

var _ config.PackagingConfigurer = config.NewAptPackagingConfigurer("some-series")
var _ config.PackagingConfigurer = config.NewYumPackagingConfigurer("some-series")
var _ config.PackagingConfigurer = config.NewZypperPackagingConfigurer("some-series")
