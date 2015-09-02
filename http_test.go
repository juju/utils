// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

func init() {
	// The default Proxy implementation for HTTP transports,
	// ProxyFromEnvironment, uses a sync.Once in Go 1.3 onwards.
	// No tests should be dialing out, so no proxy should be
	// used.
	os.Setenv("http_proxy", "")
	os.Setenv("HTTP_PROXY", "")
}

type httpSuite struct {
	testing.IsolationSuite
	Server *httptest.Server
}

var _ = gc.Suite(&httpSuite{})

func (s *httpSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	// NewTLSServer returns a server which serves TLS, but
	// its certificates are not validated by the default
	// OS certificates, so any HTTPS request will fail
	// unless a non-validating client is used.
	s.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func (s *httpSuite) TearDownTest(c *gc.C) {
	if s.Server != nil {
		s.Server.Close()
	}
	s.IsolationSuite.TearDownTest(c)
}

func (s *httpSuite) TestDefaultClientFails(c *gc.C) {
	_, err := http.Get(s.Server.URL)
	c.Assert(err, gc.ErrorMatches, "(.|\n)*x509: certificate signed by unknown authority")
}

func (s *httpSuite) TestValidatingClientGetter(c *gc.C) {
	client := utils.GetValidatingHTTPClient()
	_, err := client.Get(s.Server.URL)
	c.Assert(err, gc.ErrorMatches, "(.|\n)*x509: certificate signed by unknown authority")

	client1 := utils.GetValidatingHTTPClient()
	c.Assert(client1, gc.Not(gc.Equals), client)
}

func (s *httpSuite) TestNonValidatingClientGetter(c *gc.C) {
	client := utils.GetNonValidatingHTTPClient()
	resp, err := client.Get(s.Server.URL)
	c.Assert(err, gc.IsNil)
	resp.Body.Close()
	c.Assert(resp.StatusCode, gc.Equals, http.StatusOK)

	client1 := utils.GetNonValidatingHTTPClient()
	c.Assert(client1, gc.Not(gc.Equals), client)
}

func (s *httpSuite) TestBasicAuthHeader(c *gc.C) {
	header := utils.BasicAuthHeader("eric", "sekrit")
	c.Assert(len(header), gc.Equals, 1)
	auth := header.Get("Authorization")
	fields := strings.Fields(auth)
	c.Assert(len(fields), gc.Equals, 2)
	basic, encoded := fields[0], fields[1]
	c.Assert(basic, gc.Equals, "Basic")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	c.Assert(err, gc.IsNil)
	c.Assert(string(decoded), gc.Equals, "eric:sekrit")
}

func (s *httpSuite) TestParseBasicAuthHeader(c *gc.C) {
	tests := []struct {
		about          string
		h              http.Header
		expectUserid   string
		expectPassword string
		expectError    string
	}{{
		about:       "no Authorization header",
		h:           http.Header{},
		expectError: "invalid or missing HTTP auth header",
	}, {
		about: "empty Authorization header",
		h: http.Header{
			"Authorization": {""},
		},
		expectError: "invalid or missing HTTP auth header",
	}, {
		about: "Not basic encoding",
		h: http.Header{
			"Authorization": {"NotBasic stuff"},
		},
		expectError: "invalid or missing HTTP auth header",
	}, {
		about: "invalid base64",
		h: http.Header{
			"Authorization": {"Basic not-base64"},
		},
		expectError: "invalid HTTP auth encoding",
	}, {
		about: "no ':'",
		h: http.Header{
			"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte("aladdin"))},
		},
		expectError: "invalid HTTP auth contents",
	}, {
		about: "valid credentials",
		h: http.Header{
			"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte("aladdin:open sesame"))},
		},
		expectUserid:   "aladdin",
		expectPassword: "open sesame",
	}}
	for i, test := range tests {
		c.Logf("test %d: %s", i, test.about)
		u, p, err := utils.ParseBasicAuthHeader(test.h)
		c.Assert(u, gc.Equals, test.expectUserid)
		c.Assert(p, gc.Equals, test.expectPassword)
		if test.expectError != "" {
			c.Assert(err.Error(), gc.Equals, test.expectError)
		} else {
			c.Assert(err, gc.IsNil)
		}
	}
}

type dialSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&dialSuite{})

func (s *dialSuite) TestDialRejectsNonLocal(c *gc.C) {
	s.PatchValue(&utils.OutgoingAccessAllowed, false)
	_, err := utils.Dial("tcp", "10.0.0.1:80")
	c.Assert(err, gc.ErrorMatches, `access to address "10.0.0.1:80" not allowed`)
	_, err = utils.Dial("tcp", "somehost:80")
	c.Assert(err, gc.ErrorMatches, `access to address "somehost:80" not allowed`)
}

func (s *dialSuite) assertDial(c *gc.C, addr string) {
	dialed := false
	s.PatchValue(utils.NetDial, func(network, addr string) (net.Conn, error) {
		c.Assert(network, gc.Equals, "tcp")
		c.Assert(addr, gc.Equals, addr)
		dialed = true
		return nil, nil
	})
	_, err := utils.Dial("tcp", addr)
	c.Assert(err, gc.IsNil)
	c.Assert(dialed, jc.IsTrue)
}

func (s *dialSuite) TestDialAllowsNonLocal(c *gc.C) {
	s.PatchValue(&utils.OutgoingAccessAllowed, true)
	s.assertDial(c, "10.0.0.1:80")
}

func (s *dialSuite) TestDialAllowsLocal(c *gc.C) {
	s.PatchValue(&utils.OutgoingAccessAllowed, false)
	s.assertDial(c, "127.0.0.1:1234")
	s.assertDial(c, "localhost:1234")
}

func (s *dialSuite) TestInsecureClientNoAccess(c *gc.C) {
	s.PatchValue(&utils.OutgoingAccessAllowed, false)
	_, err := utils.GetNonValidatingHTTPClient().Get("http://10.0.0.1:1234")
	c.Assert(err, gc.ErrorMatches, `.*access to address "10.0.0.1:1234" not allowed`)
}

func (s *dialSuite) TestSecureClientNoAccess(c *gc.C) {
	s.PatchValue(&utils.OutgoingAccessAllowed, false)
	_, err := utils.GetValidatingHTTPClient().Get("http://10.0.0.1:1234")
	c.Assert(err, gc.ErrorMatches, `.*access to address "10.0.0.1:1234" not allowed`)
}
