// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package cache

var GetAtTime = (*Cache).getAtTime

func OldLen(c *Cache) int {
	return len(c.old)
}
