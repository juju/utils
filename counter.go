// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import "sync"

// ConcurrentCounter is a thread-safe type which allows users to
// increment, decrement, or retrieve its current count.
type ConcurrentCounter struct {
	countLock sync.RWMutex
	count     int64
}

// Increment increases the count by 1 and returns the current count.
func (c *ConcurrentCounter) Increment() int64 {
	return c.Add(1)
}

// Decrement decreases the count by 1 and returns the current count.
func (c *ConcurrentCounter) Decrement() int64 {
	return c.Add(-1)
}

// Add adds n to the counter and returns the current count.
func (c *ConcurrentCounter) Add(n int64) int64 {
	c.countLock.Lock()
	defer c.countLock.Unlock()

	c.count += n
	return c.count
}

// Count returns the current count.
func (c *ConcurrentCounter) Count() int64 {
	c.countLock.RLock()
	defer c.countLock.RUnlock()
	return c.count
}
