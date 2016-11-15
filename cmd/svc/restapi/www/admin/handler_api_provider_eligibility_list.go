package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/mux"
)

type providerEligibilityListAPIHandler struct {
	dataAPI api.DataAPI
}

type statePathwayMapping struct {
	StateCode   string `json:"state_code"`
	PathwayTag  string `json:"pathway_tag"`
	Notify      bool   `json:"notify"`
	Unavailable bool   `json:"unavailable"`
}

type statePathwayMappingUpdate struct {
	ID          int64 `json:"id,string"`
	Notify      *bool `json:"notify"`
	Unavailable *bool `json:"unavailable"`
}

type providerEligibilityUpdateRequest struct {
	Delete []encoding.ObjectID          `json:"delete,omitempty"`
	Create []*statePathwayMapping       `json:"create,omitempty"`
	Update []*statePathwayMappingUpdate `json:"update,omitempty"`
}

func newProviderEligibilityListAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&providerEligibilityListAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get, httputil.Patch)
}

func (h *providerEligibilityListAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r.Context())["id"], 10, 64)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	account := www.MustCtxAccount(r.Context())

	switch r.Method {
	case httputil.Get:
		audit.LogAction(account.ID, "AdminAPI", "ListDoctorEligibility", map[string]interface{}{"doctor_id": doctorID})
		h.get(w, r, doctorID)
	case httputil.Patch:
		audit.LogAction(account.ID, "AdminAPI", "UpdateDoctorEligibility", map[string]interface{}{"doctor_id": doctorID})
		h.patch(w, r, doctorID)
	}
}

func (h *providerEligibilityListAPIHandler) get(w http.ResponseWriter, r *http.Request, doctorID int64) {
	mappings, err := h.dataAPI.CareProviderStatePathwayMappings(&api.CareProviderStatePathwayMappingQuery{
		Provider: api.Provider{
			Role: api.RoleDoctor,
			ID:   doctorID,
		},
	})
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, &providerMappingsResponse{
		Mappings: mappings,
	})
}

func (h *providerEligibilityListAPIHandler) patch(w http.ResponseWriter, r *http.Request, doctorID int64) {
	var req providerEligibilityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		www.APIBadRequestError(w, r, "unable to parse body")
		return
	}

	patch := &api.CareProviderStatePathwayMappingPatch{}
	for _, id := range req.Delete {
		patch.Delete = append(patch.Delete, id.Int64())
	}
	for _, c := range req.Create {
		patch.Create = append(patch.Create, &api.CareProviderStatePathway{
			Provider: api.Provider{
				Role: api.RoleDoctor,
				ID:   doctorID,
			},
			StateCode:   c.StateCode,
			PathwayTag:  c.PathwayTag,
			Notify:      c.Notify,
			Unavailable: c.Unavailable,
		})
	}
	for _, u := range req.Update {
		patch.Update = append(patch.Update, &api.CareProviderStatePathwayMappingUpdate{
			ID:          u.ID,
			Notify:      u.Notify,
			Unavailable: u.Unavailable,
		})
	}

	if err := h.dataAPI.UpdateCareProviderStatePathwayMapping(patch); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	h.get(w, r, doctorID)
}
