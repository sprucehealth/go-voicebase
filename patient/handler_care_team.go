package patient

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/responses"
)

type careTeamHandler struct {
	dataAPI   api.DataAPI
	apiDomain string
}

func NewCareTeamHandler(dataAPI api.DataAPI, apiDomain string) http.Handler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(&careTeamHandler{
			dataAPI:   dataAPI,
			apiDomain: apiDomain,
		}), []string{"GET"})
}

type careTeamResponse struct {
	CareTeam []*responses.PatientCareTeamMember `json:"care_team"`
}

func (c *careTeamHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)
	if ctxt.Role != api.PATIENT_ROLE {
		return false, nil
	}

	return true, nil
}

func (c *careTeamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	patientID, err := c.dataAPI.GetPatientIDFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	cases, err := c.dataAPI.GetCasesForPatient(patientID, append(common.SubmittedPatientCaseStates(), common.PCStatusOpen.String()))
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	caseIDs := make([]int64, len(cases))
	for i, pc := range cases {
		caseIDs[i] = pc.ID.Int64()
	}

	careTeams, err := c.dataAPI.CaseCareTeams(caseIDs)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	members := make([]*responses.PatientCareTeamMember, 0)
	for _, careTeam := range careTeams {
		for _, member := range careTeam.Assignments {
			if member.Status == api.STATUS_ACTIVE {
				members = append(members, responses.TransformCareTeamMember(member, c.apiDomain))
			}
		}
	}

	httputil.JSONResponse(w, http.StatusOK, &careTeamResponse{
		CareTeam: members,
	})
}
