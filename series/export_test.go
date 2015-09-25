// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package series

var (
	KernelToMajor                 = kernelToMajor
	MacOSXSeriesFromKernelVersion = macOSXSeriesFromKernelVersion
	MacOSXSeriesFromMajorVersion  = macOSXSeriesFromMajorVersion
)

func SetSeriesVersions(value map[string]string) func() {
	origVersions := seriesVersions
	origUpdated := updatedseriesVersions
	seriesVersions = value
	updateVersionSeries()
	updatedseriesVersions = len(value) != 0
	return func() {
		seriesVersions = origVersions
		updateVersionSeries()
		updatedseriesVersions = origUpdated
	}
}
