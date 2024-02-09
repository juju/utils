// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/juju/errors"
	"golang.org/x/crypto/ssh"
)

// ed25519GenerateKey allows for tests to patch out ed25519 key generation
var ed25519GenerateKey = ed25519.GenerateKey

// GenerateKey makes a ED25519 no-passphrase SSH capable key.
// The private key returned is encoded to ASCII using the PKCS1 encoding.
// The public key is suitable to be added into an authorized_keys file,
// and has the comment passed in as the comment part of the key.
func GenerateKey(comment string) (private, public string, err error) {
	_, privateKey, err := ed25519GenerateKey(rand.Reader)
	if err != nil {
		return "", "", errors.Trace(err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(privateKey, comment)
	if err != nil {
		return "", "", errors.Trace(err)
	}
	identity := pem.EncodeToMemory(pemBlock)

	public, err = PublicKey(identity, comment)
	if err != nil {
		return "", "", errors.Trace(err)
	}

	return string(identity), public, nil
}

// PublicKey returns the public key for any private key. The public key is
// suitable to be added into an authorized_keys file, and has the comment
// passed in as the comment part of the key.
func PublicKey(privateKey []byte, comment string) (string, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", errors.Annotate(err, "failed to load key")
	}

	auth_key := string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
	// Strip off the trailing new line so we can add a comment.
	auth_key = strings.TrimSpace(auth_key)
	public := fmt.Sprintf("%s %s\n", auth_key, comment)

	return public, nil
}
