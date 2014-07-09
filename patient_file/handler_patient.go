package patient_file

import (
	"strconv"

	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/surescripts"

	"encoding/json"
	"net/http"

	"github.com/sprucehealth/backend/third_party/github.com/SpruceHealth/schema"
)

type doctorPatientHandler struct {
	DataApi              api.DataAPI
	ErxApi               erx.ERxAPI
	AddressValidationApi address.AddressValidationAPI
}

func NewDoctorPatientHandler(dataApi api.DataAPI, erxApi erx.ERxAPI, addressValidationApi address.AddressValidationAPI) *doctorPatientHandler {
	return &doctorPatientHandler{
		DataApi:              dataApi,
		ErxApi:               erxApi,
		AddressValidationApi: addressValidationApi,
	}
}

func (d *doctorPatientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case apiservice.HTTP_GET:
		d.getPatientInformation(w, r)
	case apiservice.HTTP_PUT:
		d.updatePatientInformation(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

type requestResponstData struct {
	Patient   *common.Patient `json:"patient"`
	PatientId string          `schema:"patient_id,required" json:"-"`
}

func (d *doctorPatientHandler) getPatientInformation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	requestData := requestResponstData{}
	if err := schema.NewDecoder().Decode(&requestData, r.Form); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	currentDoctor, err := d.DataApi.GetDoctorFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the doctor based on account id: "+err.Error())
		return
	}

	patientId, err := strconv.ParseInt(requestData.PatientId, 10, 64)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse patient id: "+err.Error())
		return
	}

	patient, err := d.DataApi.GetPatientFromId(patientId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to get patient information from id: "+err.Error())
		return
	}

	if !patient.IsUnlinked {
		if err := apiservice.ValidateDoctorAccessToPatientFile(currentDoctor.DoctorId.Int64(), patient.PatientId.Int64(), d.DataApi); err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &requestResponstData{Patient: patient})
}

func (d *doctorPatientHandler) updatePatientInformation(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	requestData := &requestResponstData{}
	if err := json.NewDecoder(r.Body).Decode(requestData); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input body that is meant to be the patient object: "+err.Error())
		return
	}

	err := surescripts.ValidatePatientInformation(requestData.Patient, d.AddressValidationApi, d.DataApi)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get the erx id for the patient, if it exists in the database
	existingPatientInfo, err := d.DataApi.GetPatientFromId(requestData.Patient.PatientId.Int64())
	if err != nil {
		apiservice.WriteUserError(w, http.StatusBadRequest, err.Error())
		return
	}

	currentDoctor, err := d.DataApi.GetDoctorFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get doctor from account id: "+err.Error())
		return
	}

	if !existingPatientInfo.IsUnlinked {
		if err := apiservice.ValidateDoctorAccessToPatientFile(currentDoctor.DoctorId.Int64(), requestData.Patient.PatientId.Int64(), d.DataApi); err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	requestData.Patient.ERxPatientId = existingPatientInfo.ERxPatientId
	requestData.Patient.Pharmacy = existingPatientInfo.Pharmacy

	if err := d.ErxApi.UpdatePatientInformation(currentDoctor.DoseSpotClinicianId, requestData.Patient); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, `Unable to update patient information on dosespot. 
			Failing to avoid out of sync issues between surescripts and our platform `+err.Error())
		return
	}

	// update the doseSpot Id for the patient in our system now that we got one
	if !existingPatientInfo.ERxPatientId.IsValid {
		if err := d.DataApi.UpdatePatientWithERxPatientId(requestData.Patient.PatientId.Int64(), requestData.Patient.ERxPatientId.Int64()); err != nil {
			apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update the patientId from dosespot: "+err.Error())
			return
		}
	}

	// go ahead and udpate the doctor's information in our system now that dosespot has it
	if err := d.DataApi.UpdatePatientInformation(requestData.Patient, true); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update patient information on our databsae: "+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, apiservice.SuccessfulGenericJSONResponse())
}
