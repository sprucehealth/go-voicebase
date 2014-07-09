package router

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/passreset"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/third_party/github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/third_party/github.com/subosito/twilio"
	"github.com/sprucehealth/backend/www"
)

func New(dataAPI api.DataAPI, authAPI api.AuthAPI, twilioCli *twilio.Client, fromNumber string, emailService email.Service, fromEmail, webSubdomain string, metricsRegistry metrics.Registry) http.Handler {
	router := mux.NewRouter()
	// Better a blank page for root than a 404
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		www.TemplateResponse(w, http.StatusOK, www.IndexTemplate, &www.BaseTemplateContext{Title: "Spruce"})
	})
	passreset.RouteResetPassword(router, dataAPI, authAPI, twilioCli, fromNumber, emailService, fromEmail, webSubdomain, metricsRegistry.Scope("reset_password"))
	return router
}
