// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"reflect"
	"sort"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

var _ = gc.Suite(&osEnvSuite{})

type osEnvSuite struct {
	testing.IsolationSuite
}

func (*osEnvSuite) TestNewOSEnvEmpty(c *gc.C) {
	env := utils.NewOSEnv()
	vars := utils.RawEnvVars(env)

	c.Check(vars, gc.HasLen, 0)
}

func (*osEnvSuite) TestNewOSEnvInitial(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	vars := utils.RawEnvVars(env)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b"})
}

func (*osEnvSuite) TestReadOSEnv(c *gc.C) {
	// TODO(ericsnow) ???
}

func (*osEnvSuite) TestOSEnvNames(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	names := env.Names()

	sort.Strings(names)
	c.Check(names, jc.DeepEquals, []string{"x", "y"})
}

func (*osEnvSuite) TestOSEnvGetOkay(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	y := env.Get("y")

	c.Check(y, gc.Equals, "b")
}

func (*osEnvSuite) TestOSEnvGetEmpty(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=")
	y := env.Get("y")

	c.Check(y, gc.Equals, "")
}

func (*osEnvSuite) TestOSEnvGetMissing(c *gc.C) {
	env := utils.NewOSEnv("x=a")
	y := env.Get("y")

	c.Check(y, gc.Equals, "")
}

func (*osEnvSuite) TestOSEnvSetOkay(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	existing := env.Set("z", "c")
	vars := utils.RawEnvVars(env)

	c.Check(existing, gc.Equals, "")
	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b", "z": "c"})
}

func (*osEnvSuite) TestOSEnvSetExists(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	existing := env.Set("x", "c")
	vars := utils.RawEnvVars(env)

	c.Check(existing, gc.Equals, "a")
	c.Check(vars, jc.DeepEquals, map[string]string{"x": "c", "y": "b"})
}

func (*osEnvSuite) TestOSEnvSetEmpty(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	env.Set("z", "")
	vars := utils.RawEnvVars(env)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b", "z": ""})
}

func (*osEnvSuite) TestOSEnvUnsetOkay(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b", "z=c")
	existing := env.Unset("y")
	vars := utils.RawEnvVars(env)

	c.Check(existing, jc.DeepEquals, []string{"b"})
	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "z": "c"})
}

func (*osEnvSuite) TestOSEnvUnsetEmpty(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=", "z=c")
	existing := env.Unset("y")
	vars := utils.RawEnvVars(env)

	c.Check(existing, jc.DeepEquals, []string{""})
	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "z": "c"})
}

func (*osEnvSuite) TestOSEnvUnsetMissing(c *gc.C) {
	env := utils.NewOSEnv("x=a", "z=c")
	existing := env.Unset("y")
	vars := utils.RawEnvVars(env)

	c.Check(existing, jc.DeepEquals, []string{""})
	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "z": "c"})
}

func (*osEnvSuite) TestOSEnvUpdateOkay(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	updated := env.Update("z=c")
	vars := utils.RawEnvVars(updated)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b", "z": "c"})
	// Ensure they aren't linked.
	c.Check(reflect.DeepEqual(vars, env.AsMap()), jc.IsFalse)
}

func (*osEnvSuite) TestOSEnvUpdateEmpty(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	updated := env.Update("z=")
	vars := utils.RawEnvVars(updated)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b", "z": ""})
}

func (*osEnvSuite) TestOSEnvUpdateReplaceOkay(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	updated := env.Update("x=c")
	vars := utils.RawEnvVars(updated)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "c", "y": "b"})
}

func (*osEnvSuite) TestOSEnvUpdateReplaceEmpty(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	updated := env.Update("x=")
	vars := utils.RawEnvVars(updated)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "", "y": "b"})
}

func (*osEnvSuite) TestOSEnvUpdateNoEqualSign(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	updated := env.Update("z")
	vars := utils.RawEnvVars(updated)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b", "z": ""})
}

func (*osEnvSuite) TestOSEnvReduceOkay(c *gc.C) {
	filter := func(name string) bool {
		return name != "y"
	}
	env := utils.NewOSEnv("x=a", "y=b", "z=c")
	reduced := env.Reduce(filter)
	vars := utils.RawEnvVars(reduced)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "z": "c"})
	// Ensure they aren't linked.
	c.Check(reflect.DeepEqual(vars, env.AsMap()), jc.IsFalse)
}

func (*osEnvSuite) TestOSEnvReduceNoFilter(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	reduced := env.Reduce()
	vars := utils.RawEnvVars(reduced)

	c.Check(vars, gc.HasLen, 0)
}

func (*osEnvSuite) TestOSEnvReduceMultipleFilters(c *gc.C) {
	noW := func(name string) bool {
		return name != "w"
	}
	noX := func(name string) bool {
		return name != "x"
	}
	noZ := func(name string) bool {
		return name != "z"
	}
	env := utils.NewOSEnv("x=a", "y=b", "z=c")
	reduced := env.Reduce(noW, noX, noZ)
	vars := utils.RawEnvVars(reduced)

	c.Check(vars, jc.DeepEquals, map[string]string{"y": "b"})
}

func (*osEnvSuite) TestOSEnvReduceNoMatch(c *gc.C) {
	filter := func(name string) bool {
		return name == "z"
	}
	env := utils.NewOSEnv("x=a", "y=b")
	reduced := env.Reduce(filter)
	vars := utils.RawEnvVars(reduced)

	c.Check(vars, gc.HasLen, 0)
}

func (*osEnvSuite) TestOSEnvReduceEmpty(c *gc.C) {
	filter := func(name string) bool {
		return name != "x"
	}
	env := utils.NewOSEnv()
	reduced := env.Reduce(filter)
	vars := utils.RawEnvVars(reduced)

	c.Check(vars, gc.HasLen, 0)
}

func (*osEnvSuite) TestOSEnvCopy(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	copied := env.Copy()
	copied.Set("z", "c")
	vars := utils.RawEnvVars(copied)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b", "z": "c"})
	// Ensure they aren't linked.
	c.Check(reflect.DeepEqual(vars, env.AsMap()), jc.IsFalse)
}

func (*osEnvSuite) TestOSEnvAsMap(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	vars := env.AsMap()
	env.Set("z", "c")
	varsOrig := utils.RawEnvVars(env)

	c.Check(vars, jc.DeepEquals, map[string]string{"x": "a", "y": "b"})
	// Ensure they aren't linked.
	c.Check(reflect.DeepEqual(vars, varsOrig), jc.IsFalse)
}

func (*osEnvSuite) TestOSEnvAsList(c *gc.C) {
	env := utils.NewOSEnv("x=a", "y=b")
	list := env.AsList()
	env.Set("z", "c") // Ensure they aren't linked.

	c.Check(list, jc.DeepEquals, []string{"x=a", "y=b"})
}

func (*osEnvSuite) TestPush(c *gc.C) {
	// TODO(ericsnow) ???
}

func (*osEnvSuite) TestPushFresh(c *gc.C) {
	// TODO(ericsnow) ???
}

func (*osEnvSuite) TestSplitEnvVarOkay(c *gc.C) {
	name, value := utils.SplitEnvVar("x=a")

	c.Check(name, gc.Equals, "x")
	c.Check(value, gc.Equals, "a")
}

func (*osEnvSuite) TestSplitEnvVarMissingValue(c *gc.C) {
	name, value := utils.SplitEnvVar("x=")

	c.Check(name, gc.Equals, "x")
	c.Check(value, gc.Equals, "")
}

func (*osEnvSuite) TestSplitEnvVarMissingName(c *gc.C) {
	name, value := utils.SplitEnvVar("=a")

	c.Check(name, gc.Equals, "")
	c.Check(value, gc.Equals, "a")
}

func (*osEnvSuite) TestSplitEnvVarNoEqualSign(c *gc.C) {
	name, value := utils.SplitEnvVar("x")

	c.Check(name, gc.Equals, "x")
	c.Check(value, gc.Equals, "")
}

func (*osEnvSuite) TestSplitEnvVarEmpty(c *gc.C) {
	name, value := utils.SplitEnvVar("")

	c.Check(name, gc.Equals, "")
	c.Check(value, gc.Equals, "")
}

func (*osEnvSuite) TestJoinEnvVarOkay(c *gc.C) {
	envVar := utils.JoinEnvVar("x", "a")

	c.Check(envVar, gc.Equals, "x=a")
}

func (*osEnvSuite) TestJoinEnvVarMissingValue(c *gc.C) {
	envVar := utils.JoinEnvVar("x", "")

	c.Check(envVar, gc.Equals, "x=")
}

func (*osEnvSuite) TestJoinEnvVarMissingName(c *gc.C) {
	envVar := utils.JoinEnvVar("", "a")

	c.Check(envVar, gc.Equals, "=a")
}
