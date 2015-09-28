package home

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cookieo9/resources-go"
	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/boot"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

const (
	invalidTimeThreshold = 60 * 60 * 24 * 30 // number of seconds after which an event is dropped
)

var once sync.Once
var logoContentType string
var logoImage []byte

type properties map[string]interface{}

func (p properties) popString(name string) string {
	s, ok := p[name].(string)
	if !ok {
		return ""
	}
	delete(p, name)
	return s
}

func (p properties) popFloat64Ptr(name string) *float64 {
	f, ok := p[name].(float64)
	if !ok {
		return nil
	}
	delete(p, name)
	return &f
}

func (p properties) popFloat64(name string) float64 {
	f := p.popFloat64Ptr(name)
	if f == nil {
		return 0.0
	}
	return *f
}

func (p properties) popInt64(name string) int64 {
	i, ok := p[name].(float64)
	if !ok {
		if s := p.popString(name); s != "" {
			if i, err := strconv.ParseInt(s, 10, 64); err == nil {
				return i
			}
		}
		return 0
	}
	delete(p, name)
	return int64(i)
}

func (p properties) popInt(name string) int {
	return int(p.popInt64(name))
}

func (p properties) popBoolPtr(name string) *bool {
	b, ok := p[name].(bool)
	if !ok {
		return nil
	}
	delete(p, name)
	return &b
}

type event struct {
	Name       string     `json:"event"`
	Properties properties `json:"properties"`
}

type analyticsHandler struct {
	logger             analytics.Logger
	statEventsReceived *metrics.Counter
	statEventsDropped  *metrics.Counter
}

type analyticsAPIRequest struct {
	CurrentTime float64 `json:"current_time"`
	Events      []event `json:"events"`
}

func newAnalyticsHandler(logger analytics.Logger, statsRegistry metrics.Registry) httputil.ContextHandler {
	once.Do(func() {
		logoContentType = "image/png"
		fi, err := resources.DefaultBundle.Open("static/img/logo-small.png")
		if err != nil {
			panic(err)
		}
		logoImage, err = ioutil.ReadAll(fi)
		if err != nil {
			panic(err)
		}
		fi.Close()
	})

	h := &analyticsHandler{
		logger:             logger,
		statEventsReceived: metrics.NewCounter(),
		statEventsDropped:  metrics.NewCounter(),
	}

	statsRegistry.Add("events/received", h.statEventsReceived)
	statsRegistry.Add("events/dropped", h.statEventsDropped)
	return h
}

func (h *analyticsHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	nowUnix := float64(now.UnixNano()) / 1e9

	var currentTime float64
	var events []event

	if r.Method == httputil.Post {
		var req analyticsAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			golog.Errorf("Failed to decode analytics POST body: %s", err.Error())
			www.APIBadRequestError(w, r, "Failed to decode body")
			return
		}
		currentTime = req.CurrentTime
		events = req.Events
	} else {
		if err := r.ParseForm(); err != nil {
			www.BadRequestError(w, r, err)
			return
		}
		prop := properties(make(map[string]interface{}, len(r.Form)))
		ev := event{
			Properties: prop,
		}
		for k, v := range r.Form {
			if k == "event" {
				ev.Name = v[0]
			} else {
				prop[k] = v[0]
			}
		}
		events = []event{ev}
	}

	h.statEventsReceived.Inc(uint64(len(events)))

	account, _ := www.CtxAccount(ctx)

	var eventsOut []analytics.Event
	for _, ev := range events {
		name, err := analytics.MangleEventName(ev.Name)
		if err != nil {
			continue
		}
		// Calculate delta time for the event from the client provided current time.
		// Use this delta to generate the absolute event time based on the server's time.
		// This accounts for the client clock being off.
		tm := now
		t := ev.Properties.popFloat64("time")
		if currentTime > 0.0 && t != 0 {
			td := currentTime - t
			if td > invalidTimeThreshold || td < 0 {
				continue
			}
			tf := nowUnix - td
			tm = time.Unix(int64(math.Floor(tf)), int64(1e9*(tf-math.Floor(tf))))
		}
		// TODO: at the moment there is no session ID for web requests so just use the remote address
		sessionID := r.RemoteAddr
		evo := &analytics.ClientEvent{
			Event:       "js:" + name,
			Timestamp:   analytics.Time(tm),
			SessionID:   sessionID,
			Error:       ev.Properties.popString("error"),
			ScreenID:    ev.Properties.popString("screen_id"),
			TimeSpent:   ev.Properties.popFloat64Ptr("time_spent"),
			AppType:     ev.Properties.popString("app_type"),
			AppVersion:  boot.GitBranch,
			AppBuild:    boot.BuildNumber,
			DeviceModel: r.UserAgent(),
		}
		if account != nil {
			evo.AccountID = account.ID
		}
		// Put anything left over into ExtraJSON if it's a valid format
		for k, v := range ev.Properties {
			switch v.(type) {
			case string, float64, int, int64, uint64, bool:
			default:
				delete(ev.Properties, k)
			}
		}
		if len(ev.Properties) != 0 {
			extraJSON, err := json.Marshal(ev.Properties)
			if err != nil {
				golog.Errorf("Failed to marshal extra properties: %s", err.Error())
			} else {
				evo.ExtraJSON = string(extraJSON)
			}
		}
		eventsOut = append(eventsOut, evo)
	}
	h.statEventsDropped.Inc(uint64(len(events) - len(eventsOut)))

	if len(eventsOut) > 0 {
		h.logger.WriteEvents(eventsOut)
	}

	if r.Method == httputil.Get {
		w.Header().Set("Content-Type", logoContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(logoImage)))
		if _, err := w.Write(logoImage); err != nil {
			golog.Errorf("Failed to write logo image: %s", err.Error())
		}
	} else {
		httputil.JSONResponse(w, http.StatusOK, struct{}{})
	}
}
