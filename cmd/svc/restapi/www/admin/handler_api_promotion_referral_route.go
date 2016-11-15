package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/mux"
)

type promotionReferralRouteHandler struct {
	dataAPI api.DataAPI
}

// PromotionReferralRoutePUTRequest represents the expected structure of a PUT request
type PromotionReferralRoutePUTRequest struct {
	Lifecycle string `json:"lifecycle"`
}

// NewPromotionReferralRouteHandler returns an initialized instance of thpromotionReferralRouteHandlere
func newPromotionReferralRouteHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&promotionReferralRouteHandler{dataAPI: dataAPI}, httputil.Put)
}

func (h *promotionReferralRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r.Context())["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	switch r.Method {
	case httputil.Put:
		req, err := h.parsePUTRequest(r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.servePUT(w, r, req, id)
	}
}

func (h *promotionReferralRouteHandler) parsePUTRequest(r *http.Request) (*PromotionReferralRoutePUTRequest, error) {
	rd := &PromotionReferralRoutePUTRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	if rd.Lifecycle == "" {
		return nil, errors.New("lifecycle required")
	}

	return rd, nil
}

func (h *promotionReferralRouteHandler) servePUT(w http.ResponseWriter, r *http.Request, req *PromotionReferralRoutePUTRequest, id int64) {
	lifecycle, err := common.ParsePRRLifecycle(req.Lifecycle)
	if err != nil {
		www.APIBadRequestError(w, r, err.Error())
		return
	}

	routeUpdate := &common.PromotionReferralRouteUpdate{
		ID:        id,
		Lifecycle: lifecycle,
	}

	if _, err := h.dataAPI.UpdatePromotionReferralRoute(routeUpdate); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}
