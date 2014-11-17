// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"fmt"
	"math"

	"github.com/juju/errors"
)

// ParseSize parses the string as a size, in mebibytes.
//
// The string must be a is a non-negative number with
// an optional multiplier suffix (M, G, T or P). If the
// suffix is not specified, "M" is implied.
func ParseSize(str string) (MB uint64, err error) {
	var val float64
	var suffix string
	n, _ := fmt.Sscanf(str, "%f%s", &val, &suffix)
	if n == 0 || val < 0 {
		return 0, errors.Errorf("expected a non-negative number with optional multiplier suffix (M/G/T/P), got %q", str)
	}
	if suffix != "" {
		multiplier, ok := mbSuffixes[suffix]
		if !ok {
			return 0, errors.Errorf("invalid multiplier suffix %q", suffix)
		}
		val *= multiplier
	}
	return uint64(math.Ceil(val)), nil
}

var mbSuffixes = map[string]float64{
	"M":   1,
	"MB":  1,
	"MiB": 1,

	"G":   1024,
	"GB":  1024,
	"GiB": 1024,

	"T":   1024 * 1024,
	"TB":  1024 * 1024,
	"TiB": 1024 * 1024,

	"P":   1024 * 1024 * 1024,
	"PB":  1024 * 1024 * 1024,
	"PiB": 1024 * 1024 * 1024,
}
