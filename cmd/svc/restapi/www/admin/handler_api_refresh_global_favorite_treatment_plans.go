package admin

import (
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/mux"
)

type syncGlobalFTPHandler struct {
	dataAPI api.DataAPI
}

func newSyncGlobalFTPHandler(
	dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(
		&syncGlobalFTPHandler{
			dataAPI: dataAPI,
		}, httputil.Post)
}

func (h *syncGlobalFTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	account := www.MustCtxAccount(r.Context())

	doctorID, err := strconv.ParseInt(mux.Vars(r.Context())["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	audit.LogAction(account.ID, "AdminAPI", "SyncGlobalFTPs", map[string]interface{}{
		"doctor_id": doctorID,
	})

	if err := h.dataAPI.SyncGlobalFTPsForDoctor(doctorID); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}
