package patient

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
)

type emergencyContactsHandler struct {
	dataAPI api.DataAPI
}

type emergencyContactsData struct {
	EmergencyContacts []*common.EmergencyContact `json:"emergency_contacts,omitempty"`
}

func NewEmergencyContactsHandler(dataAPI api.DataAPI) http.Handler {
	return &emergencyContactsHandler{
		dataAPI: dataAPI,
	}
}

func (e *emergencyContactsHandler) IsAuthorized(r *http.Request) (bool, error) {
	if apiservice.GetContext(r).Role != api.PATIENT_ROLE {
		return false, nil
	}
	return true, nil
}

func (e *emergencyContactsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case apiservice.HTTP_GET:
		e.getEmergencyContacts(w, r)
	case apiservice.HTTP_PUT:
		e.addEmergencyContacts(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (e *emergencyContactsHandler) getEmergencyContacts(w http.ResponseWriter, r *http.Request) {
	patientId, err := e.dataAPI.GetPatientIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	emergencyContacts, err := e.dataAPI.GetPatientEmergencyContacts(patientId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSON(w, &emergencyContactsData{EmergencyContacts: emergencyContacts})
}

func (e *emergencyContactsHandler) addEmergencyContacts(w http.ResponseWriter, r *http.Request) {
	requestData := &emergencyContactsData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	// validate
	for _, eContact := range requestData.EmergencyContacts {
		if eContact.FullName == "" {
			apiservice.WriteValidationError("Please enter emergency contact's name", w, r)
			return
		} else if eContact.Relationship == "" {
			apiservice.WriteValidationError("Please enter emergency contact's relationship", w, r)
			return
		} else if eContact.PhoneNumber == "" {
			apiservice.WriteValidationError("Please enter emergency contact's phone number", w, r)
			return
		}
	}

	patientId, err := e.dataAPI.GetPatientIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	if err := e.dataAPI.AddPatientEmergencyContacts(patientId, requestData.EmergencyContacts); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSON(w, requestData)
}
