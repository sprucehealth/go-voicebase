package api

import (
	"carefront/common"
	"carefront/libs/erx"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (d *DataService) AddRefillRequestStatusEvent(refillRequestStatus common.StatusEvent) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update rx_refill_status_events set status = ? where status = ? and rx_refill_request_id = ?`, STATUS_INACTIVE, STATUS_ACTIVE, refillRequestStatus.ErxRefillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	columnsAndData := map[string]interface{}{
		"rx_refill_request_id":  refillRequestStatus.ErxRefillRequestId,
		"rx_refill_status":      refillRequestStatus.Status,
		"rx_refill_status_date": time.Now(),
		"status":                STATUS_ACTIVE,
	}

	if !refillRequestStatus.ReportedTimestamp.IsZero() {
		columnsAndData["reported_timestamp"] = refillRequestStatus.ReportedTimestamp
	}

	keys, values := getKeysAndValuesFromMap(columnsAndData)
	_, err = tx.Exec(fmt.Sprintf(`insert into rx_refill_status_events (%s) values (%s)`, strings.Join(keys, ","), nReplacements(len(values))), values...)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) GetPendingRefillRequestStatusEventsForClinic() ([]common.StatusEvent, error) {
	rows, err := d.DB.Query(`select rx_refill_request_id, rx_refill_request.erx_request_queue_item_id, rx_refill_status, rx_refill_status_date   
								from rx_refill_status_events 
									inner join rx_refill_request on rx_refill_request_id = rx_refill_request.id
									where rx_refill_status_events.status = ? and rx_refill_status = ?`, STATUS_ACTIVE, RX_REFILL_STATUS_REQUESTED)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refillRequestStatuses := make([]common.StatusEvent, 0)
	for rows.Next() {
		var refillRequestStatus common.StatusEvent
		err = rows.Scan(&refillRequestStatus.ErxRefillRequestId, &refillRequestStatus.RxRequestQueueItemId, &refillRequestStatus.Status, &refillRequestStatus.StatusTimestamp)
		if err != nil {
			return nil, err
		}
		refillRequestStatuses = append(refillRequestStatuses, refillRequestStatus)
	}
	return refillRequestStatuses, rows.Err()
}

func (d *DataService) GetApprovedOrDeniedRefillRequestsForPatient(patientId int64) ([]common.StatusEvent, error) {
	rows, err := d.DB.Query(`select rx_refill_request_id, rx_refill_status, rx_refill_status_date, requested_prescription.erx_id    
									from rx_refill_status_events 
									inner join requested_prescription on requested_prescription.id = rx_refill_request.requested_prescription_id
										where rx_refill_status_events.rx_refill_status in ('Approved', 'Denied') and rx_refill_request.patient_id = ? 
										order by rx_refill_status_date desc`, patientId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refillRequestStatuses := make([]common.StatusEvent, 0)
	for rows.Next() {
		var refillRequestStatus common.StatusEvent
		err = rows.Scan(&refillRequestStatus.ErxRefillRequestId, &refillRequestStatus.Status, &refillRequestStatus.StatusTimestamp, &refillRequestStatus.PrescriptionId)
		if err != nil {
			return nil, err
		}
		refillRequestStatuses = append(refillRequestStatuses, refillRequestStatus)
	}
	return refillRequestStatuses, nil
}

func (d *DataService) LinkRequestedPrescriptionToOriginalTreatment(requestedTreatment *common.Treatment, patient *common.Patient) error {
	// lookup drug based on the drugIds
	if len(requestedTreatment.DrugDBIds) == 0 {
		// nothing to compare against to link to originating drug
		return nil
	}

	// lookup drugs prescribed to the patient within a day of the date the requestedPrescription was prescribed
	// we know that it was prescribed based on whether or not it was succesfully sent to the pharmacy
	halfDayBefore := requestedTreatment.ErxSentDate.Add(-12 * time.Hour)
	halfDayAfter := requestedTreatment.ErxSentDate.Add(12 * time.Hour)

	treatmentIds := make([]int64, 0)
	rows, err := d.DB.Query(`select treatment_id from erx_status_events 
								inner join treatment on treatment_id = treatment.id 
								inner join treatment_plan on treatment_plan_id = treatment.treatment_plan_id
								inner join patient_visit on patient_visit.id = treatment_plan.patient_visit_id
								where erx_status = ? and erx_status_events.creation_date >= ? and erx_status_events.creation_date <= ? and patient_visit.patient_id = ? `, ERX_STATUS_SENT, halfDayBefore, halfDayAfter, patient.PatientId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var treatmentId int64
		err = rows.Scan(&treatmentId)
		if err != nil {
			return err
		}
		treatmentIds = append(treatmentIds, treatmentId)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	for _, treatmentId := range treatmentIds {
		// for each of the treatments gathered for the patiend, compare the drug ids against the requested prescription to identify if they
		// match to find the originating prescritpion
		drugIds := make(map[string]string)
		drugDBIdRows, err := d.DB.Query(`select drug_db_id_tag, drug_db_id from drug_db_id where treatment_id= ?`, treatmentId)
		if err != nil {
			return err
		}
		defer drugDBIdRows.Close()

		for drugDBIdRows.Next() {
			var drugDbIdTag, drugDbId string
			err = drugDBIdRows.Scan(&drugDbIdTag, &drugDbId)
			if err != nil {
				return err
			}
			drugIds[drugDbIdTag] = drugDbId
		}
		if drugDBIdRows.Err() != nil {
			return drugDBIdRows.Err()
		}

		if requestedTreatment.DrugDBIds[erx.LexiGenProductId] == drugIds[erx.LexiGenProductId] &&
			requestedTreatment.DrugDBIds[erx.LexiDrugSynId] == drugIds[erx.LexiDrugSynId] &&
			requestedTreatment.DrugDBIds[erx.LexiSynonymTypeId] == drugIds[erx.LexiSynonymTypeId] &&
			requestedTreatment.DrugDBIds[erx.NDC] == drugIds[erx.NDC] {
			// linkage found
			requestedTreatment.OriginatingTreatmentId = treatmentId
			return nil
		}
	}

	return nil
}

func (d *DataService) CreateRefillRequest(refillRequest *common.RefillRequestItem) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err

	}

	if err := d.addRequestedTreatmentFromPharmacy(refillRequest.RequestedPrescription, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := d.addPharmacyDispensedTreatment(refillRequest.DispensedPrescription, refillRequest.RequestedPrescription, tx); err != nil {
		tx.Rollback()
		return err
	}

	columnsAndData := map[string]interface{}{
		"erx_request_queue_item_id":  refillRequest.RxRequestQueueItemId,
		"requested_drug_description": refillRequest.RequestedDrugDescription,
		"requested_refill_amount":    refillRequest.RequestedRefillAmount,
		"requested_dispense":         refillRequest.RequestedDispense,
		"patient_id":                 refillRequest.Patient.PatientId.Int64(),
		"request_date":               refillRequest.RequestDateStamp,
		"doctor_id":                  refillRequest.Doctor.DoctorId.Int64(),
		"dispensed_treatment_id":     refillRequest.DispensedPrescription.Id.Int64(),
		"requested_treatment_id":     refillRequest.RequestedPrescription.Id.Int64(),
	}

	if refillRequest.ReferenceNumber != "" {
		columnsAndData["reference_number"] = refillRequest.ReferenceNumber
	}

	if refillRequest.PharmacyRxReferenceNumber != "" {
		columnsAndData["pharmacy_rx_reference_number"] = refillRequest.PharmacyRxReferenceNumber
	}

	columns, dataForColumns := getKeysAndValuesFromMap(columnsAndData)

	lastId, err := tx.Exec(fmt.Sprintf(`insert into rx_refill_request (%s) values (%s)`,
		strings.Join(columns, ","), nReplacements(len(columns))), dataForColumns...)
	if err != nil {
		tx.Rollback()
		return err
	}

	refillRequest.Id, err = lastId.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) GetRefillRequestFromId(refillRequestId int64) (*common.RefillRequestItem, error) {
	var refillRequest common.RefillRequestItem
	var patientId, doctorId, pharmacyDispensedTreatmentId int64
	var requestedTreatmentId, approvedRefillAmount sql.NullInt64
	var refillStatus, notes, denyReason sql.NullString
	// get the refill request
	err := d.DB.QueryRow(`select rx_refill_request.id, rx_refill_request.erx_request_queue_item_id, requested_drug_description, requested_refill_amount,
		approved_refill_amount, requested_dispense, patient_id, request_date, doctor_id, requested_treatment_id, 
		dispensed_treatment_id, rx_refill_status_events.rx_refill_status, rx_refill_status_events.notes, deny_refill_reason.reason from rx_refill_request
			left outer join rx_refill_status_events on rx_refill_request.id =  rx_refill_request_id
			left outer join deny_refill_reason on reason_id = rx_refill_status_events.reason_id
				where rx_refill_request.id = ? and rx_refill_status_events.status = ?`, refillRequestId, STATUS_ACTIVE).Scan(&refillRequest.Id,
		&refillRequest.RxRequestQueueItemId, &refillRequest.RequestedDrugDescription, &refillRequest.RequestedRefillAmount, &approvedRefillAmount,
		&refillRequest.RequestedDispense, &patientId, &refillRequest.RequestDateStamp, &doctorId, &requestedTreatmentId,
		&pharmacyDispensedTreatmentId, &refillStatus, &notes, &denyReason)

	refillRequest.Status = refillStatus.String
	refillRequest.ApprovedRefillAmount = approvedRefillAmount.Int64
	refillRequest.Comments = notes.String
	refillRequest.DenialReason = denyReason.String

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// get the patient associated with the refill request
	refillRequest.Patient, err = d.GetPatientFromId(patientId)
	if err != nil {
		return nil, err
	}

	// get the doctor associated with the refill request
	refillRequest.Doctor, err = d.GetDoctorFromId(doctorId)
	if err != nil {
		return nil, err
	}

	// get the pharmacy dispensed treatment
	refillRequest.DispensedPrescription, err = d.getTreatmentForRefillRequest(table_name_pharmacy_dispensed_treatment, pharmacyDispensedTreatmentId)
	if err != nil {
		return nil, err
	}

	// get the unlinked requested treatment
	refillRequest.RequestedPrescription, err = d.getTreatmentForRefillRequest(table_name_requested_treatment, requestedTreatmentId.Int64)
	if err != nil {
		return nil, err
	}

	var originatingTreatmentId, originatingTreatmentPlanId sql.NullInt64
	err = d.DB.QueryRow(`select originating_treatment_id, treatment_plan_id from requested_treatment 
							inner join treatment on originating_treatment_id = treatment.id
								where requested_treatment.id = ?`, refillRequest.RequestedPrescription.Id.Int64()).Scan(&originatingTreatmentId, &originatingTreatmentPlanId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if originatingTreatmentId.Valid {
		refillRequest.RequestedPrescription.OriginatingTreatmentId = originatingTreatmentId.Int64
		refillRequest.TreatmentPlanId = originatingTreatmentPlanId.Int64
	}

	return &refillRequest, nil
}

func (d *DataService) getTreatmentForRefillRequest(tableName string, treatmentId int64) (*common.Treatment, error) {
	var treatment common.Treatment
	var erxId, pharmacyLocalId int64
	var doctorId sql.NullInt64
	var treatmentType string
	var drugName, drugForm, drugRoute sql.NullString

	err := d.DB.QueryRow(fmt.Sprintf(`select erx_id, drug_internal_name, 
							dosage_strength, type, dispense_value, 
							dispense_unit, refills, substitutions_allowed, 
							pharmacy_id, days_supply, pharmacy_notes, 
							patient_instructions, erx_sent_date,
							erx_last_filled_date,  status, drug_name.name, drug_route.name, drug_form.name, doctor_id from %s 
								left outer join drug_name on drug_name_id = drug_name.id
								left outer join drug_route on drug_route_id = drug_route.id
								left outer join drug_form on drug_form_id = drug_form.id
									where %s.id = ?`, tableName, tableName), treatmentId).Scan(&erxId, &treatment.DrugInternalName,
		&treatment.DosageStrength, &treatmentType, &treatment.DispenseValue,
		&treatment.DispenseUnitDescription, &treatment.NumberRefills,
		&treatment.SubstitutionsAllowed, &pharmacyLocalId,
		&treatment.DaysSupply, &treatment.PharmacyNotes,
		&treatment.PatientInstructions, &treatment.ErxSentDate,
		&treatment.ErxLastDateFilled, &treatment.Status,
		&drugName, &drugForm, &drugRoute, &doctorId)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	treatment.Id = common.NewObjectId(treatmentId)
	treatment.PrescriptionId = common.NewObjectId(erxId)
	treatment.DrugName = drugName.String
	treatment.DrugForm = drugForm.String
	treatment.DrugRoute = drugRoute.String
	treatment.OTC = treatmentType == treatment_otc
	treatment.PharmacyLocalId = common.NewObjectId(pharmacyLocalId)
	treatment.Pharmacy, err = d.GetPharmacyFromId(pharmacyLocalId)

	if err != nil {
		return nil, err
	}

	if doctorId.Valid {
		treatment.Doctor, err = d.GetDoctorFromId(doctorId.Int64)
		if err != nil {
			return nil, err
		}
	}

	return &treatment, nil
}

// this method is used to add treatments that come in from dosespot (either pharmacy dispensed medication or treatments that don't exist but
// are the basis of a refill request)
func (d *DataService) addRequestedTreatmentFromPharmacy(treatment *common.Treatment, tx *sql.Tx) error {
	substitutionsAllowedBit := 0
	if treatment.SubstitutionsAllowed {
		substitutionsAllowedBit = 1
	}

	treatmentType := treatment_rx
	if treatment.OTC {
		treatmentType = treatment_otc
	}

	columnsAndData := map[string]interface{}{
		"drug_internal_name":    treatment.DrugInternalName,
		"dosage_strength":       treatment.DosageStrength,
		"type":                  treatmentType,
		"dispense_value":        treatment.DispenseValue,
		"dispense_unit":         treatment.DispenseUnitDescription,
		"refills":               treatment.NumberRefills,
		"substitutions_allowed": substitutionsAllowedBit,
		"days_supply":           treatment.DaysSupply,
		"patient_instructions":  treatment.PatientInstructions,
		"pharmacy_notes":        treatment.PharmacyNotes,
		"status":                treatment.Status,
		"erx_id":                treatment.PrescriptionId.Int64(),
		"erx_sent_date":         treatment.ErxSentDate,
		"erx_last_filled_date":  treatment.ErxLastDateFilled,
		"pharmacy_id":           treatment.PharmacyLocalId,
		"doctor_id":             treatment.Doctor.DoctorId.Int64(),
	}

	if treatment.OriginatingTreatmentId != 0 {
		columnsAndData["originating_treatment_id"] = treatment.OriginatingTreatmentId
	}

	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugName, drug_name_table, "drug_name_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugForm, drug_form_table, "drug_form_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugRoute, drug_route_table, "drug_route_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	columns, dataForColumns := getKeysAndValuesFromMap(columnsAndData)
	res, err := tx.Exec(fmt.Sprintf(`insert into requested_treatment (%s) values (%s)`, strings.Join(columns, ","), nReplacements(len(dataForColumns))), dataForColumns...)
	if err != nil {
		tx.Rollback()
		return err
	}

	treatmentId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	treatment.Id = common.NewObjectId(treatmentId)
	// add drug db ids to the table
	for drugDbTag, drugDbId := range treatment.DrugDBIds {
		_, err := tx.Exec(`insert into requested_treatment_drug_db_id (drug_db_id_tag, drug_db_id, requested_treatment_id) values (?, ?, ?)`, drugDbTag, drugDbId, treatment.Id.Int64())
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func (d *DataService) addPharmacyDispensedTreatment(dispensedTreatment, requestedTreatment *common.Treatment, tx *sql.Tx) error {
	substitutionsAllowedBit := 0
	if dispensedTreatment.SubstitutionsAllowed {
		substitutionsAllowedBit = 1
	}

	treatmentType := treatment_rx
	if dispensedTreatment.OTC {
		treatmentType = treatment_otc
	}

	columnsAndData := map[string]interface{}{
		"drug_internal_name":     dispensedTreatment.DrugInternalName,
		"dosage_strength":        dispensedTreatment.DosageStrength,
		"type":                   treatmentType,
		"dispense_value":         dispensedTreatment.DispenseValue,
		"dispense_unit":          dispensedTreatment.DispenseUnitDescription,
		"refills":                dispensedTreatment.NumberRefills,
		"substitutions_allowed":  substitutionsAllowedBit,
		"days_supply":            dispensedTreatment.DaysSupply,
		"patient_instructions":   dispensedTreatment.PatientInstructions,
		"pharmacy_notes":         dispensedTreatment.PharmacyNotes,
		"status":                 dispensedTreatment.Status,
		"erx_id":                 dispensedTreatment.PrescriptionId.Int64(),
		"erx_sent_date":          dispensedTreatment.ErxSentDate,
		"erx_last_filled_date":   dispensedTreatment.ErxLastDateFilled,
		"pharmacy_id":            dispensedTreatment.PharmacyLocalId,
		"requested_treatment_id": requestedTreatment.Id.Int64(),
		"doctor_id":              dispensedTreatment.Doctor.DoctorId.Int64(),
	}

	if err := d.includeDrugNameComponentIfNonZero(dispensedTreatment.DrugName, drug_name_table, "drug_name_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.includeDrugNameComponentIfNonZero(dispensedTreatment.DrugForm, drug_form_table, "drug_form_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := d.includeDrugNameComponentIfNonZero(dispensedTreatment.DrugRoute, drug_route_table, "drug_route_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	columns, dataForColumns := getKeysAndValuesFromMap(columnsAndData)
	res, err := tx.Exec(fmt.Sprintf(`insert into pharmacy_dispensed_treatment (%s) values (%s)`, strings.Join(columns, ","), nReplacements(len(dataForColumns))), dataForColumns...)
	if err != nil {
		tx.Rollback()
		return err
	}

	treatmentId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	dispensedTreatment.Id = common.NewObjectId(treatmentId)
	// add drug db ids to the table
	for drugDbTag, drugDbId := range dispensedTreatment.DrugDBIds {
		_, err := tx.Exec(`insert into pharmacy_dispensed_treatment_drug_db_id (drug_db_id_tag, drug_db_id, pharmacy_dispensed_treatment_id) values (?, ?, ?)`, drugDbTag, drugDbId, dispensedTreatment.Id.Int64())
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func (d *DataService) GetRefillRequestDenialReasons() ([]*RefillRequestDenialReason, error) {
	rows, err := d.DB.Query(`select id, reason_code, reason from deny_refill_reason`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	denialReasons := make([]*RefillRequestDenialReason, 0)
	for rows.Next() {
		var denialReason RefillRequestDenialReason
		err = rows.Scan(&denialReason.Id, &denialReason.DenialCode, &denialReason.DenialReason)
		if err != nil {
			return nil, err
		}
		denialReasons = append(denialReasons, &denialReason)
	}

	return denialReasons, rows.Err()
}

func (d *DataService) MarkRefillRequestAsApproved(approvedRefillCount, rxRefillRequestId, prescriptionId int64, comments string) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update rx_refill_request set erx_id = ?, approved_refill_amount = ? where id = ?`, prescriptionId, approvedRefillCount, rxRefillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`update rx_refill_status_events set status = ? where rx_refill_request_id = ? and status = ?`, STATUS_INACTIVE, rxRefillRequestId, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`insert into rx_refill_status_events (rx_refill_request_id, rx_refill_status, status, notes, rx_refill_status_date) values (?,?,?,?, now())`, rxRefillRequestId, RX_REFILL_STATUS_APPROVED, STATUS_ACTIVE, comments)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) MarkRefillRequestAsDenied(denialReasonId, rxRefillRequestId, prescriptionId int64, comments string) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update rx_refill_request set erx_id = ? where id = ?`, prescriptionId, rxRefillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`update rx_refill_status_events set status = ? where rx_refill_request_id = ? and status = ?`, STATUS_INACTIVE, rxRefillRequestId, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`insert into rx_refill_status_events (rx_refill_request_id, rx_refill_status, reason_id,status,notes, rx_refill_status_date) values (?,?,?,?,?, now())`, rxRefillRequestId, RX_REFILL_STATUS_DENIED, denialReasonId, STATUS_ACTIVE, comments)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
