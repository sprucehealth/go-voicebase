package api

import (
	"carefront/common"
	"database/sql"
	"time"
)

type JBCQItemClaimForbidden string

func (j JBCQItemClaimForbidden) Error() string {
	return string(j)
}

func (d *DataService) temporarilyClaimCaseAndAssignDoctorToCaseAndPatient(doctorId, patientCaseId, patientId, itemId int64, eventType string, duration time.Duration) error {
	tx, err := d.db.Begin()
	if err != nil {
		return nil
	}

	// mark the case as temporarily claimed
	_, err = tx.Exec(`update patient_case set status = ? where id = ?`, common.PCStatusTempClaimed, patientCaseId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// lock the visit in the unclaimed item queue
	expiresTime := time.Now().Add(duration)
	_, err = tx.Exec(`update unclaimed_item_queue set locked = 1, expires = ?, doctor_id = ? where item_id = ? and event_type = ?`, expiresTime, doctorId, itemId, eventType)
	if err != nil {
		tx.Rollback()
		return err
	}

	// assign the doctor to the patient
	_, err = tx.Exec(`insert into patient_care_provider_assignment (role_type_id, provider_id, patient_id, status) values (?,?,?,?)`, d.roleTypeMapping[DOCTOR_ROLE], doctorId, patientId, STATUS_TEMP)
	if err != nil {
		tx.Rollback()
		return err
	}

	// assign the doctor to the patient_case
	_, err = tx.Exec(`insert into patient_case_care_provider_assignment (role_type_id, provider_id, patient_case_id, status) values (?,?,?,?)`, d.roleTypeMapping[DOCTOR_ROLE], doctorId, patientCaseId, STATUS_TEMP)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) extendClaimForDoctor(doctorId, itemId int64, eventType string, duration time.Duration) error {
	// ensure that the current doctor is the one holding on to the lock in the queue
	var currentLockHolder int64
	if err := d.db.QueryRow(`select doctor_id from unclaimed_item_queue where item_id = ? and event_type = ? and locked = ?`, itemId, eventType, true).Scan(&currentLockHolder); err == sql.ErrNoRows {
		return JBCQItemClaimForbidden("Doctor no longer listed as current claimer of case")
	} else if err != nil {
		return err
	}

	if currentLockHolder != doctorId {
		return JBCQItemClaimForbidden("Current lock holder is not the same as the doctor id provided")
	}

	// extend the claim of the doctor on the case
	expires := time.Now().Add(duration)
	_, err := d.db.Exec(`update unclaimed_item_queue set expires = ? where doctor_id = ? and item_id = ? and event_type = ? and locked = ?`, expires, doctorId, itemId, eventType, true)

	return err
}

func (d *DataService) permanentlyAssignDoctorToCaseAndPatient(doctorId, patientCaseId, patientId, itemId int64, eventType string) error {
	tx, err := d.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	// first check to ensure that the doctor is currently temporarily assigned to patient case and file
	var currentDoctorOnFile int64
	if err := tx.QueryRow(`select provider_id from patient_care_provider_assignment where role_type_id = ? and provider_id = ? and patient_id = ? and status = ?`, d.roleTypeMapping[DOCTOR_ROLE], doctorId, patientId, STATUS_TEMP).Scan(&currentDoctorOnFile); err == sql.ErrNoRows {
		return JBCQItemClaimForbidden("Expected doctor to be temporarily assigned to patient file but wasnt")
	} else if err != nil {
		return err
	}

	var currentDoctorOnCase int64
	if err := tx.QueryRow(`select provider_id from patient_case_care_provider_assignment where role_type_id = ? and provider_id = ? and patient_case_id = ? and status = ?`, d.roleTypeMapping[DOCTOR_ROLE], doctorId, patientCaseId, STATUS_TEMP).Scan(&currentDoctorOnCase); err == sql.ErrNoRows {
		return JBCQItemClaimForbidden("Expected doctor to be temporarily assigned to patient case but wasnt")
	} else if err != nil {
		return err
	}

	// delete item from unclaimed queue
	_, err = tx.Exec(`delete from unclaimed_item_queue where item_id = ? and event_type = ? and doctor_id = ? and locked = ?`, itemId, eventType, doctorId, true)
	if err != nil {
		tx.Rollback()
		return err
	}

	// permanently assign doctor to patient
	_, err = tx.Exec(`update patient_care_provider_assignment set status = ? where provider_id = ? and role_type_id = ? and patient_id = ?`, STATUS_ACTIVE, doctorId, d.roleTypeMapping[DOCTOR_ROLE], patientId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// permanent assign doctor to case
	_, err = tx.Exec(`update patient_case_care_provider_assignment set status = ? where provider_id = ? and role_type_id = ? and patient_case_id = ?`, STATUS_ACTIVE, doctorId, d.roleTypeMapping[DOCTOR_ROLE], patientCaseId)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
