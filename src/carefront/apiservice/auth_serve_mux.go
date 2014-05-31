package apiservice

import (
	"carefront/libs/golog"
	"carefront/libs/idgen"
	"carefront/thrift/api"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/samuel/go-metrics/metrics"
)

// If a handler conforms to this interface and returns true then
// non-authenticated requests will be handled. Otherwise,
// they 403 response will be returned.
type NonAuthenticated interface {
	NonAuthenticated() bool
}

type AuthServeMux struct {
	http.ServeMux
	AuthApi api.Auth

	statLatency      metrics.Histogram
	statRequests     metrics.Counter
	statAuthSuccess  metrics.Counter
	statAuthFailure  metrics.Counter
	statIDGenFailure metrics.Counter
	statIDGenSuccess metrics.Counter
}

type AuthEvent string

type AuthLog struct {
	Event AuthEvent
	Msg   string
}

type CustomResponseWriter struct {
	WrappedResponseWriter http.ResponseWriter
	StatusCode            int
	WroteHeader           bool
}

func (c *CustomResponseWriter) WriteHeader(status int) {
	c.StatusCode = status
	c.WroteHeader = true
	c.WrappedResponseWriter.WriteHeader(status)
}

func (c *CustomResponseWriter) Header() http.Header {
	return c.WrappedResponseWriter.Header()
}

func (c *CustomResponseWriter) Write(bytes []byte) (int, error) {
	if c.WroteHeader == false {
		c.WriteHeader(http.StatusOK)
	}
	return (c.WrappedResponseWriter.Write(bytes))
}

const (
	AuthEventNoSuchLogin     AuthEvent = "NoSuchLogin"
	AuthEventInvalidPassword AuthEvent = "InvalidPassword"
	AuthEventInvalidToken    AuthEvent = "InvalidToken"
)

func NewAuthServeMux(authApi api.Auth, statsRegistry metrics.Registry) *AuthServeMux {
	mux := &AuthServeMux{
		ServeMux:         *http.NewServeMux(),
		AuthApi:          authApi,
		statLatency:      metrics.NewBiasedHistogram(),
		statRequests:     metrics.NewCounter(),
		statAuthSuccess:  metrics.NewCounter(),
		statAuthFailure:  metrics.NewCounter(),
		statIDGenFailure: metrics.NewCounter(),
		statIDGenSuccess: metrics.NewCounter(),
	}
	statsRegistry.Add("requests/latency", mux.statLatency)
	statsRegistry.Add("requests/total", mux.statRequests)
	statsRegistry.Add("requests/auth/success", mux.statAuthSuccess)
	statsRegistry.Add("requests/auth/failure", mux.statAuthFailure)
	statsRegistry.Add("requests/idgen/failure", mux.statIDGenFailure)
	statsRegistry.Add("requests/idgen/success", mux.statIDGenSuccess)

	return mux
}

// Parse the "Authorization: token xxx" header and check the token for validity
func (mux *AuthServeMux) checkAuth(r *http.Request) (bool, int64, string, error) {
	if Testing {
		if idStr := r.Header.Get("AccountId"); idStr != "" {
			id, err := strconv.ParseInt(idStr, 10, 64)
			return true, id, "", err
		}
	}

	token, err := GetAuthTokenFromHeader(r)
	if err == ErrBadAuthToken {
		return false, 0, "failed to parse Authorization header", nil
	} else if err != nil {
		return false, 0, "", err
	}
	if res, err := mux.AuthApi.ValidateToken(token); err != nil {
		return false, 0, "", err
	} else {
		var accountId int64
		if res.AccountId != nil {
			accountId = *res.AccountId
		}
		return res.IsValid, accountId, res.Reason, nil
	}
}

type RequestLog struct {
	RemoteAddr   string
	RequestID    int64
	Method       string
	URL          string
	StatusCode   int
	ContentType  string
	UserAgent    string
	ResponseTime float64
}

func (mux *AuthServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.statRequests.Inc(1)

	ctx := GetContext(r)
	ctx.RequestStartTime = time.Now()
	var err error
	ctx.RequestID, err = idgen.NewID()
	if err != nil {
		golog.Errorf("Unable to generate a requestId: %s", err)
		mux.statIDGenFailure.Inc(1)
	} else {
		mux.statIDGenSuccess.Inc(1)
	}

	customResponseWriter := &CustomResponseWriter{w, 0, false}

	// Use strict transport security. Not entirely useful for a REST API, but it doesn't hurt.
	// http://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security
	customResponseWriter.Header().Set("Strict-Transport-Security", "max-age=31536000")

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			golog.Criticalf("http: panic: %v\n%s", err, buf)

			// The header may have already been written in which case
			// this will fail, but it's likely it hasn't so it's
			// good to tell the client something blew up.
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			responseTime := time.Since(ctx.RequestStartTime).Nanoseconds() / 1e3
			mux.statLatency.Update(responseTime)

			remoteAddr := r.RemoteAddr
			if idx := strings.LastIndex(remoteAddr, ":"); idx > 0 {
				remoteAddr = remoteAddr[:idx]
			}

			golog.Log("webrequest", golog.INFO, &RequestLog{
				RemoteAddr:   remoteAddr,
				RequestID:    GetContext(r).RequestID,
				Method:       r.Method,
				URL:          r.URL.String(),
				StatusCode:   customResponseWriter.StatusCode,
				ContentType:  w.Header().Get("Content-Type"),
				UserAgent:    r.UserAgent(),
				ResponseTime: float64(responseTime) / 1000.0,
			})
		}
		DeleteContext(r)
	}()
	if r.RequestURI == "*" {
		customResponseWriter.Header().Set("Connection", "close")
		customResponseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	h, pattern := mux.Handler(r)

	// these means the page is not found, in which case serve the page as we would
	// since we have a page not found handler returned
	if pattern == "" {
		h.ServeHTTP(customResponseWriter, r)
		return
	}

	if nonAuth, ok := h.(NonAuthenticated); !ok || !nonAuth.NonAuthenticated() {
		if valid, accountId, reason, err := mux.checkAuth(r); err != nil {
			customResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		} else if !valid {
			golog.Log("auth", golog.WARN, &AuthLog{
				Event: AuthEventInvalidToken,
				Msg:   reason,
			})
			mux.statAuthFailure.Inc(1)
			WriteAuthTimeoutError(customResponseWriter)
			return
		} else {
			mux.statAuthSuccess.Inc(1)
			ctx.AccountId = accountId
		}
	}
	h.ServeHTTP(customResponseWriter, r)
}
