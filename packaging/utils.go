// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package packaging

import (
	"bytes"
	"text/template"
)

// renderTemplate is a helper function which renders a given object to a given
// remplate and returns its output as a string.
func renderTemplate(temp string, obj interface{}) string {
	var buf bytes.Buffer

	t := template.Must(template.New("").Parse(temp))
	err := t.Execute(&buf, obj)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
