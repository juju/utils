// Copyright 2016 Canonical ltd.
// Copyright 2016 Cloudbase solutions
// licensed under the lgplv3, see licence file for details.

package winrm_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/juju/utils/winrm"

	gc "gopkg.in/check.v1"
)

var (
	certFileName = "winrmcert.pem"
	keyFileName  = "winrmkey.pem"
)

func (w *WinRMSuite) PathCredentials(c *gc.C) (string, string, string) {
	base := c.MkDir()
	return base, path.Join(base, certFileName), path.Join(base, keyFileName)
}
func (w *WinRMSuite) TestLoadClientCert(c *gc.C) {
	cert := winrm.NewX509()
	// path and certs dosen't exist and generate it
	base, certPath, keyPath := w.PathCredentials(c)
	err := cert.LoadClientCert(certPath, keyPath)
	c.Assert(err, gc.IsNil)

	cert.Reset()

	// read/load the already generated certs
	err = cert.LoadClientCert(certPath, keyPath)
	// check if the're the same cert and keys
	c.Assert(err, gc.IsNil)
	err = os.RemoveAll(base)
	c.Assert(err, gc.IsNil)
}

func (w *WinRMSuite) TestLoadCACert(c *gc.C) {
	cert := winrm.NewX509()
	base, _, _ := w.PathCredentials(c)
	err := os.MkdirAll(base, 0755)
	c.Assert(err, gc.IsNil)

	cacertPath := path.Join(base + "winrmcacert.pem")
	err = ioutil.WriteFile(cacertPath, []byte("content"), 0755)
	c.Assert(err, gc.IsNil)

	err = cert.LoadCACert(cacertPath)
	c.Assert(err, gc.IsNil)

	ca := cert.CACert()
	c.Assert(len(ca) > 1, gc.Equals, true)

	err = os.RemoveAll(base)
	c.Assert(err, gc.IsNil)
}
