// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"os"
	"strings"
)

// mergeEnvironment takes in a string array representing the desired environment
// and merges it with the current environment. On Windows, clearing the environment,
// or having missing environment variables, may lead to standard go packages not working
// (os.TempDir relies on $env:TEMP), and powershell erroring out
// Currently this function is only used for windows
func mergeEnvironment(env []string) []string {
	if env == nil {
		return nil
	}
	m := make(map[string]string)
	var tmpEnv []string
	for _, val := range os.Environ() {
		varSplit := strings.SplitN(val, "=", 2)
		m[varSplit[0]] = varSplit[1]
	}

	for _, val := range env {
		varSplit := strings.SplitN(val, "=", 2)
		m[varSplit[0]] = varSplit[1]
	}

	for key, val := range m {
		tmpEnv = append(tmpEnv, key+"="+val)
	}

	return tmpEnv
}
