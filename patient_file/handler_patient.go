package patient_file

import (
	"net/http"

	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/surescripts"
)

type doctorPatientHandler struct {
	dataAPI              api.DataAPI
	erxAPI               erx.ERxAPI
	addressValidationAPI address.AddressValidationAPI
}

func NewDoctorPatientHandler(
	dataAPI api.DataAPI,
	erxAPI erx.ERxAPI,
	addressValidationAPI address.AddressValidationAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(&doctorPatientHandler{
			dataAPI:              dataAPI,
			erxAPI:               erxAPI,
			addressValidationAPI: addressValidationAPI,
		}), []string{"GET", "PUT"})
}

func (d *doctorPatientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	req := ctxt.RequestCache[apiservice.RequestData].(*requestResponstData)

	if err := req.PatientUpdate.Validate(); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	switch r.Method {
	case apiservice.HTTP_GET:
		d.getPatientInformation(w, r)
	case apiservice.HTTP_PUT:
		d.updatePatientInformation(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

type patientUpdate struct {
	PatientID    int64                 `json:"id,string"`
	FirstName    string                `json:"first_name,omitempty"`
	LastName     string                `json:"last_name,omiempty"`
	MiddleName   string                `json:"middle_name,omitempty"`
	Suffix       string                `json:"suffix,omitempty"`
	Prefix       string                `json:"prefix,omitempty"`
	DOB          encoding.DOB          `json:"dob,omitempty"`
	Gender       string                `json:"gender,omitempty"`
	PhoneNumbers []*common.PhoneNumber `json:"phone_numbers,omitempty"`
	Address      *common.Address       `json:"address,omitempty"`
}

type requestResponstData struct {
	PatientUpdate *patientUpdate `json:"patient"`
	PatientID     int64          `schema:"patient_id,required" json:"-"`
}

type patientResponse struct {
	Patient *common.Patient `json:"patient"`
}

func (d *doctorPatientHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)

	requestData := &requestResponstData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error())
	}
	ctxt.RequestCache[apiservice.RequestData] = requestData

	doctor, err := d.dataAPI.GetDoctorFromAccountID(ctxt.AccountID)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.Doctor] = doctor

	patientID := requestData.PatientID
	if requestData.PatientUpdate != nil {
		patientID = requestData.PatientUpdate.PatientID
	}

	patient, err := d.dataAPI.GetPatientFromID(patientID)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.Patient] = patient

	if err := apiservice.ValidateDoctorAccessToPatientFile(r.Method,
		ctxt.Role,
		doctor.DoctorID.Int64(),
		patient.PatientID.Int64(),
		d.dataAPI); err != nil {
		return false, err
	}
	return true, nil
}

func (d *doctorPatientHandler) getPatientInformation(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	patient := ctxt.RequestCache[apiservice.Patient].(*common.Patient)

	apiservice.WriteJSON(w, &patientResponse{
		Patient: patient,
	})
}

func (d *doctorPatientHandler) updatePatientInformation(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	req := ctxt.RequestCache[apiservice.RequestData].(*requestResponstData)
	doctor := ctxt.RequestCache[apiservice.Doctor].(*common.Doctor)
	patient := ctxt.RequestCache[apiservice.Patient].(*common.Patient)

	patient.FirstName = req.PatientUpdate.FirstName
	patient.LastName = req.PatientUpdate.LastName
	patient.MiddleName = req.PatientUpdate.MiddleName
	patient.Suffix = req.PatientUpdate.Suffix
	patient.Prefix = req.PatientUpdate.Prefix
	patient.DOB = req.PatientUpdate.DOB
	patient.Gender = req.PatientUpdate.Gender
	patient.PatientAddress = req.PatientUpdate.Address
	patient.PhoneNumbers = req.PatientUpdate.PhoneNumbers

	err := surescripts.ValidatePatientInformation(patient, d.addressValidationAPI, d.dataAPI)
	if err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	if err := d.erxAPI.UpdatePatientInformation(doctor.DoseSpotClinicianID, patient); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// update the doseSpot ID for the patient in our system now that we got one
	if !patient.ERxPatientID.IsValid {
		if err := d.dataAPI.UpdatePatientWithERxPatientID(patient.PatientID.Int64(), patient.ERxPatientID.Int64()); err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	// go ahead and udpate the doctor's information in our system now that dosespot has it
	update := &api.PatientUpdate{
		FirstName:    &req.PatientUpdate.FirstName,
		MiddleName:   &req.PatientUpdate.MiddleName,
		LastName:     &req.PatientUpdate.LastName,
		Prefix:       &req.PatientUpdate.Prefix,
		Suffix:       &req.PatientUpdate.Suffix,
		DOB:          &req.PatientUpdate.DOB,
		Gender:       &req.PatientUpdate.Gender,
		PhoneNumbers: req.PatientUpdate.PhoneNumbers,
		Address:      req.PatientUpdate.Address,
	}
	if err := d.dataAPI.UpdatePatient(patient.PatientID.Int64(), update, true); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSONSuccess(w)
}

func (u *patientUpdate) Validate() error {
	if u == nil {
		return nil
	}
	if len(u.PhoneNumbers) == 0 {
		return apiservice.NewValidationError("at least one phone number is required")
	}
	if err := u.DOB.Validate(); err != nil {
		return err
	}
	if u.Address == nil {
		return apiservice.NewValidationError("address is required")
	}
	if u.FirstName == "" {
		return apiservice.NewValidationError("first name is required")
	}
	if u.LastName == "" {
		return apiservice.NewValidationError("last name is required")
	}
	return nil
}
