// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/juju/errors"
)

// OSEnv is a snapshot of an OS environment. The order of env vars
// is not preserved nor are duplicates.
type OSEnv struct {
	// Vars contains the unordered env vars.
	Vars map[string]string
}

// NewOSEnv creates a new OSEnv, prepopulated with the initial vars.
func NewOSEnv(initial ...string) *OSEnv {
	env := &OSEnv{
		Vars: make(map[string]string),
	}
	env.update(initial)
	return env
}

// ReadOSEnv creates a new OSEnv and prepoulates it with the values
// from the current OS environment.
func ReadOSEnv() *OSEnv {
	initial := os.Environ()
	env := NewOSEnv(initial...)
	return env
}

// Names returns an unsorted list of the env var names. The list
// includes those that have empty values.
func (env OSEnv) Names() []string {
	names := make([]string, 0, len(env.Vars))
	for name := range env.Vars {
		names = append(names, name)
	}
	return names
}

// Get returns the value of the named environment variable. If it is
// not set then the empty string is returned.
func (env OSEnv) Get(name string) string {
	envVar, _ := env.Vars[name]
	return envVar
}

// Set updates the value of the named environment variable. The old
// value, if any, is returned.
func (env *OSEnv) Set(name, value string) string {
	// existing is "" when not found.
	existing, _ := env.Vars[name]
	env.Vars[name] = value
	return existing
}

// Unset ensures the named env var is removed. The old values, if any,
// are returned.
func (env *OSEnv) Unset(names ...string) []string {
	var values []string
	for _, name := range names {
		value := env.unset(name)
		values = append(values, value)
	}
	return values
}

func (env *OSEnv) unset(name string) string {
	value, ok := env.Vars[name]
	if !ok {
		return ""
	}
	delete(env.Vars, name)
	return value
}

// Update sets all the provided env vars, in order, on a copy of the
// env and returns it. If any of the provided env vars is already set
// then it is overwritten, though its original order is preserved. To
// reset an env var's order, unset it before calling Update.
func (env *OSEnv) Update(vars ...string) *OSEnv {
	copied := env.Copy()
	copied.update(vars)
	return copied
}

func (env *OSEnv) update(vars []string) {
	for _, envVar := range vars {
		name, value := SplitEnvVar(envVar)
		env.Vars[name] = value
	}
}

// Reduce filters out all env vars that don't match the provided filters
// and returns the remainder in a new OSEnv.
func (env OSEnv) Reduce(filters ...func(name string) bool) *OSEnv {
	newEnv := NewOSEnv()
	for _, name := range env.Names() {
		if filtersAnd(name, filters) {
			value := env.Get(name)
			newEnv.Set(name, value)
		}
	}
	return newEnv
}

func filtersAnd(name string, filters []func(string) bool) bool {
	matched := false
	for _, filter := range filters {
		if !filter(name) {
			return false
		}
		matched = true
	}
	return matched
}

// TODO(ericsnow) Add an equivalent method to os.ExpandEnv.

// TODO(ericsnow) Drop Copy?

// Copy returns a copy of the env.
func (env OSEnv) Copy() *OSEnv {
	return NewOSEnv(env.AsList()...)
}

// AsList copies the environment variables into a new list of raw
// env var strings. The list includes those with empty values.
func (env OSEnv) AsList() []string {
	var envVars []string
	for _, name := range env.Names() {
		value := env.Vars[name]
		envVar := JoinEnvVar(name, value)
		envVars = append(envVars, envVar)
	}
	return envVars
}

// PushOSEnv updates the current OS environment.
func PushOSEnv(env *OSEnv) error {
	for _, name := range env.Names() {
		value := env.Vars[name]
		if err := os.Setenv(name, value); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// PushOSEnvFresh updates the current OS environment after clearing it.
func PushOSEnvFresh(env *OSEnv) error {
	os.Clearenv()
	if err := PushOSEnv(env); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// TempOSEnv applies the environment generated by reducing the current
// OS env using the provided filters. It returns the new env along with
// a function to restore the old env.
func TempOSEnv(filters ...func(name string) bool) (*OSEnv, func() error, error) {
	orig := ReadOSEnv()
	env := orig.Reduce(filters...)
	if err := PushOSEnvFresh(env); err != nil {
		return nil, nil, errors.Trace(err)
	}
	restore := func() error {
		if err := PushOSEnvFresh(env); err != nil {
			return errors.Trace(err)
		}
		return nil
	}
	return env, restore, nil
}

// SplitEnvVar converts a raw env var string into a (name, value) pair.
func SplitEnvVar(envVar string) (string, string) {
	parts := strings.SplitN(envVar, "=", 2)
	if len(parts) == 1 {
		return envVar, ""
	}
	return parts[0], parts[1]
}

// JoinEnvVar converts a (name, value) pair into a raw env var string.
func JoinEnvVar(name, value string) string {
	return fmt.Sprintf("%s=%s", name, value)
}
