// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// The keyvalues package implements a set of functions for parsing key=value data,
// usually passed in as command-line parameters to juju subcommands, e.g.
// juju-set mongodb logging=true
package keyvalues

import (
	"fmt"
	"sort"
	"strings"
)

// DuplicateError signals that a duplicate key was encountered while parsing
// the input into a map.
type DuplicateError string

func (e DuplicateError) Error() string {
	return string(e)
}

// Parse parses the supplied string slice into a map mapping
// keys to values. Duplicate keys cause an error to be returned.
func Parse(src []string, allowEmptyValues bool) (map[string]string, error) {
	results := map[string]string{}
	for _, kv := range src {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf(`expected "key=value", got %q`, kv)
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if len(key) == 0 || (!allowEmptyValues && len(value) == 0) {
			return nil, fmt.Errorf(`expected "key=value", got "%s=%s"`, key, value)
		}
		if _, exists := results[key]; exists {
			return nil, DuplicateError(fmt.Sprintf("key %q specified more than once", key))
		}
		results[key] = value
	}
	return results, nil
}

// Join returns a sorted slice containing each item in the given map, as
// "key=value" strings.
func Join(m map[string]string) []string {
	var result []string
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(result)
	return result
}
