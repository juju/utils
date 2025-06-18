// Copyright 2012, 2013 Canonical Ltd.
// Copyright 2016 Cloudbase solutions
// Licensed under the LGPLv3, see LICENCE file for details.

package cert_test

import (
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/v4/cert"
)

func TestAll(t *testing.T) {
	gc.TestingT(t)
}

type certSuite struct{}

var _ = gc.Suite(certSuite{})

func (certSuite) TestParseCertificate(c *gc.C) {
	xcert, err := cert.ParseCert(caCertPEM)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(xcert.Subject.CommonName, gc.Equals, `juju-generated CA for model "juju testing"`)

	xcert, err = cert.ParseCert(caKeyPEM)
	c.Check(xcert, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "no certificates found")

	xcert, err = cert.ParseCert("hello")
	c.Check(xcert, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "no certificates found")
}

func (certSuite) TestParseCertAndKey(c *gc.C) {
	xcert, key, err := cert.ParseCertAndKey(caCertPEM, caKeyPEM)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(xcert.Subject.CommonName, gc.Equals, `juju-generated CA for model "juju testing"`)
	c.Assert(key, gc.NotNil)

	c.Assert(xcert.PublicKey, gc.DeepEquals, key.Public())
}

var (
	caCertPEM = `
-----BEGIN CERTIFICATE-----
MIICHDCCAcagAwIBAgIUfzWn5ktGMxD6OiTgfiZyvKdM+ZYwDQYJKoZIhvcNAQEL
BQAwazENMAsGA1UEChMEanVqdTEzMDEGA1UEAwwqanVqdS1nZW5lcmF0ZWQgQ0Eg
Zm9yIG1vZGVsICJqdWp1IHRlc3RpbmciMSUwIwYDVQQFExwxMjM0LUFCQ0QtSVMt
Tk9ULUEtUkVBTC1VVUlEMB4XDTE2MDkyMTEwNDgyN1oXDTI2MDkyODEwNDgyN1ow
azENMAsGA1UEChMEanVqdTEzMDEGA1UEAwwqanVqdS1nZW5lcmF0ZWQgQ0EgZm9y
IG1vZGVsICJqdWp1IHRlc3RpbmciMSUwIwYDVQQFExwxMjM0LUFCQ0QtSVMtTk9U
LUEtUkVBTC1VVUlEMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAL+0X+1zl2vt1wI4
1Q+RnlltJyaJmtwCbHRhREXVGU7t0kTMMNERxqLnuNUyWRz90Rg8s9XvOtCqNYW7
mypGrFECAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB/wQFMAMBAf8w
HQYDVR0OBBYEFHueMLZ1QJ/2sKiPIJ28TzjIMRENMA0GCSqGSIb3DQEBCwUAA0EA
ovZN0RbUHrO8q9Eazh0qPO4mwW9jbGTDz126uNrLoz1g3TyWxIas1wRJ8IbCgxLy
XUrBZO5UPZab66lJWXyseA==
-----END CERTIFICATE-----
`

	caKeyPEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAL+0X+1zl2vt1wI41Q+RnlltJyaJmtwCbHRhREXVGU7t0kTMMNER
xqLnuNUyWRz90Rg8s9XvOtCqNYW7mypGrFECAwEAAQJAMPa+JaUHgO6foxam/LIB
0u95N3OgFR+dWeBaEsgKDclpREdJ0rXNI+3C3kwqeEZR4omoPlBeSEewSkwHxpmI
0QIhAOjKiHZ5v6R8haleipbDzkGUnZW07hEwL5Ld4MNx/QQ1AiEA0tEzSSNAdM0C
M/vY0x5mekIYai8/tFSEG9PJ3ZkpEy0CIQCo9B3YxwI1Un777vbs903iQQeiWP+U
EAHnOQvhLgDxpQIgGkpml+9igW5zoOH+h02aQBLwEoXz7tw/YW0HFrCcE70CIGkS
ve4WjiEqnQaHNAPy0hY/1DfIgBOSpOfnkFHOk9vX
-----END RSA PRIVATE KEY-----
`
)
