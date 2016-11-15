package admin

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/treatment_plan"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/mux"
)

type rxGuidesAPIHandler struct {
	dataAPI api.DataAPI
}

func newRXGuideAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&rxGuidesAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get)
}

func (h *rxGuidesAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r.Context())["id"], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	account := www.MustCtxAccount(r.Context())
	audit.LogAction(account.ID, "AdminAPI", "GetRXGuide", map[string]interface{}{"id": id})

	details, err := h.dataAPI.DrugDetails(id)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	var html string

	if r.FormValue("with_html") != "" {
		// Fill in some sample content
		treatment := &common.Treatment{
			DrugName:            details.Name,
			PatientInstructions: "The doctors instructions will go here. This text is just to show what the RX guide will visually look like.",
			Doctor: &common.Doctor{
				ShortDisplayName: "Dr. Kohen",
			},
		}

		b := &bytes.Buffer{}
		if err := treatment_plan.RenderRXGuide(b, details, treatment, nil); err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		html = b.String()
	}

	httputil.JSONResponse(w, http.StatusOK, &struct {
		Guide *common.DrugDetails `json:"guide"`
		HTML  string              `json:"html"`
	}{
		Guide: details,
		HTML:  html,
	})
}
