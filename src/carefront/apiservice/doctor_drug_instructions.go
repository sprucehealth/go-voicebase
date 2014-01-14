package apiservice

import (
	"carefront/api"
	"carefront/common"
	"encoding/json"
	"github.com/gorilla/schema"
	"net/http"
)

type DoctorDrugInstructionsHandler struct {
	DataApi api.DataAPI
}

type GetDoctorDrugInstructionsRequestData struct {
	DrugInternalName string `schema:"drug_internal_name"`
}

type DeleteDrugInstructionsResponse struct {
	Result string `json:"result"`
}

type DoctorDrugInstructionsRequestResponse struct {
	AllSupplementalInstructions      []*common.DoctorInstructionItem `json:"all_supplemental_instructions"`
	DrugInternalName                 string                          `json:"drug_internal_name"`
	TreatmentId                      int64                           `json:"treatment_id,string,omitempty"`
	PatientVisitId                   int64                           `json:"patient_visit_id,string,omitempty"`
	SelectedSupplementalInstructions []*common.DoctorInstructionItem `json:"selected_supplemental_instructions,omitempty"`
}

func NewDoctorDrugInstructionsHandler(dataApi api.DataAPI) *DoctorDrugInstructionsHandler {
	return &DoctorDrugInstructionsHandler{DataApi: dataApi}
}

func (d *DoctorDrugInstructionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		d.getDrugInstructions(w, r)
	case "POST":
		d.addDrugInstructions(w, r)
	}
}

func (d *DoctorDrugInstructionsHandler) addDrugInstructions(w http.ResponseWriter, r *http.Request) {
	jsonDecoder := json.NewDecoder(r.Body)
	addInstructionsRequestBody := &DoctorDrugInstructionsRequestResponse{}

	err := jsonDecoder.Decode(addInstructionsRequestBody)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse json request body for adding instructions: "+err.Error())
		return
	}

	drugName, drugForm, drugRoute := breakDrugInternalNameIntoComponents(addInstructionsRequestBody.DrugInternalName)

	doctorId, _, _, statusCode, err := ValidateDoctorAccessToPatientVisitAndGetRelevantData(addInstructionsRequestBody.PatientVisitId, GetContext(r).AccountId, d.DataApi)
	if err != nil {
		WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	// go update the drug instructions based on the global list
	newOrUpdatedInstructionToIdMapping := make(map[string]int64)
	updatedInstructionList := make([]*common.DoctorInstructionItem, 0)
	for _, drugInstructionItem := range addInstructionsRequestBody.AllSupplementalInstructions {
		switch drugInstructionItem.State {
		case common.STATE_ADDED, common.STATE_MODIFIED:
			err = d.DataApi.AddOrUpdateDrugInstructionForDoctor(drugName, drugForm, drugRoute, drugInstructionItem, doctorId)
			if err != nil {
				WriteDeveloperError(w, http.StatusInternalServerError, "Unable to add instructions for doctor: "+err.Error())
				return
			}
			newOrUpdatedInstructionToIdMapping[drugInstructionItem.Text] = drugInstructionItem.Id
			updatedInstructionList = append(updatedInstructionList, drugInstructionItem)
		case common.STATE_DELETED:
			err := d.DataApi.DeleteDrugInstructionForDoctor(drugInstructionItem, doctorId)
			if err != nil {
				WriteDeveloperError(w, http.StatusInternalServerError, "Unable to add instruction for doctor: "+err.Error())
				return
			}
		default:
			updatedInstructionList = append(updatedInstructionList, drugInstructionItem)
		}
		// empty out the state now that it has been taken care of
		drugInstructionItem.State = ""
	}

	// go through the selected supplemental instructions to assign ids to them
	for _, selectedInstructionItem := range addInstructionsRequestBody.SelectedSupplementalInstructions {
		updatedOrNewId := newOrUpdatedInstructionToIdMapping[selectedInstructionItem.Text]
		if updatedOrNewId != 0 {
			selectedInstructionItem.Id = updatedOrNewId
		}
	}

	err = d.DataApi.AddDrugInstructionsToTreatment(drugName, drugForm, drugRoute, addInstructionsRequestBody.SelectedSupplementalInstructions, addInstructionsRequestBody.TreatmentId, doctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to add drug instructions to treatment: "+err.Error())
		return
	}

	addInstructionsRequestBody.AllSupplementalInstructions = updatedInstructionList

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, addInstructionsRequestBody)
}

func (d *DoctorDrugInstructionsHandler) getDrugInstructions(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestData := new(GetDoctorDrugInstructionsRequestData)
	decoder := schema.NewDecoder()
	err := decoder.Decode(requestData, r.Form)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	doctorId, err := d.DataApi.GetDoctorIdFromAccountId(GetContext(r).AccountId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the doctor id from the account id "+err.Error())
		return
	}

	drugName, drugForm, drugRoute := breakDrugInternalNameIntoComponents(requestData.DrugInternalName)
	// break up drug name into its components
	drugInstructions, err := d.DataApi.GetDrugInstructionsForDoctor(drugName, drugForm, drugRoute, doctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get drug instructions for doctor: "+err.Error())
		return
	}
	WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorDrugInstructionsRequestResponse{AllSupplementalInstructions: drugInstructions, DrugInternalName: requestData.DrugInternalName})
}
