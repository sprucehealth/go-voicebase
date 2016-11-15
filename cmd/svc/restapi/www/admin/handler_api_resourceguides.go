package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/mux"
)

type resourceGuidesAPIHandler struct {
	dataAPI api.DataAPI
}

func newResourceGuidesAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&resourceGuidesAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get, httputil.Patch)
}

func (h *resourceGuidesAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r.Context())["id"], 10, 64)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	account := www.MustCtxAccount(r.Context())
	if r.Method == "PATCH" {
		audit.LogAction(account.ID, "AdminAPI", "UpdateResourceGuide", map[string]interface{}{"guide_id": id})

		var update api.ResourceGuideUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		if err := h.dataAPI.UpdateResourceGuide(id, &update); err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		httputil.JSONResponse(w, http.StatusOK, true)
		return
	}

	audit.LogAction(account.ID, "AdminAPI", "GetResourceGuide", map[string]interface{}{"guide_id": id})

	guide, err := h.dataAPI.GetResourceGuide(id)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, guide)
}
