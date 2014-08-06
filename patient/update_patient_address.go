package patient

import (
	"net/http"

	"github.com/sprucehealth/backend/api"

	"github.com/sprucehealth/backend/third_party/github.com/SpruceHealth/schema"
)

const (
	BILLING_ADDRESS_TYPE = "BILLING"
)

type UpdatePatientAddressHandler struct {
	DataApi     api.DataAPI
	AddressType string
}

type UpdatePatientAddressRequestData struct {
	AddressLine1 string `schema:"address_line_1,required"`
	AddressLine2 string `schema:"address_line_2"`
	City         string `schema:"city,required"`
	State        string `schema:"state,required"`
	Zipcode      string `schema:"zip_code,required"`
}

func (u *UpdatePatientAddressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case HTTP_POST:
		u.updatePatientAddress(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (u *UpdatePatientAddressHandler) updatePatientAddress(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	var requestData UpdatePatientAddressRequestData
	if err := schema.NewDecoder().Decode(&requestData, r.Form); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	patientId, err := u.DataApi.GetPatientIdFromAccountId(GetContext(r).AccountId)
	if err != nil {
		WriteDeveloperError(w, http.StatusOK, "Unable to get patient id from account id: "+err.Error())
		return
	}

	err = u.DataApi.UpdatePatientAddress(patientId, requestData.AddressLine1, requestData.AddressLine2, requestData.City, requestData.State, requestData.Zipcode, u.AddressType)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update patient address: "+err.Error())
		return
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, SuccessfulGenericJSONResponse())
}
