// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.
package series

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// fallbackSeries is used if we cannot determine
// the series by using distro-info.
const fallbackLtsSeries = "xenial"

// latestLtsSeries is used to ensure we only do
// the work to determine the latest lts series once.
var latestLtsSeries string

// LatestLtsSeries returns the LatestLtsSeries found in distro-info
func LatestLts() string {
	if latestLtsSeries != "" {
		return latestLtsSeries
	}
	series, err := distroLtsSeries()
	if err != nil {
		latestLtsSeries = fallbackLtsSeries
	} else {
		latestLtsSeries = series
	}
	return latestLtsSeries
}

var distroLtsSeries = distroLtsSeriesFunc

// distInfoCmd allows replacing the exec.Command call in tests.
var distInfoCmd = func() ([]byte, error) {
	return exec.Command("distro-info", "--lts").Output()
}

// distroLtsSeriesFunc returns the latest LTS series, if this information is
// available on this system.
func distroLtsSeriesFunc() (string, error) {
	out, err := distInfoCmd()
	if err != nil {
		return "", err
	}
	series := strings.TrimSpace(string(out))
	if !isValidLts(series) {
		return "", fmt.Errorf("not a valid LTS series: %q", series)
	}

	return series, nil
}

// Evaluates if an ubuntu y.m string is a valid/supported LTS.
func isValidLts(s string) bool {
	ver := ubuntuSeries[s]
	ym := strings.Split(ver, ".")
	if len(ym) != 2 {
		return false
	}
	m, err := strconv.Atoi(ym[1])
	if err != nil {
		return false
	}
	if m != 4 {
		return false
	}
	y, err := strconv.Atoi(ym[0])
	if err != nil {
		return false
	}
	year := time.Now().Year()
	if y%2 != 0 || year-y >= year-5 {
		return false
	}
	return true
}

// SetLatestLts is provided to allow tests to
// override the lts series used and decouple the tests
// from the host by avoiding calling out to distro-ifno.
func SetLatestLts(series string) {
	latestLtsSeries = series
}
