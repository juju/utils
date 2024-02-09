// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"crypto/dsa"
	"crypto/ed25519"
	"io"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/v4/ssh"
)

type GenerateSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&GenerateSuite{})

var (
	pregeneratedKey ed25519.PrivateKey
)

// overrideGenerateKey patches out rsa.GenerateKey to create a single testing
// key which is saved and used between tests to save computation time.
func overrideGenerateKey() testing.Restorer {
	restorer := testing.PatchValue(ssh.ED25519GenerateKey, func(random io.Reader) (ed25519.PublicKey, ed25519.PrivateKey, error) {
		if pregeneratedKey != nil {
			return ed25519.PublicKey{}, pregeneratedKey, nil
		}
		public, private, err := generateED25519Key(random)
		if err != nil {
			return nil, nil, err
		}
		pregeneratedKey = private
		return public, private, nil
	})
	return restorer
}

func generateED25519Key(random io.Reader) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	// Ignore requested bits and just use 512 bits for speed
	public, private, err := ed25519.GenerateKey(random)
	if err != nil {
		return nil, nil, err
	}
	return public, private, nil
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
	defer overrideGenerateKey().Restore()
	private, public, err := ssh.GenerateKey("some-comment")
	c.Check(err, jc.ErrorIsNil)
	c.Check(private, jc.HasPrefix, "-----BEGIN OPENSSH PRIVATE KEY-----\n")
	c.Check(private, jc.HasSuffix, "-----END OPENSSH PRIVATE KEY-----\n")
	c.Check(public, jc.HasPrefix, "ssh-ed25519 ")
	c.Check(public, jc.HasSuffix, " some-comment\n")
}
