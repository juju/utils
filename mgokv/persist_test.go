// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgokv_test

import (
	"sync"
	"time"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"
	"gopkg.in/errgo.v1"
	"gopkg.in/mgo.v2/bson"

	"github.com/juju/utils/mgokv"
)

type suite struct {
	testing.MgoSuite
}

var _ = gc.Suite(&suite{})

func (s *suite) TestPutInitial(c *gc.C) {
	type val struct {
		A, B int
	}
	store := mgokv.NewStore(10*time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	ok, err := store.PutInitial("key", val{1, 2})
	c.Assert(err, gc.Equals, nil)
	c.Assert(ok, gc.Equals, true)

	var v val
	err = store.Get("key", &v)
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	// Check that it really is stored in the database by
	// using a fresh store to access it.
	store = mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	v = val{}
	err = store.Get("key", &v)
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	// The second time PutInitial is called, it should do nothing.
	ok, err = store.PutInitial("key", val{99, 100})
	c.Assert(err, gc.Equals, nil)
	c.Assert(ok, gc.Equals, false)

	// The value should not have changed in the cache...
	v = val{}
	err = store.Get("key", &v)
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	// ... or in the database itself.
	store = mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	v = val{}
	err = store.Get("key", &v)
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})
}

func (s *suite) TestPutGet(c *gc.C) {
	type val struct {
		A, B int
	}
	t0 := time.Now()
	store := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	err := mgokv.PutAtTime(store, "key", val{1, 2}, t0)
	c.Assert(err, gc.Equals, nil)

	var v val
	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Millisecond))
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	// If we put a value into the database in another store, the value
	// in the original store will persist until the cache expires.
	store1 := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	err = mgokv.PutAtTime(store1, "key", val{88, 99}, t0)
	c.Assert(err, gc.Equals, nil)

	// Just before the deadline we still see the original value.
	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Second-1))
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	// After the deadline, we see the new value.
	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Second))
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{88, 99})
}

func (s *suite) TestNotFound(c *gc.C) {
	type val struct {
		A, B int
	}
	t0 := time.Now()
	store := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	var v val
	err := mgokv.GetAtTime(store, "key", &v, t0)
	c.Assert(errgo.Cause(err), gc.Equals, mgokv.ErrNotFound)
	c.Assert(v, gc.Equals, val{})

	// Use another store to store a value. The original store should
	// not see the new value until the not-found entry has expired.
	store1 := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	err = mgokv.PutAtTime(store1, "key", val{1, 2}, t0)
	c.Assert(err, gc.Equals, nil)

	// Just before the deadline we still see the not-found error.
	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Second-1))
	c.Assert(errgo.Cause(err), gc.Equals, mgokv.ErrNotFound)
	c.Assert(v, gc.Equals, val{})

	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Second))
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})
}

func (s *suite) TestMultipleKeys(c *gc.C) {
	type val struct {
		A, B int
	}
	store := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)

	err := store.Put("key1", val{1, 2})
	c.Assert(err, gc.Equals, nil)

	err = store.Put("key2", val{3, 4})
	c.Assert(err, gc.Equals, nil)

	var v val
	err = store.Get("key1", &v)
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	err = store.Get("key2", &v)
	c.Assert(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{3, 4})
}

func (s *suite) TestConcurrentGet(c *gc.C) {
	// This test is designed to run with the race detector enabled.

	type val struct {
		A, B int
	}
	// Put a value into the store.
	store := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	err := store.Put("key", val{1, 2})
	c.Check(err, gc.Equals, nil)

	// Use a new store so that we haven't already got a cache entry.
	store = mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var v val
			err := store.Get("key", &v)
			c.Check(err, gc.Equals, nil)
			c.Check(v, gc.Equals, val{1, 2})
		}()
	}
	wg.Wait()
}

func (s *suite) TestRefresh(c *gc.C) {
	type val struct {
		A, B int
	}
	// Put a value into the store.
	store := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	t0 := time.Now()
	err := mgokv.PutAtTime(store, "key", val{1, 2}, t0)
	c.Check(err, gc.Equals, nil)

	// Put a different value using a different store.
	store1 := mgokv.NewStore(time.Second, s.Session.DB("foo").C("x")).Session(s.Session)
	err = store1.Put("key", val{88, 99})
	c.Check(err, gc.Equals, nil)

	// Sanity check that the first store still retains the cached value.
	var v val
	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Millisecond))
	c.Check(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{1, 2})

	// Refresh the store and we should now see the new value.
	store.Refresh()

	err = mgokv.GetAtTime(store, "key", &v, t0.Add(time.Millisecond))
	c.Check(err, gc.Equals, nil)
	c.Assert(v, gc.Equals, val{88, 99})
}

func (s *suite) TestUpdate(c *gc.C) {
	type val struct {
		N int
	}
	store := mgokv.NewStore(time.Minute, s.Session.DB("foo").C("x")).Session(s.Session)
	_, err := store.PutInitial("somekey", val{2})
	c.Assert(err, gc.Equals, nil)
	err = store.Update("somekey", bson.M{"$inc": bson.M{"value.n": 1}})
	c.Assert(err, gc.Equals, nil)

	var v val
	err = store.Get("somekey", &v)
	c.Assert(err, gc.IsNil)
	c.Assert(v, gc.Equals, val{N: 3})
}

func (s *suite) TestUpdateNotFound(c *gc.C) {
	store := mgokv.NewStore(time.Minute, s.Session.DB("foo").C("x")).Session(s.Session)
	err := store.Update("somekey", bson.M{"$inc": bson.M{"value.n": 1}})
	c.Assert(err, gc.Equals, mgokv.ErrNotFound)
}
