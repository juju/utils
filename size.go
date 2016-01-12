// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/juju/errors"
)

// ParseSize parses the string as a size, in mebibytes.
//
// The string must be a is a non-negative number with
// an optional multiplier suffix (M, G, T, P, E, Z, or Y).
// If the suffix is not specified, "M" is implied.
func ParseSize(str string) (MB uint64, err error) {
	// Find the first non-digit/period:
	i := strings.IndexFunc(str, func(r rune) bool {
		return r != '.' && !unicode.IsDigit(r)
	})
	var multiplier float64 = 1
	if i > 0 {
		suffix := str[i:]
		multiplier = 0
		for j := 0; j < len(sizeSuffixes); j++ {
			base := string(sizeSuffixes[j])
			// M, MB, or MiB are all valid.
			switch suffix {
			case base, base + "B", base + "iB":
				multiplier = float64(sizeSuffixMultiplier(j))
				break
			}
		}
		if multiplier == 0 {
			return 0, errors.Errorf("invalid multiplier suffix %q, expected one of %s", suffix, []byte(sizeSuffixes))
		}
		str = str[:i]
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil || val < 0 {
		return 0, errors.Errorf("expected a non-negative number, got %q", str)
	}
	val *= multiplier
	return uint64(math.Ceil(val)), nil
}

var sizeSuffixes = "MGTPEZY"

func sizeSuffixMultiplier(i int) int {
	return 1 << uint(i*10)
}

// sizeTracker tracks the number of bytes passing through
// a read/write func.
type sizeTracker struct {
	rawOp func(data []byte) (n int, err error)
	size  int64
}

// Size returns the number of bytes read so far.
func (st sizeTracker) Size() int64 {
	return st.size
}

// Reset sets the number of bytes read to zero.
func (st *sizeTracker) Reset() {
	st.size = 0
}

// op implements io.Reader/io.Writer.
func (st *sizeTracker) op(data []byte) (n int, err error) {
	n, err = st.rawOp(data)
	if err != nil {
		// No trace because some callers, like ioutil.ReadAll(), won't work.
		return n, err
	}
	st.size += int64(n)
	return n, nil
}

// SizingReader is a reader that tracks the number of bytes read.
type SizingReader struct {
	sizeTracker
}

// NewSizingReader wraps the provided reader in a SizingReader.
func NewSizingReader(raw io.Reader) *SizingReader {
	return &SizingReader{
		sizeTracker: sizeTracker{
			rawOp: raw.Read,
		},
	}
}

// Read implements io.Reader.
func (sr *SizingReader) Read(data []byte) (n int, err error) {
	return sr.op(data)
}

// SizingWriter is a reader that tracks the number of bytes read.
type SizingWriter struct {
	sizeTracker
}

// NewSizingWriter wraps the provided reader in a SizingWriter.
func NewSizingWriter(raw io.Writer) *SizingWriter {
	return &SizingWriter{
		sizeTracker: sizeTracker{
			rawOp: raw.Write,
		},
	}
}

// Write implements io.Writer.
func (sw *SizingWriter) Write(data []byte) (n int, err error) {
	return sw.op(data)
}
