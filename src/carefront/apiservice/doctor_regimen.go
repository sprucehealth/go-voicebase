package apiservice

import (
	"carefront/api"
	"carefront/common"
	"carefront/encoding"
	"carefront/libs/dispatch"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/schema"
)

type DoctorRegimenHandler struct {
	DataApi api.DataAPI
}

type GetDoctorRegimenRequestData struct {
	PatientVisitId  int64 `schema:"patient_visit_id"`
	TreatmentPlanId int64 `schema:"treatment_plan_id"`
}

type DoctorRegimenRequestResponse struct {
	RegimenSteps     []*common.DoctorInstructionItem `json:"regimen_steps"`
	DrugInternalName string                          `json:"drug_internal_name,omitempty"`
	PatientVisitId   int64                           `json:"patient_visit_id,string,omitempty"`
}

func NewDoctorRegimenHandler(dataApi api.DataAPI) *DoctorRegimenHandler {
	return &DoctorRegimenHandler{DataApi: dataApi}
}

func (d *DoctorRegimenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case HTTP_GET:
		d.getRegimenSteps(w, r)
	case HTTP_POST:
		d.updateRegimenSteps(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (d *DoctorRegimenHandler) getRegimenSteps(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	requestData := new(GetDoctorRegimenRequestData)

	if err := schema.NewDecoder().Decode(requestData, r.Form); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	patientVisitId := requestData.PatientVisitId
	treatmentPlanId := requestData.TreatmentPlanId
	if err := ensureTreatmentPlanOrPatientVisitIdPresent(d.DataApi, treatmentPlanId, &patientVisitId); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	patientVisitReviewData, statusCode, err := ValidateDoctorAccessToPatientVisitAndGetRelevantData(patientVisitId, GetContext(r).AccountId, d.DataApi)
	if err != nil {
		WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	regimenSteps, err := d.DataApi.GetRegimenStepsForDoctor(patientVisitReviewData.DoctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get regimen steps for doctor: "+err.Error())
		return
	}

	if treatmentPlanId == 0 {
		treatmentPlanId, err = d.DataApi.GetActiveTreatmentPlanForPatientVisit(patientVisitReviewData.DoctorId, patientVisitId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get treatment plan for patient visit: "+err.Error())
			return
		}
	}

	regimenPlan, err := d.DataApi.GetRegimenPlanForTreatmentPlan(treatmentPlanId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to lookup regimen plan for patient visit: "+err.Error())
		return
	}

	if regimenPlan == nil {
		regimenPlan = &common.RegimenPlan{}
	}
	regimenPlan.AllRegimenSteps = regimenSteps

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, regimenPlan)
}

func (d *DoctorRegimenHandler) updateRegimenSteps(w http.ResponseWriter, r *http.Request) {
	jsonDecoder := json.NewDecoder(r.Body)
	requestData := &common.RegimenPlan{}

	err := jsonDecoder.Decode(requestData)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse json request body for updating regimen steps: "+err.Error())
		return
	}

	patientVisitReviewData, statusCode, err := ValidateDoctorAccessToPatientVisitAndGetRelevantData(requestData.PatientVisitId.Int64(), GetContext(r).AccountId, d.DataApi)
	if err != nil {
		WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	err = EnsurePatientVisitInExpectedStatus(d.DataApi, requestData.PatientVisitId.Int64(), api.CASE_STATUS_REVIEWING)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	// ensure that all regimen steps in the regimen sections actually exist in the client global list
	for _, regimenSection := range requestData.RegimenSections {
		for _, regimenStep := range regimenSection.RegimenSteps {
			if httpStatusCode, err := d.ensureRegimenStepExistsInMasterList(regimenStep, requestData, patientVisitReviewData.DoctorId); err != nil {
				WriteDeveloperError(w, httpStatusCode, err.Error())
				return
			}
		}
	}

	// compare the master list of regimen steps from the client with the active list
	// that we have stored on the server
	currentActiveRegimenSteps, err := d.DataApi.GetRegimenStepsForDoctor(patientVisitReviewData.DoctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get regimen steps for doctor: "+err.Error())
		return
	}
	regimenStepsToDelete := make([]*common.DoctorInstructionItem, 0, len(currentActiveRegimenSteps))
	for _, currentRegimenStep := range currentActiveRegimenSteps {
		regimenStepFound := false
		for _, regimenStep := range requestData.AllRegimenSteps {
			if regimenStep.Id.Int64() == currentRegimenStep.Id.Int64() {
				regimenStepFound = true
				break
			}
		}
		if !regimenStepFound {
			regimenStepsToDelete = append(regimenStepsToDelete, currentRegimenStep)
		}
	}
	err = d.DataApi.MarkRegimenStepsToBeDeleted(regimenStepsToDelete, patientVisitReviewData.DoctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to delete regimen steps that are no longer in the client list: "+err.Error())
		return
	}

	// Go through regimen steps to add and update regimen steps before creating the regimen plan
	// for the user
	newStepToIdMapping := make(map[string][]int64)
	// keep track of the multiple items that could have the exact same text associated with it
	updatedStepToIdMapping := make(map[int64]int64)
	updatedAllRegimenSteps := make([]*common.DoctorInstructionItem, 0)
	for _, regimenStep := range requestData.AllRegimenSteps {
		switch regimenStep.State {
		case common.STATE_ADDED:
			err = d.DataApi.AddRegimenStepForDoctor(regimenStep, patientVisitReviewData.DoctorId)
			if err != nil {
				WriteDeveloperError(w, http.StatusInternalServerError, "Unable to add reigmen step to doctor. Application may be left in inconsistent state. Error = "+err.Error())
				return
			}
			newStepToIdMapping[regimenStep.Text] = append(newStepToIdMapping[regimenStep.Text], regimenStep.Id.Int64())
			updatedAllRegimenSteps = append(updatedAllRegimenSteps, regimenStep)
		case common.STATE_MODIFIED:
			previousRegimenStepId := regimenStep.Id.Int64()
			err = d.DataApi.UpdateRegimenStepForDoctor(regimenStep, patientVisitReviewData.DoctorId)
			if err != nil {
				WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update regimen step for doctor: "+err.Error())
				return
			}
			// keep track of the new id for updated regimen steps so that we can update the regimen step in the
			// regimen section
			updatedStepToIdMapping[previousRegimenStepId] = regimenStep.Id.Int64()
			updatedAllRegimenSteps = append(updatedAllRegimenSteps, regimenStep)
		default:
			updatedAllRegimenSteps = append(updatedAllRegimenSteps, regimenStep)
		}
	}

	// go through regimen steps within the regimen sections to assign ids to the new steps that dont have them
	for _, regimenSection := range requestData.RegimenSections {
		for _, regimenStep := range regimenSection.RegimenSteps {

			if newIds, ok := newStepToIdMapping[regimenStep.Text]; ok {
				regimenStep.ParentId = encoding.NewObjectId(newIds[0])
				// update the list to remove the item just used
				newStepToIdMapping[regimenStep.Text] = newIds[1:]
			}

			if updatedId, ok := updatedStepToIdMapping[regimenStep.ParentId.Int64()]; ok {
				// update the parentId to point to the new updated regimen step
				regimenStep.ParentId = encoding.NewObjectId(updatedId)
			}
		}
	}

	treatmentPlanId := requestData.TreatmentPlanId.Int64()
	if treatmentPlanId == 0 {
		treatmentPlanId, err = d.DataApi.GetActiveTreatmentPlanForPatientVisit(patientVisitReviewData.DoctorId, requestData.PatientVisitId.Int64())
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get treatment plan for patient visit: "+err.Error())
			return
		}
	}

	requestData.TreatmentPlanId = encoding.NewObjectId(treatmentPlanId)
	err = d.DataApi.CreateRegimenPlanForPatientVisit(requestData)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to create regimen plan for patient visit: "+err.Error())
		return
	}

	// fetch all regimen steps in the treatment plan and the global regimen steps to
	// return an updated view of the world to the client
	regimenPlan, err := d.DataApi.GetRegimenPlanForTreatmentPlan(treatmentPlanId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the regimen plan for treatment plan: "+err.Error())
		return
	}

	allRegimenSteps, err := d.DataApi.GetRegimenStepsForDoctor(patientVisitReviewData.DoctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the list of regimen steps for doctor: "+err.Error())
		return
	}

	regimenPlan = &common.RegimenPlan{
		RegimenSections: regimenPlan.RegimenSections,
		AllRegimenSteps: allRegimenSteps,
		TreatmentPlanId: encoding.NewObjectId(treatmentPlanId),
		PatientVisitId:  requestData.PatientVisitId,
	}

	dispatch.Default.PublishAsync(&RegimenPlanAddedEvent{
		PatientVisitId: patientVisitReviewData.PatientVisit.PatientVisitId.Int64(),
		RegimenPlan:    requestData,
		DoctorId:       patientVisitReviewData.DoctorId,
	})

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, regimenPlan)
}

func (d *DoctorRegimenHandler) ensureRegimenStepExistsInMasterList(regimenStep *common.DoctorInstructionItem, regimenPlan *common.RegimenPlan, doctorId int64) (int, error) {
	regimenStepFound := false

	for _, globalRegimenStep := range regimenPlan.AllRegimenSteps {
		// if we are working with new items in the list do a text comparison to check for the existence of the step
		if globalRegimenStep.Id.Int64() == 0 && regimenStep.ParentId.Int64() == 0 {
			if globalRegimenStep.Text == regimenStep.Text {
				regimenStepFound = true
				break
			}
		} else if globalRegimenStep.Id.Int64() == regimenStep.ParentId.Int64() {
			if globalRegimenStep.Text != regimenStep.Text {
				// check the text to ensure that its the same
				return http.StatusBadRequest, errors.New("The text of an item in the regimen section cannot be different from that in the global list if they are considered linked.")
			}
			regimenStepFound = true
			break
		} else if regimenStep.ParentId.Int64() != 0 {
			// its possible that the step is not present in the active global list but exists as a
			// step from the past
			parentRegimenStep, err := d.DataApi.GetRegimenStepForDoctor(regimenStep.ParentId.Int64(), doctorId)
			if err == api.NoRowsError {
				return http.StatusBadRequest, errors.New("Cannot have a step in a regimen section that does not link to a regimen step the doctor created at some point.")
			} else if err != nil {
				return http.StatusInternalServerError, errors.New("Unable to get a regimen step for a doctor: " + err.Error())
			}
			// if the parent regimen step does exist, ensure that the text matches up
			if parentRegimenStep.Text != regimenStep.Text && regimenStep.State != common.STATE_MODIFIED {
				return http.StatusBadRequest, errors.New("Cannot modify the text of a regimen step that is linked to a parent regimen step without indicating intent via STATE=MODIFIED")
			}
			regimenStepFound = true
			break
		}
	}

	if !regimenStepFound {
		return http.StatusBadRequest, errors.New("Regimen step in the section for the patient visit not found in the global list of all regimen steps")
	}

	return 0, nil
}
