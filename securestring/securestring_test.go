// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL.
// Licensed under the AGPLv3, see LICENCE file for details.
//
// +build windows

package securestring_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	gss "github.com/juju/utils/securestring"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

type SecureStringSuite struct{}

var _ = gc.Suite(&SecureStringSuite{})

// testInputs is the set of inputs we will be using in our tests.
var testInputs []string = []string{
	"Simple",
	"A longer string",
	"A!string%with(a4239lot#of$&*special@characters{[]})",
	"Quite a very much longer string meant to push the envelope",
	"fsdafsgdfgdfgdfgdfgsdfgdgdfgdmmghnh kv dfv dj fkvjjenrwenvfvvslfvnsljfvnlsfvlnsfjlvnssdwoewivdsvmxxvsdvsdv",
}

// TestEncryptDecryptSymmetry tests whether encryption and decryption
// are perfectly symmetrical operations.
func (s *SecureStringSuite) TestEncryptDecryptSymmetry(c *gc.C) {
	for _, input := range testInputs {
		enc, err := gss.Encrypt(input)
		c.Assert(err, jc.ErrorIsNil)

		dec, err := gss.Decrypt(enc)
		c.Assert(err, jc.ErrorIsNil)

		c.Assert(dec, gc.Equals, input)
	}
}

// invokePowerShellParams is the standard set of parameters
// use to invoke a PoweShell session.
var invokePowerShellParams []string = []string{
	"-NoProfile",
	"-NonInteractive",
	"-Command",
	"try{$input|iex; exit $LastExitCode}catch{Write-Error -Message $Error[0]; exit 1}",
}

// runPowerShellCommand is a helper function which invokes a PowerShell session
// and runs the given command.
func runPowerShellCommands(cmds string) (string, error) {
	ps := exec.Command("powershell.exe", invokePowerShellParams...)

	ps.Stdin = strings.NewReader(cmds)
	stdout := &bytes.Buffer{}
	ps.Stdout = stdout

	err := ps.Run()
	if err != nil {
		return "", err
	}

	output := string(stdout.String())
	return strings.TrimSpace(output), nil
}

// TestDecryptFromCFSS tests whether the output of ConvertFrom-SecureString
// is compatible with this module's Decrypt function and can be
// succesfully decrypted.
func (s *SecureStringSuite) TestDecryptFromCFSS(c *gc.C) {
	for _, input := range testInputs {
		psenc, err := runPowerShellCommands(fmt.Sprintf("ConvertTo-SecureString \"%s\" -AsPlainText -Force | ConvertFrom-SecureString", input))
		c.Assert(err, jc.ErrorIsNil)

		dec, err := gss.Decrypt(psenc)
		c.Assert(err, jc.ErrorIsNil)

		c.Assert(dec, gc.Equals, input)
	}
}

// TestConvertEncryptedToPowerShellSS tests whether the output of the module's
// Encrypt function is compatible with PowerShell's SecureString and is accepted
// as valid input by being taken in as a System.Security.SecureString internal.
func (s *SecureStringSuite) TestConvertEncryptedToPowerShellSS(c *gc.C) {
	for _, input := range testInputs {
		enc, err := gss.Encrypt(input)
		c.Assert(err, jc.ErrorIsNil)

		psresp, err := runPowerShellCommands(fmt.Sprintf("\"%s\" | ConvertTo-SecureString", enc))
		c.Assert(err, jc.ErrorIsNil)

		c.Assert(psresp, gc.Equals, "System.Security.SecureString")
	}
}
