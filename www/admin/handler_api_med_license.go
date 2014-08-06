package admin

import (
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/www"
)

type medicalLicenseAPIHandler struct {
	dataAPI api.DataAPI
}

func NewMedicalLicenseAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&medicalLicenseAPIHandler{
		dataAPI: dataAPI,
	}, []string{"GET"})
}

func (h *medicalLicenseAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	licenses, err := h.dataAPI.MedicalLicenses(doctorID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	www.JSONResponse(w, r, http.StatusOK, licenses)
}
