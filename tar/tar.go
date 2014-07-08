// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package tar

import (
	"archive/tar"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarFiles creates a tar archive at targetPath holding the files listed
// in fileList. If compress is true, the archive will also be gzip
// compressed.
// TarFiles writes a tar stream into target holding the files listed
// in fileList. strip will be removed from the beginning of all the paths
// when stored (much like gnu tar -C option)
func TarFiles(fileList []string, target io.Writer, strip string) (shaSum string, err error) {
	shahash := sha1.New()
	if err := tarAndHashFiles(fileList, target, strip, shahash); err != nil {
		return "", err
	}
	// we use a base64 encoded sha1 hash, because this is the hash
	// used by RFC 3230 Digest headers in http responses
	encodedHash := base64.StdEncoding.EncodeToString(shahash.Sum(nil))
	return encodedHash, nil
}

func tarAndHashFiles(fileList []string, target io.Writer, strip string, hashw io.Writer) (err error) {
	checkClose := func(w io.Closer) {
		if closeErr := w.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing backup file: %v", closeErr)
		}
	}

	w := io.MultiWriter(target, hashw)
	tarw := tar.NewWriter(w)
	defer checkClose(tarw)
	for _, ent := range fileList {
		if err := writeContents(ent, strip, tarw); err != nil {
			return fmt.Errorf("backup failed: %v", err)
		}
	}
	return nil
}

// writeContents creates an entry for the given file
// or directory in the given tar archive.
func writeContents(fileName, strip string, tarw *tar.Writer) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	fInfo, err := f.Stat()
	if err != nil {
		return err
	}
	h, err := tar.FileInfoHeader(fInfo, "")
	if err != nil {
		return fmt.Errorf("cannot create tar header for %q: %v", fileName, err)
	}
	h.Name = filepath.ToSlash(strings.TrimPrefix(fileName, strip))
	if err := tarw.WriteHeader(h); err != nil {
		return fmt.Errorf("cannot write header for %q: %v", fileName, err)
	}
	if !fInfo.IsDir() {
		if _, err := io.Copy(tarw, f); err != nil {
			return fmt.Errorf("failed to write %q: %v", fileName, err)
		}
		return nil
	}
	if !strings.HasSuffix(fileName, string(os.PathSeparator)) {
		fileName = fileName + string(os.PathSeparator)
	}

	for {
		names, err := f.Readdirnames(100)
		// this can happen if there are less than 100 names remaining
		if err == io.EOF && len(names) == 0 {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error reading directory %q: %v", fileName, err)
		}
		for _, name := range names {
			if err := writeContents(filepath.Join(fileName, name), strip, tarw); err != nil {
				return err
			}
		}
	}

}

// UntarFiles will extract the contents of tarFile using
// outputFolder as root
func UntarFiles(tarFile io.Reader, outputFolder string) error {
	tr := tar.NewReader(tarFile)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return fmt.Errorf("failed while reading tar header: %v", err)
		}
		fullPath := filepath.Join(outputFolder, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			if err = os.MkdirAll(fullPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("cannot extract directory %q: %v", fullPath, err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("failed while reading tar contents: %v", err)
		}
		fh, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("some of the tar contents cannot be written to disk: %v", err)
		}
		_, err = io.Copy(fh, tr)
		if err != nil {
			fh.Close()
			return fmt.Errorf("failed while reading tar contents: %v", err)
		}
		err = fh.Chmod(os.FileMode(hdr.Mode))
		fh.Close()
		if err != nil {
			return fmt.Errorf("cannot set proper mode on file %q: %v", fullPath, err)
		}

	}
	return nil
}
