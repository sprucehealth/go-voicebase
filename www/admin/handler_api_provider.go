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

type providerAPIHandler struct {
	dataAPI api.DataAPI
}

func NewProviderAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&providerAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get)
}

func (h *providerAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	account := context.Get(r, www.CKAccount).(*common.Account)
	audit.LogAction(account.ID, "AdminAPI", "GetDoctor", map[string]interface{}{"doctor_id": doctorID})

	doctor, err := h.dataAPI.GetDoctorFromID(doctorID)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, &struct {
		Doctor *common.Doctor `json:"doctor"`
	}{
		Doctor: doctor,
	})
}
