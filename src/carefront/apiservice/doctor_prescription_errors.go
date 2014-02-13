package apiservice

import (
	"carefront/api"
	"carefront/common"
	"carefront/libs/erx"
	"carefront/libs/pharmacy"
	"net/http"
	"strconv"
)

type DoctorPrescriptionsErrorsHandler struct {
	DataApi api.DataAPI
	ErxApi  erx.ERxAPI
}

type DoctorPrescriptionErrorsResponse struct {
	TransmissionErrors []*transmissionErrorItem `json:"transmission_errors"`
}

type transmissionErrorItem struct {
	Treatment *common.Treatment      `json:"treatment,omitempty"`
	Patient   *common.Patient        `json:"patient,omitempty"`
	Pharmacy  *pharmacy.PharmacyData `json:"pharmacy,omitempty"`
}

func (d *DoctorPrescriptionsErrorsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	medicationsWithErrors, err := d.ErxApi.GetTransmissionErrorDetails()
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get prescription related errors: "+err.Error())
		return
	}

	transmissionErrors := make([]*transmissionErrorItem, 0)
	uniquePatientIdsBookKeeping := make(map[int64]bool)
	uniquePatientIds := make([]int64, 0)
	pharmacyIdToTransmissionErrorMapping := make(map[int64]*transmissionErrorItem)
	for _, medicationWithError := range medicationsWithErrors {
		treatment, err := d.DataApi.GetTreatmentBasedOnPrescriptionId(medicationWithError.DoseSpotPrescriptionId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get treatment based on prescription: "+err.Error())
		}

		// there is no treatment in our system based on the prescription.
		// lets not ignore the error, instead lets just show the data as we have it from dosespot
		// without linking it to patient data. This can happen in the event that the prescription id
		// did not get stored for some reason or if we have multiple doctors using the same account (incorrectly)
		if treatment == nil {
			dispenseValue, _ := strconv.ParseInt(medicationWithError.Dispense, 0, 64)
			treatment = &common.Treatment{
				PrescriptionId:          medicationWithError.DoseSpotPrescriptionId,
				ErxSentDate:             medicationWithError.PrescriptionDate,
				DrugDBIds:               medicationWithError.DrugDBIds,
				DispenseUnitId:          medicationWithError.DispenseUnitId,
				DispenseUnitDescription: medicationWithError.DispenseUnitDescription,
				DrugName:                medicationWithError.DisplayName,
				DosageStrength:          medicationWithError.Strength,
				NumberRefills:           medicationWithError.Refills,
				DaysSupply:              medicationWithError.DaysSupply,
				DispenseValue:           dispenseValue,
				PatientInstructions:     medicationWithError.Instructions,
				PharmacyNotes:           medicationWithError.PharmacyNotes,
				SubstitutionsAllowed:    !medicationWithError.NoSubstitutions,
			}
		} else {
			if !uniquePatientIdsBookKeeping[treatment.PatientId] {
				uniquePatientIdsBookKeeping[treatment.PatientId] = true
				uniquePatientIds = append(uniquePatientIds, treatment.PatientId)
			}
		}

		treatment.PrescriptionStatus = medicationWithError.PrescriptionStatus
		treatment.TransmissionErrorDate = medicationWithError.ErrorTimeStamp
		treatment.StatusDetails = medicationWithError.ErrorDetails
		if treatment.ErxSentDate == nil {
			treatment.ErxSentDate = medicationWithError.PrescriptionDate
		}

		transmissionError := &transmissionErrorItem{
			Treatment: treatment,
		}

		// keep track of which pharmacy Id maps to which transmissionError so that we can assign the pharmacy to the transmissionError
		pharmacyIdToTransmissionErrorMapping[medicationWithError.PharmacyId] = transmissionError
		transmissionErrors = append(transmissionErrors, transmissionError)
	}

	patients, err := d.DataApi.GetPatientsForIds(uniquePatientIds)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patients based on ids: "+err.Error())
		return
	}

	pharmacies, err := d.DataApi.GetPharmacySelectionForPatients(uniquePatientIds)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get pharmacies for patients based on ids: "+err.Error())
		return
	}

	for pharmacyId, transmissionError := range pharmacyIdToTransmissionErrorMapping {
		// check if the pharmacy exists in the pharmacies returned
		for _, pharmacySelection := range pharmacies {
			pharmacyIdInt, _ := strconv.ParseInt(pharmacySelection.Id, 0, 64)
			if pharmacySelection.Source != pharmacy.PHARMACY_SOURCE_SURESCRIPTS && pharmacyIdInt == pharmacyId {
				transmissionError.Pharmacy = pharmacySelection
			} else {
				// TODO lookup pharmacy from surescripts based on id and assign it here
			}
		}
	}

	for _, transmissionError := range transmissionErrors {
		for _, patient := range patients {
			if patient.PatientId == transmissionError.Treatment.PatientId {
				transmissionError.Patient = patient
			}
		}
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorPrescriptionErrorsResponse{TransmissionErrors: transmissionErrors})
}
