package patient_file

import (
	"net/http"
	"sort"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/responses"
)

type alertsHandler struct {
	dataAPI api.DataAPI
}

func NewAlertsHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.RequestCacheHandler(
				apiservice.AuthorizationRequired(
					&alertsHandler{
						dataAPI: dataAPI,
					})),
			api.RoleDoctor),
		httputil.Get)

}

type alertsRequestData struct {
	PatientID int64 `schema:"patient_id"`
	CaseID    int64 `schema:"case_id"`
	VisitID   int64 `schema:"patient_visit_id"`
}

type alertsResponse struct {
	Alerts []*responses.Alert `json:"alerts"`
}

func (a *alertsHandler) IsAuthorized(ctx context.Context, r *http.Request) (bool, error) {
	requestCache := apiservice.MustCtxCache(ctx)
	account := apiservice.MustCtxAccount(ctx)

	requestData := &alertsRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error())
	}
	requestCache[apiservice.CKRequestData] = requestData

	if requestData.PatientID == 0 && requestData.CaseID == 0 && requestData.VisitID == 0 {
		return false, apiservice.NewValidationError("patient_id or case_id or patient_visit_id must be specified")
	}

	doctorID, err := a.dataAPI.GetDoctorIDFromAccountID(account.ID)
	if err != nil {
		return false, err
	}

	switch {
	case requestData.CaseID > 0:
		pc, err := a.dataAPI.GetPatientCaseFromID(requestData.CaseID)
		if err != nil {
			return false, err
		}

		if err := apiservice.ValidateAccessToPatientCase(
			r.Method,
			account.Role,
			doctorID,
			pc.PatientID.Int64(),
			pc.ID.Int64(),
			a.dataAPI); err != nil {
			return false, err
		}
	case requestData.VisitID > 0:
		visit, err := a.dataAPI.GetPatientVisitFromID(requestData.VisitID)
		if err != nil {
			return false, err
		}

		if err := apiservice.ValidateAccessToPatientCase(
			r.Method,
			account.Role,
			doctorID,
			visit.PatientID.Int64(),
			visit.PatientCaseID.Int64(),
			a.dataAPI); err != nil {
			return false, err
		}
	case requestData.PatientID > 0:
		if err := apiservice.ValidateDoctorAccessToPatientFile(
			r.Method,
			account.Role,
			doctorID,
			requestData.PatientID,
			a.dataAPI); err != nil {
			return false, err
		}
	}

	requestCache[apiservice.CKRequestData] = requestData

	return true, nil
}

func (a *alertsHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	rd := requestCache[apiservice.CKRequestData].(*alertsRequestData)

	visitID := rd.VisitID
	var err error

	// fall back to caseID or patientID if the visitID is not specified
	switch {
	case visitID > 0:

	case rd.CaseID > 0:
		// get the alerts for the latest visitID pertaining to the case

		visitID, err = a.getVisitIDFromCaseID(rd.CaseID)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}

	case rd.PatientID > 0:
		// get the alerts for the latest visitID of the latest submitted case for the patient

		cases, err := a.dataAPI.GetCasesForPatient(rd.PatientID, common.SubmittedPatientCaseStates())
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}

		if len(cases) > 0 {

			sort.Sort(sort.Reverse(common.ByPatientCaseCreationDate(cases)))
			caseID := cases[0].ID.Int64()

			visitID, err = a.getVisitIDFromCaseID(caseID)
			if err != nil {
				apiservice.WriteError(ctx, err, w, r)
				return
			}
		}
	}

	alerts, err := a.dataAPI.AlertsForVisit(visitID)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	response := &alertsResponse{
		Alerts: make([]*responses.Alert, len(alerts)),
	}

	for i, alert := range alerts {
		response.Alerts[i] = responses.TransformAlert(alert)
	}

	httputil.JSONResponse(w, http.StatusOK, response)
}

func (a *alertsHandler) getVisitIDFromCaseID(caseID int64) (int64, error) {
	visits, err := a.dataAPI.GetVisitsForCase(caseID, common.NonOpenPatientVisitStates())
	if err != nil {
		return 0, err
	}

	var visitID int64
	if len(visits) > 0 {
		sort.Sort(sort.Reverse(common.ByPatientVisitCreationDate(visits)))
		visitID = visits[0].ID.Int64()
	}

	return visitID, nil
}
