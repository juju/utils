// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package cache_test

import (
	"fmt"
	"sync"
	"time"

	gc "gopkg.in/check.v1"
	"gopkg.in/errgo.v1"

	"github.com/juju/utils/cache"
)

type suite struct{}

var _ = gc.Suite(&suite{})

func (*suite) TestSimpleGet(c *gc.C) {
	p := cache.New(time.Hour)
	v, err := p.Get("a", fetchValue(2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)
}

func (*suite) TestEvict(c *gc.C) {
	p := cache.New(time.Hour)
	v, err := p.Get("a", fetchValue(2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)

	v, err = p.Get("a", fetchValue(4))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)

	p.Evict("a")
	v, err = p.Get("a", fetchValue(3))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 3)

	v, err = p.Get("a", fetchValue(4))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 3)
}

func (*suite) TestEvictOld(c *gc.C) {
	// Test that evict removes entries even when they're
	// in the old map.

	now := time.Now()
	p := cache.New(time.Minute)

	// Populate the cache with an initial entry.
	v, err := cache.GetAtTime(p, "a", fetchValue("a"), now)
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "a")
	c.Assert(p.Len(), gc.Equals, 1)

	v, err = cache.GetAtTime(p, "b", fetchValue("b"), now.Add(time.Minute/2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "b")
	c.Assert(p.Len(), gc.Equals, 2)

	// Fetch an item after the expiry time,
	// causing current entries to be moved to old.
	v, err = cache.GetAtTime(p, "a", fetchValue("a1"), now.Add(time.Minute+1))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "a1")
	c.Assert(p.Len(), gc.Equals, 2)
	c.Assert(cache.OldLen(p), gc.Equals, 1)

	p.Evict("b")
	v, err = cache.GetAtTime(p, "b", fetchValue("b1"), now.Add(time.Minute+2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "b1")
}

func (*suite) TestFetchError(c *gc.C) {
	p := cache.New(time.Hour)
	expectErr := errgo.New("hello")
	v, err := p.Get("a", fetchError(expectErr))
	c.Assert(err, gc.ErrorMatches, "hello")
	c.Assert(errgo.Cause(err), gc.Equals, expectErr)
	c.Assert(v, gc.Equals, nil)
}

func (*suite) TestFetchOnlyOnce(c *gc.C) {
	p := cache.New(time.Hour)
	v, err := p.Get("a", fetchValue(2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)

	v, err = p.Get("a", fetchError(errUnexpectedFetch))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)
}

func (*suite) TestEntryExpiresAfterMaxEntryAge(c *gc.C) {
	now := time.Now()
	p := cache.New(time.Minute)
	v, err := cache.GetAtTime(p, "a", fetchValue(2), now)
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)

	// Entry is definitely not expired before half the entry expiry time.
	v, err = cache.GetAtTime(p, "a", fetchError(errUnexpectedFetch), now.Add(time.Minute/2-1))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, 2)

	// Entry is definitely expired after the entry expiry time
	v, err = cache.GetAtTime(p, "a", fetchValue(3), now.Add(time.Minute+1))
	c.Assert(v, gc.Equals, 3)
}

func (*suite) TestEntriesRemovedWhenNotRetrieved(c *gc.C) {
	now := time.Now()
	p := cache.New(time.Minute)

	// Populate the cache with an initial entry.
	v, err := cache.GetAtTime(p, "a", fetchValue("a"), now)
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "a")
	c.Assert(p.Len(), gc.Equals, 1)

	// Fetch another item after the expiry time,
	// causing current entries to be moved to old.
	v, err = cache.GetAtTime(p, "b", fetchValue("b"), now.Add(time.Minute+1))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "b")
	c.Assert(p.Len(), gc.Equals, 2)
	c.Assert(cache.OldLen(p), gc.Equals, 1)

	// Fetch the other item after another expiry time
	// causing the old entries to be discarded because
	// nothing has fetched them.
	v, err = cache.GetAtTime(p, "b", fetchValue("b"), now.Add(time.Minute*2+2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "b")
	c.Assert(p.Len(), gc.Equals, 1)
}

// TestRefreshedEntry tests the code path where a value is moved
// from the old map to new.
func (*suite) TestRefreshedEntry(c *gc.C) {
	now := time.Now()
	p := cache.New(time.Minute)

	// Populate the cache with an initial entry.
	v, err := cache.GetAtTime(p, "a", fetchValue("a"), now)
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "a")
	c.Assert(p.Len(), gc.Equals, 1)

	// Fetch another item very close to the expiry time.
	v, err = cache.GetAtTime(p, "b", fetchValue("b"), now.Add(time.Minute-1))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "b")
	c.Assert(p.Len(), gc.Equals, 2)

	// Fetch it again just after the expiry time,
	// which should move it into the new map.
	v, err = cache.GetAtTime(p, "b", fetchError(errUnexpectedFetch), now.Add(time.Minute+1))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "b")
	c.Assert(p.Len(), gc.Equals, 2)

	// Fetch another item, causing "a" to be removed from the cache
	// and keeping "b" in there.
	v, err = cache.GetAtTime(p, "c", fetchValue("c"), now.Add(time.Minute*2+2))
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, "c")
	c.Assert(p.Len(), gc.Equals, 2)
}

// TestConcurrentFetch checks that the cache is safe
// to use concurrently. It is designed to fail when
// tested with the race detector enabled.
func (*suite) TestConcurrentFetch(c *gc.C) {
	p := cache.New(time.Minute)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		v, err := p.Get("a", fetchValue("a"))
		c.Check(err, gc.IsNil)
		c.Check(v, gc.Equals, "a")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		v, err := p.Get("b", fetchValue("b"))
		c.Check(err, gc.IsNil)
		c.Check(v, gc.Equals, "b")
	}()
	wg.Wait()
}

func (*suite) TestRefreshSpread(c *gc.C) {
	now := time.Now()
	p := cache.New(time.Minute)
	// Get all values to start with.
	const N = 100
	for i := 0; i < N; i++ {
		v, err := cache.GetAtTime(p, fmt.Sprint(i), fetchValue(i), now)
		c.Assert(err, gc.IsNil)
		c.Assert(v, gc.Equals, i)
	}
	counts := make([]int, time.Minute/time.Millisecond/10+1)

	// Continually get values over the course of the
	// expiry time; the fetches should be spread out.
	slot := 0
	for t := now.Add(0); t.Before(now.Add(time.Minute + 1)); t = t.Add(time.Millisecond * 10) {
		for i := 0; i < N; i++ {
			cache.GetAtTime(p, fmt.Sprint(i), func() (interface{}, error) {
				counts[slot]++
				return i, nil
			}, t)
		}
		slot++
	}

	// There should be no fetches in the first half of the cycle.
	for i := 0; i < len(counts)/2; i++ {
		c.Assert(counts[i], gc.Equals, 0, gc.Commentf("slot %d", i))
	}

	max := 0
	total := 0
	for _, count := range counts {
		if count > max {
			max = count
		}
		total += count
	}
	if max > 10 {
		c.Errorf("requests grouped too closely (max %d)", max)
	}
	c.Assert(total, gc.Equals, N)
}

var errUnexpectedFetch = errgo.New("fetch called unexpectedly")

func fetchError(err error) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, err
	}
}

func fetchValue(val interface{}) func() (interface{}, error) {
	return func() (interface{}, error) {
		return val, nil
	}
}
