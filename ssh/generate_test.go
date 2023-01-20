// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"encoding/pem"
	"strings"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/v3/ssh"
	"golang.org/x/crypto/ed25519"
	cssh "golang.org/x/crypto/ssh"
	gc "gopkg.in/check.v1"
)

type GenerateSuite struct{}

var _ = gc.Suite(&GenerateSuite{})

func (s *GenerateSuite) TestGenerate(c *gc.C) {
	private, public, err := ssh.GenerateKey("some-comment")

	c.Check(err, jc.ErrorIsNil)
	c.Check(private, jc.HasPrefix, "-----BEGIN OPENSSH PRIVATE KEY-----\n")
	c.Check(private, jc.HasSuffix, "-----END OPENSSH PRIVATE KEY-----\n")
	c.Check(public, jc.HasPrefix, "ssh-ed25519 ")
	c.Check(public, jc.HasSuffix, " some-comment\n")
}

func (s *GenerateSuite) TestKeysMatch(c *gc.C) {
	private, authKey, err := ssh.GenerateKey("some-comment")
	c.Assert(err, jc.ErrorIsNil)

	block, _ := pem.Decode([]byte(private))
	privKey := parseEd25519PrivateKey(block.Bytes)

	publicKeyfromPriv, _ := cssh.NewPublicKey(privKey.Public())
	authKeyFromPriv := string(cssh.MarshalAuthorizedKey(publicKeyfromPriv))

	c.Assert(authKey, jc.HasPrefix, strings.TrimSpace(authKeyFromPriv))
}

func parseEd25519PrivateKey(der []byte) ed25519.PrivateKey {
	headerSize := len(append([]byte("openssh-key-v1"), 0))
	der = der[headerSize:]

	meta := struct {
		CipherName   string
		KdfName      string
		KdfOpts      string
		NumKeys      uint32
		PubKey       []byte
		PrivKeyBlock []byte
	}{}
	cssh.Unmarshal(der, &meta)

	keyData := struct {
		Check1  uint32
		Check2  uint32
		Keytype string
		Pub     []byte
		Priv    []byte
		Comment string
		Pad     []byte `ssh:"rest"`
	}{}
	cssh.Unmarshal(meta.PrivKeyBlock, &keyData)

	return ed25519.PrivateKey(keyData.Priv)
}
