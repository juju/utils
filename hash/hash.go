// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"

	"github.com/juju/errors"
)

type hashSum struct {
	raw hash.Hash
}

// Sum returns the raw checksum.
func (hs hashSum) Sum() []byte {
	return hs.raw.Sum(nil)
}

// Base64Sum returns the base64 encoded hash.
func (hs hashSum) Base64Sum() string {
	raw := hs.raw.Sum(nil)
	return base64.StdEncoding.EncodeToString(raw)
}

// HexSum returns the hex-ified checksum.
func (hs hashSum) HexSum() string {
	raw := hs.raw.Sum(nil)
	return fmt.Sprintf("%x", raw)
}

// Fingerprint returns the fingerprint corresponding to this hash.
func (hs hashSum) Fingerprint() Fingerprint {
	return NewValidFingerprint(hs.raw)
}

// SHA384 returns the newHash and validate functions for use
// with SHA384 hashes. SHA384 is used in several key places in Juju.
func SHA384() (newHash func() hash.Hash, validate func([]byte) error) {
	const digestLenBytes = 384 / 8
	validate = newSizeChecker(digestLenBytes)
	return sha512.New384, validate
}

func newSizeChecker(size int) func([]byte) error {
	return func(sum []byte) error {
		if len(sum) < size {
			return errors.NewNotValid(nil, "invalid fingerprint (too small)")
		}
		if len(sum) > size {
			return errors.NewNotValid(nil, "invalid fingerprint (too big)")
		}
		return nil
	}
}
