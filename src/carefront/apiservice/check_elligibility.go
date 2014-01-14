package apiservice

import (
	"carefront/api"
	"carefront/common"
	"carefront/libs/maps"
	"fmt"
	"github.com/gorilla/schema"
	"net/http"
	"strings"
)

type CheckCareProvidingElligibilityHandler struct {
	DataApi     api.DataAPI
	MapsService maps.MapsService
}

type CheckCareProvidingElligibilityRequestData struct {
	Zipcode string `schema:"zip_code,required"`
}

type CheckCareProvidingElligibilityResponse struct {
	Doctor *common.Doctor `json:"doctor"`
}

func (c *CheckCareProvidingElligibilityHandler) NonAuthenticated() bool {
	return true
}

func (c *CheckCareProvidingElligibilityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestData := new(CheckCareProvidingElligibilityRequestData)
	decoder := schema.NewDecoder()
	err := decoder.Decode(requestData, r.Form)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input to check elligibility: "+err.Error())
		return
	}

	// given the zipcode, cover to city and state info
	cityStateInfo, err := c.MapsService.ConvertZipcodeToCityState(requestData.Zipcode)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to use the maps service to reverse geocode the given zipcode to city and state information: "+err.Error())
		return
	}

	doctorId, err := c.DataApi.CheckCareProvidingElligibility(cityStateInfo.ShortStateName, HEALTH_CONDITION_ACNE_ID)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to check elligiblity for the patient to be seen by doctor: "+err.Error())
		return
	}

	if doctorId != 0 {
		doctor, err := c.DataApi.GetDoctorFromId(doctorId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get doctor from id: "+err.Error())
			return
		}
		doctor.ThumbnailUrl = strings.ToLower(fmt.Sprintf("%s%s.%s_thumbnail", api.SpruceImageBaseUrl, doctor.FirstName, doctor.LastName))
		WriteJSONToHTTPResponseWriter(w, http.StatusOK, &CheckCareProvidingElligibilityResponse{Doctor: doctor})
	} else {
		WriteUserError(w, http.StatusForbidden, "Patient cannot be seen in this state.")
	}
}
