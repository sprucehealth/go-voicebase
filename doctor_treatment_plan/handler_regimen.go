package doctor_treatment_plan

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/httputil"
)

type regimenHandler struct {
	dataAPI    api.DataAPI
	dispatcher *dispatch.Dispatcher
}

type DoctorRegimenRequestResponse struct {
	RegimenSteps     []*common.DoctorInstructionItem `json:"regimen_steps"`
	DrugInternalName string                          `json:"drug_internal_name,omitempty"`
}

func NewRegimenHandler(dataAPI api.DataAPI, dispatcher *dispatch.Dispatcher) http.Handler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(&regimenHandler{
			dataAPI:    dataAPI,
			dispatcher: dispatcher,
		}), []string{"POST"})
}

func (d *regimenHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)

	requestData := &common.RegimenPlan{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error())
	} else if requestData.TreatmentPlanID.Int64() == 0 {
		return false, apiservice.NewValidationError("treatment_plan_id must be specified")
	}
	ctxt.RequestCache[apiservice.RequestData] = requestData

	doctorID, err := d.dataAPI.GetDoctorIDFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.DoctorID] = doctorID

	patientID, err := d.dataAPI.GetPatientIDFromTreatmentPlanID(requestData.TreatmentPlanID.Int64())
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.PatientID] = patientID

	// can only add regimen for a treatment that is a draft
	treatmentPlan, err := d.dataAPI.GetAbridgedTreatmentPlan(requestData.TreatmentPlanID.Int64(), doctorID)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.TreatmentPlan] = treatmentPlan

	if err := apiservice.ValidateAccessToPatientCase(r.Method, ctxt.Role, doctorID, patientID, treatmentPlan.PatientCaseID.Int64(), d.dataAPI); err != nil {
		return false, err
	}

	return true, nil
}

func (d *regimenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	treatmentPlan := ctxt.RequestCache[apiservice.TreatmentPlan].(*common.TreatmentPlan)
	doctorID := ctxt.RequestCache[apiservice.DoctorID].(int64)
	requestData := ctxt.RequestCache[apiservice.RequestData].(*common.RegimenPlan)

	if !treatmentPlan.InDraftMode() {
		apiservice.WriteValidationError("treatment plan must be in draft mode", w, r)
		return
	}

	// ensure that all regimen steps in the regimen sections actually exist in the client global list
	for _, regimenSection := range requestData.Sections {
		for _, regimenStep := range regimenSection.Steps {
			if httpStatusCode, err := d.ensureLinkedRegimenStepExistsInMasterList(regimenStep, requestData, doctorID); err != nil {
				apiservice.WriteDeveloperError(w, httpStatusCode, err.Error())
				return
			}
		}
	}

	// compare the master list of regimen steps from the client with the active list
	// that we have stored on the server
	currentActiveRegimenSteps, err := d.dataAPI.GetRegimenStepsForDoctor(doctorID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}
	regimenStepsToDelete := make([]*common.DoctorInstructionItem, 0, len(currentActiveRegimenSteps))
	for _, currentRegimenStep := range currentActiveRegimenSteps {
		regimenStepFound := false
		for _, regimenStep := range requestData.AllSteps {
			if regimenStep.ID.Int64() == currentRegimenStep.ID.Int64() {
				regimenStepFound = true
				break
			}
		}
		if !regimenStepFound {
			regimenStepsToDelete = append(regimenStepsToDelete, currentRegimenStep)
		}
	}
	err = d.dataAPI.MarkRegimenStepsToBeDeleted(regimenStepsToDelete, doctorID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// Go through regimen steps to add and update regimen steps before creating the regimen plan
	// for the user
	newStepToIdMapping := make(map[string][]int64)
	// keep track of the multiple items that could have the exact same text associated with it
	updatedStepToIdMapping := make(map[int64]int64)
	updatedAllRegimenSteps := make([]*common.DoctorInstructionItem, 0)
	for _, regimenStep := range requestData.AllSteps {
		switch regimenStep.State {
		case common.STATE_ADDED:
			err = d.dataAPI.AddRegimenStepForDoctor(regimenStep, doctorID)
			if err != nil {
				apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to add reigmen step to doctor. Application may be left in inconsistent state. Error = "+err.Error())
				return
			}
			newStepToIdMapping[regimenStep.Text] = append(newStepToIdMapping[regimenStep.Text], regimenStep.ID.Int64())
			updatedAllRegimenSteps = append(updatedAllRegimenSteps, regimenStep)
		case common.STATE_MODIFIED:
			previousRegimenStepId := regimenStep.ID.Int64()
			err = d.dataAPI.UpdateRegimenStepForDoctor(regimenStep, doctorID)
			if err != nil {
				apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update regimen step for doctor: "+err.Error())
				return
			}
			// keep track of the new id for updated regimen steps so that we can update the regimen step in the
			// regimen section
			updatedStepToIdMapping[previousRegimenStepId] = regimenStep.ID.Int64()
			updatedAllRegimenSteps = append(updatedAllRegimenSteps, regimenStep)
		default:
			updatedAllRegimenSteps = append(updatedAllRegimenSteps, regimenStep)
		}
	}

	// go through regimen steps within the regimen sections to assign ids to the new steps that dont have them
	for _, regimenSection := range requestData.Sections {

		for _, regimenStep := range regimenSection.Steps {

			if newIds, ok := newStepToIdMapping[regimenStep.Text]; ok {
				regimenStep.ParentID = encoding.NewObjectID(newIds[0])
				// update the list to move the item just used to the back of the queue
				newStepToIdMapping[regimenStep.Text] = append(newIds[1:], newIds[0])
			} else if updatedID, ok := updatedStepToIdMapping[regimenStep.ParentID.Int64()]; ok {
				// update the parentId to point to the new updated regimen step
				regimenStep.ParentID = encoding.NewObjectID(updatedID)
			} else if regimenStep.State == common.STATE_MODIFIED || regimenStep.State == common.STATE_ADDED {
				// break any linkage to the parent step because the text is no longer the same and the regimen step does
				// not exist in the master list
				regimenStep.ParentID = encoding.ObjectID{}
			}
		}
	}

	err = d.dataAPI.CreateRegimenPlanForTreatmentPlan(requestData)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to create regimen plan for patient visit: "+err.Error())
		return
	}

	// fetch all regimen steps in the treatment plan and the global regimen steps to
	// return an updated view of the world to the client
	regimenPlan, err := d.dataAPI.GetRegimenPlanForTreatmentPlan(requestData.TreatmentPlanID.Int64())
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the regimen plan for treatment plan: "+err.Error())
		return
	}

	allRegimenSteps, err := d.dataAPI.GetRegimenStepsForDoctor(doctorID)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the list of regimen steps for doctor: "+err.Error())
		return
	}

	regimenPlan = &common.RegimenPlan{
		Sections:        regimenPlan.Sections,
		AllSteps:        allRegimenSteps,
		TreatmentPlanID: requestData.TreatmentPlanID,
		Status:          api.STATUS_COMMITTED,
	}

	d.dispatcher.PublishAsync(&RegimenPlanAddedEvent{
		TreatmentPlanID: requestData.TreatmentPlanID.Int64(),
		RegimenPlan:     requestData,
		DoctorID:        doctorID,
	})

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, regimenPlan)
}

func (d *regimenHandler) ensureLinkedRegimenStepExistsInMasterList(regimenStep *common.DoctorInstructionItem, regimenPlan *common.RegimenPlan, doctorID int64) (int, error) {
	// no need to check if the regimen step does not indicate that it exists in the master list
	if !regimenStep.ParentID.IsValid {
		return http.StatusOK, nil
	}

	// search for the regimen step against the current master list returned from the client
	for _, globalRegimenStep := range regimenPlan.AllSteps {

		if !globalRegimenStep.ID.IsValid {
			continue
		}

		// break the linkage if the text doesn't match
		if globalRegimenStep.ID.Int64() == regimenStep.ParentID.Int64() {
			if globalRegimenStep.Text != regimenStep.Text {
				regimenStep.ParentID = encoding.ObjectID{}
			}
			return http.StatusOK, nil
		}
	}

	// its possible that the step is not present in the active global list but exists as a
	// step from the past
	parentRegimenStep, err := d.dataAPI.GetRegimenStepForDoctor(regimenStep.ParentID.Int64(), doctorID)
	if err != nil {
		regimenStep.ParentID = encoding.ObjectID{}
	}

	// if the parent regimen step does exist, ensure that the text matches up, and if not break the linkage
	if parentRegimenStep.Text != regimenStep.Text && regimenStep.State != common.STATE_MODIFIED {
		regimenStep.ParentID = encoding.ObjectID{}
	}

	return http.StatusOK, nil
}
