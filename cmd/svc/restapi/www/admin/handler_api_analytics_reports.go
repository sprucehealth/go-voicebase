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

type analyticsReportsAPIHandler struct {
	dataAPI api.DataAPI
}

func newAnalyticsReportsAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&analyticsReportsAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get, httputil.Post)
}

func (h *analyticsReportsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r.Context())["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	account := www.MustCtxAccount(r.Context())

	if r.Method == httputil.Post {
		audit.LogAction(account.ID, "AdminAPI", "UpdateAnalyticsReport", map[string]interface{}{"report_id": id})

		updateReq := &struct {
			Name         *string `json:"name"`
			Query        *string `json:"query"`
			Presentation *string `json:"presentation"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(updateReq); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		if err := h.dataAPI.UpdateAnalyticsReport(id, updateReq.Name, updateReq.Query, updateReq.Presentation); api.IsErrNotFound(err) {
			www.APINotFound(w, r)
			return
		} else if err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, true)
		return
	}

	audit.LogAction(account.ID, "AdminAPI", "GetAnalyticsReport", map[string]interface{}{"report_id": id})

	report, err := h.dataAPI.AnalyticsReport(id)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, report)
}
