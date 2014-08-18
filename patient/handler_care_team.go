package patient

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
)

type careTeamHandler struct {
	dataAPI api.DataAPI
}

func NewCareTeamHandler(dataAPI api.DataAPI) http.Handler {
	return &careTeamHandler{
		dataAPI: dataAPI,
	}
}

func (c *careTeamHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)
	if ctxt.Role != api.PATIENT_ROLE {
		return false, nil
	}

	return true, nil
}

func (c *careTeamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_GET {
		http.NotFound(w, r)
		return
	}

	patientId, err := c.dataAPI.GetPatientIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	careTeam, err := c.dataAPI.GetActiveMembersOfCareTeamForPatient(patientId, true)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSON(w, map[string]interface{}{
		"care_team": careTeam,
	})
}
