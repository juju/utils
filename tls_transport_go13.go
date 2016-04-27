// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build go1.3

package utils

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHttpTLSTransport returns a new http.Transport constructed with the TLS config
// and the necessary parameters for Juju.
func NewHttpTLSTransport(tlsConfig *tls.Config) *http.Transport {
	// See https://code.google.com/p/go/issues/detail?id=4677
	// We need to force the connection to close each time so that we don't
	// hit the above Go bug.
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSClientConfig:     tlsConfig,
		DisableKeepAlives:   true,
		Dial:                dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	registerFileProtocol(transport)
	return transport
}

// knownGoodCipherSuites contains the list of secure cipher suites to use
// with tls.Config.  This list currently differs from the list in crypto/tls by
// excluding all RC4 implementations, due to known security vulnerabilities in
// RC4 - CVE-2013-2566, CVE-2015-2808.  We also exclude ciphersuites which do
// not provide forward secrecy.
var knownGoodCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
}

// SecureTLSConfig returns a tls.Config that conforms to Juju's security
// standards, so as to avoid known security vulnerabilities in certain
// configurations.
//
// Currently it excludes RC4 implementations from the available ciphersuites,
// requires ciphersuites that provide forward secrecy, and sets the minimum TLS
// version to 1.2.
func SecureTLSConfig() *tls.Config {
	return &tls.Config{
		CipherSuites: knownGoodCipherSuites,
		MinVersion:   tls.VersionTLS12,
	}
}
