// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package packaging

import (
	"bytes"
	"text/template"
)

// renderTemplate is a helper function which renders a given object to a given
// template and returns its output as a string.
func renderTemplate(temp string, obj interface{}) (string, error) {
	var buf bytes.Buffer

	t, err := template.New("").Parse(temp)
	if err != nil {
		return "", err
	}

	err = t.Execute(&buf, obj)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
