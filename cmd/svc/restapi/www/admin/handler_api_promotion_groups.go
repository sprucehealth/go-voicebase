package admin

import (
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/responses"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/httputil"
)

type promotionGroupsHandler struct {
	dataAPI api.DataAPI
}

// PromotionGroupsGETResponse represents the data returned by sucessful GET requests to promotionHandler
type PromotionGroupsGETResponse struct {
	PromotionGroups []*responses.PromotionGroup `json:"promotion_groups"`
}

// newPromotionGroupsHandler returns a new initialized instance of promotionGroupsHandler
func newPromotionGroupsHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&promotionGroupsHandler{dataAPI: dataAPI}, httputil.Get)
}

func (h *promotionGroupsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httputil.Get:
		h.serveGET(w, r)
	}
}

func (h *promotionGroupsHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	promotionGroups, err := h.dataAPI.PromotionGroups()
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	resps := make([]*responses.PromotionGroup, len(promotionGroups))
	for i, v := range promotionGroups {
		resps[i] = responses.TransformPromotionGroup(v)
	}
	httputil.JSONResponse(w, http.StatusOK, &PromotionGroupsGETResponse{PromotionGroups: resps})
}
