// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

var _ = gc.Suite(&usernameSuite{})

type usernameSuite struct {
	testing.IsolationSuite
}

func (s *usernameSuite) TestResolveUsername(c *gc.C) {
	type test struct {
		userEnv  string
		sudoEnv  string
		userOS   string
		expected string
		err      string
	}
	tests := []test{{
		userEnv:  "someone",
		sudoEnv:  "notroot",
		userOS:   "other",
		expected: "someone",
	}, {
		userOS:   "other",
		expected: "other",
	}, {
		userEnv:  "root",
		expected: "root",
	}, {
		userEnv:  "root",
		sudoEnv:  "other",
		expected: "other",
	}, {
		err: "failed to determine username for namespace: oh noes",
	}}

	resolveUsername := func(t test) (string, error) {
		if t.err != "" {
			return "", errors.Errorf(t.err)
		}

		var funcs []func() (string, error)
		if t.userEnv != "" {
			funcs = append(funcs, func() (string, error) {
				return t.userEnv, nil
			})
		}
		if t.userOS != "" {
			funcs = append(funcs, func() (string, error) {
				return t.userOS, nil
			})
		}

		resolveSudo := func(username string) string {
			return utils.ResolveSudoByFunc(username, func(string) string {
				return t.sudoEnv
			})
		}

		return utils.ResolveUsername(resolveSudo, funcs...)
	}

	for i, test := range tests {
		c.Logf("test %d: %v", i, test)

		username, err := resolveUsername(test)

		if test.err == "" {
			if c.Check(err, jc.ErrorIsNil) {
				c.Check(username, gc.Equals, test.expected)
			}
		} else {
			c.Check(err, gc.ErrorMatches, test.err)
		}
	}
}
