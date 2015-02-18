package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/www"
)

type medicalLicenseAPIHandler struct {
	dataAPI api.DataAPI
}

type license struct {
	State      string                      `json:"state"`
	Number     string                      `json:"number"`
	Expiration encoding.Date               `json:"expiration"`
	Status     common.MedicalLicenseStatus `json:"status"`
}

type licenseReqRes struct {
	Licenses []*license `json:"licenses"`
}

func NewMedicalLicenseAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&medicalLicenseAPIHandler{
		dataAPI: dataAPI,
	}, []string{"GET", "PUT"})
}

func (h *medicalLicenseAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	account := context.Get(r, www.CKAccount).(*common.Account)

	if r.Method == "GET" {
		audit.LogAction(account.ID, "AdminAPI", "GetDoctorMedicalLicenses", map[string]interface{}{"doctor_id": doctorID})
	} else {
		audit.LogAction(account.ID, "AdminAPI", "UpdateDoctorMedicalLicenses", map[string]interface{}{"doctor_id": doctorID})
		var req licenseReqRes
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			www.APIBadRequestError(w, r, "Failed to decode body")
			return
		}
		licenses := make([]*common.MedicalLicense, len(req.Licenses))
		for i, l := range req.Licenses {
			ll := &common.MedicalLicense{
				DoctorID: doctorID,
				State:    l.State,
				Status:   l.Status,
				Number:   l.Number,
			}
			if !l.Expiration.IsZero() {
				ll.Expiration = &l.Expiration
			}
			licenses[i] = ll
		}
		if err := h.dataAPI.UpdateMedicalLicenses(doctorID, licenses); err != nil {
			www.APIInternalError(w, r, err)
			return
		}
	}

	licenses, err := h.dataAPI.MedicalLicenses(doctorID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	res := &licenseReqRes{
		Licenses: make([]*license, len(licenses)),
	}
	for i, l := range licenses {
		ll := &license{
			State:  l.State,
			Number: l.Number,
			Status: l.Status,
		}
		if l.Expiration != nil {
			ll.Expiration = *l.Expiration
		}
		res.Licenses[i] = ll
	}

	httputil.JSONResponse(w, http.StatusOK, res)
}
