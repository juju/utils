// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// Package mgokv defines cached MongoDB-backed global persistent storage for
// key-value pairs.
//
// It is designed to be used when there is a small set of attributes that change infrequently.
// It shouldn't be used when there's an unbounded set of keys, as key
// entries are not deleted.
package mgokv

import (
	"sync"
	"time"

	"gopkg.in/errgo.v1"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ErrNotFound is returned as the cause of the error
// when an entry is not found.
var ErrNotFound = errgo.New("persistent data entry not found")

type entry struct {
	// value holds the BSON-marshaled value. It will be empty
	// if the entry was not found when fetched.
	value []byte
	// expire holds when the value will expire
	// from the cache. This will be zero when the
	// value does not exist.
	expire time.Time
}

// Store represents a cached set of key-value pairs.
type Store struct {
	cacheLifetime time.Duration
	mu            sync.RWMutex
	// entries holds all the cached entries.
	entries map[string]entry
	coll    *mgo.Collection
}

// Refresh forgets all cached items.
func (s *Store) Refresh() {
	s.mu.Lock()
	s.entries = make(map[string]entry)
	s.mu.Unlock()
}

// NewStore returns a Store that will cache items for at most the given
// time in the given collection. The session in the collection will not
// be used - the session passed to Store.Session will be used instead.
func NewStore(cacheLifetime time.Duration, c *mgo.Collection) *Store {
	return &Store{
		entries:       make(map[string]entry),
		cacheLifetime: cacheLifetime,
		coll:          c,
	}
}

// Session associates a Store instance with a mongo session.
type Session struct {
	*Store
	coll *mgo.Collection
}

// entryDoc holds the document that's stored in MongoDB.
type entryDoc struct {
	Key string `bson:"_id"`
	// Value holds the value. We store it as a raw value
	// so that it looks nice when looking at the collection directly.
	Value bson.Raw `bson:"value"`
}

// Session returns a store session that uses the given
// session for storage. Each store entry is stored
// in a document in the collection.
func (s *Store) Session(session *mgo.Session) *Session {
	return &Session{
		Store: s,
		coll:  s.coll.With(session),
	}
}

// Put stores the given value for the given key. The value must be a struct type that is
// marshalable as BSON (see http://gopkg.in/mgo.v2/bson).
func (s *Session) Put(key string, val interface{}) error {
	return s.putAtTime(key, val, time.Now())
}

// putAtTime is the internal version of Put - it takes the current time
// as an argument for testing.
func (s *Session) putAtTime(key string, val interface{}, now time.Time) error {
	data, err := bson.Marshal(val)
	if err != nil {
		return errgo.Mask(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.coll.UpsertId(key, bson.D{{
		"$set", bson.D{{"value", bson.Raw{
			Kind: 3,
			Data: data,
		}}},
	}})
	if err != nil {
		return errgo.Notef(err, "cannot put %q", key)
	}
	s.entries[key] = entry{
		expire: now.Add(s.cacheLifetime),
		value:  data,
	}
	return nil
}

// PutInitial puts an initial value for the given key. It does
// nothing if there is already a value stored for the key.
// It reports whether the value was actually set.
func (s *Session) PutInitial(key string, val interface{}) (bool, error) {
	return s.putInitialAtTime(key, val, time.Now())
}

// Update updates the value using the MongoDB update operation
// specified in update. The value is stored in the "value" field
// in the document.
//
// For example, if a value of type struct { N int } is associated
// with a key, then:
//
//	s.Update(key, bson.M{"$inc": bson.M{"value.n": 1}})
//
// will atomically increment the N value.
//
// If there is no value associated with the key, Update
// returns ErrNotFound.
func (s *Session) Update(key string, update interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.coll.UpdateId(key, update); err != nil {
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return errgo.Mask(err)
	}
	// We can't easily find the new value so just delete the
	// item from the cache so it will be fetched next time.
	delete(s.entries, key)
	return nil
}

// putInitialAtTime is the internal version of PutInitial - it takes the current time
// as an argument for testing.
func (s *Session) putInitialAtTime(key string, val interface{}, now time.Time) (bool, error) {
	data, err := bson.Marshal(val)
	if err != nil {
		return false, errgo.Mask(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err = s.coll.Insert(&entryDoc{
		Key: key,
		Value: bson.Raw{
			Kind: 3,
			Data: data,
		},
	})
	if mgo.IsDup(err) {
		return false, nil
	}
	if err != nil {
		return false, errgo.Mask(err)
	}
	s.entries[key] = entry{
		expire: now.Add(s.cacheLifetime),
		value:  data,
	}
	return true, nil
}

// Get gets the value associated with the given key into the
// value pointed to by v, which should be a pointer to
// the same struct type used to put the value originally.
//
// If the value is not found, it returns ErrNotFound.
func (s *Session) Get(key string, v interface{}) error {
	return s.getAtTime(key, v, time.Now())
}

// getAtTime is the internal version of Get - it takes the current time
// as an argument for testing.
func (s *Session) getAtTime(key string, v interface{}, now time.Time) error {
	e, err := s.getEntryAtTime(key, now)
	if err != nil {
		return errgo.Mask(err)
	}
	if e.value == nil {
		return ErrNotFound
	}
	if err := bson.Unmarshal(e.value, v); err != nil {
		return errgo.Notef(err, "cannot unmarshal data for key %q into %T", key, v)
	}
	return nil
}

func (s *Session) getEntryAtTime(key string, now time.Time) (entry, error) {
	s.mu.RLock()
	e, ok := s.entries[key]
	s.mu.RUnlock()
	if ok && now.Before(e.expire) {
		return e, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok = s.entries[key]
	if ok && now.Before(e.expire) {
		return e, nil
	}
	var doc entryDoc
	if err := s.coll.FindId(key).One(&doc); err != nil {
		if err != mgo.ErrNotFound {
			return entry{}, errgo.Notef(err, "cannot retrieve data for key %q", key)
		}
	}
	e = entry{
		value:  doc.Value.Data,
		expire: now.Add(s.cacheLifetime),
	}
	s.entries[key] = e
	return e, nil
}
