// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package utils

import (
	stdfmt "fmt"
	"regexp"
	"time"
)

var re = regexp.MustCompile("%.")

var tsReplacements = map[string]string{
	"Y": "%04d",
	"M": "%02d",
	"D": "%02d",
	"h": "%02d",
	"m": "%02d",
	"s": "%02d",
}

// FormatTimestamp returns a copy of the format string with date-related
// fields replaced with their values from the timestamp.
func FormatTimestamp(format string, timestamp *time.Time) string {
	if timestamp == nil {
		ts := time.Now().UTC()
		timestamp = &ts
	}
	return re.ReplaceAllStringFunc(format, func(sub string) string {
		id := sub[1:]
		fmt, ok := tsReplacements[id]
		if !ok {
			return sub
		}
		var val interface{}
		switch id {
		case "Y":
			val = timestamp.Year()
		case "M":
			val = timestamp.Month()
		case "D":
			val = timestamp.Day()
		case "h":
			val = timestamp.Hour()
		case "m":
			val = timestamp.Minute()
		case "s":
			val = timestamp.Second()
		}
		return stdfmt.Sprintf(fmt, val)
	})
}

// ParseTimestamp turns a timestamp string into a time.Time following
// the provided format.
func ParseTimestamp(format, timestamp string) *time.Time {
	var Y, M, D, h, m, s int

	fields := make([]interface{}, 0)
	format = re.ReplaceAllStringFunc(format, func(sub string) string {
		id := sub[1:]
		fmt, ok := tsReplacements[id]
		if !ok {
			return sub
		}
		switch id {
		case "Y":
			fields = append(fields, &Y)
		case "M":
			fields = append(fields, &M)
		case "D":
			fields = append(fields, &D)
		case "h":
			fields = append(fields, &h)
		case "m":
			fields = append(fields, &m)
		case "s":
			fields = append(fields, &s)
		}
		return fmt
	})

	_, err := stdfmt.Sscanf(timestamp, format, fields...)
	if err != nil {
		return nil
	}

	ts := time.Date(Y, time.Month(M), D, h, m, s, 0, time.UTC)
	return &ts
}
