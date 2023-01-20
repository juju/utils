// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	mathrand "math/rand"
	"strings"

	"github.com/juju/errors"
	"golang.org/x/crypto/ssh"
)

// PublicKey returns the public key for any private key. The public key is
// suitable to be added into an authorized_keys file, and has the comment
// passed in as the comment part of the key.
func PublicKey(privateKey []byte, comment string) (string, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", errors.Annotate(err, "failed to load key")
	}

	authKey := string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
	// Strip off the trailing new line so we can add a comment.
	authKey = strings.TrimSpace(authKey)
	public := fmt.Sprintf("%s %s\n", authKey, comment)

	return public, nil
}

// GenerateKey generated an Ed25519 no-passphrase SSH capable key. The private
// key is returned in an OpenSSH compatible format. The public key is suitable
// to be added into an authorized_keys file. The comment is encoded into the
// private key, and passed into the comment part of the public key
func GenerateKey(comment string) (private, public string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", errors.Trace(err)
	}

	publicKey, _ := ssh.NewPublicKey(pub)
	public = string(ssh.MarshalAuthorizedKey(publicKey))
	// insert comment to end of the public key in the comment space
	public = fmt.Sprintf("%s %s\n", strings.TrimSpace(public), comment)

	identity := pem.EncodeToMemory(
		&pem.Block{
			Type:  "OPENSSH PRIVATE KEY",
			Bytes: MarshalEd25519PrivateKey(priv, pub, comment),
		},
	)
	return string(identity), public, nil
}

// MarshalEd25519PrivateKey formats an Ed25519 private key into a OpenSSH
// compatible format
//
// NOTE(jack-w-shaw, 2023-01-20) Ideally we would a standard library function
// to do this, as we can with RSA and ECDSA algorithms. But this doesn't seem
// possible for Ed25519 as an exception. My implementation is based on
// https://github.com/mikesmitty/edkey
func MarshalEd25519PrivateKey(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, comment string) []byte {
	// Add our key header (followed by a null byte)
	out := append([]byte("openssh-key-v1"), 0)

	// Set our check ints
	ci := mathrand.Uint32()

	keyData := struct {
		Check1  uint32
		Check2  uint32
		Keytype string
		Pub     []byte
		Priv    []byte
		Comment string
		Pad     []byte `ssh:"rest"`
	}{
		Check1:  ci,
		Check2:  ci,
		Keytype: ssh.KeyAlgoED25519,
		Pub:     []byte(publicKey),
		Priv:    []byte(privateKey),
		Comment: comment,
	}

	// Add some padding to match the encryption block size within PrivKeyBlock (without Pad field)
	// 8 doesn't match the documentation, but that's what ssh-keygen uses for unencrypted keys.
	bs := 8
	blockLen := len(ssh.Marshal(keyData))
	padLen := (bs - (blockLen % bs)) % bs
	keyData.Pad = make([]byte, padLen)

	// Padding is a sequence of bytes like: 1, 2, 3...
	for i := 0; i < padLen; i++ {
		keyData.Pad[i] = byte(i + 1)
	}

	// Generate the pubkey prefix "\0\0\0\nssh-ed25519\0\0\0 "
	prefix := []byte{0x0, 0x0, 0x0, 0x0b}
	prefix = append(prefix, []byte(ssh.KeyAlgoED25519)...)
	prefix = append(prefix, 0x0, 0x0, 0x0, 0x20)

	keyMeta := struct {
		CipherName   string
		KdfName      string
		KdfOpts      string
		NumKeys      uint32
		PubKey       []byte
		PrivKeyBlock []byte
	}{
		// Don't encrypt key
		CipherName:   "none",
		KdfName:      "none",
		KdfOpts:      "",
		NumKeys:      1,
		PubKey:       append(prefix, []byte(publicKey)...),
		PrivKeyBlock: ssh.Marshal(keyData),
	}

	out = append(out, ssh.Marshal(keyMeta)...)
	return out
}
