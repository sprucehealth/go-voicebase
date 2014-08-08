package patient

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
)

type UpdateHandler struct {
	dataApi api.DataAPI
}

func NewUpdateHandler(dataApi api.DataAPI) *UpdateHandler {
	return &UpdateHandler{
		dataApi: dataApi,
	}
}

func (u *UpdateHandler) IsAuthorized(r *http.Request) (bool, error) {
	if apiservice.GetContext(r).Role != api.PATIENT_ROLE {
		return false, apiservice.NewAccessForbiddenError()
	}
	return true, nil
}

func (u *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_PUT {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	patient, err := u.dataApi.GetPatientFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patient from id: "+err.Error())
		return
	}

	// identify the fields that the caller wants updated
	if firstName := r.FormValue("first_name"); firstName != "" {
		patient.FirstName = firstName
	}

	if lastName := r.FormValue("last_name"); lastName != "" {
		patient.LastName = lastName
	}

	if email := r.FormValue("email"); email != "" {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Updating email for patient currently not supported")
		return
	}

	if zipcode := r.FormValue("zip_code"); zipcode != "" {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Updating zipcode for patient current not supported")
		return
	}

	if gender := r.FormValue("gender"); gender != "" {
		patient.Gender = gender
	}

	if phoneNumber := r.FormValue("phone"); phoneNumber != "" {
		if len(patient.PhoneNumbers) == 0 {
			patient.PhoneNumbers = make([]*common.PhoneNumber, 1)
		}
		patient.PhoneNumbers[0] = &common.PhoneNumber{
			Phone: phoneNumber,
			Type:  api.PHONE_CELL,
		}
	}

	if dobString := r.FormValue("dob"); dobString != "" {
		patient.DOB, err = encoding.NewDOBFromString(dobString)
		if err != nil {
			apiservice.WriteUserError(w, http.StatusBadRequest, "Unable to parse dob: "+err.Error())
			return
		}
	}

	if err := u.dataApi.UpdateTopLevelPatientInformation(patient); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update top level patient information: "+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, apiservice.SuccessfulGenericJSONResponse())
}
