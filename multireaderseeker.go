package utils

import (
	"io"
	"os"

	"github.com/juju/errors"
)

type multiReaderSeeker struct {
	readers []io.ReadSeeker
	sizes   []int64
	index   int
	offset  int64 // offset in current file.
}

// NewMultiReaderSeeker returns an io.ReadSeeker that combines
// all the given readers into a single one. It assumes that
// all the seekers are initially positioned at the start.
func NewMultiReaderSeeker(readers ...io.ReadSeeker) io.ReadSeeker {
	r := &multiReaderSeeker{
		readers: readers,
		sizes:   make([]int64, len(readers)),
	}
	for i := range r.sizes {
		r.sizes[i] = -1
	}
	return r
}

// Read implements io.Reader.Read.
func (r *multiReaderSeeker) Read(buf []byte) (int, error) {
	if r.index >= len(r.readers) {
		return 0, io.EOF
	}
	n, err := r.readers[r.index].Read(buf)
	r.offset += int64(n)
	if err == io.EOF {
		// We've got to the end of a file so we
		// now know how big it is.
		r.sizes[r.index] = r.offset
		r.index++
		r.offset = 0
		err = nil
	}
	return n, err
}

// Seek implements io.Seeker.Seek. It can only be used to seek to the
// start.
func (r *multiReaderSeeker) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == 0 {
		// Easy special case: seeking to the very start.
		for _, reader := range r.readers {
			if _, err := reader.Seek(0, 0); err != nil {
				return 0, errors.Trace(err)
			}
		}
	}
	// Find all the file sizes because we may need them.
	// Technically we could avoid some seeks here, but
	// it's probably not worth it.
	for i, size := range r.sizes {
		if size != -1 {
			continue
		}
		size, err := r.readers[i].Seek(0, 2)
		if err != nil {
			return 0, errors.Annotate(err, "cannot seek to end")
		}
		r.sizes[i] = size
	}
	switch whence {
	case os.SEEK_SET:
		// Nothing to do.
	case os.SEEK_END:
		totalSize := int64(0)
		for _, size := range r.sizes {
			totalSize += size
		}
		offset = totalSize + offset
	case os.SEEK_CUR:
		size := int64(0)
		for i := 0; i < r.index; i++ {
			size += r.sizes[i]
		}
		offset = size + r.offset + offset
	default:
		return 0, errors.New("unknown whence value in seek")
	}
	if offset < 0 {
		return 0, errors.New("negative position")
	}
	start := int64(0)
	for i, size := range r.sizes {
		if offset < start+size {
			var err error
			_, err = r.readers[i].Seek(offset-start, 0)
			if err != nil {
				return 0, errors.Annotate(err, "cannot seek into file")
			}
			// Make sure that all the subsequent readers are
			// positioned at the start.
			for _, rr := range r.readers[i+1:] {
				if _, err := rr.Seek(0, os.SEEK_SET); err != nil {
					return 0, errors.Annotate(err, "cannot seek to start of file")
				}
			}
			r.index = i
			r.offset = offset - start
			return offset, nil
		}
		start += size
	}
	r.index = len(r.readers)
	r.offset = offset - start
	return offset, nil
}
