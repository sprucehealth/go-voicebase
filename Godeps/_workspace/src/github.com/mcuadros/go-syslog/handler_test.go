package syslog

import (
	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/jeromer/syslogparser"
	. "launchpad.net/gocheck"
)

type HandlerSuite struct{}

var _ = Suite(&HandlerSuite{})

func (s *HandlerSuite) TestHandle(c *C) {
	logPart := syslogparser.LogParts{"tag": "foo"}

	channel := make(LogPartsChannel, 1)
	handler := NewChannelHandler(channel)
	handler.Handle(logPart)

	fromChan := <-channel
	c.Check(fromChan["tag"], Equals, logPart["tag"])
}
