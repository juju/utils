// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package manager_test

import (
	"os"
	"os/exec"
	"strings"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/proxy"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/packaging/commands"
	"github.com/juju/utils/packaging/manager"
)

var _ = gc.Suite(&ManagerSuite{})

type ManagerSuite struct {
	apt, yum manager.PackageManager
	testing.IsolationSuite
	calledCommand string
}

func (s *ManagerSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.apt = manager.NewAptPackageManager()
	s.yum = manager.NewYumPackageManager()
}

func (s *ManagerSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
}

func (s *ManagerSuite) TearDownTest(c *gc.C) {
	s.IsolationSuite.TearDownTest(c)
}

func (s *ManagerSuite) TearDownSuite(c *gc.C) {
	s.IsolationSuite.TearDownSuite(c)
}

var (
	// aptCmder is the commands.PackageCommander for apt-based
	// systems whose commands will be checked against.
	aptCmder = commands.NewAptPackageCommander()

	// yumCmder is the commands.PackageCommander for yum-based
	// systems whose commands will be checked against.
	yumCmder = commands.NewYumPackageCommander()

	// testedPackageName is the package name used in all
	// single-package testing scenarios.
	testedPackageName = "test-package"

	// testedRepoName is the repository name used in all
	// repository-related tests.
	testedRepoName = "some-repo"

	// testedProxySettings is the set of proxy settings used in
	// all proxy-related tests.
	testedProxySettings = proxy.Settings{
		Http:  "http://some-proxy.domain",
		Https: "https://some-proxy.domain",
		Ftp:   "ftp://some-proxy.domain",
	}

	// testedPackageNames is a list of package names used in all
	// multiple-package testing scenarions.
	testedPackageNames = []string{
		"first-test-package",
		"second-test-package",
		"third-test-package",
	}
)

// getMockRunCommandWithRetry returns a function with the same signature as
// RunCommandWithRetry which saves the command it recieves in the provided
// string whilst always returning no output, 0 error code and nil error.
func getMockRunCommandWithRetry(stor *string) func(string) (string, int, error) {
	return func(cmd string) (string, int, error) {
		*stor = cmd
		return "", 0, nil
	}
}

// getMockRunCommand returns a function with the same signature as RunCommand
// which saves the command it revieves in the provided string whilst always
// returning empty output and no error.
func getMockRunCommand(stor *string) func(string, ...string) (string, error) {
	return func(cmd string, args ...string) (string, error) {
		*stor = strings.Join(append([]string{cmd}, args...), " ")
		return "", nil
	}
}

// simpleTestCase is a struct containing all the information required for
// running a simple error/no error test case.
type simpleTestCase struct {
	// description of the test:
	desc string

	// the expected apt command which will get executed:
	expectedAptCmd string

	// the expected result of the given apt operation:
	expectedAptResult interface{}

	// the expected yum command which will get executed:
	expectedYumCmd string

	// the expected result of the given yum operation:
	expectedYumResult interface{}

	// the function to be applied on the package manager.
	// returns the result of the operation and the error.
	operation func(manager.PackageManager) (interface{}, error)
}

var simpleTestCases = []*simpleTestCase{
	&simpleTestCase{
		"Test install prerequisites.",
		aptCmder.InstallPrerequisiteCmd(),
		nil,
		yumCmder.InstallPrerequisiteCmd(),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.InstallPrerequisite()
		},
	},
	&simpleTestCase{
		"Test system update.",
		aptCmder.UpdateCmd(),
		nil,
		yumCmder.UpdateCmd(),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.Update()
		},
	},
	&simpleTestCase{
		"Test system upgrade.",
		aptCmder.UpgradeCmd(),
		nil,
		yumCmder.UpgradeCmd(),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.Upgrade()
		},
	},
	&simpleTestCase{
		"Test install packages.",
		aptCmder.InstallCmd(testedPackageNames...),
		nil,
		yumCmder.InstallCmd(testedPackageNames...),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.Install(testedPackageNames...)
		},
	},
	&simpleTestCase{
		"Test remove packages.",
		aptCmder.RemoveCmd(testedPackageNames...),
		nil,
		yumCmder.RemoveCmd(testedPackageNames...),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.Remove(testedPackageNames...)
		},
	},
	&simpleTestCase{
		"Test purge packages.",
		aptCmder.PurgeCmd(testedPackageNames...),
		nil,
		yumCmder.PurgeCmd(testedPackageNames...),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.Purge(testedPackageNames...)
		},
	},
	&simpleTestCase{
		"Test repository addition.",
		aptCmder.AddRepositoryCmd(testedRepoName),
		nil,
		yumCmder.AddRepositoryCmd(testedRepoName),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.AddRepository(testedRepoName)
		},
	},
	&simpleTestCase{
		"Test repository removal.",
		aptCmder.RemoveRepositoryCmd(testedRepoName),
		nil,
		yumCmder.RemoveRepositoryCmd(testedRepoName),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.RemoveRepository(testedRepoName)
		},
	},
	&simpleTestCase{
		"Test running cleanup.",
		aptCmder.CleanupCmd(),
		nil,
		yumCmder.CleanupCmd(),
		nil,
		func(pacman manager.PackageManager) (interface{}, error) {
			return nil, pacman.Cleanup()
		},
	},
}

// searchingTestCases are a couple of simple test cases which search for a
// given package; either localy or remotely, and need to be tested seperately
// on the case of their return value being a boolean.
var searchingTestCases = []*simpleTestCase{
	&simpleTestCase{
		"Test package search.",
		aptCmder.SearchCmd(testedPackageName),
		false,
		yumCmder.SearchCmd(testedPackageName),
		true,
		func(pacman manager.PackageManager) (interface{}, error) {
			return pacman.Search(testedPackageName)
		},
	},
	&simpleTestCase{
		"Test local package search.",
		aptCmder.IsInstalledCmd(testedPackageName),
		true,
		yumCmder.IsInstalledCmd(testedPackageName),
		true,
		func(pacman manager.PackageManager) (interface{}, error) {
			return pacman.IsInstalled(testedPackageName), nil
		},
	},
}

func (s *ManagerSuite) TestSimpleCases(c *gc.C) {
	s.PatchValue(&manager.RunCommand, getMockRunCommand(&s.calledCommand))
	s.PatchValue(&manager.RunCommandWithRetry, getMockRunCommandWithRetry(&s.calledCommand))

	testCases := append(simpleTestCases, searchingTestCases...)
	for i, testCase := range testCases {
		c.Logf("Simple test %d: %s", i+1, testCase.desc)

		// run for the apt PackageManager implementation:
		res, err := testCase.operation(s.apt)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(s.calledCommand, gc.Equals, testCase.expectedAptCmd)
		c.Assert(res, jc.DeepEquals, testCase.expectedAptResult)

		// run for the yum PackageManager implementation.
		res, err = testCase.operation(s.yum)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(s.calledCommand, gc.Equals, testCase.expectedYumCmd)
		c.Assert(res, jc.DeepEquals, testCase.expectedYumResult)
	}
}

func (s *ManagerSuite) TestSimpleErrorCases(c *gc.C) {
	const (
		expectedErr    = "packaging command failed: exit status 0"
		expectedErrMsg = `E: I done failed :(`
	)
	state := os.ProcessState{}
	cmdError := &exec.ExitError{&state}

	cmdChan := s.HookCommandOutput(&manager.CommandOutput, []byte(expectedErrMsg), error(cmdError))

	for i, testCase := range simpleTestCases {
		c.Logf("Error'd test %d: %s", i+1, testCase.desc)
		s.PatchValue(&manager.ProcessStateSys, func(*os.ProcessState) interface{} {
			return mockExitStatuser(i + 1)
		})

		// run for the apt PackageManager implementation:
		_, err := testCase.operation(s.apt)
		c.Assert(err, gc.ErrorMatches, expectedErr)

		cmd := <-cmdChan
		c.Assert(strings.Join(cmd.Args, " "), gc.DeepEquals, testCase.expectedAptCmd)

		// run for the yum PackageManager implementation:
		_, err = testCase.operation(s.yum)
		c.Assert(err, gc.ErrorMatches, expectedErr)

		cmd = <-cmdChan
		c.Assert(strings.Join(cmd.Args, " "), gc.DeepEquals, testCase.expectedYumCmd)
	}
}
