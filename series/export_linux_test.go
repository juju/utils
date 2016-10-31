// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package series

var (
	DistroInfo    = &distroInfo
	ReadSeries    = readSeries
	OSReleaseFile = &osReleaseFile
)

func SetUbuntuSeries(value map[string]string) func() {
	origSeries := ubuntuSeries
	ubuntuSeries = value
	return func() {
		ubuntuSeries = origSeries
	}
}
