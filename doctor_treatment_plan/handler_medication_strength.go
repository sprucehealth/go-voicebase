package doctor_treatment_plan

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/httputil"
)

type medicationStrengthSearchHandler struct {
	erxAPI  erx.ERxAPI
	dataAPI api.DataAPI
}

func NewMedicationStrengthSearchHandler(dataAPI api.DataAPI, erxAPI erx.ERxAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(&medicationStrengthSearchHandler{
			dataAPI: dataAPI,
			erxAPI:  erxAPI,
		}), httputil.Get)
}

type MedicationStrengthRequestData struct {
	MedicationName string `schema:"drug_internal_name,required"`
}

type MedicationStrengthSearchResponse struct {
	MedicationStrengths []string `json:"dosage_strength_options"`
}

func (m *medicationStrengthSearchHandler) IsAuthorized(r *http.Request) (bool, error) {
	if apiservice.GetContext(r).Role != api.RoleDoctor {
		return false, apiservice.NewAccessForbiddenError()
	}
	return true, nil
}

func (m *medicationStrengthSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestData := new(MedicationStrengthRequestData)
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	doctor, err := m.dataAPI.GetDoctorFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	medicationStrengths, err := m.erxAPI.SearchForMedicationStrength(doctor.DoseSpotClinicianID, requestData.MedicationName)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	medicationStrengthResponse := &MedicationStrengthSearchResponse{MedicationStrengths: medicationStrengths}
	httputil.JSONResponse(w, http.StatusOK, medicationStrengthResponse)
}
