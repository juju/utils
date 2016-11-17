// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ssh_test

import (
	"io/ioutil"
	"strings"
	"testing/iotest"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/ssh"
)

type SSHStreamSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&SSHStreamSuite{})

func (s *SSHStreamSuite) TestNewStripCRNil(c *gc.C) {
	reader := ssh.StripCRReader(nil)
	c.Assert(reader, gc.IsNil)
}

func (s *SSHStreamSuite) TestStripCR(c *gc.C) {
	reader := ssh.StripCRReader(strings.NewReader("One\r\nTwo"))
	output, err := ioutil.ReadAll(reader)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(string(output), gc.Equals, "One\nTwo")
}

func (s *SSHStreamSuite) TestStripCROneByte(c *gc.C) {
	reader := ssh.StripCRReader(strings.NewReader("One\r\r\rTwo"))
	output, err := ioutil.ReadAll(iotest.OneByteReader(reader))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(string(output), gc.Equals, "OneTwo")
}

func (s *SSHStreamSuite) TestStripCRError(c *gc.C) {
	reader := ssh.StripCRReader(strings.NewReader("One\r\r\rTwo"))
	_, err := ioutil.ReadAll(iotest.TimeoutReader(reader))
	c.Assert(err.Error(), gc.Equals, "timeout")
}
