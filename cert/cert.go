// Copyright 2012, 2013 Canonical Ltd.
// Copyright 2016 Cloudbase solutions
// Licensed under the LGPLv3, see LICENCE file for details.

package cert

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"

	"github.com/juju/errors"
)

// OtherName type for asn1 encoding
type OtherName struct {
	A string `asn1:"utf8"`
}

// GeneralName type for asn1 encoding
type GeneralName struct {
	OID       asn1.ObjectIdentifier
	OtherName `asn1:"tag:0"`
}

// GeneralNames type for asn1 encoding
type GeneralNames struct {
	GeneralName `asn1:"tag:0"`
}

var (
	// https://support.microsoft.com/en-us/kb/287547
	//  szOID_NT_PRINCIPAL_NAME 1.3.6.1.4.1.311.20.2.3
	szOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 20, 2, 3}
	// http://www.umich.edu/~x509/ssleay/asn1-oids.html
	// 2 5 29 17  subjectAltName
	subjAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
)

// getUPNExtensionValue returns marsheled asn1 encoded info
func getUPNExtensionValue(subject pkix.Name) ([]byte, error) {
	// returns the ASN.1 encoding of val
	// in addition to the struct tags recognized
	// we used:
	// utf8 => causes string to be marsheled as ASN.1, UTF8 strings
	// tag:x => specifies the ASN.1 tag number; imples ASN.1 CONTEXT SPECIFIC
	return asn1.Marshal(GeneralNames{
		GeneralName: GeneralName{
			// init our ASN.1 object identifier
			OID: szOID,
			OtherName: OtherName{
				A: subject.CommonName,
			},
		},
	})
}

// ParseCert parses the given PEM-formatted X509 certificate.
func ParseCert(certPEM string) (*x509.Certificate, error) {
	certPEMData := []byte(certPEM)
	for len(certPEMData) > 0 {
		var certBlock *pem.Block
		certBlock, certPEMData = pem.Decode(certPEMData)
		if certBlock == nil {
			break
		}
		if certBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(certBlock.Bytes)
			return cert, err
		}
	}
	return nil, errors.New("no certificates found")
}

// ParseCertAndKey parses the given PEM-formatted X509 certificate
// and RSA private key.
func ParseCertAndKey(certPEM, keyPEM string) (*x509.Certificate, crypto.Signer, error) {
	tlsCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, nil, err
	}

	key, ok := tlsCert.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key with unexpected type %T", tlsCert.PrivateKey)
	}
	return cert, key, nil
}
