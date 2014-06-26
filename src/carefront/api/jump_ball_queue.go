package api

import (
	"carefront/common"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type JBCQItemClaimForbidden string

func (j JBCQItemClaimForbidden) Error() string {
	return string(j)
}

func (d *DataService) InsertUnclaimedItemIntoQueue(queueItem *DoctorQueueItem) error {
	_, err := d.db.Exec(`insert into unclaimed_item_queue (care_providing_state_id, item_id, patient_case_id, event_type, status) values (?,?,?,?,?)`, queueItem.CareProvidingStateId, queueItem.ItemId, queueItem.PatientCaseId, queueItem.EventType, queueItem.Status)
	return err
}

func (d *DataService) TemporarilyClaimCaseAndAssignDoctorToCaseAndPatient(doctorId, patientCaseId, patientId int64, duration time.Duration) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// mark the case as temporarily claimed
	_, err = tx.Exec(`update patient_case set status = ? where id = ?`, common.PCStatusTempClaimed, patientCaseId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// lock the visit in the unclaimed item queue
	_, err = tx.Exec(`update unclaimed_item_queue set locked = 1, doctor_id = ? where patient_case_id = ?`, doctorId, patientCaseId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// temporarily assign the doctor to the patient
	expiresTime := time.Now().Add(duration)

	var count int64
	if err := tx.QueryRow(`select count(*) from patient_care_provider_assignment where provider_id = ?  and role_type_id = ? and patient_id=?`, doctorId, d.roleTypeMapping[DOCTOR_ROLE], patientId).Scan(&count); err != nil {
		tx.Rollback()
		return err
	}

	if count == 0 {
		// give temp access for the doctor to the patient file only if the doctor does not already have access to the patient file
		_, err = tx.Exec(`insert into patient_care_provider_assignment (role_type_id, provider_id, patient_id, status, expires) values (?,?,?,?,?)`, d.roleTypeMapping[DOCTOR_ROLE], doctorId, patientId, STATUS_TEMP, expiresTime)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// temporarily assign the doctor to the patient_case
	_, err = tx.Exec(`replace into patient_case_care_provider_assignment (role_type_id, provider_id, patient_case_id, status, expires) values (?,?,?,?,?)`, d.roleTypeMapping[DOCTOR_ROLE], doctorId, patientCaseId, STATUS_TEMP, expiresTime)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) ExtendClaimForDoctor(doctorId, patientCaseId int64, duration time.Duration) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// ensure that the current doctor is the one holding on to the lock in the queue
	var currentLockHolder int64
	if err := tx.QueryRow(`select doctor_id from unclaimed_item_queue where patient_case_id = ? and locked = ?`, patientCaseId, true).Scan(&currentLockHolder); err == sql.ErrNoRows {
		tx.Rollback()
		return JBCQItemClaimForbidden("Doctor no longer listed as current claimer of case")
	} else if err != nil {
		tx.Rollback()
		return err
	}

	if currentLockHolder != doctorId {
		return JBCQItemClaimForbidden("Current lock holder is not the same as the doctor id provided")
	}

	// extend the claim of the doctor on the case and the patient file
	expires := time.Now().Add(duration)
	_, err = tx.Exec(`update patient_care_provider_assignment set expires = ? where provider_id = ? and role_type_id = ? and status = ?`, expires, doctorId, d.roleTypeMapping[DOCTOR_ROLE], STATUS_TEMP)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`update patient_case_care_provider_assignment set expires = ? where provider_id = ? and role_type_id = ? and status = ?`, expires, doctorId, d.roleTypeMapping[DOCTOR_ROLE], STATUS_TEMP)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) PermanentlyAssignDoctorToCaseAndPatient(doctorId, patientCaseId, patientId int64) error {
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
	_, err = tx.Exec(`delete from unclaimed_item_queue where patient_case_id = ? and doctor_id = ? and locked = ?`, patientCaseId, doctorId, true)
	if err != nil {
		tx.Rollback()
		return err
	}

	// permanently assign doctor to patient
	_, err = tx.Exec(`update patient_care_provider_assignment set status = ?, expires = NULL where provider_id = ? and role_type_id = ? and patient_id = ? and status = ?`, STATUS_ACTIVE, doctorId, d.roleTypeMapping[DOCTOR_ROLE], patientId, STATUS_TEMP)
	if err != nil {
		tx.Rollback()
		return err
	}

	// permanent assign doctor to case
	_, err = tx.Exec(`update patient_case_care_provider_assignment set status = ?, expires = NULL where provider_id = ? and role_type_id = ? and patient_case_id = ?`, STATUS_ACTIVE, doctorId, d.roleTypeMapping[DOCTOR_ROLE], patientCaseId, STATUS_TEMP)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) GetClaimedItemsInQueue() ([]*DoctorQueueItem, error) {
	rows, err := d.db.Query(`select id, event_type, item_id, patient_case_id, enqueue_date, status from unclaimed_item_queue where locked = ?`, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claimedItemsQueue := make([]*DoctorQueueItem, 0)
	for rows.Next() {
		var queueItem DoctorQueueItem
		if err := rows.Scan(&queueItem.Id,
			&queueItem.EventType,
			&queueItem.ItemId,
			&queueItem.PatientCaseId,
			&queueItem.EnqueueDate,
			&queueItem.Status); err != nil {
			return nil, err
		}
		claimedItemsQueue = append(claimedItemsQueue, &queueItem)
	}
	return claimedItemsQueue, rows.Err()
}

func (d *DataService) GetElligibleItemsInUnclaimedQueue(doctorId int64) ([]*DoctorQueueItem, error) {
	// first get the list of care providing state ids where the doctor is registered to serve
	rows, err := d.db.Query(`select care_providing_state_id from care_provider_state_elligibility where provider_id = ? and role_type_id = ?`, doctorId, d.roleTypeMapping[DOCTOR_ROLE])
	if err != nil {
		return nil, err
	}

	var careProvidingStateIds []int64
	for rows.Next() {
		var careProvidingStateId int64
		if err := rows.Scan(&careProvidingStateId); err != nil {
			return nil, err
		}
		careProvidingStateIds = append(careProvidingStateIds, careProvidingStateId)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	rows.Close()

	if len(careProvidingStateIds) == 0 {
		return nil, errors.New("Doctor is not elligible to provide care for any health condition in any state")
	}

	// then get the items in the unclaimed queue that are not currently locked by another doctor
	params := appendInt64sToInterfaceSlice(nil, careProvidingStateIds)
	params = append(params, []interface{}{false, true, doctorId}...)
	rows, err = d.db.Query(fmt.Sprintf(`select id, event_type, item_id, enqueue_date, status from unclaimed_item_queue where care_providing_state_id in (%s) and locked = ? or (locked = ? and doctor_id = ?) order by enqueue_date`, nReplacements(len(careProvidingStateIds))), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queueItems []*DoctorQueueItem
	for rows.Next() {
		var queueItem DoctorQueueItem
		if err := rows.Scan(
			&queueItem.Id,
			&queueItem.EventType,
			&queueItem.ItemId,
			&queueItem.EnqueueDate,
			&queueItem.Status); err != nil {
			return nil, err
		}
		queueItems = append(queueItems, &queueItem)
	}

	return queueItems, rows.Err()
}
