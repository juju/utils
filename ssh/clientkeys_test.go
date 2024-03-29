// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"io/ioutil"
	"os"

	gitjujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/v4"
	"github.com/juju/utils/v4/ssh"
)

type ClientKeysSuite struct {
	gitjujutesting.FakeHomeSuite
}

var _ = gc.Suite(&ClientKeysSuite{})

func (s *ClientKeysSuite) SetUpTest(c *gc.C) {
	s.FakeHomeSuite.SetUpTest(c)
	s.AddCleanup(func(*gc.C) { ssh.ClearClientKeys() })
	generateKeyRestorer := overrideGenerateKey()
	s.AddCleanup(func(*gc.C) { generateKeyRestorer.Restore() })
}

func checkFiles(c *gc.C, obtained, expected []string) {
	var err error
	for i, e := range expected {
		expected[i], err = utils.NormalizePath(e)
		c.Assert(err, jc.ErrorIsNil)
	}
	c.Assert(obtained, jc.SameContents, expected)
}

func checkPublicKeyFiles(c *gc.C, expected ...string) {
	keys := ssh.PublicKeyFiles()
	checkFiles(c, keys, expected)
}

func checkPrivateKeyFiles(c *gc.C, expected ...string) {
	keys := ssh.PrivateKeyFiles()
	checkFiles(c, keys, expected)
}

func (s *ClientKeysSuite) TestPublicKeyFiles(c *gc.C) {
	// LoadClientKeys will create the specified directory
	// and populate it with a key pair.
	err := ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	checkPublicKeyFiles(c, "~/.juju/ssh/juju_id_ed25519.pub")
	// All files ending with .pub in the client key dir get picked up.
	priv, pub, err := ssh.GenerateKey("whatever")
	c.Assert(err, jc.ErrorIsNil)
	err = ioutil.WriteFile(gitjujutesting.HomePath(".juju", "ssh", "whatever.pub"), []byte(pub), 0600)
	c.Assert(err, jc.ErrorIsNil)
	err = ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	// The new public key won't be observed until the
	// corresponding private key exists.
	checkPublicKeyFiles(c, "~/.juju/ssh/juju_id_ed25519.pub")
	err = ioutil.WriteFile(gitjujutesting.HomePath(".juju", "ssh", "whatever"), []byte(priv), 0600)
	c.Assert(err, jc.ErrorIsNil)
	err = ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	checkPublicKeyFiles(c, "~/.juju/ssh/juju_id_ed25519.pub", "~/.juju/ssh/whatever.pub")
}

func (s *ClientKeysSuite) TestPrivateKeyFiles(c *gc.C) {
	// Create/load client keys. They will be cached in memory:
	// any files added to the directory will not be considered
	// unless LoadClientKeys is called again.
	err := ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	checkPrivateKeyFiles(c, "~/.juju/ssh/juju_id_ed25519")
	priv, pub, err := ssh.GenerateKey("whatever")
	c.Assert(err, jc.ErrorIsNil)
	err = ioutil.WriteFile(gitjujutesting.HomePath(".juju", "ssh", "whatever"), []byte(priv), 0600)
	c.Assert(err, jc.ErrorIsNil)
	err = ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	// The new private key won't be observed until the
	// corresponding public key exists.
	checkPrivateKeyFiles(c, "~/.juju/ssh/juju_id_ed25519")
	err = ioutil.WriteFile(gitjujutesting.HomePath(".juju", "ssh", "whatever.pub"), []byte(pub), 0600)
	c.Assert(err, jc.ErrorIsNil)
	// new keys won't be reported until we call LoadClientKeys again
	checkPublicKeyFiles(c, "~/.juju/ssh/juju_id_ed25519.pub")
	checkPrivateKeyFiles(c, "~/.juju/ssh/juju_id_ed25519")
	err = ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	checkPublicKeyFiles(c, "~/.juju/ssh/juju_id_ed25519.pub", "~/.juju/ssh/whatever.pub")
	checkPrivateKeyFiles(c, "~/.juju/ssh/juju_id_ed25519", "~/.juju/ssh/whatever")
}

func (s *ClientKeysSuite) TestLoadClientKeysDirExists(c *gc.C) {
	err := os.MkdirAll(gitjujutesting.HomePath(".juju", "ssh"), 0755)
	c.Assert(err, jc.ErrorIsNil)
	err = ssh.LoadClientKeys("~/.juju/ssh")
	c.Assert(err, jc.ErrorIsNil)
	checkPrivateKeyFiles(c, "~/.juju/ssh/juju_id_ed25519")
}
