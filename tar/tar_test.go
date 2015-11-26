// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package tar

import (
	"archive/tar"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	stdtesting "testing"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"
)

func TestPackage(t *stdtesting.T) {
	gc.TestingT(t)
}

var _ = gc.Suite(&TarSuite{})

type TarSuite struct {
	testing.IsolationSuite
	cwd       string
	testFiles []string
}

func (t *TarSuite) SetUpTest(c *gc.C) {
	t.cwd = c.MkDir()
	t.IsolationSuite.SetUpTest(c)
}

func (t *TarSuite) createTestFiles(c *gc.C) {
	tarDirE := filepath.Join(t.cwd, "TarDirectoryEmpty")
	err := os.Mkdir(tarDirE, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarDirP := filepath.Join(t.cwd, "TarDirectoryPopulated")
	err = os.Mkdir(tarDirP, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarlink1 := filepath.Join(t.cwd, "TarLink")
	err = os.Symlink(tarDirP, tarlink1)
	c.Check(err, gc.IsNil)

	tarSubFile1 := filepath.Join(tarDirP, "TarSubFile1")
	tarSubFile1Handle, err := os.Create(tarSubFile1)
	c.Check(err, gc.IsNil)
	tarSubFile1Handle.WriteString("TarSubFile1")
	tarSubFile1Handle.Close()

	tarSublink1 := filepath.Join(tarDirP, "TarSubLink")
	err = os.Symlink(tarSubFile1, tarSublink1)
	c.Check(err, gc.IsNil)

	tarSubDir := filepath.Join(tarDirP, "TarDirectoryPopulatedSubDirectory")
	err = os.Mkdir(tarSubDir, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarFile1 := filepath.Join(t.cwd, "TarFile1")
	tarFile1Handle, err := os.Create(tarFile1)
	c.Check(err, gc.IsNil)
	tarFile1Handle.WriteString("TarFile1")
	tarFile1Handle.Close()

	tarFile2 := filepath.Join(t.cwd, "TarFile2")
	tarFile2Handle, err := os.Create(tarFile2)
	c.Check(err, gc.IsNil)
	tarFile2Handle.WriteString("TarFile2")
	tarFile2Handle.Close()
	t.testFiles = []string{tarDirE, tarDirP, tarlink1, tarFile1, tarFile2}

}

func (t *TarSuite) removeTestFiles(c *gc.C) {
	for _, removable := range t.testFiles {
		err := os.RemoveAll(removable)
		c.Assert(err, gc.IsNil)
	}
}

type expectedTarContents struct {
	Name string
	Body string
}

var testExpectedTarContents = []expectedTarContents{
	{"TarDirectoryEmpty", ""},
	{"TarDirectoryPopulated", ""},
	{"TarLink", ""},
	{"TarDirectoryPopulated/TarSubFile1", "TarSubFile1"},
	{"TarDirectoryPopulated/TarSubLink", ""},
	{"TarDirectoryPopulated/TarDirectoryPopulatedSubDirectory", ""},
	{"TarFile1", "TarFile1"},
	{"TarFile2", "TarFile2"},
}

// Assert thar contents checks that the tar reader provided contains the
// Expected files
// expectedContents: is a slice of the filenames with relative paths that are
// expected to be on the tar file
// tarFile: is the path of the file to be checked
func (t *TarSuite) assertTarContents(c *gc.C, expectedContents []expectedTarContents,
	tarFile io.Reader) {
	tr := tar.NewReader(tarFile)
	tarContents := make(map[string]string)
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		c.Assert(err, gc.IsNil)
		buf, err := ioutil.ReadAll(tr)
		c.Assert(err, gc.IsNil)
		tarContents[hdr.Name] = string(buf)
	}
	for _, expectedContent := range expectedContents {
		fullExpectedContent := strings.TrimPrefix(expectedContent.Name, string(os.PathSeparator))
		body, ok := tarContents[fullExpectedContent]
		c.Log(tarContents)
		c.Log(expectedContents)
		c.Log(fmt.Sprintf("checking for presence of %q on tar file", fullExpectedContent))
		c.Assert(ok, gc.Equals, true)
		if expectedContent.Body != "" {
			c.Log("Also checking the file contents")
			c.Assert(body, gc.Equals, expectedContent.Body)
		}
	}

}

func (t *TarSuite) assertFilesWhereUntared(c *gc.C,
	expectedContents []expectedTarContents,
	tarOutputFolder string) {
	tarContents := make(map[string]string)
	var walkFn filepath.WalkFunc
	walkFn = func(path string, finfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fileName := strings.TrimPrefix(path, tarOutputFolder)
		fileName = strings.TrimPrefix(fileName, string(os.PathSeparator))
		c.Log(fileName)
		if fileName == "" {
			return nil
		}
		if finfo.IsDir() || finfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			tarContents[fileName] = ""
		} else {
			readable, err := os.Open(path)
			if err != nil {
				return err
			}
			defer readable.Close()
			buf, err := ioutil.ReadAll(readable)
			c.Assert(err, gc.IsNil)
			tarContents[fileName] = string(buf)
		}
		return nil
	}
	filepath.Walk(tarOutputFolder, walkFn)
	for _, expectedContent := range expectedContents {
		fullExpectedContent := strings.TrimPrefix(expectedContent.Name, string(os.PathSeparator))
		expectedPath := filepath.Join(tarOutputFolder, fullExpectedContent)
		_, err := os.Lstat(expectedPath)
		c.Assert(err, gc.Equals, nil)
		body, ok := tarContents[fullExpectedContent]
		c.Log(fmt.Sprintf("checking for presence of %q on untar files", fullExpectedContent))
		c.Assert(ok, gc.Equals, true)
		if expectedContent.Body != "" {
			c.Log("Also checking the file contents")
			c.Assert(body, gc.Equals, expectedContent.Body)
		}
	}

}

func shaSumFile(c *gc.C, fileToSum io.Reader) string {
	shahash := sha1.New()
	_, err := io.Copy(shahash, fileToSum)
	c.Assert(err, gc.IsNil)
	return base64.StdEncoding.EncodeToString(shahash.Sum(nil))
}

// Tar
func (t *TarSuite) TestTarFiles(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	shaSum, err := TarFiles(t.testFiles, &outputTar, trimPath)
	c.Check(err, gc.IsNil)
	outputBytes := outputTar.Bytes()
	fileShaSum := shaSumFile(c, bytes.NewBuffer(outputBytes))
	c.Assert(shaSum, gc.Equals, fileShaSum)
	t.removeTestFiles(c)
	t.assertTarContents(c, testExpectedTarContents, bytes.NewBuffer(outputBytes))
}

func (t *TarSuite) TestSymlinksTar(c *gc.C) {
	tarDirP := filepath.Join(t.cwd, "TarDirectory")
	err := os.Mkdir(tarDirP, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarlink1 := filepath.Join(t.cwd, "TarLink")
	err = os.Symlink(tarDirP, tarlink1)
	c.Check(err, gc.IsNil)
	testFiles := []string{tarDirP, tarlink1}

	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err = TarFiles(testFiles, &outputTar, trimPath)
	c.Check(err, gc.IsNil)

	outputBytes := outputTar.Bytes()
	tr := tar.NewReader(bytes.NewBuffer(outputBytes))
	symlinks := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		c.Assert(err, gc.IsNil)
		if hdr.Typeflag == tar.TypeSymlink {
			symlinks += 1
			c.Assert(hdr.Linkname, gc.Equals, tarDirP)
		}
	}
	c.Assert(symlinks, gc.Equals, 1)

}

// UnTar
func (t *TarSuite) TestUnTarFilesUncompressed(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := TarFiles(t.testFiles, &outputTar, trimPath)
	c.Check(err, gc.IsNil)
	t.removeTestFiles(c)

	outputDir := filepath.Join(t.cwd, "TarOuputFolder")
	err = os.Mkdir(outputDir, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	UntarFiles(&outputTar, outputDir)
	t.assertFilesWhereUntared(c, testExpectedTarContents, outputDir)
}

func (t *TarSuite) TestFindFileFound(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, gc.IsNil)
	t.removeTestFiles(c)

	_, file, err := FindFile(&outputTar, "TarDirectoryPopulated/TarSubFile1")
	c.Assert(err, gc.IsNil)

	data, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)

	c.Check(string(data), gc.Equals, "TarSubFile1")
}

func (t *TarSuite) TestFindFileNotFound(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, gc.IsNil)
	t.removeTestFiles(c)

	_, _, err = FindFile(&outputTar, "does_not_exist")

	c.Check(err, gc.ErrorMatches, "does_not_exist not found")
}

func (t *TarSuite) TestUntarFilesHeadersIgnored(c *gc.C) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	err := w.WriteHeader(&tar.Header{
		Name:     "pax_global_header",
		Typeflag: tar.TypeXGlobalHeader,
	})
	c.Assert(err, gc.IsNil)
	err = w.Flush()
	c.Assert(err, gc.IsNil)

	err = UntarFiles(&buf, t.cwd)
	err = filepath.Walk(t.cwd, func(path string, finfo os.FileInfo, err error) error {
		if path != t.cwd {
			return fmt.Errorf("unexpected file: %v", path)
		}
		return err
	})
	c.Assert(err, gc.IsNil)
}
