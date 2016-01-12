// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash_test

import (
	"crypto/sha512"
	"encoding/hex"
	stdhash "hash"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/filetesting"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/hash"
)

var _ = gc.Suite(&FingerprintSuite{})

type FingerprintSuite struct {
	stub *testing.Stub
	hash *filetesting.StubHash
}

func (s *FingerprintSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}
	s.hash = filetesting.NewStubHash(s.stub, nil)
}

func (s *FingerprintSuite) newHash() stdhash.Hash {
	s.stub.AddCall("newHash")
	s.stub.NextErr() // Pop one off.

	return s.hash
}

func (s *FingerprintSuite) validate(sum []byte) error {
	s.stub.AddCall("validate", sum)
	if err := s.stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *FingerprintSuite) TestNewFingerprintOkay(c *gc.C) {
	expected, _ := newFingerprint(c, "spamspamspam")

	fp, err := hash.NewFingerprint(expected, s.validate)
	c.Assert(err, jc.ErrorIsNil)
	sum := fp.Bytes()

	s.stub.CheckCallNames(c, "validate")
	c.Check(sum, jc.DeepEquals, expected)
}

func (s *FingerprintSuite) TestNewFingerprintInvalid(c *gc.C) {
	expected, _ := newFingerprint(c, "spamspamspam")
	failure := errors.NewNotValid(nil, "bogus!!!")
	s.stub.SetErrors(failure)

	_, err := hash.NewFingerprint(expected, s.validate)

	s.stub.CheckCallNames(c, "validate")
	c.Check(errors.Cause(err), gc.Equals, failure)
}

func (s *FingerprintSuite) TestNewValidFingerprint(c *gc.C) {
	expected, _ := newFingerprint(c, "spamspamspam")
	s.hash.ReturnSum = expected

	fp := hash.NewValidFingerprint(s.hash)
	sum := fp.Bytes()

	s.stub.CheckCallNames(c, "Sum")
	c.Check(sum, jc.DeepEquals, expected)
}

func (s *FingerprintSuite) TestGenerateFingerprintOkay(c *gc.C) {
	expected, _ := newFingerprint(c, "spamspamspam")
	s.hash.ReturnSum = expected
	s.hash.Writer, _ = filetesting.NewStubWriter(s.stub)
	reader := filetesting.NewStubReader(s.stub, "spamspamspam")

	fp, err := hash.GenerateFingerprint(reader, s.newHash)
	c.Assert(err, jc.ErrorIsNil)
	sum := fp.Bytes()

	s.stub.CheckCallNames(c, "newHash", "Read", "Write", "Read", "Sum")
	c.Check(sum, jc.DeepEquals, expected)
}

func (s *FingerprintSuite) TestGenerateFingerprintNil(c *gc.C) {
	_, err := hash.GenerateFingerprint(nil, s.newHash)

	s.stub.CheckNoCalls(c)
	c.Check(err, gc.ErrorMatches, `missing reader`)
}

func (s *FingerprintSuite) TestParseHexFingerprint(c *gc.C) {
	expected, hexSum := newFingerprint(c, "spamspamspam")

	fp, err := hash.ParseHexFingerprint(hexSum, s.validate)
	c.Assert(err, jc.ErrorIsNil)
	sum := fp.Bytes()

	s.stub.CheckCallNames(c, "validate")
	c.Check(sum, jc.DeepEquals, expected)
}

func (s *FingerprintSuite) TestString(c *gc.C) {
	sum, expected := newFingerprint(c, "spamspamspam")
	fp, err := hash.NewFingerprint(sum, s.validate)
	c.Assert(err, jc.ErrorIsNil)

	hex := fp.String()

	c.Check(hex, gc.Equals, expected)
}

func (s *FingerprintSuite) TestHex(c *gc.C) {
	sum, expected := newFingerprint(c, "spamspamspam")
	fp, err := hash.NewFingerprint(sum, s.validate)
	c.Assert(err, jc.ErrorIsNil)

	hex := fp.String()

	c.Check(hex, gc.Equals, expected)
}

func (s *FingerprintSuite) TestBytes(c *gc.C) {
	expected, _ := newFingerprint(c, "spamspamspam")
	fp, err := hash.NewFingerprint(expected, s.validate)
	c.Assert(err, jc.ErrorIsNil)

	sum := fp.Bytes()

	c.Check(sum, jc.DeepEquals, expected)
}

func (s *FingerprintSuite) TestValidateOkay(c *gc.C) {
	sum, _ := newFingerprint(c, "spamspamspam")
	fp, err := hash.NewFingerprint(sum, s.validate)
	c.Assert(err, jc.ErrorIsNil)

	err = fp.Validate()

	c.Check(err, jc.ErrorIsNil)
}

func (s *FingerprintSuite) TestValidateZero(c *gc.C) {
	var fp hash.Fingerprint
	err := fp.Validate()

	c.Check(err, jc.Satisfies, errors.IsNotValid)
	c.Check(err, gc.ErrorMatches, `zero-value fingerprint not valid`)
}

func newFingerprint(c *gc.C, data string) ([]byte, string) {
	hash := sha512.New384()
	_, err := hash.Write([]byte(data))
	c.Assert(err, jc.ErrorIsNil)
	sum := hash.Sum(nil)

	hexStr := hex.EncodeToString(sum)
	return sum, hexStr
}
