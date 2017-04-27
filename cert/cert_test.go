// Copyright 2012, 2013 Canonical Ltd.
// Copyright 2016 Cloudbase solutions
// Licensed under the AGPLv3, see LICENCE file for details.

package cert_test

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	"github.com/juju/utils/cert"
	gc "gopkg.in/check.v1"
)

func TestAll(t *testing.T) {
	gc.TestingT(t)
}

type certSuite struct{}

var _ = gc.Suite(certSuite{})

func checkNotBefore(c *gc.C, cert *x509.Certificate, now time.Time) {
	// Check that the certificate is valid from one week before today.
	c.Check(cert.NotBefore.Before(now), jc.IsTrue)
	c.Check(cert.NotBefore.Before(now.AddDate(0, 0, -6)), jc.IsTrue)
	c.Check(cert.NotBefore.After(now.AddDate(0, 0, -8)), jc.IsTrue)
}

func checkNotAfter(c *gc.C, cert *x509.Certificate, expiry time.Time) {
	// Check the surrounding day.
	c.Assert(cert.NotAfter.Before(expiry.AddDate(0, 0, 1)), jc.IsTrue)
	c.Assert(cert.NotAfter.After(expiry.AddDate(0, 0, -1)), jc.IsTrue)
}

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

	c.Assert(xcert.PublicKey.(*rsa.PublicKey), gc.DeepEquals, &key.PublicKey)
}

func (certSuite) TestNewCA(c *gc.C) {
	now := time.Now()
	expiry := roundTime(now.AddDate(0, 0, 1))
	caCertPEM, caKeyPEM, err := cert.NewCA(
		fmt.Sprintf("juju-generated CA for model %s", "foo"),
		"1", expiry, 0,
	)
	c.Assert(err, jc.ErrorIsNil)

	caCert, caKey, err := cert.ParseCertAndKey(caCertPEM, caKeyPEM)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(caKey, gc.FitsTypeOf, (*rsa.PrivateKey)(nil))
	c.Check(caCert.Subject.CommonName, gc.Equals, `juju-generated CA for model foo`)
	checkNotBefore(c, caCert, now)
	checkNotAfter(c, caCert, expiry)
	c.Check(caCert.BasicConstraintsValid, jc.IsTrue)
	c.Check(caCert.IsCA, jc.IsTrue)
	//c.Assert(caCert.MaxPathLen, Equals, 0)	TODO it ends up as -1 - check that this is ok.
}

func checkCertificate(c *gc.C, caCert *x509.Certificate, srvCertPEM, srvKeyPEM string, now, expiry time.Time) {
	srvCert, srvKey, err := cert.ParseCertAndKey(srvCertPEM, srvKeyPEM)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(srvCert.Subject.CommonName, gc.Equals, "*")
	checkNotBefore(c, srvCert, now)
	checkNotAfter(c, srvCert, expiry)
	c.Assert(srvCert.BasicConstraintsValid, jc.IsFalse)
	c.Assert(srvCert.IsCA, jc.IsFalse)
	c.Assert(srvCert.ExtKeyUsage, gc.DeepEquals, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	c.Assert(srvCert.SerialNumber, gc.NotNil)
	if srvCert.SerialNumber.Cmp(big.NewInt(0)) == 0 {
		c.Fatalf("zero serial number")
	}

	checkTLSConnection(c, caCert, srvCert, srvKey)
}

// checkTLSConnection checks that we can correctly perform a TLS
// handshake using the given credentials.
func checkTLSConnection(c *gc.C, caCert, srvCert *x509.Certificate, srvKey *rsa.PrivateKey) (caName string) {
	clientCertPool := x509.NewCertPool()
	clientCertPool.AddCert(caCert)

	var outBytes bytes.Buffer

	const msg = "hello to the server"
	p0, p1 := net.Pipe()
	p0 = &recordingConn{
		Conn:   p0,
		Writer: io.MultiWriter(p0, &outBytes),
	}

	var clientState tls.ConnectionState
	done := make(chan error)
	go func() {
		config := utils.SecureTLSConfig()
		config.Certificates = []tls.Certificate{{
			Certificate: [][]byte{srvCert.Raw},
			PrivateKey:  srvKey,
		}}

		conn := tls.Server(p1, config)
		defer conn.Close()
		data, err := ioutil.ReadAll(conn)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(string(data), gc.Equals, msg)
		close(done)
	}()

	tlsConfig := utils.SecureTLSConfig()
	tlsConfig.ServerName = "anyServer"
	tlsConfig.RootCAs = clientCertPool
	clientConn := tls.Client(p0, tlsConfig)
	defer clientConn.Close()

	_, err := clientConn.Write([]byte(msg))
	c.Assert(err, jc.ErrorIsNil)
	clientState = clientConn.ConnectionState()
	clientConn.Close()

	// wait for server to exit
	<-done

	outData := outBytes.String()
	c.Assert(outData, gc.Not(gc.HasLen), 0)
	if strings.Index(outData, msg) != -1 {
		c.Fatalf("TLS connection not encrypted")
	}
	c.Assert(clientState.VerifiedChains, gc.HasLen, 1)
	c.Assert(clientState.VerifiedChains[0], gc.HasLen, 2)
	return clientState.VerifiedChains[0][1].Subject.CommonName
}

type recordingConn struct {
	net.Conn
	io.Writer
}

func (c recordingConn) Write(buf []byte) (int, error) {
	return c.Writer.Write(buf)
}

// roundTime returns t rounded to the previous whole second.
func roundTime(t time.Time) time.Time {
	return t.Add(time.Duration(-t.Nanosecond()))
}

var rsaByteSizes = []int{512, 1024, 2048, 4096}

func (certSuite) TestNewClientCertRSASize(c *gc.C) {
	for _, size := range rsaByteSizes {
		now := time.Now()
		expiry := roundTime(now.AddDate(0, 0, 1))
		certPem, privPem, err := cert.NewClientCert(
			fmt.Sprintf("juju-generated CA for model %s", "foo"), "1", expiry, size)

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(certPem, gc.NotNil)
		c.Assert(privPem, gc.NotNil)

		caCert, caKey, err := cert.ParseCertAndKey(certPem, privPem)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(caCert.Subject.CommonName, gc.Equals, "juju-generated CA for model foo")
		c.Check(caCert.Subject.Organization, gc.DeepEquals, []string{"juju"})
		c.Check(caCert.Subject.SerialNumber, gc.DeepEquals, "1")

		c.Check(caKey, gc.FitsTypeOf, (*rsa.PrivateKey)(nil))
		c.Check(caCert.Version, gc.Equals, 3)

		value, err := cert.CertGetUPNExtenstionValue(caCert.Subject)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(value, gc.Not(gc.IsNil))

		c.Assert(caCert.Extensions[len(caCert.Extensions)-1], jc.DeepEquals, pkix.Extension{
			Id:       cert.CertSubjAltName,
			Value:    value,
			Critical: false,
		})
		c.Assert(caCert.PublicKeyAlgorithm, gc.Equals, x509.RSA)
		c.Assert(caCert.ExtKeyUsage[0], gc.Equals, x509.ExtKeyUsageClientAuth)
		checkNotBefore(c, caCert, now)
		checkNotAfter(c, caCert, expiry)

	}
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
