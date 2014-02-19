package apiservice

import (
	"carefront/libs/erx"
	"net/http"

	"github.com/gorilla/schema"
)

type MedicationStrengthSearchHandler struct {
	ERxApi erx.ERxAPI
}

type MedicationStrengthRequestData struct {
	MedicationName string `schema:"drug_internal_name,required"`
}

type MedicationStrengthSearchResponse struct {
	MedicationStrengths []string `json:"dosage_strength_options"`
}

func (m *MedicationStrengthSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != HTTP_GET {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	requestData := new(MedicationStrengthRequestData)
	decoder := schema.NewDecoder()
	err := decoder.Decode(requestData, r.Form)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	medicationStrengths, err := m.ERxApi.SearchForMedicationStrength(requestData.MedicationName)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get medication strength results for given drug: "+err.Error())
		return
	}

	medicationStrengthResponse := &MedicationStrengthSearchResponse{MedicationStrengths: medicationStrengths}
	WriteJSONToHTTPResponseWriter(w, http.StatusOK, medicationStrengthResponse)
}
