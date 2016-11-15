package patient_visit

import (
	"encoding/json"
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
)

type pathwaySTPHandler struct {
	dataAPI api.DataAPI
}

type pathwaySTPRequest struct {
	PathwayTag string `schema:"pathway_id"`
}

type pathwaySTPResponse struct {
	SampleTreatmentPlan interface{} `json:"sample_treatment_plan"`
}

func NewPathwaySTPHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.NoAuthorizationRequired(
			&pathwaySTPHandler{
				dataAPI: dataAPI,
			}), httputil.Get)
}

func (p *pathwaySTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var rd pathwaySTPRequest
	if err := apiservice.DecodeRequestData(&rd, r); err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	stp, err := p.dataAPI.PathwaySTP(rd.PathwayTag)
	if api.IsErrNotFound(err) {
		apiservice.WriteResourceNotFoundError("Oops! Something went wrong and we couldn't find the correct Sample Treatment Plan.", w, r)
		return
	} else if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	var stpJSON map[string]interface{}
	if err := json.Unmarshal(stp, &stpJSON); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, pathwaySTPResponse{
		SampleTreatmentPlan: stpJSON,
	})
}
