// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package configurer

// flipMap is a helper function which flips a strings map, making the
// keys of the initial one the values and vice-versa.
func flipMap(m map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		res[v] = k
	}
	return res
}

// sourceWithPrefs is a simple wrapper struct meant to unify a PackageSource
// with PackagePreferences.
type sourceWithPrefs struct {
	src   *PackageSource
	prefs *PackagePreferences
}
