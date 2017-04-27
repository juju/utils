// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"crypto/dsa"
	"crypto/rsa"
	"io"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/ssh"
)

type GenerateSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&GenerateSuite{})

var (
	pregeneratedKey   *rsa.PrivateKey
	alternativeRSAKey *rsa.PrivateKey
)

// overrideGenerateKey patches out rsa.GenerateKey to create a single testing
// key which is saved and used between tests to save computation time.
func overrideGenerateKey(c *gc.C) testing.Restorer {
	restorer := testing.PatchValue(ssh.RSAGenerateKey, func(random io.Reader, bits int) (*rsa.PrivateKey, error) {
		if pregeneratedKey != nil {
			return pregeneratedKey, nil
		}
		key, err := generateRSAKey(random)
		if err != nil {
			return nil, err
		}
		pregeneratedKey = key
		return key, nil
	})
	return restorer
}

func generateRSAKey(random io.Reader) (*rsa.PrivateKey, error) {
	// Ignore requested bits and just use 512 bits for speed
	key, err := rsa.GenerateKey(random, 512)
	if err != nil {
		return nil, err
	}
	key.Precompute()
	return key, nil
}

func generateDSAKey(random io.Reader) (*dsa.PrivateKey, error) {
	var privKey dsa.PrivateKey
	if err := dsa.GenerateParameters(&privKey.Parameters, random, dsa.L1024N160); err != nil {
		return nil, err
	}
	if err := dsa.GenerateKey(&privKey, random); err != nil {
		return nil, err
	}
	return &privKey, nil
}

func (s *GenerateSuite) TestGenerate(c *gc.C) {
	defer overrideGenerateKey(c).Restore()
	private, public, err := ssh.GenerateKey("some-comment")

	c.Check(err, jc.ErrorIsNil)
	c.Check(private, jc.HasPrefix, "-----BEGIN RSA PRIVATE KEY-----\n")
	c.Check(private, jc.HasSuffix, "-----END RSA PRIVATE KEY-----\n")
	c.Check(public, jc.HasPrefix, "ssh-rsa ")
	c.Check(public, jc.HasSuffix, " some-comment\n")
}
