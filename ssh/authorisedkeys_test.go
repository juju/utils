// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"encoding/base64"
	"strings"

	gitjujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/v3/ssh"
	sshtesting "github.com/juju/utils/v3/ssh/testing"
)

type AuthorisedKeysKeysSuite struct {
	gitjujutesting.FakeHomeSuite
}

const (
	// We'll use the current user for ssh tests.
	testSSHUser          = ""
	authKeysFile         = "authorized_keys"
	alternativeKeysFile2 = "authorized_keys2"
	alternativeKeysFile3 = "authorized_keys3"
)

var _ = gc.Suite(&AuthorisedKeysKeysSuite{})

func writeAuthKeysFile(c *gc.C, keys []string, file string) {
	err := ssh.WriteAuthorisedKeys(testSSHUser, file, keys)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *AuthorisedKeysKeysSuite) TestListKeys(c *gc.C) {
	keys := []string{
		sshtesting.ValidKeyOne.Key + " user@host",
		sshtesting.ValidKeyTwo.Key,
	}
	writeAuthKeysFile(c, keys, authKeysFile)
	keys, err := ssh.ListKeys(testSSHUser, ssh.Fingerprints)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(
		keys, gc.DeepEquals,
		[]string{sshtesting.ValidKeyOne.Fingerprint + " (user@host)", sshtesting.ValidKeyTwo.Fingerprint})
}

func (s *AuthorisedKeysKeysSuite) TestListKeysFull(c *gc.C) {
	keys := []string{
		sshtesting.ValidKeyOne.Key + " user@host",
		sshtesting.ValidKeyTwo.Key + " anotheruser@host",
	}
	writeAuthKeysFile(c, keys, authKeysFile)
	actual, err := ssh.ListKeys(testSSHUser, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, keys)
}

func (s *AuthorisedKeysKeysSuite) TestAddNewKey(c *gc.C) {
	key := sshtesting.ValidKeyOne.Key + " user@host"
	err := ssh.AddKeys(testSSHUser, key)
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ListKeys(testSSHUser, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, []string{key})
}

func (s *AuthorisedKeysKeysSuite) TestAddMoreKeys(c *gc.C) {
	firstKey := sshtesting.ValidKeyOne.Key + " user@host"
	writeAuthKeysFile(c, []string{firstKey}, authKeysFile)
	moreKeys := []string{
		sshtesting.ValidKeyTwo.Key + " anotheruser@host",
		sshtesting.ValidKeyThree.Key + " yetanotheruser@host",
	}
	err := ssh.AddKeys(testSSHUser, moreKeys...)
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ListKeys(testSSHUser, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, append([]string{firstKey}, moreKeys...))
}

func (s *AuthorisedKeysKeysSuite) TestAddDuplicateKey(c *gc.C) {
	key := sshtesting.ValidKeyOne.Key + " user@host"
	err := ssh.AddKeys(testSSHUser, key)
	c.Assert(err, jc.ErrorIsNil)
	moreKeys := []string{
		sshtesting.ValidKeyOne.Key + " user@host",
		sshtesting.ValidKeyTwo.Key + " yetanotheruser@host",
	}
	err = ssh.AddKeys(testSSHUser, moreKeys...)
	c.Assert(err, gc.ErrorMatches, "cannot add duplicate ssh key: "+sshtesting.ValidKeyOne.Fingerprint)
}

func (s *AuthorisedKeysKeysSuite) TestAddDuplicateComment(c *gc.C) {
	key := sshtesting.ValidKeyOne.Key + " user@host"
	err := ssh.AddKeys(testSSHUser, key)
	c.Assert(err, jc.ErrorIsNil)
	moreKeys := []string{
		sshtesting.ValidKeyTwo.Key + " user@host",
		sshtesting.ValidKeyThree.Key + " yetanotheruser@host",
	}
	err = ssh.AddKeys(testSSHUser, moreKeys...)
	c.Assert(err, gc.ErrorMatches, "cannot add ssh key with duplicate comment: user@host")
}

func (s *AuthorisedKeysKeysSuite) TestAddKeyWithoutComment(c *gc.C) {
	keys := []string{
		sshtesting.ValidKeyOne.Key + " user@host",
		sshtesting.ValidKeyTwo.Key,
	}
	err := ssh.AddKeys(testSSHUser, keys...)
	c.Assert(err, gc.ErrorMatches, "cannot add ssh key without comment")
}

func (s *AuthorisedKeysKeysSuite) TestAddKeepsUnrecognised(c *gc.C) {
	writeAuthKeysFile(c, []string{sshtesting.ValidKeyOne.Key, "invalid-key"}, authKeysFile)
	anotherKey := sshtesting.ValidKeyTwo.Key + " anotheruser@host"
	err := ssh.AddKeys(testSSHUser, anotherKey)
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ReadAuthorisedKeys(testSSHUser, authKeysFile)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, []string{sshtesting.ValidKeyOne.Key, "invalid-key", anotherKey})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteKeys(c *gc.C) {
	firstKey := sshtesting.ValidKeyOne.Key + " user@host"
	anotherKey := sshtesting.ValidKeyTwo.Key
	thirdKey := sshtesting.ValidKeyThree.Key + " anotheruser@host"
	writeAuthKeysFile(c, []string{firstKey, anotherKey, thirdKey}, authKeysFile)
	err := ssh.DeleteKeys(testSSHUser, "user@host", sshtesting.ValidKeyTwo.Fingerprint)
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ListKeys(testSSHUser, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, []string{thirdKey})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteKeysKeepsUnrecognised(c *gc.C) {
	firstKey := sshtesting.ValidKeyOne.Key + " user@host"
	writeAuthKeysFile(c, []string{firstKey, sshtesting.ValidKeyTwo.Key, "invalid-key"}, authKeysFile)
	err := ssh.DeleteKeys(testSSHUser, "user@host")
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ReadAuthorisedKeys(testSSHUser, authKeysFile)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, []string{"invalid-key", sshtesting.ValidKeyTwo.Key})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteNonExistentComment(c *gc.C) {
	firstKey := sshtesting.ValidKeyOne.Key + " user@host"
	writeAuthKeysFile(c, []string{firstKey}, authKeysFile)
	err := ssh.DeleteKeys(testSSHUser, "someone@host")
	c.Assert(err, gc.ErrorMatches, "cannot delete non existent key: someone@host")
}

func (s *AuthorisedKeysKeysSuite) TestDeleteNonExistentFingerprint(c *gc.C) {
	firstKey := sshtesting.ValidKeyOne.Key + " user@host"
	writeAuthKeysFile(c, []string{firstKey}, authKeysFile)
	err := ssh.DeleteKeys(testSSHUser, sshtesting.ValidKeyTwo.Fingerprint)
	c.Assert(err, gc.ErrorMatches, "cannot delete non existent key: "+sshtesting.ValidKeyTwo.Fingerprint)
}

func (s *AuthorisedKeysKeysSuite) TestDeleteLastKeyForbidden(c *gc.C) {
	keys := []string{
		sshtesting.ValidKeyOne.Key + " user@host",
		sshtesting.ValidKeyTwo.Key + " yetanotheruser@host",
	}
	writeAuthKeysFile(c, keys, authKeysFile)
	err := ssh.DeleteKeys(testSSHUser, "user@host", sshtesting.ValidKeyTwo.Fingerprint)
	c.Assert(err, gc.ErrorMatches, "cannot delete all keys")
}

func (s *AuthorisedKeysKeysSuite) TestReplaceKeys(c *gc.C) {
	firstKey := sshtesting.ValidKeyOne.Key + " user@host"
	anotherKey := sshtesting.ValidKeyTwo.Key
	writeAuthKeysFile(c, []string{firstKey, anotherKey}, authKeysFile)

	// replaceKey is created without a comment so test that
	// ReplaceKeys handles keys without comments. This is
	// because existing keys may not have a comment and
	// ReplaceKeys is used to rewrite the entire authorized_keys
	// file when adding new keys.
	replaceKey := sshtesting.ValidKeyThree.Key
	err := ssh.ReplaceKeys(testSSHUser, replaceKey)
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ListKeys(testSSHUser, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, []string{replaceKey})
}

func (s *AuthorisedKeysKeysSuite) TestReplaceKeepsUnrecognised(c *gc.C) {
	writeAuthKeysFile(c, []string{sshtesting.ValidKeyOne.Key, "invalid-key"}, authKeysFile)
	anotherKey := sshtesting.ValidKeyTwo.Key + " anotheruser@host"
	err := ssh.ReplaceKeys(testSSHUser, anotherKey)
	c.Assert(err, jc.ErrorIsNil)
	actual, err := ssh.ReadAuthorisedKeys(testSSHUser, authKeysFile)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actual, gc.DeepEquals, []string{"invalid-key", anotherKey})
}

func (s *AuthorisedKeysKeysSuite) TestEnsureJujuComment(c *gc.C) {
	sshKey := sshtesting.ValidKeyOne.Key
	for _, test := range []struct {
		key      string
		expected string
	}{
		{"invalid-key", "invalid-key"},
		{sshKey, sshKey + " Juju:sshkey"},
		{sshKey + " user@host", sshKey + " Juju:user@host"},
		{sshKey + " Juju:user@host", sshKey + " Juju:user@host"},
		{sshKey + " " + sshKey[3:5], sshKey + " Juju:" + sshKey[3:5]},
	} {
		actual := ssh.EnsureJujuComment(test.key)
		c.Assert(actual, gc.Equals, test.expected)
	}
}

func (s *AuthorisedKeysKeysSuite) TestSplitAuthorisedKeys(c *gc.C) {
	sshKey := sshtesting.ValidKeyOne.Key
	for _, test := range []struct {
		keyData  string
		expected []string
	}{
		{"", nil},
		{sshKey, []string{sshKey}},
		{sshKey + "\n", []string{sshKey}},
		{sshKey + "\n\n", []string{sshKey}},
		{sshKey + "\n#comment\n", []string{sshKey}},
		{sshKey + "\n #comment\n", []string{sshKey}},
		{sshKey + "\ninvalid\n", []string{sshKey, "invalid"}},
	} {
		actual := ssh.SplitAuthorisedKeys(test.keyData)
		c.Assert(actual, gc.DeepEquals, test.expected)
	}
}

func b64decode(c *gc.C, s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	c.Assert(err, jc.ErrorIsNil)
	return b
}

func (s *AuthorisedKeysKeysSuite) TestParseAuthorisedKey(c *gc.C) {
	for i, test := range []struct {
		line    string
		key     []byte
		comment string
		err     string
	}{{
		line: sshtesting.ValidKeyOne.Key,
		key:  b64decode(c, strings.Fields(sshtesting.ValidKeyOne.Key)[1]),
	}, {
		line:    sshtesting.ValidKeyOne.Key + " a b c",
		key:     b64decode(c, strings.Fields(sshtesting.ValidKeyOne.Key)[1]),
		comment: "a b c",
	}, {
		line: "ssh-xsa blah",
		err:  "invalid authorized_key \"ssh-xsa blah\"",
	}, {
		// options should be skipped
		line: `no-pty,principals="\"",command="\!" ` + sshtesting.ValidKeyOne.Key,
		key:  b64decode(c, strings.Fields(sshtesting.ValidKeyOne.Key)[1]),
	}, {
		line: "ssh-rsa",
		err:  "invalid authorized_key \"ssh-rsa\"",
	}, {
		line: sshtesting.ValidKeyOne.Key + " line1\nline2",
		err:  "newline in authorized_key \".*",
	}} {
		c.Logf("test %d: %s", i, test.line)
		ak, err := ssh.ParseAuthorisedKey(test.line)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(ak, gc.Not(gc.IsNil))
			c.Assert(ak.Key, gc.DeepEquals, test.key)
			c.Assert(ak.Comment, gc.Equals, test.comment)
		}
	}
}

func (s *AuthorisedKeysKeysSuite) TestConcatAuthorisedKeys(c *gc.C) {
	for _, test := range []struct{ a, b, result string }{
		{"a", "", "a"},
		{"", "b", "b"},
		{"a", "b", "a\nb"},
		{"a\n", "b", "a\nb"},
	} {
		c.Check(ssh.ConcatAuthorisedKeys(test.a, test.b), gc.Equals, test.result)
	}
}

func (s *AuthorisedKeysKeysSuite) TestAddKeysToFileToDifferentFiles(c *gc.C) {
	key1 := sshtesting.ValidKeyOne.Key + " user@host"
	err := ssh.AddKeysToFile(testSSHUser, alternativeKeysFile2, []string{key1})
	c.Assert(err, jc.ErrorIsNil)

	list1, err := ssh.ListKeysFromFile(testSSHUser, alternativeKeysFile2, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(list1, gc.DeepEquals, []string{key1})

	key2 := sshtesting.ValidKeyTwo.Key + " user@host"
	err = ssh.AddKeysToFile(testSSHUser, alternativeKeysFile3, []string{key2})
	c.Assert(err, jc.ErrorIsNil)

	list2, err := ssh.ListKeysFromFile(testSSHUser, alternativeKeysFile3, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(list2, gc.DeepEquals, []string{key2})
}

func (s *AuthorisedKeysKeysSuite) TestAddKeysToFileMultipleKeys(c *gc.C) {
	key1 := sshtesting.ValidKeyOne.Key + " user@host"
	key2 := sshtesting.ValidKeyTwo.Key + " alice@host"
	err := ssh.AddKeysToFile(testSSHUser, alternativeKeysFile2, []string{key1, key2})
	c.Assert(err, jc.ErrorIsNil)

	list, err := ssh.ListKeysFromFile(testSSHUser, alternativeKeysFile2, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(list, jc.DeepEquals, []string{key1, key2})
}

func (s *AuthorisedKeysKeysSuite) TestDeleteAllKeysFromFile(c *gc.C) {
	key1 := sshtesting.ValidKeyOne.Key + " user@host"
	writeAuthKeysFile(c, []string{key1}, alternativeKeysFile2)

	err := ssh.DeleteKeysFromFile(testSSHUser, alternativeKeysFile2, []string{sshtesting.ValidKeyOne.Fingerprint})
	c.Assert(err, jc.ErrorIsNil)

	emptyList, err := ssh.ListKeysFromFile(testSSHUser, alternativeKeysFile2, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(emptyList, gc.HasLen, 0)
}

func (s *AuthorisedKeysKeysSuite) TestDeleteSomeKeysFromFile(c *gc.C) {
	key1 := sshtesting.ValidKeyOne.Key + " user@host"
	key2 := sshtesting.ValidKeyTwo.Key + " alice@host"
	key3 := sshtesting.ValidKeyThree.Key + " bob@host"
	writeAuthKeysFile(c, []string{key1, key2, key3}, alternativeKeysFile2)

	err := ssh.DeleteKeysFromFile(testSSHUser, alternativeKeysFile2, []string{sshtesting.ValidKeyTwo.Fingerprint})
	c.Assert(err, jc.ErrorIsNil)

	keys, err := ssh.ListKeysFromFile(testSSHUser, alternativeKeysFile2, ssh.FullKeys)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(keys, gc.HasLen, 2)
	c.Assert(keys, jc.DeepEquals, []string{key1, key3})
}
