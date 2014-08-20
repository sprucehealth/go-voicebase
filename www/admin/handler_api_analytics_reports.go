package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/context"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/www"
)

type analyticsReportsAPIHandler struct {
	dataAPI api.DataAPI
}

func NewAnalyticsReportsAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&analyticsReportsAPIHandler{
		dataAPI: dataAPI,
	}, []string{"GET", "POST"})
}

func (h *analyticsReportsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	account := context.Get(r, www.CKAccount).(*common.Account)

	if r.Method == "POST" {
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

		if err := h.dataAPI.UpdateAnalyticsReport(id, updateReq.Name, updateReq.Query, updateReq.Presentation); err == api.NoRowsError {
			www.APINotFound(w, r)
			return
		} else if err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		www.JSONResponse(w, r, http.StatusOK, true)
		return
	} else {
		audit.LogAction(account.ID, "AdminAPI", "GetAnalyticsReport", map[string]interface{}{"report_id": id})
	}

	report, err := h.dataAPI.AnalyticsReport(id)
	if err == api.NoRowsError {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	www.JSONResponse(w, r, http.StatusOK, report)
}
