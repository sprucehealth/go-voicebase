package doctor_treatment_plan

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
)

type treatmentTemplatesHandler struct {
	dataAPI api.DataAPI
}

func NewTreatmentTemplatesHandler(dataApi api.DataAPI) *treatmentTemplatesHandler {
	return &treatmentTemplatesHandler{
		dataAPI: dataApi,
	}
}

type DoctorTreatmentTemplatesRequest struct {
	TreatmentPlanId    encoding.ObjectId                 `json:"treatment_plan_id"`
	TreatmentTemplates []*common.DoctorTreatmentTemplate `json:"treatment_templates"`
}

type DoctorTreatmentTemplatesResponse struct {
	TreatmentTemplates []*common.DoctorTreatmentTemplate `json:"treatment_templates"`
	Treatments         []*common.Treatment               `json:"treatments,omitempty"`
}

func (t *treatmentTemplatesHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)
	if ctxt.Role != api.DOCTOR_ROLE {
		return false, apiservice.NewAccessForbiddenError()
	}

	if r.Method == apiservice.HTTP_GET {
		return true, nil
	}

	requestData := &DoctorTreatmentTemplatesRequest{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, err
	} else if requestData.TreatmentPlanId.Int64() == 0 {
		return false, apiservice.NewValidationError("treatment_plan_id must be specified", r)
	}
	ctxt.RequestCache[apiservice.RequestData] = requestData

	doctorId, err := t.dataAPI.GetDoctorIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.DoctorId] = doctorId

	treatmentPlan, err := t.dataAPI.GetAbridgedTreatmentPlan(requestData.TreatmentPlanId.Int64(), doctorId)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.TreatmentPlan] = treatmentPlan

	if err := apiservice.ValidateAccessToPatientCase(r.Method, doctorId, treatmentPlan.PatientId, treatmentPlan.PatientCaseId.Int64(), t.dataAPI); err != nil {
		return false, err
	}

	return true, nil
}

func (t *treatmentTemplatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case apiservice.HTTP_GET:
		t.getTreatmentTemplates(w, r)
	case apiservice.HTTP_POST:
		t.addTreatmentTemplates(w, r)
	case apiservice.HTTP_DELETE:
		t.deleteTreatmentTemplates(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (t *treatmentTemplatesHandler) getTreatmentTemplates(w http.ResponseWriter, r *http.Request) {
	doctorId, err := t.dataAPI.GetDoctorIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get doctor from account id: "+err.Error())
		return
	}

	doctorTreatmentTemplates, err := t.dataAPI.GetTreatmentTemplates(doctorId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get favorite treatments for doctor: "+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorTreatmentTemplatesResponse{TreatmentTemplates: doctorTreatmentTemplates})
}

func (t *treatmentTemplatesHandler) deleteTreatmentTemplates(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	doctorId := ctxt.RequestCache[apiservice.DoctorId].(int64)
	requestData := ctxt.RequestCache[apiservice.RequestData].(*DoctorTreatmentTemplatesRequest)

	for _, favoriteTreatment := range requestData.TreatmentTemplates {
		if favoriteTreatment.Id.Int64() == 0 {
			apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to delete a treatment that does not have an id associated with it")
			return
		}
	}

	err := t.dataAPI.DeleteTreatmentTemplates(requestData.TreatmentTemplates, doctorId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to delete favorited treatment: "+err.Error())
		return
	}

	treatmentTemplates, err := t.dataAPI.GetTreatmentTemplates(doctorId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get favorite treatments for doctor: "+err.Error())
		return
	}

	treatmentsInTreatmentPlan, err := t.dataAPI.GetTreatmentsBasedOnTreatmentPlanId(requestData.TreatmentPlanId.Int64())
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get treatments based on treatment plan id: "+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorTreatmentTemplatesResponse{
		TreatmentTemplates: treatmentTemplates,
		Treatments:         treatmentsInTreatmentPlan,
	})
}

func (t *treatmentTemplatesHandler) addTreatmentTemplates(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	doctorId := ctxt.RequestCache[apiservice.DoctorId].(int64)
	requestData := ctxt.RequestCache[apiservice.RequestData].(*DoctorTreatmentTemplatesRequest)

	for _, treatmentTemplate := range requestData.TreatmentTemplates {
		err := apiservice.ValidateTreatment(treatmentTemplate.Treatment)
		if err != nil {
			apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
			return
		}

		// break up the name into its components so that it can be saved into the database as its components
		treatmentTemplate.Treatment.DrugName, treatmentTemplate.Treatment.DrugForm, treatmentTemplate.Treatment.DrugRoute = apiservice.BreakDrugInternalNameIntoComponents(treatmentTemplate.Treatment.DrugInternalName)
	}

	err := t.dataAPI.AddTreatmentTemplates(requestData.TreatmentTemplates, doctorId, requestData.TreatmentPlanId.Int64())
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to favorite treatment: "+err.Error())
		return
	}

	treatmentTemplates, err := t.dataAPI.GetTreatmentTemplates(doctorId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get favorited treatments for doctor: "+err.Error())
		return
	}

	treatmentsInTreatmentPlan, err := t.dataAPI.GetTreatmentsBasedOnTreatmentPlanId(requestData.TreatmentPlanId.Int64())
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorTreatmentTemplatesResponse{
		TreatmentTemplates: treatmentTemplates,
		Treatments:         treatmentsInTreatmentPlan,
	})
}
