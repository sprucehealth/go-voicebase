// Package apiservice contains the PingHandler
//	Description:
//		PingHandler is an HTTP handler for processing a request to a basic healt-check request
//	Request:
//		GET /v1/ping
//	Response:
//		Content-Type: text/plain
//		Content: pong
//		Status: HTTP/1.1 200 OK
package handlers

import (
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/libs/httputil"
)

const (
	pong = "pong"
)

type pingHandler struct{}

// NewPingHandler returns an initialized instance of pingHandler
func NewPingHandler() httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.NoAuthorizationRequired(
			pingHandler{}), httputil.Get)
}

func (pingHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(pong)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
