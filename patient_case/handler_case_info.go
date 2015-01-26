package patient_case

import (
	"fmt"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/responses"
)

type caseInfoHandler struct {
	dataAPI   api.DataAPI
	apiDomain string
}

func NewCaseInfoHandler(dataAPI api.DataAPI, apiDomain string) http.Handler {
	return httputil.SupportedMethods(
		apiservice.NoAuthorizationRequired(
			apiservice.SupportedRoles(
				&caseInfoHandler{
					dataAPI:   dataAPI,
					apiDomain: apiDomain,
				}, []string{api.PATIENT_ROLE, api.DOCTOR_ROLE})),
		[]string{"GET"})
}

type caseInfoRequestData struct {
	CaseID int64 `schema:"case_id"`
}

type caseInfoResponseData struct {
	Case       *responses.Case `json:"case"`
	CaseConfig struct {
		MessagingEnabled            bool   `json:"messaging_enabled"`
		MessagingDisabledReason     string `json:"messaging_disabled_reason"`
		TreatmentPlanEnabled        bool   `json:"treatment_plan_enabled"`
		TreatmentPlanDisabledReason string `json:"treatment_plan_disabled_reason"`
	} `json:"case_config"`
}

func (c *caseInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestData := &caseInfoRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	} else if requestData.CaseID == 0 {
		apiservice.WriteValidationError("case_id must be specified", w, r)
		return
	}

	patientCase, err := c.dataAPI.GetPatientCaseFromID(requestData.CaseID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	responseData := &caseInfoResponseData{}

	ctxt := apiservice.GetContext(r)
	switch ctxt.Role {
	case api.PATIENT_ROLE:
		patientID, err := c.dataAPI.GetPatientIDFromAccountID(ctxt.AccountID)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		// ensure that the case is owned by the patient
		if patientID != patientCase.PatientID.Int64() {
			apiservice.WriteAccessNotAllowedError(w, r)
			return
		}

		// messaging is enabled even if the case is not claimed as the patient should be able to message with the MA
		// at any time
		responseData.CaseConfig.MessagingEnabled = true

		// treatment plan is enabled if one exists
		activeTreatmentPlanExists, err := c.dataAPI.DoesActiveTreatmentPlanForCaseExist(patientCase.ID.Int64())
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		if !activeTreatmentPlanExists {
			responseData.CaseConfig.TreatmentPlanDisabledReason = "Your doctor will create a custom treatment plan just for you."
		} else {
			responseData.CaseConfig.TreatmentPlanEnabled = true
		}

	case api.DOCTOR_ROLE:
		doctorID, err := c.dataAPI.GetDoctorIDFromAccountID(ctxt.AccountID)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		if err := apiservice.ValidateAccessToPatientCase(r.Method, ctxt.Role, doctorID, patientCase.PatientID.Int64(), requestData.CaseID, c.dataAPI); err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	patientVisits, err := c.dataAPI.GetVisitsForCase(patientCase.ID.Int64(), nil)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// set the case level diagnosis to be that of the latest treated patient visit
	var diagnosis string
	for _, visit := range patientVisits {
		if visit.Status == common.PVStatusTreated {
			diagnosis, err = c.dataAPI.DiagnosisForVisit(visit.PatientVisitID.Int64())
			if !api.IsErrNotFound(err) && err != nil {
				apiservice.WriteError(err, w, r)
				return
			}
			break
		}
	}

	if patientCase.Status == common.PCStatusUnsuitable {
		diagnosis = "Unsuitable for Spruce"
	} else if diagnosis == "" {
		diagnosis = "Pending"
	}

	// only set the care team if the patient has been claimed or the case has been marked as unsuitable
	var careTeamMembers []*responses.PatientCareTeamMember
	if patientCase.Status == common.PCStatusClaimed || patientCase.Status == common.PCStatusUnsuitable {
		// get the care team for case
		members, err := c.dataAPI.GetActiveMembersOfCareTeamForCase(requestData.CaseID, true)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		careTeamMembers = make([]*responses.PatientCareTeamMember, len(members))
		for i, member := range members {
			careTeamMembers[i] = responses.TransformCareTeamMember(member, c.apiDomain)
		}
	}
	responseData.Case = &responses.Case{
		ID:           patientCase.ID.Int64(),
		CaseID:       patientCase.ID.Int64(),
		PathwayTag:   patientCase.PathwayTag,
		Title:        fmt.Sprintf("%s Case", patientCase.Name),
		CreationDate: &patientCase.CreationDate,
		Status:       patientCase.Status.String(),
		Diagnosis:    diagnosis,
		CareTeam:     careTeamMembers,
	}

	apiservice.WriteJSON(w, &responseData)
}
