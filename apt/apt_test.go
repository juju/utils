// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package apt_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
	"github.com/juju/utils/apt"
	"github.com/juju/utils/proxy"
)

type AptSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&AptSuite{})

func (s *AptSuite) TestOnePackage(c *gc.C) {
	cmdChan := s.HookCommandOutput(&apt.CommandOutput, []byte{}, nil)
	err := apt.GetInstall("test-package")
	c.Assert(err, gc.IsNil)
	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, []string{
		"apt-get", "--option=Dpkg::Options::=--force-confold",
		"--option=Dpkg::options::=--force-unsafe-io", "--assume-yes", "--quiet",
		"install", "test-package",
	})
	c.Assert(cmd.Env[len(cmd.Env)-1], gc.Equals, "DEBIAN_FRONTEND=noninteractive")
}

func (s *AptSuite) TestAptGetPreparePackages(c *gc.C) {
	packagesList := apt.GetPreparePackages([]string{"lxc", "bridge-utils", "git", "mongodb"}, "precise")
	c.Assert(packagesList[0], gc.DeepEquals, []string{"--target-release", "precise-updates/cloud-tools", "lxc", "mongodb"})
	c.Assert(packagesList[1], gc.DeepEquals, []string{"bridge-utils", "git"})
}

func (s *AptSuite) TestAptGetError(c *gc.C) {
	const expected = `E: frobnicator failure detected`
	state := os.ProcessState{}
	cmdError := &exec.ExitError{&state}

	cmdChan := s.HookCommandOutput(&apt.CommandOutput, []byte(expected), error(cmdError))
	err := apt.GetInstall("foo")
	c.Assert(err, gc.ErrorMatches, "apt-get failed: exit status 0")
	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, []string{
		"apt-get", "--option=Dpkg::Options::=--force-confold",
		"--option=Dpkg::options::=--force-unsafe-io", "--assume-yes", "--quiet",
		"install", "foo",
	})
}

type mockExitStatuser int

func (m mockExitStatuser) ExitStatus() int {
	return int(m)
}

func (s *AptSuite) TestAptGetUnexpectedError(c *gc.C) {
	cmdError := errors.New("whatever")
	_ = s.HookCommandOutput(&apt.CommandOutput, []byte{}, cmdError)
	err := apt.GetInstall("test-package")
	c.Assert(err, gc.ErrorMatches, "apt-get failed: unexpected error type \\*errors\\.Err: whatever")
}

func (s *AptSuite) TestAptGetRetry(c *gc.C) {
	var calls int
	state := os.ProcessState{}
	cmdError := &exec.ExitError{&state}
	s.PatchValue(apt.InstallAttemptStrategy, utils.AttemptStrategy{Min: 3})
	s.PatchValue(apt.ProcessStateSys, func(*os.ProcessState) interface{} {
		calls++
		return mockExitStatuser(100 + calls - 1) // 100 is retried
	})

	_ = s.HookCommandOutput(&apt.CommandOutput, []byte{}, cmdError)
	err := apt.GetInstall("test-package")
	c.Check(err, gc.ErrorMatches, "apt-get failed: exit status.*")
	c.Check(calls, gc.Equals, 2) // only 2 because second exit status != 100
}

func (s *AptSuite) TestAptGetRetryDoesNotCallCombinedOutputTwice(c *gc.C) {
	const minRetries = 3
	var calls int
	state := os.ProcessState{}
	cmdError := &exec.ExitError{&state}
	s.PatchValue(apt.InstallAttemptStrategy, utils.AttemptStrategy{Min: minRetries})
	s.PatchValue(apt.ProcessStateSys, func(*os.ProcessState) interface{} {
		return mockExitStatuser(100) // retry each time.
	})
	s.PatchValue(&apt.CommandOutput, func(cmd *exec.Cmd) ([]byte, error) {
		calls++
		// Replace the command path and args so it's a no-op.
		cmd.Path = ""
		cmd.Args = []string{"version"}
		// Call the real cmd.CombinedOutput to simulate better what
		// happens in production. See also http://pad.lv/1394524.
		output, err := cmd.CombinedOutput()
		if _, ok := err.(*exec.Error); err != nil && !ok {
			c.Check(err, gc.ErrorMatches, "exec: Stdout already set")
			c.Fatalf("CommandOutput called twice unexpectedly")
		}
		return output, cmdError
	})
	err := apt.GetInstall("test-package")
	c.Check(err, gc.ErrorMatches, "apt-get failed: exit status.*")
	c.Check(calls, gc.Equals, minRetries)
}

func (s *AptSuite) TestConfigProxyEmpty(c *gc.C) {
	cmdChan := s.HookCommandOutput(&apt.CommandOutput, []byte{}, nil)
	out, err := apt.ConfigProxy()
	c.Assert(err, gc.IsNil)
	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, []string{
		"apt-config", "dump", "Acquire::http::Proxy",
		"Acquire::https::Proxy", "Acquire::ftp::Proxy",
	})
	c.Assert(out, gc.Equals, "")
}

func (s *AptSuite) TestConfigProxyConfigured(c *gc.C) {
	const expected = `Acquire::http::Proxy "10.0.3.1:3142";
Acquire::https::Proxy "false";`
	cmdChan := s.HookCommandOutput(&apt.CommandOutput, []byte(expected), nil)
	out, err := apt.ConfigProxy()
	c.Assert(err, gc.IsNil)
	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, []string{
		"apt-config", "dump", "Acquire::http::Proxy",
		"Acquire::https::Proxy", "Acquire::ftp::Proxy",
	})
	c.Assert(out, gc.Equals, expected)
}

func (s *AptSuite) TestDetectAptProxy(c *gc.C) {
	const output = `CommandLine::AsString "apt-config dump";
Acquire::http::Proxy  "10.0.3.1:3142";
Acquire::https::Proxy "false";
Acquire::ftp::Proxy "none";
Acquire::magic::Proxy "none";
`
	_ = s.HookCommandOutput(&apt.CommandOutput, []byte(output), nil)

	proxySettings, err := apt.DetectProxies()
	c.Assert(err, gc.IsNil)
	c.Assert(proxySettings, gc.DeepEquals, proxy.Settings{
		Http:  "10.0.3.1:3142",
		Https: "false",
		Ftp:   "none",
	})
}

func (s *AptSuite) TestDetectAptProxyNone(c *gc.C) {
	_ = s.HookCommandOutput(&apt.CommandOutput, []byte{}, nil)
	proxySettings, err := apt.DetectProxies()
	c.Assert(err, gc.IsNil)
	c.Assert(proxySettings, gc.DeepEquals, proxy.Settings{})
}

func (s *AptSuite) TestDetectAptProxyPartial(c *gc.C) {
	const output = `CommandLine::AsString "apt-config dump";
Acquire::http::Proxy  "10.0.3.1:3142";
Acquire::ftp::Proxy "here-it-is";
Acquire::magic::Proxy "none";
`
	_ = s.HookCommandOutput(&apt.CommandOutput, []byte(output), nil)

	proxySettings, err := apt.DetectProxies()
	c.Assert(err, gc.IsNil)
	c.Assert(proxySettings, gc.DeepEquals, proxy.Settings{
		Http: "10.0.3.1:3142",
		Ftp:  "here-it-is",
	})
}

func (s *AptSuite) TestAptProxyContentEmpty(c *gc.C) {
	output := apt.ProxyContent(proxy.Settings{})
	c.Assert(output, gc.Equals, "")
}

func (s *AptSuite) TestAptProxyContentPartial(c *gc.C) {
	proxySettings := proxy.Settings{
		Http: "user@10.0.0.1",
	}
	output := apt.ProxyContent(proxySettings)
	expected := `Acquire::http::Proxy "user@10.0.0.1";`
	c.Assert(output, gc.Equals, expected)
}

func (s *AptSuite) TestAptProxyContentRoundtrip(c *gc.C) {
	proxySettings := proxy.Settings{
		Http:  "http://user@10.0.0.1",
		Https: "https://user@10.0.0.1",
		Ftp:   "ftp://user@10.0.0.1",
	}
	output := apt.ProxyContent(proxySettings)

	s.HookCommandOutput(&apt.CommandOutput, []byte(output), nil)

	detected, err := apt.DetectProxies()
	c.Assert(err, gc.IsNil)
	c.Assert(detected, gc.DeepEquals, proxySettings)
}

func (s *AptSuite) TestConfigProxyConfiguredFilterOutput(c *gc.C) {
	const (
		output = `CommandLine::AsString "apt-config dump";
Acquire::http::Proxy  "10.0.3.1:3142";
Acquire::https::Proxy "false";`
		expected = `Acquire::http::Proxy  "10.0.3.1:3142";
Acquire::https::Proxy "false";`
	)
	cmdChan := s.HookCommandOutput(&apt.CommandOutput, []byte(output), nil)
	out, err := apt.ConfigProxy()
	c.Assert(err, gc.IsNil)
	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, []string{
		"apt-config", "dump", "Acquire::http::Proxy",
		"Acquire::https::Proxy", "Acquire::ftp::Proxy",
	})
	c.Assert(out, gc.Equals, expected)
}

func (s *AptSuite) TestConfigProxyError(c *gc.C) {
	const expected = `E: frobnicator failure detected`
	cmdError := fmt.Errorf("error")
	cmdExpectedError := fmt.Errorf("apt-config failed: error")
	cmdChan := s.HookCommandOutput(&apt.CommandOutput, []byte(expected), cmdError)
	out, err := apt.ConfigProxy()
	c.Assert(err, gc.DeepEquals, cmdExpectedError)
	cmd := <-cmdChan
	c.Assert(cmd.Args, gc.DeepEquals, []string{
		"apt-config", "dump", "Acquire::http::Proxy",
		"Acquire::https::Proxy", "Acquire::ftp::Proxy",
	})
	c.Assert(out, gc.Equals, "")
}

func (s *AptSuite) patchDpkgQuery(c *gc.C, installed bool) {
	rc := 0
	if !installed {
		rc = 1
	}
	content := fmt.Sprintf("#!/bin/bash --norc\nexit %v", rc)
	patchExecutable(s, c.MkDir(), "dpkg-query", content)
}

func (s *AptSuite) TestIsPackageInstalled(c *gc.C) {
	s.patchDpkgQuery(c, true)
	c.Assert(apt.IsPackageInstalled("foo-bar"), jc.IsTrue)
}

func (s *AptSuite) TestIsPackageNotInstalled(c *gc.C) {
	s.patchDpkgQuery(c, false)
	c.Assert(apt.IsPackageInstalled("foo-bar"), jc.IsFalse)
}

type EnvironmentPatcher interface {
	PatchEnvironment(name, value string)
}

func patchExecutable(patcher EnvironmentPatcher, dir, execName, script string) {
	patcher.PatchEnvironment("PATH", dir)
	filename := filepath.Join(dir, execName)
	ioutil.WriteFile(filename, []byte(script), 0755)
}
