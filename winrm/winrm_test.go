// Copyright 2016 Canonical ltd.
// Copyright 2016 Cloudbase solutions
// Licensed under the lgplv3, see licence file for details.

package winrm_test

import (
	"os"
	"testing"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/winrm"
	gc "gopkg.in/check.v1"
)

func TestAll(t *testing.T) {
	gc.TestingT(t)
}

type WinRMSuite struct{}

func (w *WinRMSuite) GetPasswd(password string, c *gc.C) winrm.GetPasswd {
	return func() (string, error) {
		return password, nil
	}
}

var _ = gc.Suite(&WinRMSuite{})

func (w *WinRMSuite) TestValidateClient(c *gc.C) {
	config := winrm.ClientConfig{}

	err := config.Validate()
	// Empty host in client config
	c.Assert(err, gc.NotNil)

	hostname, err := os.Hostname()
	c.Assert(err, gc.IsNil)
	config.Host = hostname
	err = config.Validate()
	// Nil password getter, unable to retrive password
	c.Assert(err, gc.NotNil)

	config.Password = w.GetPasswd("Password123", c)
	err = config.Validate()
	c.Assert(err, gc.IsNil)

	config.Password = nil
	config.Cert = []byte("smth")
	config.Key = []byte("smth")
	err = config.Validate()
	// Cannot use cert auth with http connection
	c.Assert(err, gc.NotNil)

	config.Secure = true
	err = config.Validate()
	// Empty CA cert passed in client config
	c.Assert(err, gc.NotNil)

	config.Password = w.GetPasswd("Password123", c)
	err = config.Validate()
	// Empty CA cert passed in client config
	c.Assert(err, gc.NotNil)

	config.Cert = nil
	config.Key = nil
	config.Password = nil
	// empty key or cert in client config
	config.Insecure = false
	err = config.Validate()
	// Empty CA cert passed in client config
	c.Assert(err, gc.NotNil)

	config.Password = w.GetPasswd("Password123", c)
	config.CACert = []byte("smth")
	err = config.Validate()
	c.Assert(err, gc.IsNil)

}
func (w *WinRMSuite) TestWinrmClient(c *gc.C) {
	hostname, err := os.Hostname()
	c.Assert(err, gc.IsNil)

	config := winrm.ClientConfig{
		User:     "Administrator",
		Host:     hostname,
		Password: w.GetPasswd("Password123", c),
		Secure:   false,
	}

	cli, err := winrm.NewClient(config)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cli.Password(), jc.DeepEquals, "Password123")
	c.Assert(cli.Secure(), jc.IsFalse)
}

func (w *WinRMSuite) TestWinrmSecureClient(c *gc.C) {
	hostname, err := os.Hostname()
	c.Assert(err, gc.IsNil)

	config := winrm.ClientConfig{
		User:     "Administrator",
		Host:     hostname,
		Cert:     []byte(clientCert),
		Key:      []byte(clientKey),
		Insecure: true,
		Secure:   true,
	}

	cli, err := winrm.NewClient(config)
	c.Assert(err, gc.IsNil)
	c.Assert("", jc.DeepEquals, cli.Password())
	c.Assert(cli.Secure(), jc.IsTrue)
}

func (w *WinRMSuite) TestWinrmClientPasswd(c *gc.C) {
	hostname, err := os.Hostname()
	c.Assert(err, gc.IsNil)
	config := winrm.ClientConfig{
		User:     "Administrator",
		Host:     hostname,
		Password: w.GetPasswd("Password123", c),
	}

	cli, err := winrm.NewClient(config)
	c.Assert(err, gc.IsNil)
	c.Assert(cli, gc.NotNil)
	c.Assert("Password123", jc.DeepEquals, cli.Password())
	c.Assert(cli.Secure(), jc.IsFalse)
	config.Password = winrm.TTYGetPasswd
	cli, err = winrm.NewClient(config)
	c.Assert(err, gc.NotNil)
	c.Assert(cli, gc.IsNil)
}

func (w *WinRMSuite) TestDefaultClient(c *gc.C) {
	hostname, err := os.Hostname()
	c.Assert(err, gc.IsNil)

	config := winrm.ClientConfig{
		User:     "Administrator",
		Host:     hostname,
		Password: w.GetPasswd("Password123", c),
	}

	cli, err := winrm.NewClient(config)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cli, gc.NotNil)
	c.Assert(cli.Password(), jc.DeepEquals, "Password123")
	c.Assert(cli.Secure(), jc.IsFalse)

	configSecure := winrm.ClientConfig{
		User:     "Administrator",
		Host:     hostname,
		Password: nil,
		Cert:     []byte(clientCert),
		Key:      []byte(clientKey),
		Insecure: true,
		Secure:   true,
	}

	cli, err = winrm.NewClient(configSecure)
	c.Assert(cli, gc.NotNil)
	c.Assert("", jc.DeepEquals, cli.Password())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cli.Secure(), jc.IsTrue)
}

const clientKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAxsHQSbJCvWlF+BOn0QDRxJ0aKCAcFWvQryfqMKo+HUMgHQe7
1kgxXNTryi9uuGVTcxo1Jk9RVYBjbLYF56gNWdBv8lHz+K3CGVYJSXF2FTEvV56x
VdZKy3i1Ss8nm8RlOZN6qp8Ko0FJoKXi8IDYcoeNzEQeL/sZl6PUumTe92NQU4Df
NM+W/+p5hFSVAn0Hh/ICdAecInEiwCQinE1ZZcapTR2z8zjWi5zbzpzJM9e8ner1
pMHhGY/JbWqbjKnI7ZMR0qL6Vfc9QlMnuVe9B/1jDnieosUwnG9NTqeQsLCONuEP
0KRv9Dx8MjVYu26NAtspNpMVop+7hT+Mg/o4UQIDAQABAoIBAGCOe3+x6BZYdDNm
aRpyCXY41CI2Jy6I6CKPg4gMzIwutgUkYq5g+AofLkVU8bcHEplSXAu2cM7WxnZj
n07BJ2tAhYH1Kk7fKvJgB7b0Nedps5Qfdgs7Ra6/f2NOa/GSYZKtAOSuyt4ws3Il
5K/uCDjrfYmEdh3dILu4TXoX0vHfIxk2ZUj7FHAXFhyl8+kl2NvwRehJbGFx0JoK
Rm9ku8PtHEO6Geh4lj04qIydDqClRRy7xWyqov6cU8aqVHIOXtCW7siAXo3mrbLi
jXcgaM+gzRDGVvv8yVYhVVWWjJe6KLrDsagG/vquo+wItO6tUbeww08Lut2o6L7G
BFg9aDkCgYEA+tbPCOLESqKBWVExlSuNFZoYKV4RQnEETOFiOOmQ3oPH2HlzDXc5
iG4271hAIkLLjgnaQrc0imiom7nXhUAUtJMBHkM/OSf3wY3k3ev7NkmH6QZvkaXm
2+zaXdtW4/cg1L/smAAcx7yZb9vpkKVKBxxVz1jeHmncJK0BlhXOSV8CgYEAytiv
NkBJjdrl0cXYz7L3Eazv8fFyYYSKVA/pSt1f8DD/Di/PZjfrsjkhoLVrmQy1XD/j
VP4uBt5zeDsXfhXiBaXNDj5k9+LvpXCId0yRACBNI5Cm1+68GVJuz4V88gR2bLua
ZITkzwIe+5dGpNMteitD3AfQ1ePOx1UsIqsS7E8CgYEAkD6syehVhrHSfkFRqP1l
YVG+qTM9655AIdHOAPpXY44Wgya8AbdY71qp3pM6NjmBAsopqAngfeNXak3BYRAL
mBedIgD7v2t7buOhA/kq+fno3RjlWbU0f63BmQ2D9w3q5E0FyhbudfG/rnKg6pwS
aOpjchwhhw3LGZAfhGY/vTMCgYA8pXAtHidfnBSeFTLvVih8RmIuyetSsJfS7jbn
xSwL2fpHuY+elhWH4YDmVZdn2N7YR9ml7aDBOPz482HgtpYu7hVSruDtJBJWOkDy
uheYHBA0E+luIdhnEbhDnztt+FuXwrc0Wm82XQH6Yo4idWjhX9IYFNYhPMzz18ks
TE2KDQKBgQCXEOd26yjQd3yUiNiWy7dj/SjgjakKJ3EfUNU9Jn5d2OpCRKZpue3v
kwNtCKrmdl71oSlUtLxMOg5b5aoaKc52+NXQ6YII4gKUFSELEiPXUQiFNLzPeVF9
3vOGxdNw61G2F/vfoTN6EhmEHZySjU+nRInwET/eGyE3nVgepxj96w==
-----END RSA PRIVATE KEY-----
`

const clientCert = `
-----BEGIN CERTIFICATE-----
MIID3zCCAsegAwIBAgIVANcq1/xuRPDyKLvhiL2hnHnCYd4PMA0GCSqGSIb3DQEB
CwUAMEwxDTALBgNVBAoTBGp1anUxOzA5BgNVBAMTMmp1anUtZ2VuZXJhdGVkIGNs
aWVudCBjZXJ0IGZvciBtb2RlbCBBZG1pbmlzdHJhdG9yMB4XDTE2MTExNzE4MzIw
NVoXDTI2MTEyNDE4MzIwNVowTDENMAsGA1UEChMEanVqdTE7MDkGA1UEAxMyanVq
dS1nZW5lcmF0ZWQgY2xpZW50IGNlcnQgZm9yIG1vZGVsIEFkbWluaXN0cmF0b3Iw
ggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDGwdBJskK9aUX4E6fRANHE
nRooIBwVa9CvJ+owqj4dQyAdB7vWSDFc1OvKL264ZVNzGjUmT1FVgGNstgXnqA1Z
0G/yUfP4rcIZVglJcXYVMS9XnrFV1krLeLVKzyebxGU5k3qqnwqjQUmgpeLwgNhy
h43MRB4v+xmXo9S6ZN73Y1BTgN80z5b/6nmEVJUCfQeH8gJ0B5wicSLAJCKcTVll
xqlNHbPzONaLnNvOnMkz17yd6vWkweEZj8ltapuMqcjtkxHSovpV9z1CUye5V70H
/WMOeJ6ixTCcb01Op5CwsI424Q/QpG/0PHwyNVi7bo0C2yk2kxWin7uFP4yD+jhR
AgMBAAGjgbcwgbQwDgYDVR0PAQH/BAQDAgOoMBMGA1UdJQQMMAoGCCsGAQUFBwMC
MB0GA1UdDgQWBBSe8QgQMBjotVFMquerN06PYaB04TAfBgNVHSMEGDAWgBSe8QgQ
MBjotVFMquerN06PYaB04TBNBgNVHREERjBEoEIGCisGAQQBgjcUAgOgNAwyanVq
dS1nZW5lcmF0ZWQgY2xpZW50IGNlcnQgZm9yIG1vZGVsIEFkbWluaXN0cmF0b3Iw
DQYJKoZIhvcNAQELBQADggEBAGBSX1K5n10mBlD/HPcxUrwQ+yaZuQNihrC+nELX
BLLGEhKCz/6+15B8uvNTTePt0C73F/Gnp1376r1Q901J2Ec/z4X7w6VSig87/QgG
Yb0Ct/xU9mMqW0FpxdGNoFCt0nfBN9SXFpqyKwWup+R13XqDqoi1MQgNv4aiK5Yd
Xj8BUlGgJAmXHHsrHn2//m8+4C+3SIF76WKBYw8kgsy6W4+pd11iClyCiteNCmqT
OukwGPUibJ1CQesqnS6faYnI8cMSTM5ntFic1kr80IlmpLRnaQneiAUssd40XoRW
DS94Bhc+lI61ZWy7CWcy20ZKABNRfzVjoSoBXdSZnCihNSg=
-----END CERTIFICATE-----

`
