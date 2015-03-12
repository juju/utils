// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filepath

import (
	"strings"
)

// A substantial portion of this code comes from the Go stdlib code.

const (
	WindowsSeparator     = '\\' // OS-specific path separator
	WindowsListSeparator = ';'  // OS-specific path list separator
)

type windowsRenderer struct{}

// Base implements Renderer.
func (ur windowsRenderer) Base(path string) string {
	return Base(WindowsSeparator, ur.VolumeName, path)
}

// Clean implements Renderer.
func (ur windowsRenderer) Clean(path string) string {
	return Clean(WindowsSeparator, ur.VolumeName, path)
}

// Dir implements Renderer.
func (ur windowsRenderer) Dir(path string) string {
	return Dir(WindowsSeparator, ur.VolumeName, path)
}

// Ext implements Renderer.
func (windowsRenderer) Ext(path string) string {
	return Ext(WindowsSeparator, path)
}

// FromSlash implements Renderer.
func (windowsRenderer) FromSlash(path string) string {
	return FromSlash(WindowsSeparator, path)
}

// IsAbs implements Renderer.
func (windowsRenderer) IsAbs(path string) bool {
	l := volumeNameLen(path)
	if l == 0 {
		return false
	}
	path = path[l:]
	if path == "" {
		return false
	}
	return isSlash(path[0])
}

// Join implements Renderer.
func (ur windowsRenderer) Join(path ...string) string {
	return Join(WindowsSeparator, ur.VolumeName, path...)
}

// Match implements Renderer.
func (windowsRenderer) Match(pattern, name string) (matched bool, err error) {
	return Match(WindowsSeparator, pattern, name)
}

// Split implements Renderer.
func (ur windowsRenderer) Split(path string) (dir, file string) {
	return Split(WindowsSeparator, ur.VolumeName, path)
}

// SplitList implements Renderer.
func (windowsRenderer) SplitList(path string) []string {
	if path == "" {
		return []string{}
	}

	// Split path, respecting but preserving quotes.
	list := []string{}
	start := 0
	quo := false
	for i := 0; i < len(path); i++ {
		switch c := path[i]; {
		case c == '"':
			quo = !quo
		case c == WindowsListSeparator && !quo:
			list = append(list, path[start:i])
			start = i + 1
		}
	}
	list = append(list, path[start:])

	// Remove quotes.
	for i, s := range list {
		if strings.Contains(s, `"`) {
			list[i] = strings.Replace(s, `"`, ``, -1)
		}
	}

	return list
}

// ToSlash implements Renderer.
func (windowsRenderer) ToSlash(path string) string {
	return ToSlash(WindowsSeparator, path)
}

// VolumeName implements Renderer.
func (windowsRenderer) VolumeName(path string) string {
	return path[:volumeNameLen(path)]
}

func isSlash(c uint8) bool {
	return c == WindowsSeparator || c == '/'
}

// volumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
func volumeNameLen(path string) int {
	if len(path) < 2 {
		return 0
	}
	// with drive letter
	c := path[0]
	if path[1] == ':' && ('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') {
		return 2
	}
	// is it UNC
	if l := len(path); l >= 5 && isSlash(path[0]) && isSlash(path[1]) &&
		!isSlash(path[2]) && path[2] != '.' {
		// first, leading `\\` and next shouldn't be `\`. its server name.
		for n := 3; n < l-1; n++ {
			// second, next '\' shouldn't be repeated.
			if isSlash(path[n]) {
				n++
				// third, following something characters. its share name.
				if !isSlash(path[n]) {
					if path[n] == '.' {
						break
					}
					for ; n < l; n++ {
						if isSlash(path[n]) {
							break
						}
					}
					return n
				}
				break
			}
		}
	}
	return 0
}
