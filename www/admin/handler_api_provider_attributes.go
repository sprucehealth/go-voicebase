package admin

import (
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/www"
)

type providerAttributesAPIHandler struct {
	dataAPI api.DataAPI
}

func NewProviderAttributesAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&providerAttributesAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get)
}

func (h *providerAttributesAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	account := context.Get(r, www.CKAccount).(*common.Account)
	audit.LogAction(account.ID, "AdminAPI", "GetDoctorAttributes", map[string]interface{}{"doctor_id": doctorID})

	attributes, err := h.dataAPI.DoctorAttributes(doctorID, nil)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, attributes)
}
