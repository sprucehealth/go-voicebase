package doctor_treatment_plan

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"net/http"
)

type listHandler struct {
	dataApi api.DataAPI
}

type listHandlerRequestData struct {
	PatientId int64 `schema:"patient_id"`
}

type treatmentPlansResponseData struct {
	DraftTreatmentPlans    []*common.DoctorTreatmentPlan `json:"draft_treatment_plans,omitempty"`
	ActiveTreatmentPlans   []*common.DoctorTreatmentPlan `json:"active_treatment_plans,omitempty"`
	InactiveTreatmentPlans []*common.DoctorTreatmentPlan `json:"inactive_treatment_plans,omitempty"`
}

func NewListHandler(dataApi api.DataAPI) *listHandler {
	return &listHandler{
		dataApi: dataApi,
	}
}

func (l *listHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestData := &listHandlerRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if requestData.PatientId == 0 {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "PatientId required")
		return
	}

	doctorId, err := l.dataApi.GetDoctorIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	activeTreatmentPlans, err := l.dataApi.GetAbridgedTreatmentPlanListForPatient(requestData.PatientId, api.STATUS_ACTIVE)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	inactiveTreatmentPlans, err := l.dataApi.GetAbridgedTreatmentPlanListForPatient(requestData.PatientId, api.STATUS_INACTIVE)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	draftTreatmentPlans, err := l.dataApi.GetAbridgedTreatmentPlanListInDraftForDoctor(doctorId, requestData.PatientId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &treatmentPlansResponseData{
		DraftTreatmentPlans:    draftTreatmentPlans,
		ActiveTreatmentPlans:   activeTreatmentPlans,
		InactiveTreatmentPlans: inactiveTreatmentPlans,
	})
}
