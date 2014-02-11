package apiservice

import (
	"carefront/api"
	"carefront/common"
	"github.com/gorilla/schema"
	"net/http"
	"time"
)

type DoctorPrescriptionsHandler struct {
	DataApi api.DataAPI
}

type DoctorPrescriptionsRequestData struct {
	FromTimeUnix int64 `schema:"from"`
	ToTimeUnix   int64 `schema:"to"`
}

type DoctorPrescriptionsResponse struct {
	TreatmentPlans []*common.TreatmentPlan `json:"treatment_plans"`
}

func (d *DoctorPrescriptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestData := new(DoctorPrescriptionsRequestData)
	decoder := schema.NewDecoder()
	err := decoder.Decode(requestData, r.Form)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	// ensure that from and to date are specified
	if requestData.FromTimeUnix == 0 || requestData.ToTimeUnix == 0 {
		WriteDeveloperError(w, http.StatusBadRequest, "From and to times (in time since epoch) need to be specified!")
		return
	}

	doctorId, err := d.DataApi.GetDoctorIdFromAccountId(GetContext(r).AccountId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get doctor id from account id: "+err.Error())
		return
	}

	fromTime := time.Unix(requestData.FromTimeUnix, 0)
	toTime := time.Unix(requestData.ToTimeUnix, 0)

	treatmentPlans, err := d.DataApi.GetCompletedPrescriptionsForDoctor(fromTime, toTime, doctorId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get completed prescriptions for doctor: "+err.Error())
		return
	}

	// find a list of unique patients for which to get information
	uniquePatientIdsBookKeeping := make(map[int64]bool)
	uniquePatientIds := make([]int64, 0)
	for _, treatmentPlan := range treatmentPlans {
		if !uniquePatientIdsBookKeeping[treatmentPlan.PatientId] {
			uniquePatientIds = append(uniquePatientIds, treatmentPlan.PatientId)
			uniquePatientIdsBookKeeping[treatmentPlan.PatientId] = true
		}
	}

	patients, err := d.DataApi.GetPatientsForIds(uniquePatientIds)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the patients based on ids: "+err.Error())
		return
	}

	pharmacies, err := d.DataApi.GetPatientPharmacySelectionForPatients(uniquePatientIds)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get pharmacies for patients based on idsL "+err.Error())
		return
	}

	for _, pharmacySelection := range pharmacies {
		for _, patient := range patients {
			if patient.PatientId == pharmacySelection.PatientId {
				patient.Pharmacy = pharmacySelection
			}
		}
	}

	for _, treatmentPlan := range treatmentPlans {
		for _, patient := range patients {
			if patient.PatientId == treatmentPlan.PatientId {
				treatmentPlan.PatientInfo = patient
			}
		}
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorPrescriptionsResponse{TreatmentPlans: treatmentPlans})

}
