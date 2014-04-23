package api

import (
	"carefront/common"
	"carefront/encoding"
	"carefront/libs/erx"
	"carefront/libs/pharmacy"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

func (d *DataService) AddRefillRequestStatusEvent(refillRequestStatus common.StatusEvent) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update rx_refill_status_events set status = ? where status = ? and rx_refill_request_id = ?`, STATUS_INACTIVE, STATUS_ACTIVE, refillRequestStatus.ItemId)
	if err != nil {
		tx.Rollback()
		return err
	}

	columnsAndData := map[string]interface{}{
		"rx_refill_request_id": refillRequestStatus.ItemId,
		"rx_refill_status":     refillRequestStatus.Status,
		"status":               STATUS_ACTIVE,
		"event_details":        refillRequestStatus.StatusDetails,
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
	rows, err := d.DB.Query(`select rx_refill_request_id, rx_refill_status, rx_refill_status_date, 
								event_details, erx_id  
								from rx_refill_status_events 
									inner join rx_refill_request on rx_refill_request_id = rx_refill_request.id
									where rx_refill_status_events.status = ? and rx_refill_status = ?`, STATUS_ACTIVE, RX_REFILL_STATUS_REQUESTED)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getRefillStatusEventsFromRows(rows)
}

func (d *DataService) GetApprovedOrDeniedRefillRequestsForPatient(patientId int64) ([]common.StatusEvent, error) {
	rows, err := d.DB.Query(`select rx_refill_request_id, rx_refill_status, rx_refill_status_date, 
									event_details, erx_id    
									from rx_refill_status_events 
										inner join rx_refill_request on rx_refill_request_id = rx_refill_request.id
										where rx_refill_status_events.rx_refill_status in (?, ?) and rx_refill_request.patient_id = ?
										and status = ?
											order by rx_refill_status_date desc`, RX_REFILL_STATUS_APPROVED, RX_REFILL_STATUS_DENIED, patientId, STATUS_ACTIVE)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getRefillStatusEventsFromRows(rows)
}

func (d *DataService) GetRefillStatusEventsForRefillRequest(refillRequestId int64) ([]common.StatusEvent, error) {
	rows, err := d.DB.Query(`select rx_refill_request_id,rx_refill_status, rx_refill_status_date, 
									event_details, erx_id    
									from rx_refill_status_events 
										inner join rx_refill_request on rx_refill_request_id = rx_refill_request.id
										where rx_refill_status_events.rx_refill_request_id = ?
										order by rx_refill_status_date desc`, refillRequestId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getRefillStatusEventsFromRows(rows)
}

func getRefillStatusEventsFromRows(rows *sql.Rows) ([]common.StatusEvent, error) {
	refillRequestStatuses := make([]common.StatusEvent, 0)
	for rows.Next() {
		var refillRequestStatus common.StatusEvent
		var prescriptionId sql.NullInt64
		err := rows.Scan(&refillRequestStatus.ItemId, &refillRequestStatus.Status,
			&refillRequestStatus.StatusTimestamp, &refillRequestStatus.StatusDetails, &prescriptionId)
		if err != nil {
			return nil, err
		}
		refillRequestStatus.PrescriptionId = prescriptionId.Int64
		refillRequestStatuses = append(refillRequestStatuses, refillRequestStatus)
	}
	return refillRequestStatuses, rows.Err()
}

func (d *DataService) LinkRequestedPrescriptionToOriginalTreatment(requestedTreatment *common.Treatment, patient *common.Patient) error {
	// lookup drug based on the drugIds
	if len(requestedTreatment.DrugDBIds) == 0 {
		// nothing to compare against to link to originating drug
		return nil
	}

	// lookup drugs prescribed to the patient within a day of the date the requestedPrescription was prescribed
	// we know that it was prescribed based on whether or not it was succesfully sent to the pharmacy
	halfDayBefore := requestedTreatment.ERx.ErxSentDate.Add(-12 * time.Hour)
	halfDayAfter := requestedTreatment.ERx.ErxSentDate.Add(12 * time.Hour)

	treatmentIds := make([]int64, 0)
	rows, err := d.DB.Query(`select treatment_id from erx_status_events 
								inner join treatment on treatment_id = treatment.id 
								inner join treatment_plan on treatment_plan_id = treatment.treatment_plan_id
								inner join patient_visit on patient_visit.id = treatment_plan.patient_visit_id
								where erx_status = ? and erx_status_events.creation_date >= ? 
								and erx_status_events.creation_date <= ? and patient_visit.patient_id = ? `, ERX_STATUS_SENT, halfDayBefore, halfDayAfter, patient.PatientId.Int64())
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
			requestedTreatment.DrugDBIds[erx.LexiSynonymTypeId] == drugIds[erx.LexiSynonymTypeId] {
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
		"erx_request_queue_item_id": refillRequest.RxRequestQueueItemId,
		"patient_id":                refillRequest.Patient.PatientId.Int64(),
		"request_date":              refillRequest.RequestDateStamp,
		"doctor_id":                 refillRequest.Doctor.DoctorId.Int64(),
		"dispensed_treatment_id":    refillRequest.DispensedPrescription.Id.Int64(),
		"requested_treatment_id":    refillRequest.RequestedPrescription.Id.Int64(),
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
	// get the refill request
	rows, err := d.DB.Query(`select rx_refill_request.id, rx_refill_request.erx_request_queue_item_id,rx_refill_request.reference_number, rx_refill_request.erx_id,
		approved_refill_amount, patient_id, request_date, doctor_id, requested_treatment_id, 
		dispensed_treatment_id, comments, deny_refill_reason.reason from rx_refill_request
				left outer join deny_refill_reason on deny_refill_reason.id = denial_reason_id
				where rx_refill_request.id = ?`, refillRequestId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refillRequests, err := d.getRefillRequestsFromRow(rows)
	if err != nil {
		return nil, err
	}

	if len(refillRequests) > 1 {
		return nil, fmt.Errorf("Expected just one refill request instead got %d", len(refillRequests))
	}

	if len(refillRequests) == 0 {
		return nil, nil
	}

	return refillRequests[0], nil
}

func (d *DataService) GetRefillRequestsForPatient(patientId int64) ([]*common.RefillRequestItem, error) {
	// get the refill request
	rows, err := d.DB.Query(`select rx_refill_request.id, rx_refill_request.erx_request_queue_item_id,rx_refill_request.reference_number, rx_refill_request.erx_id,
		approved_refill_amount, patient_id, request_date, doctor_id, requested_treatment_id, 
		dispensed_treatment_id, comments, deny_refill_reason.reason from rx_refill_request
				left outer join deny_refill_reason on deny_refill_reason.id = denial_reason_id
				where patient_id = ? order by rx_refill_request.request_date desc`, patientId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refillRequests, err := d.getRefillRequestsFromRow(rows)

	return refillRequests, err
}

func (d *DataService) getRefillRequestsFromRow(rows *sql.Rows) ([]*common.RefillRequestItem, error) {

	refillRequests := make([]*common.RefillRequestItem, 0)

	for rows.Next() {
		var refillRequest common.RefillRequestItem
		var patientId, doctorId, pharmacyDispensedTreatmentId int64
		var requestedTreatmentId, approvedRefillAmount, prescriptionId sql.NullInt64
		var denyReason, comments sql.NullString

		err := rows.Scan(&refillRequest.Id,
			&refillRequest.RxRequestQueueItemId, &refillRequest.ReferenceNumber, &prescriptionId, &approvedRefillAmount,
			&patientId, &refillRequest.RequestDateStamp, &doctorId, &requestedTreatmentId,
			&pharmacyDispensedTreatmentId, &comments, &denyReason)
		if err != nil {
			return nil, err
		}

		refillRequest.PrescriptionId = prescriptionId.Int64
		refillRequest.ApprovedRefillAmount = approvedRefillAmount.Int64
		refillRequest.DenialReason = denyReason.String
		refillRequest.Comments = comments.String

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
		refillRequest.DispensedPrescription, err = d.getTreatmentForRefillRequest(pharmacyDispensedTreatmentTable, pharmacyDispensedTreatmentId)
		if err != nil {
			return nil, err
		}

		// get the unlinked requested treatment
		refillRequest.RequestedPrescription, err = d.getTreatmentForRefillRequest(requestedTreatmentTable, requestedTreatmentId.Int64)
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
			refillRequest.TreatmentPlanId = encoding.NewObjectId(originatingTreatmentPlanId.Int64)
		}

		refillRequest.RxHistory, err = d.GetRefillStatusEventsForRefillRequest(refillRequest.Id)
		if err != nil {
			return nil, err
		}

		refillRequests = append(refillRequests, &refillRequest)
	}

	return refillRequests, rows.Err()
}

func (d *DataService) getTreatmentForRefillRequest(tableName string, treatmentId int64) (*common.Treatment, error) {
	var treatment common.Treatment
	treatment.ERx = &common.ERxData{}
	var erxId, pharmacyLocalId int64
	var daysSupply, refills encoding.NullInt64
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
		&treatment.DispenseUnitDescription, &refills,
		&treatment.SubstitutionsAllowed, &pharmacyLocalId,
		&daysSupply, &treatment.PharmacyNotes,
		&treatment.PatientInstructions, &treatment.ERx.ErxSentDate,
		&treatment.ERx.ErxLastDateFilled, &treatment.Status,
		&drugName, &drugForm, &drugRoute, &doctorId)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	treatment.Id = encoding.NewObjectId(treatmentId)
	treatment.ERx.PrescriptionId = encoding.NewObjectId(erxId)
	treatment.DrugName = drugName.String
	treatment.DrugForm = drugForm.String
	treatment.DrugRoute = drugRoute.String
	treatment.OTC = treatmentType == treatmentOTC
	treatment.DaysSupply = daysSupply
	treatment.NumberRefills = refills
	treatment.ERx.PharmacyLocalId = encoding.NewObjectId(pharmacyLocalId)
	treatment.ERx.Pharmacy, err = d.GetPharmacyFromId(pharmacyLocalId)

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

	treatmentType := treatmentRX
	if treatment.OTC {
		treatmentType = treatmentOTC
	}

	columnsAndData := map[string]interface{}{
		"drug_internal_name":    treatment.DrugInternalName,
		"dosage_strength":       treatment.DosageStrength,
		"type":                  treatmentType,
		"dispense_value":        treatment.DispenseValue,
		"dispense_unit":         treatment.DispenseUnitDescription,
		"refills":               treatment.NumberRefills.Int64Value,
		"substitutions_allowed": treatment.SubstitutionsAllowed,
		"patient_instructions":  treatment.PatientInstructions,
		"pharmacy_notes":        treatment.PharmacyNotes,
		"status":                treatment.Status,
		"erx_id":                treatment.ERx.PrescriptionId.Int64(),
		"erx_sent_date":         treatment.ERx.ErxSentDate,
		"erx_last_filled_date":  treatment.ERx.ErxLastDateFilled,
		"pharmacy_id":           treatment.ERx.PharmacyLocalId,
		"doctor_id":             treatment.Doctor.DoctorId.Int64(),
	}

	if treatment.DaysSupply.IsValid {
		columnsAndData["days_supply"] = treatment.DaysSupply.Int64Value
	}

	if treatment.OriginatingTreatmentId != 0 {
		columnsAndData["originating_treatment_id"] = treatment.OriginatingTreatmentId
	}

	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugName, drugNameTable, "drug_name_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugForm, drugFormTable, "drug_form_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugRoute, drugRouteTable, "drug_route_id", columnsAndData, tx); err != nil {
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

	treatment.Id = encoding.NewObjectId(treatmentId)
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

	treatmentType := treatmentRX
	if dispensedTreatment.OTC {
		treatmentType = treatmentOTC
	}

	columnsAndData := map[string]interface{}{
		"drug_internal_name":     dispensedTreatment.DrugInternalName,
		"dosage_strength":        dispensedTreatment.DosageStrength,
		"type":                   treatmentType,
		"dispense_value":         dispensedTreatment.DispenseValue,
		"dispense_unit":          dispensedTreatment.DispenseUnitDescription,
		"refills":                dispensedTreatment.NumberRefills.Int64Value,
		"substitutions_allowed":  dispensedTreatment.SubstitutionsAllowed,
		"patient_instructions":   dispensedTreatment.PatientInstructions,
		"pharmacy_notes":         dispensedTreatment.PharmacyNotes,
		"status":                 dispensedTreatment.Status,
		"erx_id":                 dispensedTreatment.ERx.PrescriptionId.Int64(),
		"erx_sent_date":          dispensedTreatment.ERx.ErxSentDate,
		"erx_last_filled_date":   dispensedTreatment.ERx.ErxLastDateFilled,
		"pharmacy_id":            dispensedTreatment.ERx.PharmacyLocalId,
		"requested_treatment_id": requestedTreatment.Id.Int64(),
		"doctor_id":              dispensedTreatment.Doctor.DoctorId.Int64(),
	}

	if dispensedTreatment.DaysSupply.IsValid {
		columnsAndData["days_supply"] = dispensedTreatment.DaysSupply.Int64Value
	}

	if err := d.includeDrugNameComponentIfNonZero(dispensedTreatment.DrugName, drugNameTable, "drug_name_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.includeDrugNameComponentIfNonZero(dispensedTreatment.DrugForm, drugFormTable, "drug_form_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := d.includeDrugNameComponentIfNonZero(dispensedTreatment.DrugRoute, drugRouteTable, "drug_route_id", columnsAndData, tx); err != nil {
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

	dispensedTreatment.Id = encoding.NewObjectId(treatmentId)
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

func (d *DataService) includeDrugNameComponentIfNonZero(drugNameComponent, tableName, columnName string, columnsAndData map[string]interface{}, tx *sql.Tx) error {
	if drugNameComponent != "" {
		componentId, err := d.getOrInsertNameInTable(tx, tableName, drugNameComponent)
		if err != nil {
			return err
		}
		columnsAndData[columnName] = componentId
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

func (d *DataService) MarkRefillRequestAsApproved(prescriptionId, approvedRefillCount, rxRefillRequestId int64, comments string) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update rx_refill_request set erx_id = ?, approved_refill_amount = ?, comments = ? where id = ?`, prescriptionId, approvedRefillCount, comments, rxRefillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`update rx_refill_status_events set status = ? where rx_refill_request_id = ? and status = ?`, STATUS_INACTIVE, rxRefillRequestId, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`insert into rx_refill_status_events (rx_refill_request_id, rx_refill_status, status) values (?,?,?)`, rxRefillRequestId, RX_REFILL_STATUS_APPROVED, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) MarkRefillRequestAsDenied(prescriptionId, denialReasonId, rxRefillRequestId int64, comments string) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update rx_refill_request set erx_id = ?, comments = ?, denial_reason_id = ? where id = ?`, prescriptionId, comments, denialReasonId, rxRefillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`update rx_refill_status_events set status = ? where rx_refill_request_id = ? and status = ?`, STATUS_INACTIVE, rxRefillRequestId, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`insert into rx_refill_status_events (rx_refill_request_id, rx_refill_status, status) values (?,?,?)`, rxRefillRequestId, RX_REFILL_STATUS_DENIED, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) UpdateUnlinkedDNTFTreatmentWithPharmacyAndErxId(treatment *common.Treatment, pharmacySentTo *pharmacy.PharmacyData, doctorId int64) error {
	if treatment.ERx.PrescriptionId.Int64() != 0 {
		_, err := d.DB.Exec(`update unlinked_dntf_treatment set erx_id = ?, pharmacy_id = ?, erx_sent_date=now() where id = ?`, treatment.ERx.PrescriptionId, pharmacySentTo.LocalId, treatment.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DataService) AddUnlinkedTreatmentInEventOfDNTF(treatment *common.Treatment, refillRequestId int64) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	treatmentType := treatmentRX
	if treatment.OTC {
		treatmentType = treatmentOTC
	}

	columnsAndData := map[string]interface{}{
		"drug_internal_name":    treatment.DrugInternalName,
		"dosage_strength":       treatment.DosageStrength,
		"type":                  treatmentType,
		"dispense_value":        treatment.DispenseValue,
		"dispense_unit_id":      treatment.DispenseUnitId.Int64(),
		"refills":               treatment.NumberRefills.Int64Value,
		"substitutions_allowed": treatment.SubstitutionsAllowed,
		"patient_instructions":  treatment.PatientInstructions,
		"pharmacy_notes":        treatment.PharmacyNotes,
		"status":                treatment.Status,
		"doctor_id":             treatment.DoctorId.Int64(),
		"patient_id":            treatment.PatientId.Int64(),
	}

	if treatment.DaysSupply.IsValid {
		columnsAndData["days_supply"] = treatment.DaysSupply.Int64Value
	}

	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugName, drugNameTable, "drug_name_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugForm, drugFormTable, "drug_form_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := d.includeDrugNameComponentIfNonZero(treatment.DrugRoute, drugRouteTable, "drug_route_id", columnsAndData, tx); err != nil {
		tx.Rollback()
		return err
	}

	columns, dataForColumns := getKeysAndValuesFromMap(columnsAndData)
	res, err := tx.Exec(fmt.Sprintf(`insert into unlinked_dntf_treatment (%s) values (%s)`, strings.Join(columns, ","), nReplacements(len(dataForColumns))), dataForColumns...)
	if err != nil {
		tx.Rollback()
		return err
	}

	treatmentId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	treatment.Id = encoding.NewObjectId(treatmentId)

	_, err = tx.Exec(`insert into unlinked_dntf_treatment_status_events (unlinked_dntf_treatment_id, erx_status, status) values (?,?,?)`, treatmentId, ERX_STATUS_NEW_RX_FROM_DNTF, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`insert into dntf_mapping (unlinked_dntf_treatment_id, rx_refill_request_id) values (?,?)`, treatmentId, refillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) GetUnlinkedDNTFTreatment(treatmentId int64) (*common.Treatment, error) {
	rows, err := d.DB.Query(`select unlinked_dntf_treatment.id, unlinked_dntf_treatment.erx_id, unlinked_dntf_treatment.drug_internal_name, unlinked_dntf_treatment.dosage_strength, unlinked_dntf_treatment.type,
			unlinked_dntf_treatment.dispense_value, unlinked_dntf_treatment.dispense_unit_id, ltext, unlinked_dntf_treatment.refills, unlinked_dntf_treatment.substitutions_allowed, 
			unlinked_dntf_treatment.days_supply, unlinked_dntf_treatment.pharmacy_id, unlinked_dntf_treatment.pharmacy_notes, unlinked_dntf_treatment.patient_instructions, unlinked_dntf_treatment.creation_date, unlinked_dntf_treatment.erx_sent_date,
			unlinked_dntf_treatment.erx_last_filled_date, unlinked_dntf_treatment.status, drug_name.name, drug_route.name, drug_form.name,
			patient_id, unlinked_dntf_treatment.doctor_id from unlinked_dntf_treatment 
				inner join dispense_unit on unlinked_dntf_treatment.dispense_unit_id = dispense_unit.id
				inner join localized_text on localized_text.app_text_id = dispense_unit.dispense_unit_text_id
				left outer join drug_name on drug_name_id = drug_name.id
				left outer join drug_route on drug_route_id = drug_route.id
				left outer join drug_form on drug_form_id = drug_form.id
				where unlinked_dntf_treatment.id = ? and localized_text.language_id = ?`, treatmentId, EN_LANGUAGE_ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	treatments, err := d.getUnlinkedDNTFTreatmentsFromRow(rows)
	if err != nil {
		return nil, err
	}

	if len(treatments) > 1 {
		return nil, fmt.Errorf("Expected just one unlinked dntf treatment but got back %d", len(treatments))
	}

	if len(treatments) == 0 {
		return nil, nil
	}

	return treatments[0], err
}

func (d *DataService) GetUnlinkedDNTFTreatmentsForPatient(patientId int64) ([]*common.Treatment, error) {
	rows, err := d.DB.Query(`select unlinked_dntf_treatment.id, unlinked_dntf_treatment.erx_id, unlinked_dntf_treatment.drug_internal_name, unlinked_dntf_treatment.dosage_strength, unlinked_dntf_treatment.type,
			unlinked_dntf_treatment.dispense_value, unlinked_dntf_treatment.dispense_unit_id, ltext, unlinked_dntf_treatment.refills, unlinked_dntf_treatment.substitutions_allowed, 
			unlinked_dntf_treatment.days_supply, unlinked_dntf_treatment.pharmacy_id, unlinked_dntf_treatment.pharmacy_notes, unlinked_dntf_treatment.patient_instructions, unlinked_dntf_treatment.creation_date, unlinked_dntf_treatment.erx_sent_date,
			unlinked_dntf_treatment.erx_last_filled_date, unlinked_dntf_treatment.status, drug_name.name, drug_route.name, drug_form.name,
			patient_id, unlinked_dntf_treatment.doctor_id from unlinked_dntf_treatment 
				inner join dispense_unit on unlinked_dntf_treatment.dispense_unit_id = dispense_unit.id
				inner join localized_text on localized_text.app_text_id = dispense_unit.dispense_unit_text_id
				left outer join drug_name on drug_name_id = drug_name.id
				left outer join drug_route on drug_route_id = drug_route.id
				left outer join drug_form on drug_form_id = drug_form.id
				where patient_id = ? and localized_text.language_id = ? order by unlinked_dntf_treatment.creation_date desc`, patientId, EN_LANGUAGE_ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	treatments, err := d.getUnlinkedDNTFTreatmentsFromRow(rows)
	if err != nil {
		return nil, err
	}

	return treatments, err
}

func (d *DataService) getUnlinkedDNTFTreatmentsFromRow(rows *sql.Rows) ([]*common.Treatment, error) {
	treatments := make([]*common.Treatment, 0)
	for rows.Next() {
		var dispenseUnitId, doctorId, patientId, unlinkedDntfTreatmentId int64
		var dispenseValue encoding.HighPrecisionFloat64
		var drugInternalName, dosageStrength, treatmentType, dispenseUnitDescription, pharmacyNotes, patientInstructions, status string
		var creationDate time.Time
		var erxId, pharmacyId sql.NullInt64
		var daysSupply, refills encoding.NullInt64
		var erxSentDate, erxLastFilledDate mysql.NullTime
		var drugName, drugRoute, drugForm sql.NullString
		var substitutionsAllowed bool
		err := rows.Scan(&unlinkedDntfTreatmentId, &erxId, &drugInternalName, &dosageStrength, &treatmentType, &dispenseValue, &dispenseUnitId, &dispenseUnitDescription,
			&refills, &substitutionsAllowed, &daysSupply, &pharmacyId, &pharmacyNotes, &patientInstructions, &creationDate, &erxSentDate, &erxLastFilledDate, &status, &drugName, &drugRoute, &drugForm, &patientId, &doctorId)
		if err != nil {
			return nil, err
		}

		treatment := &common.Treatment{
			Id:                      encoding.NewObjectId(unlinkedDntfTreatmentId),
			PatientId:               encoding.NewObjectId(patientId),
			DoctorId:                encoding.NewObjectId(doctorId),
			DrugInternalName:        drugInternalName,
			DosageStrength:          dosageStrength,
			DispenseValue:           dispenseValue,
			DispenseUnitId:          encoding.NewObjectId(dispenseUnitId),
			DispenseUnitDescription: dispenseUnitDescription,
			NumberRefills:           refills,
			SubstitutionsAllowed:    substitutionsAllowed,
			DaysSupply:              daysSupply,
			DrugName:                drugName.String,
			DrugForm:                drugForm.String,
			DrugRoute:               drugRoute.String,
			PatientInstructions:     patientInstructions,
			CreationDate:            &creationDate,
			Status:                  status,
			PharmacyNotes:           pharmacyNotes,
			OTC:                     treatmentType == treatmentOTC,
			ERx: &common.ERxData{
				ErxLastDateFilled: &erxLastFilledDate.Time,
				ErxSentDate:       &erxSentDate.Time,
			},
		}

		if pharmacyId.Valid {
			treatment.ERx.PharmacyLocalId = encoding.NewObjectId(pharmacyId.Int64)
			treatment.ERx.Pharmacy, err = d.GetPharmacyFromId(pharmacyId.Int64)
			if err != nil {
				return nil, err
			}
		}

		if erxId.Valid {
			treatment.ERx.PrescriptionId = encoding.NewObjectId(erxId.Int64)
		}

		treatment.Doctor, err = d.GetDoctorFromId(treatment.DoctorId.Int64())
		if err != nil {
			return nil, err
		}

		treatment.Patient, err = d.GetPatientFromId(treatment.PatientId.Int64())
		if err != nil {
			return nil, err
		}

		treatment.ERx.RxHistory, err = d.GetErxStatusEventsForDNTFTreatment(unlinkedDntfTreatmentId)
		if err != nil {
			return nil, err
		}
		treatments = append(treatments, treatment)

	}
	return treatments, rows.Err()
}

func (d *DataService) AddTreatmentToTreatmentPlanInEventOfDNTF(treatment *common.Treatment, refillRequestId int64) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	if err := d.addTreatment(treatment, asPatientTreatment, tx); err != nil {
		tx.Rollback()
		return err
	}

	if treatment.DoctorTreatmentTemplateId.Int64() != 0 {
		_, err = tx.Exec(`insert into treatment_dr_template_selection (treatment_id, dr_treatment_template_id) values (?,?)`, treatment.Id, treatment.DoctorTreatmentTemplateId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_, err = tx.Exec(`insert into erx_status_events (treatment_id, erx_status, status) values (?,?,?)`, treatment.Id.Int64(), ERX_STATUS_NEW_RX_FROM_DNTF, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`insert into dntf_mapping (treatment_id, rx_refill_request_id) values (?,?)`, treatment.Id.Int64(), refillRequestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) AddErxStatusEventForDNTFTreatment(statusEvent common.StatusEvent) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update unlinked_dntf_treatment_status_events set status = ? where unlinked_dntf_treatment_id = ? and status = ?`, STATUS_INACTIVE, statusEvent.ItemId, STATUS_ACTIVE)
	if err != nil {
		tx.Rollback()
		return err
	}

	columnsAndData := map[string]interface{}{
		"unlinked_dntf_treatment_id": statusEvent.ItemId,
		"erx_status":                 statusEvent.Status,
		"status":                     STATUS_ACTIVE,
	}

	if statusEvent.StatusDetails != "" {
		columnsAndData["event_details"] = statusEvent.StatusDetails
	}

	if !statusEvent.ReportedTimestamp.IsZero() {
		columnsAndData["reported_timestamp"] = statusEvent.ReportedTimestamp
	}

	columns, values := getKeysAndValuesFromMap(columnsAndData)

	_, err = tx.Exec(fmt.Sprintf(`insert into unlinked_dntf_treatment_status_events (%s) values (%s)`, strings.Join(columns, ","), nReplacements(len(values))), values...)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) GetErxStatusEventsForDNTFTreatment(treatmentId int64) ([]common.StatusEvent, error) {
	rows, err := d.DB.Query(`select unlinked_dntf_treatment_status_events.unlinked_dntf_treatment_id, unlinked_dntf_treatment.erx_id, unlinked_dntf_treatment_status_events.erx_status, unlinked_dntf_treatment_status_events.event_details, unlinked_dntf_treatment_status_events.status, unlinked_dntf_treatment_status_events.creation_date from unlinked_dntf_treatment_status_events 
								inner join unlinked_dntf_treatment on unlinked_dntf_treatment_id = unlinked_dntf_treatment.id
									where unlinked_dntf_treatment.id = ? order by unlinked_dntf_treatment_status_events.creation_date desc`, treatmentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusEvents := make([]common.StatusEvent, 0)
	for rows.Next() {
		var statusDetails sql.NullString
		var statusEventItem common.StatusEvent
		if err := rows.Scan(&statusEventItem.ItemId, &statusEventItem.PrescriptionId, &statusEventItem.Status, &statusDetails, &statusEventItem.InternalStatus, &statusEventItem.StatusTimestamp); err != nil {
			return nil, err
		}
		statusEventItem.StatusDetails = statusDetails.String
		statusEvents = append(statusEvents, statusEventItem)
	}
	return statusEvents, rows.Err()
}

func (d *DataService) GetErxStatusEventsForDNTFTreatmentBasedOnPatientId(patientId int64) ([]common.StatusEvent, error) {
	rows, err := d.DB.Query(`select unlinked_dntf_treatment_status_events.unlinked_dntf_treatment_id, unlinked_dntf_treatment.erx_id, unlinked_dntf_treatment_status_events.erx_status,unlinked_dntf_treatment_status_events.status, unlinked_dntf_treatment_status_events.creation_date from unlinked_dntf_treatment_status_events 
								inner join unlinked_dntf_treatment on unlinked_dntf_treatment_id = unlinked_dntf_treatment.id
									where unlinked_dntf_treatment.patient_id = ? order by unlinked_dntf_treatment_status_events.creation_date desc`, patientId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusEvents := make([]common.StatusEvent, 0)
	for rows.Next() {
		var statusEventItem common.StatusEvent
		if err := rows.Scan(&statusEventItem.ItemId, &statusEventItem.PrescriptionId, &statusEventItem.Status, &statusEventItem.InternalStatus, &statusEventItem.StatusTimestamp); err != nil {
			return nil, err
		}
		statusEvents = append(statusEvents, statusEventItem)
	}

	return statusEvents, rows.Err()
}
