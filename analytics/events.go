package analytics

import (
	"regexp"
	"time"
)

const ValidEventNameChar = `A-Za-z0-9_\-\.`

var (
	EventNameRE = regexp.MustCompile(`^[` + ValidEventNameChar + `]+$`)
)

type Eventer interface {
	Events() []Event
}

type Events []Event

func (es Events) Events() []Event {
	return []Event(es)
}

type ClientEvent struct {
	Event            string   `json:"event"`
	Timestamp        Time     `json:"time"`
	Error            string   `json:"error,omitempty"`
	SessionID        string   `json:"session_id"`
	DeviceID         string   `json:"device_id"`
	AccountID        int64    `json:"account_id,omitempty"`
	PatientID        int64    `json:"patient_id,omitempty"`
	DoctorID         int64    `json:"doctor_id,omitempty"`
	CaseID           int64    `json:"case_id,omitempty"`
	VisitID          int64    `json:"visit_id,omitempty"`
	ScreenID         string   `json:"screen_id,omitempty"`
	QuestionID       string   `json:"question_id,omitempty"`
	TimeSpent        *float64 `json:"time_spent,omitempty"`
	AppType          string   `json:"app_type,omitempty"`
	AppEnv           string   `json:"app_env,omitempty"`
	AppVersion       string   `json:"app_version,omitempty"`
	AppBuild         string   `json:"app_build,omitempty"`
	Platform         string   `json:"platform,omitempty"`
	PlatformVersion  string   `json:"platform_version,omitempty"`
	DeviceType       string   `json:"device_type,omitempty"`
	DeviceModel      string   `json:"device_model,omitempty"`
	ScreenWidth      int      `json:"screen_width,omitempty"`
	ScreenHeight     int      `json:"screen_height,omitempty"`
	ScreenResolution string   `json:"screen_resolution,omitempty"`
	ExtraJSON        string   `json:"extra_json,omitempty"`
}

func (*ClientEvent) Category() string {
	return "client"
}

func (e *ClientEvent) Time() time.Time {
	return time.Time(e.Timestamp)
}

func (e *ClientEvent) Events() []Event {
	return []Event{e}
}

type ServerEvent struct {
	Application     string `json:"application"`
	Event           string `json:"event"`
	Timestamp       Time   `json:"time"`
	SessionID       string `json:"session_id,omitempty"`
	AccountID       int64  `json:"account_id,omitempty"`
	PatientID       int64  `json:"patient_id,omitempty"`
	DoctorID        int64  `json:"doctor_id,omitempty"`
	VisitID         int64  `json:"visit_id,omitempty"`
	CaseID          int64  `json:"case_id,omitempty"`
	TreatmentPlanID int64  `json:"treatment_plan_id,omitempty"`
	Role            string `json:"role,omitempty"`
	ExtraJSON       string `json:"extra_json,omitempty"`
}

func (*ServerEvent) Category() string {
	return "server"
}

func (e *ServerEvent) Time() time.Time {
	return time.Time(e.Timestamp)
}

func (e *ServerEvent) Events() []Event {
	return []Event{e}
}

type WebRequestEvent struct {
	Service      string `json:"service"`
	Path         string `json:"path"`
	Timestamp    Time   `json:"time"`
	RequestID    uint64 `json:"request_id"`
	StatusCode   int    `json:"status_code"`
	Method       string `json:"method"`
	URL          string `json:"url"`
	RemoteAddr   string `json:"remote_addr,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	Referrer     string `json:"referrer,omitempty"`
	ResponseTime int    `json:"response_time"` // microseconds
	Server       string `json:"server"`
	AccountID    int64  `json:"account_id,omitempty"`
	DeviceID     string `json:"device_id,omitempty"`
}

func (*WebRequestEvent) Category() string {
	return "webrequest"
}

func (e *WebRequestEvent) Time() time.Time {
	return time.Time(e.Timestamp)
}

func (e *WebRequestEvent) Events() []Event {
	return []Event{e}
}
