// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/juju/errors"
	"github.com/juju/loggo/v2"
	"golang.org/x/crypto/ssh"

	"github.com/juju/utils/v4"
)

var logger = loggo.GetLogger("juju.utils.ssh")

type ListMode bool

var (
	FullKeys     ListMode = true
	Fingerprints ListMode = false
)

const (
	defaultAuthKeysFile = "authorized_keys"
)

type AuthorisedKey struct {
	Type    string
	Key     []byte
	Comment string
}

func authKeysDir(username string) (string, error) {
	homeDir, err := utils.UserHomeDir(username)
	if err != nil {
		return "", err
	}
	homeDir, err = utils.NormalizePath(homeDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".ssh"), nil
}

// ParseAuthorisedKey parses a non-comment line from an
// authorized_keys file and returns the constituent parts.
// Based on description in "man sshd".
func ParseAuthorisedKey(line string) (*AuthorisedKey, error) {
	if strings.Contains(line, "\n") {
		return nil, errors.NotValidf("newline in authorized_key %q", line)
	}
	key, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
	if err != nil {
		return nil, errors.Errorf("invalid authorized_key %q", line)
	}
	return &AuthorisedKey{
		Type:    key.Type(),
		Key:     key.Marshal(),
		Comment: comment,
	}, nil
}

// ConcatAuthorisedKeys will joing two or more authorised keys together to form
// a string based list of authorised keys that can be read by ssh programs. Keys
// joined with a newline as the separator.
func ConcatAuthorisedKeys(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if a[len(a)-1] != '\n' {
		return a + "\n" + b
	}
	return a + b
}

// SplitAuthorisedKeys extracts a key slice from the specified key data,
// by splitting the key data into lines and ignoring comments and blank lines.
func SplitAuthorisedKeys(keyData string) []string {
	var keys []string
	for _, key := range strings.Split(string(keyData), "\n") {
		key = strings.Trim(key, " \r")
		if len(key) == 0 {
			continue
		}
		if key[0] == '#' {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}

func readAuthorisedKeys(username, filename string) ([]string, error) {
	keyDir, err := authKeysDir(username)
	if err != nil {
		return nil, err
	}
	sshKeyFile := filepath.Join(keyDir, filename)
	logger.Debugf("reading authorised keys file %s", sshKeyFile)
	keyData, err := ioutil.ReadFile(sshKeyFile)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, errors.Annotate(err, "reading ssh authorised keys file")
	}
	var keys []string
	for _, key := range strings.Split(string(keyData), "\n") {
		if len(strings.Trim(key, " \r")) == 0 {
			continue
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func writeAuthorisedKeys(username, filename string, keys []string) error {
	keyDir, err := authKeysDir(username)
	if err != nil {
		return err
	}
	err = os.MkdirAll(keyDir, os.FileMode(0755))
	if err != nil {
		return errors.Annotate(err, "cannot create ssh key directory")
	}
	keyData := strings.Join(keys, "\n") + "\n"

	// Get perms to use on auth keys file
	sshKeyFile := filepath.Join(keyDir, filename)
	perms := os.FileMode(0644)
	info, err := os.Stat(sshKeyFile)
	if err == nil {
		perms = info.Mode().Perm()
	}

	logger.Debugf("writing authorised keys file %s", sshKeyFile)
	err = utils.AtomicWriteFile(sshKeyFile, []byte(keyData), perms)
	if err != nil {
		return err
	}

	// TODO (wallyworld) - what to do on windows (if anything)
	// TODO(dimitern) - no need to use user.Current() if username
	// is "" - it will use the current user anyway.
	if runtime.GOOS != "windows" {
		// Ensure the resulting authorised keys file has its ownership
		// set to the specified username.
		var u *user.User
		if username == "" {
			u, err = user.Current()
		} else {
			u, err = user.Lookup(username)
		}
		if err != nil {
			return err
		}
		// chown requires ints but user.User has strings for windows.
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return err
		}
		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return err
		}
		err = os.Chown(sshKeyFile, uid, gid)
		if err != nil {
			return err
		}
	}
	return nil
}

// We need a mutex because updates to the authorised keys file are done by
// reading the contents, updating, and writing back out. So only one caller
// at a time can use either Add, Delete, List.
var keysMutex sync.Mutex

// AddKeys adds the specified ssh keys to the authorized_keys file for user.
// Returns an error if there is an issue with *any* of the supplied keys.
func AddKeys(user string, newKeys ...string) error {
	keysMutex.Lock()
	defer keysMutex.Unlock()
	existingKeys, err := readAuthorisedKeys(user, defaultAuthKeysFile)
	if err != nil {
		return err
	}
	return addKeys(user, defaultAuthKeysFile, newKeys, existingKeys)
}

// DeleteKeys removes the specified ssh keys from the authorized ssh keys file for user.
// keyIds may be either key comments or fingerprints.
// Returns an error if there is an issue with *any* of the keys to delete.
func DeleteKeys(user string, keyIds ...string) error {
	keysMutex.Lock()
	defer keysMutex.Unlock()
	existingKeys, err := readAuthorisedKeys(user, defaultAuthKeysFile)
	if err != nil {
		return err
	}
	return deleteKeys(user, defaultAuthKeysFile, existingKeys, keyIds, false)
}

// ReplaceKeys writes the specified ssh keys to the authorized_keys file for user,
// replacing any that are already there.
// Returns an error if there is an issue with *any* of the supplied keys.
func ReplaceKeys(user string, newKeys ...string) error {
	keysMutex.Lock()
	defer keysMutex.Unlock()

	existingKeyData, err := readAuthorisedKeys(user, defaultAuthKeysFile)
	if err != nil {
		return err
	}
	var existingNonKeyLines []string
	for _, line := range existingKeyData {
		_, _, err := KeyFingerprint(line)
		if err != nil {
			existingNonKeyLines = append(existingNonKeyLines, line)
		}
	}
	return writeAuthorisedKeys(user, defaultAuthKeysFile, append(existingNonKeyLines, newKeys...))
}

// ListKeys returns either the full keys or key comments from the authorized ssh keys file for user.
func ListKeys(user string, mode ListMode) ([]string, error) {
	keysMutex.Lock()
	defer keysMutex.Unlock()
	keyData, err := readAuthorisedKeys(user, defaultAuthKeysFile)
	if err != nil {
		return nil, err
	}
	return listKeys(keyData, mode)
}

// Any ssh key added to the authorised keys list by Juju will have this prefix.
// This allows Juju to know which keys have been added externally and any such keys
// will always be retained by Juju when updating the authorised keys file.
const JujuCommentPrefix = "Juju:"

func EnsureJujuComment(key string) string {
	ak, err := ParseAuthorisedKey(key)
	// Just return an invalid key as is.
	if err != nil {
		logger.Warningf("invalid Juju ssh key %s: %v", key, err)
		return key
	}
	if ak.Comment == "" {
		return key + " " + JujuCommentPrefix + "sshkey"
	} else {
		// Add the Juju prefix to the comment if necessary.
		if !strings.HasPrefix(ak.Comment, JujuCommentPrefix) {
			commentIndex := strings.LastIndex(key, ak.Comment)
			return key[:commentIndex] + JujuCommentPrefix + ak.Comment
		}
	}
	return key
}

// AddKeysToFile adds the specified ssh keys to the specified file for user.
// Returns an error if there is an issue with *any* of the supplied keys.
func AddKeysToFile(user, file string, newKeys []string) error {
	keysMutex.Lock()
	defer keysMutex.Unlock()
	existingKeys, err := readAuthorisedKeys(user, file)
	if err != nil {
		return err
	}
	return addKeys(user, file, newKeys, existingKeys)
}

// DeleteKeysFromFile removes the specified ssh keys from the authorized ssh keys file for user.
// keyIds may be either key comments or fingerprints.
// Returns an error if there is an issue with *any* of the keys to delete.
//
// Unlike DeleteKeys, this version can delete ALL keys from the target file.
func DeleteKeysFromFile(user, file string, keyIds []string) error {
	keysMutex.Lock()
	defer keysMutex.Unlock()
	existingKeys, err := readAuthorisedKeys(user, file)
	if err != nil {
		return err
	}
	return deleteKeys(user, file, existingKeys, keyIds, true)
}

// ListKeys returns either the full keys or key comments from the authorized ssh keys file for user.
func ListKeysFromFile(user, file string, mode ListMode) ([]string, error) {
	keysMutex.Lock()
	defer keysMutex.Unlock()
	keyData, err := readAuthorisedKeys(user, file)
	if err != nil {
		return nil, err
	}
	return listKeys(keyData, mode)
}

func addKeys(user, file string, newKeys, existingKeys []string) error {
	for _, newKey := range newKeys {
		fingerprint, comment, err := KeyFingerprint(newKey)
		if err != nil {
			return err
		}
		if comment == "" {
			return errors.Errorf("cannot add ssh key without comment")
		}
		for _, key := range existingKeys {
			existingFingerprint, existingComment, err := KeyFingerprint(key)
			if err != nil {
				// Only log a warning if the unrecognised key line is not a comment.
				if key[0] != '#' {
					logger.Warningf("invalid existing ssh key %q: %v", key, err)
				}
				continue
			}
			if existingFingerprint == fingerprint {
				return errors.Errorf("cannot add duplicate ssh key: %v", fingerprint)
			}
			if existingComment == comment {
				return errors.Errorf("cannot add ssh key with duplicate comment: %v", comment)
			}
		}
	}
	sshKeys := append(existingKeys, newKeys...)
	return writeAuthorisedKeys(user, file, sshKeys)
}

func deleteKeys(user, file string, existingKeys, keyIdsToDelete []string, deleteAll bool) error {
	// Build up a map of keys indexed by fingerprint, and fingerprints indexed by comment
	// so we can easily get the key represented by each keyId, which may be either a fingerprint
	// or comment.
	var keysToWrite []string
	var sshKeys = make(map[string]string)
	var keyComments = make(map[string]string)
	for _, key := range existingKeys {
		fingerprint, comment, err := KeyFingerprint(key)
		if err != nil {
			logger.Debugf("keeping unrecognised existing ssh key %q: %v", key, err)
			keysToWrite = append(keysToWrite, key)
			continue
		}
		sshKeys[fingerprint] = key
		if comment != "" {
			keyComments[comment] = fingerprint
		}
	}
	for _, keyId := range keyIdsToDelete {
		// assume keyId may be a fingerprint
		fingerprint := keyId
		_, ok := sshKeys[keyId]
		if !ok {
			// keyId is a comment
			fingerprint, ok = keyComments[keyId]
		}
		if !ok {
			return errors.Errorf("cannot delete non existent key: %v", keyId)
		}
		delete(sshKeys, fingerprint)
	}
	for _, key := range sshKeys {
		keysToWrite = append(keysToWrite, key)
	}
	if len(keysToWrite) == 0 && !deleteAll {
		return errors.Errorf("cannot delete all keys")
	}
	return writeAuthorisedKeys(user, file, keysToWrite)
}

func listKeys(existingKeys []string, mode ListMode) ([]string, error) {
	var keys []string
	for _, key := range existingKeys {
		fingerprint, comment, err := KeyFingerprint(key)
		if err != nil {
			// Only log a warning if the unrecognised key line is not a comment.
			if key[0] != '#' {
				logger.Warningf("ignoring invalid ssh key %q: %v", key, err)
			}
			continue
		}
		if mode == FullKeys {
			keys = append(keys, key)
		} else {
			shortKey := fingerprint
			if comment != "" {
				shortKey += fmt.Sprintf(" (%s)", comment)
			}
			keys = append(keys, shortKey)
		}
	}
	return keys, nil
}
