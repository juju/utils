// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"github.com/juju/utils/packaging/manager"
	"github.com/juju/utils/packaging/manager/testing"
)

var _ manager.PackageManager = &testing.MockPackageManager{}
