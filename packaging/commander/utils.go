// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package commander

import "strings"

// join is a helper function which joins its attributes with a space
func join(args ...string) string {
	return strings.Join(args, " ")
}
