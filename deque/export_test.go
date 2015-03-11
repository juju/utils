// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package deque

// GetDequeBlocks returns the number of internal blocks that the Deque
// is using.
func GetDequeBlocks(d *Deque) int {
	return d.blocks.Len()
}
