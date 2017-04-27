// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	cryptossh "golang.org/x/crypto/ssh"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/ssh"
)

var (
	testCommand     = []string{"echo", "$abc"}
	testCommandFlat = `echo "\$abc"`
)

type sshServer struct {
	cfg      *cryptossh.ServerConfig
	listener net.Listener
	client   *cryptossh.Client
}

func newServer(c *gc.C, serverConfig cryptossh.ServerConfig) (*sshServer, cryptossh.PublicKey) {
	private, _, err := ssh.GenerateKey("test-server")
	c.Assert(err, jc.ErrorIsNil)
	key, err := cryptossh.ParsePrivateKey([]byte(private))
	c.Assert(err, jc.ErrorIsNil)
	server := &sshServer{cfg: &serverConfig}
	server.cfg.AddHostKey(key)
	server.listener, err = net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, jc.ErrorIsNil)
	return server, key.PublicKey()
}

func (s *sshServer) run(c *gc.C) {
	netconn, err := s.listener.Accept()
	c.Assert(err, jc.ErrorIsNil)
	defer netconn.Close()
	conn, chans, reqs, err := cryptossh.NewServerConn(netconn, s.cfg)
	c.Assert(err, jc.ErrorIsNil)
	s.client = cryptossh.NewClient(conn, chans, reqs)
	var wg sync.WaitGroup
	defer wg.Wait()
	sessionChannels := s.client.HandleChannelOpen("session")
	c.Assert(sessionChannels, gc.NotNil)
	for newChannel := range sessionChannels {
		c.Assert(newChannel.ChannelType(), gc.Equals, "session")
		channel, reqs, err := newChannel.Accept()
		c.Assert(err, jc.ErrorIsNil)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer channel.Close()
			for req := range reqs {
				switch req.Type {
				case "exec":
					c.Assert(req.WantReply, jc.IsTrue)
					n := binary.BigEndian.Uint32(req.Payload[:4])
					command := string(req.Payload[4 : n+4])
					c.Assert(command, gc.Equals, testCommandFlat)
					req.Reply(true, nil)
					channel.Write([]byte("abc value\n"))
					_, err := channel.SendRequest("exit-status", false, cryptossh.Marshal(&struct{ n uint32 }{0}))
					c.Check(err, jc.ErrorIsNil)
					return
				default:
					c.Errorf("Unexpected request type: %v", req.Type)
					return
				}
			}
		}()
	}
}

func newClient(c *gc.C) (*ssh.GoCryptoClient, cryptossh.PublicKey) {
	private, _, err := ssh.GenerateKey("test-client")
	c.Assert(err, jc.ErrorIsNil)
	key, err := cryptossh.ParsePrivateKey([]byte(private))
	client, err := ssh.NewGoCryptoClient(key)
	c.Assert(err, jc.ErrorIsNil)
	return client, key.PublicKey()
}

type SSHGoCryptoCommandSuite struct {
	testing.IsolationSuite
	client         ssh.Client
	knownHostsFile string
}

var _ = gc.Suite(&SSHGoCryptoCommandSuite{})

func (s *SSHGoCryptoCommandSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	generateKeyRestorer := overrideGenerateKey(c)
	s.AddCleanup(func(*gc.C) { generateKeyRestorer.Restore() })
	client, err := ssh.NewGoCryptoClient()
	c.Assert(err, jc.ErrorIsNil)
	s.client = client

	s.knownHostsFile = filepath.Join(c.MkDir(), "known_hosts")
	ssh.SetGoCryptoKnownHostsFile(s.knownHostsFile)
}

func (s *SSHGoCryptoCommandSuite) TestNewGoCryptoClient(c *gc.C) {
	_, err := ssh.NewGoCryptoClient()
	c.Assert(err, jc.ErrorIsNil)
	private, _, err := ssh.GenerateKey("test-client")
	c.Assert(err, jc.ErrorIsNil)
	key, err := cryptossh.ParsePrivateKey([]byte(private))
	c.Assert(err, jc.ErrorIsNil)
	_, err = ssh.NewGoCryptoClient(key)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *SSHGoCryptoCommandSuite) TestClientNoKeys(c *gc.C) {
	client, err := ssh.NewGoCryptoClient()
	c.Assert(err, jc.ErrorIsNil)
	cmd := client.Command("0.1.2.3", []string{"echo", "123"}, nil)
	_, err = cmd.Output()
	c.Assert(err, gc.ErrorMatches, "no private keys available")
	defer ssh.ClearClientKeys()
	err = ssh.LoadClientKeys(c.MkDir())
	c.Assert(err, jc.ErrorIsNil)

	s.PatchValue(ssh.SSHDial, func(network, address string, cfg *cryptossh.ClientConfig) (*cryptossh.Client, error) {
		return nil, errors.New("ssh.Dial failed")
	})
	cmd = client.Command("0.1.2.3", []string{"echo", "123"}, nil)
	_, err = cmd.Output()
	// error message differs based on whether using cgo or not
	c.Assert(err, gc.ErrorMatches, "ssh.Dial failed")
}

func (s *SSHGoCryptoCommandSuite) TestCommand(c *gc.C) {
	client, clientKey := newClient(c)
	server, serverKey := newServer(c, cryptossh.ServerConfig{})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksNo)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	checkedKey := false
	server.cfg.PublicKeyCallback = func(conn cryptossh.ConnMetadata, pubkey cryptossh.PublicKey) (*cryptossh.Permissions, error) {
		c.Check(pubkey, gc.DeepEquals, clientKey)
		checkedKey = true
		return nil, nil
	}
	go server.run(c)
	out, err := cmd.Output()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), gc.Equals, "abc value\n")
	c.Assert(checkedKey, jc.IsTrue)

	knownHosts, err := ioutil.ReadFile(s.knownHostsFile)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(knownHosts), gc.Equals, fmt.Sprintf(
		"[127.0.0.1]:%d %s",
		serverPort,
		cryptossh.MarshalAuthorizedKey(serverKey)),
	)
}

func (s *SSHGoCryptoCommandSuite) TestCopy(c *gc.C) {
	client, err := ssh.NewGoCryptoClient()
	c.Assert(err, jc.ErrorIsNil)
	err = client.Copy([]string{"0.1.2.3:b", c.MkDir()}, nil)
	c.Assert(err, gc.ErrorMatches, `scp command is not implemented \(OpenSSH scp not available in PATH\)`)
}

func (s *SSHGoCryptoCommandSuite) TestProxyCommand(c *gc.C) {
	realNetcat, err := exec.LookPath("nc")
	if err != nil {
		c.Skip("skipping test, couldn't find netcat: %v")
		return
	}
	netcat := filepath.Join(c.MkDir(), "nc")
	err = ioutil.WriteFile(netcat, []byte("#!/bin/sh\necho $0 \"$@\" > $0.args && exec "+realNetcat+" \"$@\""), 0755)
	c.Assert(err, jc.ErrorIsNil)

	client, _ := newClient(c)
	server, _ := newServer(c, cryptossh.ServerConfig{})
	var opts ssh.Options
	port := server.listener.Addr().(*net.TCPAddr).Port
	opts.SetProxyCommand(netcat, "-q0", "%h", "%p")
	opts.SetPort(port)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	server.cfg.PublicKeyCallback = func(_ cryptossh.ConnMetadata, pubkey cryptossh.PublicKey) (*cryptossh.Permissions, error) {
		return nil, nil
	}
	go server.run(c)
	out, err := cmd.Output()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), gc.Equals, "abc value\n")
	// Ensure the proxy command was executed with the appropriate arguments.
	data, err := ioutil.ReadFile(netcat + ".args")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(data), gc.Equals, fmt.Sprintf("%s -q0 127.0.0.1 %v\n", netcat, port))
}

func (s *SSHGoCryptoCommandSuite) TestStrictHostChecksYes(c *gc.C) {
	server, _ := newServer(c, cryptossh.ServerConfig{NoClientAuth: true})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	go server.run(c)

	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksYes)
	client, _ := newClient(c)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	_, err := cmd.Output()
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf(
		"ssh: handshake failed: no ssh-rsa host key is known for 127.0.0.1:%d and you have requested strict checking",
		serverPort,
	))
	_, err = os.Stat(s.knownHostsFile)
	c.Assert(err, jc.Satisfies, os.IsNotExist)
}

func (s *SSHGoCryptoCommandSuite) TestStrictHostChecksAskNonTerminal(c *gc.C) {
	server, _ := newServer(c, cryptossh.ServerConfig{NoClientAuth: true})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	go server.run(c)

	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksAsk)
	client, _ := newClient(c)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	_, err := cmd.Output()
	c.Assert(err, gc.ErrorMatches, "ssh: handshake failed: not running in a terminal, cannot prompt for verification")
	_, err = os.Stat(s.knownHostsFile)
	c.Assert(err, jc.Satisfies, os.IsNotExist)
}

func (s *SSHGoCryptoCommandSuite) TestStrictHostChecksAskTerminalYes(c *gc.C) {
	var readLineWriter mockReadLineWriter
	ssh.PatchTerminal(&s.CleanupSuite, &readLineWriter)
	readLineWriter.addLine("")
	readLineWriter.addLine("yes")

	server, serverKey := newServer(c, cryptossh.ServerConfig{NoClientAuth: true})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	go server.run(c)

	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksAsk)
	client, _ := newClient(c)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	_, err := cmd.Output()
	c.Assert(err, jc.ErrorIsNil)

	knownHosts, err := ioutil.ReadFile(s.knownHostsFile)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(knownHosts), gc.Equals, fmt.Sprintf(
		"[127.0.0.1]:%d %s",
		serverPort,
		cryptossh.MarshalAuthorizedKey(serverKey),
	))

	c.Assert(readLineWriter.written.String(), gc.Equals, fmt.Sprintf(`
The authenticity of host '127.0.0.1:%[1]d (127.0.0.1:%[1]d)' can't be established.
ssh-rsa key fingerprint is %[2]s.
Are you sure you want to continue connecting (yes/no)? Please type 'yes' or 'no': `[1:],
		serverPort, cryptossh.FingerprintSHA256(serverKey)))
}

func (s *SSHGoCryptoCommandSuite) TestStrictHostChecksAskTerminalNo(c *gc.C) {
	var readLineWriter mockReadLineWriter
	ssh.PatchTerminal(&s.CleanupSuite, &readLineWriter)
	readLineWriter.addLine("no")

	server, serverKey := newServer(c, cryptossh.ServerConfig{NoClientAuth: true})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	go server.run(c)

	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksAsk)
	client, _ := newClient(c)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	_, err := cmd.Output()
	c.Assert(err, gc.ErrorMatches, "ssh: handshake failed: Host key verification failed.")

	_, err = os.Stat(s.knownHostsFile)
	c.Assert(err, jc.Satisfies, os.IsNotExist)

	c.Assert(readLineWriter.written.String(), gc.Equals, fmt.Sprintf(`
The authenticity of host '127.0.0.1:%[1]d (127.0.0.1:%[1]d)' can't be established.
ssh-rsa key fingerprint is %[2]s.
Are you sure you want to continue connecting (yes/no)? `[1:],
		serverPort, cryptossh.FingerprintSHA256(serverKey)))
}

func (s *SSHGoCryptoCommandSuite) TestStrictHostChecksNoMismatch(c *gc.C) {
	var readLineWriter mockReadLineWriter
	ssh.PatchTerminal(&s.CleanupSuite, &readLineWriter)

	server, serverKey := newServer(c, cryptossh.ServerConfig{NoClientAuth: true})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	go server.run(c)

	// Write a mismatching key to the known_hosts file. Even with
	// StrictHostChecksNo, we should be verifying against an existing
	// host key.
	alternativeRSAKey, err := generateRSAKey(rand.Reader)
	c.Assert(err, jc.ErrorIsNil)
	alternativePublicKey, err := cryptossh.NewPublicKey(alternativeRSAKey.Public())
	c.Assert(err, jc.ErrorIsNil)
	err = ioutil.WriteFile(s.knownHostsFile, []byte(fmt.Sprintf(
		"[127.0.0.1]:%d %s",
		serverPort,
		cryptossh.MarshalAuthorizedKey(alternativePublicKey),
	)), 0600)
	c.Assert(err, jc.ErrorIsNil)

	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksNo)
	client, _ := newClient(c)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	_, err = cmd.Output()
	c.Assert(err, gc.ErrorMatches, "ssh: handshake failed: knownhosts: key mismatch")

	c.Assert(readLineWriter.written.String(), gc.Matches, fmt.Sprintf(`
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now \(man-in-the-middle attack\)!
It is also possible that a host key has just been changed.
The fingerprint for the ssh-rsa key sent by the remote host is
%s.
Please contact your system administrator.
Add correct host key in .*/known_hosts to get rid of this message.
Offending ssh-rsa key in .*/known_hosts:1
`[1:], regexp.QuoteMeta(cryptossh.FingerprintSHA256(serverKey))))
}

func (s *SSHGoCryptoCommandSuite) TestStrictHostChecksDifferentKeyTypes(c *gc.C) {
	var readLineWriter mockReadLineWriter
	ssh.PatchTerminal(&s.CleanupSuite, &readLineWriter)

	server, serverKey := newServer(c, cryptossh.ServerConfig{NoClientAuth: true})
	serverPort := server.listener.Addr().(*net.TCPAddr).Port
	go server.run(c)

	// Write a mismatching key to the known_hosts file with a different
	// key type. Even with StrictHostChecksNo, we should be verifying
	// against an existing host key.
	dsaKey, err := generateDSAKey(rand.Reader)
	c.Assert(err, jc.ErrorIsNil)
	alternativePublicKey, err := cryptossh.NewPublicKey(&dsaKey.PublicKey)
	c.Assert(err, jc.ErrorIsNil)
	err = ioutil.WriteFile(s.knownHostsFile, []byte(fmt.Sprintf(
		"[127.0.0.1]:%d %s",
		serverPort,
		cryptossh.MarshalAuthorizedKey(alternativePublicKey),
	)), 0600)
	c.Assert(err, jc.ErrorIsNil)

	var opts ssh.Options
	opts.SetPort(serverPort)
	opts.SetStrictHostKeyChecking(ssh.StrictHostChecksNo)
	client, _ := newClient(c)
	cmd := client.Command("127.0.0.1", testCommand, &opts)
	_, err = cmd.Output()
	c.Assert(err, gc.ErrorMatches, "ssh: handshake failed: knownhosts: key mismatch")

	c.Assert(readLineWriter.written.String(), gc.Matches, fmt.Sprintf(`
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now \(man-in-the-middle attack\)!
It is also possible that a host key has just been changed.
The fingerprint for the ssh-rsa key sent by the remote host is
%s.
Please contact your system administrator.
Add correct host key in .*/known_hosts to get rid of this message.
Host was previously using different host key algorithms:
 - ssh-dss key in .*/known_hosts:1
`[1:], regexp.QuoteMeta(cryptossh.FingerprintSHA256(serverKey))))
}

type mockReadLineWriter struct {
	testing.Stub
	lines   []string
	written bytes.Buffer
}

func (m *mockReadLineWriter) addLine(line string) {
	m.lines = append(m.lines, line)
}

func (m *mockReadLineWriter) ReadLine() (string, error) {
	m.MethodCall(m, "ReadLine")
	if len(m.lines) == 0 {
		return "", io.EOF
	}
	line := m.lines[0]
	m.lines = m.lines[1:]
	return line, nil
}

func (m *mockReadLineWriter) Write(data []byte) (int, error) {
	m.MethodCall(m, "Write", data)
	return m.written.Write(data)
}
