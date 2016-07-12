package doctor

import (
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/libs/httputil"
)

type refillRxDenialReasonsHandler struct {
	dataAPI api.DataAPI
}

func NewRefillRxDenialReasonsHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(&refillRxDenialReasonsHandler{
				dataAPI: dataAPI,
			}), api.RoleDoctor, api.RoleCC),
		httputil.Get)
}

type RefillRequestDenialReasonsResponse struct {
	DenialReasons []*api.RefillRequestDenialReason `json:"refill_request_denial_reasons"`
}

func (d *refillRxDenialReasonsHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	denialReasons, err := d.dataAPI.GetRefillRequestDenialReasons()
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, &RefillRequestDenialReasonsResponse{DenialReasons: denialReasons})
}
