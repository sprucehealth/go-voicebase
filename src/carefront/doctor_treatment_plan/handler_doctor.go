package doctor_treatment_plan

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/libs/dispatch"
	"fmt"
	"net/http"
)

type doctorTreatmentPlanHandler struct {
	dataApi api.DataAPI
}

func NewDoctorTreatmentPlanHandler(dataApi api.DataAPI) *doctorTreatmentPlanHandler {
	return &doctorTreatmentPlanHandler{
		dataApi: dataApi,
	}
}

type DoctorTreatmentPlanRequestData struct {
	DoctorFavoriteTreatmentPlanId int64 `schema:"dr_favorite_treatment_plan_id" json:"dr_favorite_treatment_plan_id,string"`
	TreatmentPlanId               int64 `schema:"treatment_plan_id" json:"treatment_plan_id,string"`
	PatientVisitId                int64 `schema:"patient_visit_id" json:"patient_visit_id,string"`
	Abridged                      bool  `schema:"abridged" json:"abridged"`
}

type DoctorTreatmentPlanResponse struct {
	TreatmentPlan *common.DoctorTreatmentPlan `json:"treatment_plan"`
}

func (d *doctorTreatmentPlanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case apiservice.HTTP_GET:
		d.getTreatmentPlan(w, r)
	case apiservice.HTTP_PUT:
		d.pickATreatmentPlan(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (d *doctorTreatmentPlanHandler) getTreatmentPlan(w http.ResponseWriter, r *http.Request) {
	requestData := &DoctorTreatmentPlanRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	} else if requestData.TreatmentPlanId == 0 {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "treatment_plan_id not specified")
		return
	}

	doctorId, err := d.dataApi.GetDoctorIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	drTreatmentPlan, err := d.dataApi.GetAbridgedTreatmentPlan(requestData.TreatmentPlanId, doctorId)
	if err == api.NoRowsError {
		apiservice.WriteDeveloperError(w, http.StatusNotFound, "No treatment plan exists for patient visit")
		return
	} else if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get treatment plan for patient visit: "+err.Error())
		return
	}

	// if we are dealing with a draft, and the owner of the treatment plan does not match the doctor requesting it,
	// return an error because this should never be the case
	if drTreatmentPlan.Status == api.STATUS_DRAFT && drTreatmentPlan.DoctorId.Int64() != doctorId {
		apiservice.WriteValidationError("Cannot retrieve draft treatment plan owned by different doctor", w, r)
		return
	}

	// only return the small amount of information retreived about the treatment plan
	if requestData.Abridged {
		apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorTreatmentPlanResponse{TreatmentPlan: drTreatmentPlan})
		return
	}

	if err := fillInTreatmentPlan(drTreatmentPlan, doctorId, d.dataApi); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorTreatmentPlanResponse{TreatmentPlan: drTreatmentPlan})
}

func (d *doctorTreatmentPlanHandler) pickATreatmentPlan(w http.ResponseWriter, r *http.Request) {
	requestData := &DoctorTreatmentPlanRequestData{}

	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	} else if requestData.PatientVisitId == 0 {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "PatientVisitId not specified")
		return
	}

	patientVisitReviewData, statusCode, err := apiservice.ValidateDoctorAccessToPatientVisitAndGetRelevantData(requestData.PatientVisitId, apiservice.GetContext(r).AccountId, d.dataApi)
	if err != nil {
		apiservice.WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	patientVisitStatus := patientVisitReviewData.PatientVisit.Status
	if patientVisitStatus != api.CASE_STATUS_REVIEWING && patientVisitStatus != api.CASE_STATUS_SUBMITTED {
		apiservice.WriteDeveloperError(w, http.StatusForbidden, fmt.Sprintf("Unable to start a new treatment plan for a patient visit that is in the %s state", patientVisitReviewData.PatientVisit.Status))
		return
	}

	// Start new treatment plan for patient visit (indicate favorite treatment plan if indicated)
	// Note that this method deletes any pre-existing treatment plan
	treatmentPlanId, err := d.dataApi.StartNewTreatmentPlanForPatientVisit(patientVisitReviewData.PatientVisit.PatientId.Int64(),
		patientVisitReviewData.PatientVisit.PatientVisitId.Int64(), patientVisitReviewData.DoctorId, requestData.DoctorFavoriteTreatmentPlanId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to start new treatment plan for patient visit: "+err.Error())
		return
	}

	// get the treatment plan just created
	drTreatmentPlan, err := d.dataApi.GetAbridgedTreatmentPlan(treatmentPlanId, patientVisitReviewData.DoctorId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	if err := fillInTreatmentPlan(drTreatmentPlan, patientVisitReviewData.DoctorId, d.dataApi); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	dispatch.Default.Publish(&NewTreatmentPlanStartedEvent{
		DoctorId:        patientVisitReviewData.DoctorId,
		PatientVisitId:  requestData.PatientVisitId,
		TreatmentPlanId: treatmentPlanId,
	})

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorTreatmentPlanResponse{TreatmentPlan: drTreatmentPlan})
}
