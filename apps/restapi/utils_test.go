package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/test"
)

func init() {
	conc.Testing = true
}

type mockSNS_logHandler struct {
	snsiface.SNSAPI
	msg *sns.PublishInput
}

func (m *mockSNS_logHandler) Publish(msg *sns.PublishInput) (*sns.PublishOutput, error) {
	m.msg = msg
	return &sns.PublishOutput{}, nil
}

func TestSNSLogHandler(t *testing.T) {
	snsCli := &mockSNS_logHandler{}
	h := snsLogHandler(snsCli, "topic", "test", nil, nil, metrics.NewRegistry())
	test.OK(t, h.Log(&golog.Entry{Lvl: golog.INFO, Msg: "Danger Danger", Src: "somewhere:123"}))
	test.Assert(t, snsCli.msg == nil, "INFO events shoudldn't be published")
	test.OK(t, h.Log(&golog.Entry{Lvl: golog.WARN, Msg: "High Voltage", Src: "somewhere:123"}))
	test.Assert(t, snsCli.msg == nil, "WARN events shoudldn't be published")
	test.OK(t, h.Log(&golog.Entry{Lvl: golog.ERR, Msg: "Danger Danger", Src: "somewhere:123"}))
	test.Assert(t, snsCli.msg != nil, "ERR event not published")
	snsCli.msg = nil
	test.OK(t, h.Log(&golog.Entry{Lvl: golog.CRIT, Msg: "Danger Danger", Src: "somewhere:123"}))
	test.Assert(t, snsCli.msg != nil, "CRIT event not published")
}

func TestSanitizeSNSSubject(t *testing.T) {
	cases := map[string]string{
		"basic": "basic",
		"no\nnew\rlines\ttabs\aor\bother\fcontrol": "no new lines tabs or other control",
		"these!characters?are*likely(ok)":          "these!characters?are*likely(ok)",
	}
	for s, e := range cases {
		test.Equals(t, e, sanitizeSNSSubject(s))
	}
}
