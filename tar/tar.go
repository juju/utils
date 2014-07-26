// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// This package provides convenience helpers on top of archive/tar
// to be able to tar/untar files with a functionality closer
// to gnu tar command.
package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/juju/utils/hash"
)

//---------------------------
// packing

// TarFiles writes a tar stream into target holding the files listed
// in fileList. strip will be removed from the beginning of all the paths
// when stored (much like gnu tar -C option)
// Returns a Sha sum of the tar and nil if everything went well
// or empty sting and error in case of error.
// We use a base64 encoded sha1 hash, because this is the hash
// used by RFC 3230 Digest headers in http responses
func TarFiles(
	fileList []string, target io.Writer, strip string,
) (shaSum string, err error) {
	ar := Archiver{fileList, strip}

	// Write out the archive.
	proxy := hash.NewSHA1Proxy(target)
	err = ar.Write(proxy)
	if err != nil {
		return "", err
	}
	return proxy.Hash(), nil
}

//---------------------------
// unpacking

func createAndFill(filePath string, mode int64, content io.Reader) error {
	fh, err := os.Create(filePath)
	defer fh.Close()
	if err != nil {
		return fmt.Errorf("some of the tar contents cannot be written to disk: %v", err)
	}
	_, err = io.Copy(fh, content)
	if err != nil {
		return fmt.Errorf("failed while reading tar contents: %v", err)
	}
	err = fh.Chmod(os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("cannot set proper mode on file %q: %v", filePath, err)
	}
	return nil
}

// UntarFiles will extract the contents of tarFile using
// outputFolder as root
func UntarFiles(tarFile io.Reader, outputFolder string) error {
	tr := tar.NewReader(tarFile)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			return nil
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
		if err = createAndFill(fullPath, hdr.Mode, tr); err != nil {
			return fmt.Errorf("cannot extract file %q: %v", fullPath, err)
		}

	}
	return nil
}
