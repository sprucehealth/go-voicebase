package demo

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/libs/httputil"
)

type demoVisitHandler struct {
	dataAPI api.DataAPI
}

func NewTrainingCasesHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(
				&demoVisitHandler{
					dataAPI: dataAPI,
				}), []string{api.DOCTOR_ROLE}),
		[]string{"POST"})
}

func (d *demoVisitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doctorID, err := d.dataAPI.GetDoctorIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	if err := d.dataAPI.ClaimTrainingSet(doctorID, api.HEALTH_CONDITION_ACNE_ID); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSONSuccess(w)
}
