package api

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/dbutil"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/transactional/tsql"
)

func (d *dataService) RegisterProvider(provider *common.Doctor, role string) (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}

	switch role {
	case RoleCC:
		if provider.ShortDisplayName == "" {
			provider.ShortDisplayName = provider.FirstName + " " + provider.LastName
		}
		if provider.LongDisplayName == "" {
			provider.LongDisplayName = provider.FirstName + " " + provider.LastName
		}
		if provider.ShortTitle == "" {
			provider.ShortTitle = "Care Provider"
		}
		if provider.LongTitle == "" {
			provider.LongTitle = "Care Provider"
		}
	case RoleDoctor:
		if provider.ShortDisplayName == "" {
			provider.ShortDisplayName = "Dr. " + provider.LastName
		}
		if provider.LongDisplayName == "" {
			provider.LongDisplayName = "Dr. " + provider.FirstName + " " + provider.LastName
		}
		if provider.ShortTitle == "" {
			provider.ShortTitle = "Doctor"
		}
		if provider.LongTitle == "" {
			provider.LongTitle = "Doctor"
		}
	}

	res, err := tx.Exec(`
		INSERT INTO doctor (
			account_id, first_name, last_name, short_title, long_title, short_display_name,	long_display_name,
			suffix, prefix, middle_name, gender, dob_year, dob_month, dob_day, status, clinician_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		provider.AccountID.Int64(), provider.FirstName, provider.LastName, provider.ShortTitle,
		provider.LongTitle, provider.ShortDisplayName, provider.LongDisplayName, provider.MiddleName,
		provider.Suffix, provider.Prefix, provider.Gender, provider.DOB.Year, provider.DOB.Month, provider.DOB.Day,
		DoctorRegistered, provider.DoseSpotClinicianID)
	if err != nil {
		tx.Rollback()
		return 0, errors.Trace(err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, errors.Trace(err)
	}

	provider.ID = encoding.DeprecatedNewObjectID(lastID)

	// Initialize the providers practice model records with nothing enabled
	if err := d.initializePracticeModelInAllStates(provider.ID.Int64(), tx); err != nil {
		tx.Rollback()
		return 0, errors.Trace(err)
	}

	if provider.Address != nil {
		provider.Address.ID, err = addAddress(tx, provider.Address)
		if err != nil {
			tx.Rollback()
			return 0, errors.Trace(err)
		}

		_, err = tx.Exec(`INSERT INTO doctor_address_selection (doctor_id, address_id) VALUES (?,?)`, lastID, provider.Address.ID)
		if err != nil {
			tx.Rollback()
			return 0, errors.Trace(err)
		}
	}

	if provider.CellPhone != "" {
		_, err = tx.Exec(`INSERT INTO account_phone (phone, phone_type, account_id, status) VALUES (?,?,?,?) `,
			provider.CellPhone.String(), common.PNTCell.String(), provider.AccountID.Int64(), StatusActive)
		if err != nil {
			tx.Rollback()
			return 0, errors.Trace(err)
		}
	}

	res, err = tx.Exec(`INSERT INTO person (role_type_id, role_id) VALUES (?, ?)`, d.roleTypeMapping[role], lastID)
	if err != nil {
		tx.Rollback()
		return 0, errors.Trace(err)
	}
	provider.PersonID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, errors.Trace(err)
	}

	return lastID, errors.Trace(tx.Commit())
}

func (d *dataService) GetDoctorFromID(doctorID int64) (*common.Doctor, error) {
	return d.queryDoctor(`doctor.id = ? AND (account_phone.phone IS NULL OR account_phone.phone_type = ?)`,
		doctorID, common.PNTCell.String())
}

func (d *dataService) Doctor(id int64, basicInfoOnly bool) (*common.Doctor, error) {
	if !basicInfoOnly {
		return d.GetDoctorFromID(id)
	}

	return d.scanDoctor(d.db.QueryRow(`
		SELECT d.id, account_id, first_name, last_name, short_title, long_title, short_display_name, long_display_name, gender,
			dob_year, dob_month, dob_day, status, clinician_id, small_thumbnail_id, large_thumbnail_id, hero_image_id, npi_number,
			dea_number, a.role_type_id, d.primary_cc
		FROM doctor d
		INNER JOIN account a ON a.id = d.account_id
		WHERE d.id = ?`, id))
}

func (d *dataService) Doctors(ids []int64) ([]*common.Doctor, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := d.db.Query(`
		SELECT d.id, account_id, first_name, last_name, short_title, long_title, short_display_name, long_display_name, gender,
			dob_year, dob_month, dob_day, status, clinician_id, small_thumbnail_id, large_thumbnail_id, hero_image_id, npi_number,
			dea_number, a.role_type_id, d.primary_cc
		FROM doctor d
		INNER JOIN account a ON a.id = d.account_id
		WHERE d.id in (`+dbutil.MySQLArgs(len(ids))+`)`,
		dbutil.AppendInt64sToInterfaceSlice(nil, ids)...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()

	doctorMap := make(map[int64]*common.Doctor)
	for rows.Next() {
		doctor, err := d.scanDoctor(rows)
		if err != nil {
			return nil, errors.Trace(err)
		}
		doctorMap[doctor.ID.Int64()] = doctor
	}

	doctors := make([]*common.Doctor, len(ids))
	for i, doctorID := range ids {
		doctors[i] = doctorMap[doctorID]
	}

	return doctors, errors.Trace(rows.Err())
}

func (d *dataService) ListCareProviders(opt ListCareProvidersOption) ([]*common.Doctor, error) {
	var vals []interface{}
	var where string
	if opt.Has(LCPOptDoctorsOnly) {
		where = "WHERE a.role_type_id = ?"
		vals = append(vals, d.roleTypeMapping[RoleDoctor])
	} else if opt.Has(LCPOptPrimaryCCOnly) {
		where = "WHERE a.role_type_id = ? AND d.primary_cc = ?"
		vals = append(vals, d.roleTypeMapping[RoleCC], true)
	} else if opt.Has(LCPOptCCOnly) {
		where = "WHERE a.role_type_id = ?"
		vals = append(vals, d.roleTypeMapping[RoleCC])
	}
	rows, err := d.db.Query(`
		SELECT d.id, account_id, first_name, last_name, short_title, long_title, short_display_name, long_display_name, gender,
			dob_year, dob_month, dob_day, status, clinician_id, small_thumbnail_id, large_thumbnail_id, hero_image_id, npi_number,
			dea_number, a.role_type_id, d.primary_cc
		FROM doctor d
		INNER JOIN account a ON a.id = d.account_id `+where, vals...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()
	var doctors []*common.Doctor
	for rows.Next() {
		dr, err := d.scanDoctor(rows)
		if err != nil {
			return nil, errors.Trace(err)
		}
		doctors = append(doctors, dr)
	}
	return doctors, errors.Trace(rows.Err())
}

func (d *dataService) scanDoctor(s scannable) (*common.Doctor, error) {
	var doctor common.Doctor
	var smallThumbnailID, largeThumbnailID, heroImageID sql.NullString
	var shortTitle, longTitle, shortDisplayName, longDisplayName sql.NullString
	var NPI, DEA sql.NullString
	var clinicianID sql.NullInt64
	var roleTypeID int64
	err := s.Scan(
		&doctor.ID, &doctor.AccountID, &doctor.FirstName, &doctor.LastName,
		&shortTitle, &longTitle, &shortDisplayName, &longDisplayName,
		&doctor.Gender, &doctor.DOB.Year, &doctor.DOB.Month, &doctor.DOB.Day,
		&doctor.Status, &clinicianID, &smallThumbnailID, &largeThumbnailID,
		&heroImageID, &NPI, &DEA, &roleTypeID, &doctor.IsPrimaryCC)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound("doctor")
	} else if err != nil {
		return nil, err
	}
	doctor.ShortTitle = shortTitle.String
	doctor.LongTitle = longTitle.String
	doctor.ShortDisplayName = shortDisplayName.String
	doctor.LongDisplayName = longDisplayName.String
	doctor.SmallThumbnailID = smallThumbnailID.String
	doctor.DoseSpotClinicianID = clinicianID.Int64
	doctor.LargeThumbnailID = largeThumbnailID.String
	doctor.HeroImageID = heroImageID.String
	doctor.IsCC = roleTypeID != d.roleTypeMapping[RoleDoctor]
	return &doctor, nil
}

func (d *dataService) GetDoctorFromAccountID(accountID int64) (*common.Doctor, error) {
	return d.queryDoctor(`doctor.account_id = ? AND (account_phone.phone IS NULL OR account_phone.phone_type = ?)`,
		accountID, common.PNTCell.String())
}

func (d *dataService) GetDoctorFromDoseSpotClinicianID(clinicianID int64) (*common.Doctor, error) {
	return d.queryDoctor(`doctor.clinician_id = ? AND (account_phone.phone IS NULL OR account_phone.phone_type = ?)`,
		clinicianID, common.PNTCell.String())
}

func (d *dataService) GetAccountIDFromDoctorID(doctorID int64) (int64, error) {
	var accountID int64
	err := d.db.QueryRow(`select account_id from doctor where id = ?`, doctorID).Scan(&accountID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return accountID, err
}

func (d *dataService) GetFirstDoctorWithAClinicianID() (*common.Doctor, error) {
	return d.queryDoctor(`doctor.clinician_id is not null AND (account_phone.phone IS NULL OR account_phone.phone_type = ?) LIMIT 1`, common.PNTCell.String())
}

func (d *dataService) queryDoctor(where string, queryParams ...interface{}) (*common.Doctor, error) {
	row := d.db.QueryRow(fmt.Sprintf(`
		SELECT doctor.id, doctor.account_id, phone, first_name, last_name, middle_name, suffix,
			prefix, short_title, long_title, short_display_name, long_display_name, account.email,
			gender, dob_year, dob_month, dob_day, doctor.status, clinician_id,
			address.address_line_1,	address.address_line_2, address.city, address.state,
			address.zip_code, person.id, npi_number, dea_number, account.role_type_id,
			doctor.small_thumbnail_id, doctor.large_thumbnail_id, doctor.hero_image_id,
			doctor.primary_cc
		FROM doctor
		INNER JOIN account ON account.id = doctor.account_id
		INNER JOIN person ON person.role_type_id = account.role_type_id AND person.role_id = doctor.id
		LEFT OUTER JOIN account_phone ON account_phone.account_id = doctor.account_id
		LEFT OUTER JOIN doctor_address_selection ON doctor_address_selection.doctor_id = doctor.id
		LEFT OUTER JOIN address ON doctor_address_selection.address_id = address.id
		WHERE %s`, where),
		queryParams...)

	var firstName, lastName, status, gender, email string
	var addressLine1, addressLine2, city, state, zipCode sql.NullString
	var middleName, suffix, prefix, shortTitle, longTitle sql.NullString
	var smallThumbnailID, largeThumbnailID, heroImageID sql.NullString
	var cellPhoneNumber common.Phone
	var doctorID, accountID encoding.ObjectID
	var dobYear, dobMonth, dobDay int
	var personID, roleTypeID int64
	var clinicianID sql.NullInt64
	var NPI, DEA, shortDisplayName, longDisplayName sql.NullString
	var primaryCC bool

	err := row.Scan(
		&doctorID, &accountID, &cellPhoneNumber, &firstName, &lastName,
		&middleName, &suffix, &prefix, &shortTitle, &longTitle, &shortDisplayName,
		&longDisplayName, &email, &gender, &dobYear, &dobMonth,
		&dobDay, &status, &clinicianID, &addressLine1, &addressLine2,
		&city, &state, &zipCode, &personID, &NPI, &DEA, &roleTypeID,
		&smallThumbnailID, &largeThumbnailID, &heroImageID, &primaryCC)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound("doctor")
	} else if err != nil {
		return nil, err
	}

	doctor := &common.Doctor{
		ID:                  doctorID,
		AccountID:           accountID,
		FirstName:           firstName,
		LastName:            lastName,
		MiddleName:          middleName.String,
		Suffix:              suffix.String,
		Prefix:              prefix.String,
		ShortTitle:          shortTitle.String,
		LongTitle:           longTitle.String,
		ShortDisplayName:    shortDisplayName.String,
		LongDisplayName:     longDisplayName.String,
		SmallThumbnailID:    smallThumbnailID.String,
		LargeThumbnailID:    largeThumbnailID.String,
		HeroImageID:         heroImageID.String,
		Status:              status,
		Gender:              gender,
		Email:               email,
		CellPhone:           cellPhoneNumber,
		DoseSpotClinicianID: clinicianID.Int64,
		Address: &common.Address{
			AddressLine1: addressLine1.String,
			AddressLine2: addressLine2.String,
			City:         city.String,
			State:        state.String,
			ZipCode:      zipCode.String,
		},
		DOB:         encoding.Date{Year: dobYear, Month: dobMonth, Day: dobDay},
		PersonID:    personID,
		NPI:         NPI.String,
		DEA:         DEA.String,
		IsCC:        d.roleTypeMapping[RoleCC] == roleTypeID,
		IsPrimaryCC: primaryCC,
	}

	doctor.PromptStatus, err = d.GetPushPromptStatus(doctor.AccountID.Int64())
	if err != nil {
		return nil, err
	}

	return doctor, nil
}

func (d *dataService) GetDoctorIDFromAccountID(accountID int64) (int64, error) {
	var doctorID int64
	err := d.db.QueryRow("SELECT id FROM doctor WHERE account_id = ?", accountID).Scan(&doctorID)
	return doctorID, err
}

func (d *dataService) GetRegimenStepsForDoctor(doctorID int64) ([]*common.DoctorInstructionItem, error) {
	rows, err := d.db.Query(`
		SELECT id, text, status
		FROM dr_regimen_step where doctor_id = ? AND status = ?`, doctorID, StatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*common.DoctorInstructionItem
	for rows.Next() {
		var step common.DoctorInstructionItem
		if err := rows.Scan(
			&step.ID,
			&step.Text,
			&step.Status); err != nil {
			return nil, err
		}
		steps = append(steps, &step)
	}

	return steps, rows.Err()
}

func (d *dataService) GetRegimenStepForDoctor(regimenStepID, doctorID int64) (*common.DoctorInstructionItem, error) {
	var regimenStep common.DoctorInstructionItem
	err := d.db.QueryRow(`
		SELECT id, text, status
		FROM dr_regimen_step
		WHERE id = ? AND doctor_id = ?`, regimenStepID, doctorID,
	).Scan(&regimenStep.ID, &regimenStep.Text, &regimenStep.Status)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound("dr_regimen_step")
	}

	return &regimenStep, err
}

func (d *dataService) AddRegimenStepForDoctor(regimenStep *common.DoctorInstructionItem, doctorID int64) error {
	res, err := d.db.Exec(`insert into dr_regimen_step (text, doctor_id,status) values (?,?,?)`, regimenStep.Text, doctorID, StatusActive)
	if err != nil {
		return err
	}
	instructionID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// assign an id given that its a new regimen step
	regimenStep.ID = encoding.DeprecatedNewObjectID(instructionID)
	return nil
}

func (d *dataService) UpdateRegimenStepForDoctor(regimenStep *common.DoctorInstructionItem, doctorID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// lookup the sourceId and status for the current regimen step if it exists
	var sourceID sql.NullInt64
	var status string
	if err := tx.QueryRow(`
		SELECT source_id, status
		FROM dr_regimen_step
		WHERE id = ? AND doctor_id = ?`,
		regimenStep.ID.Int64(), doctorID,
	).Scan(&sourceID, &status); err != nil {
		return err
	}

	// if the source id does not exist for the step, this means that
	// this step is the source itself. tracking the source id helps for
	// tracking revision from the beginning of time.
	sourceIDForUpdatedStep := regimenStep.ID.Int64()
	if sourceID.Valid {
		sourceIDForUpdatedStep = sourceID.Int64
	}

	// update the current regimen step to be inactive
	_, err = tx.Exec(`UPDATE dr_regimen_step SET status = ? WHERE id = ? AND doctor_id = ?`,
		StatusInactive, regimenStep.ID.Int64(), doctorID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// insert a new active regimen step in its place
	res, err := tx.Exec(`INSERT INTO dr_regimen_step (text, doctor_id, source_id, status) VALUES (?, ?, ?, ?)`,
		regimenStep.Text, doctorID, sourceIDForUpdatedStep, status)
	if err != nil {
		tx.Rollback()
		return err
	}

	instructionID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	// update the regimenStep Id
	regimenStep.ID = encoding.DeprecatedNewObjectID(instructionID)
	return tx.Commit()
}

func (d *dataService) MarkRegimenStepsToBeDeleted(regimenSteps []*common.DoctorInstructionItem, doctorID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	for _, regimenStep := range regimenSteps {
		_, err = tx.Exec(`UPDATE dr_regimen_step SET status = ? WHERE id = ? AND doctor_id=?`,
			StatusDeleted, regimenStep.ID.Int64(), doctorID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// DoctorQueueAction captures the possible actions that can take place on the doctor queue.
type DoctorQueueAction int

const (
	// DQActionInsert is an action that represents inserting of a single item into the doctor queue.
	DQActionInsert DoctorQueueAction = iota + 1

	// DQActionReplace is an action that represents replacing a single item in the doctor queue with another item. The item is only replaced if an item in
	// the specified current state is found in the doctor queue.
	DQActionReplace

	// DQActionRemove is an action that represents removing of a single item in the doctor queue.
	DQActionRemove
)

func (a DoctorQueueAction) String() string {
	switch a {
	case DQActionInsert:
		return "DQActionInsert"
	case DQActionReplace:
		return "DQActionReplace"
	case DQActionRemove:
		return "DQActionRemove"
	}
	return fmt.Sprintf("DoctorQueueAction(%d)", int(a))
}

// DoctorQueueUpdate represents a single update to undertake on the doctor queue with the provided action and item
type DoctorQueueUpdate struct {
	// Action represents the action to undertake.
	Action DoctorQueueAction

	// QueueItem represents the unique item with which to undertake the action on the doctor queue.
	QueueItem *DoctorQueueItem

	// Dedupe is only applicable for an insert action and indicates whether or not to dedupe on an insert
	// into the doctor's inbox.
	Dedupe bool

	// CurrentState is only applicable for a replace and indicates the current state an item is expected to be in the
	// doctor's inbox before replacing it with another item.
	CurrentState string
}

func (d *dataService) UpdateDoctorQueue(updates []*DoctorQueueUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	if err := updateDoctorQueue(tx, updates); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func updateDoctorQueue(tx tsql.Tx, updates []*DoctorQueueUpdate) error {
	for _, update := range updates {
		switch update.Action {
		case DQActionInsert:
			if err := insertItemIntoDoctorQueue(tx, update.QueueItem, update.Dedupe); err != nil {
				return err
			}
		case DQActionRemove:
			if err := deleteItemFromDoctorQueue(tx, update.QueueItem); err != nil {
				return err
			}
		case DQActionReplace:
			if err := replaceItemInDoctorQueue(tx, update.QueueItem, update.CurrentState); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown action type: %d", update.Action)
		}
	}

	return nil
}

func deleteItemFromDoctorQueue(tx tsql.Tx, doctorQueueItem *DoctorQueueItem) error {
	var err error
	if doctorQueueItem.QueueType == DQTUnclaimedQueue {
		_, err = tx.Exec(`
			DELETE FROM unclaimed_case_queue
			WHERE item_id = ? AND event_type = ? AND status = ?`,
			doctorQueueItem.ItemID,
			doctorQueueItem.EventType,
			doctorQueueItem.Status)
	} else {
		if doctorQueueItem.DoctorID == 0 {
			return errors.New("api: doctor ID required to delete doctor queue item")
		}
		_, err = tx.Exec(`
			DELETE FROM doctor_queue
			WHERE doctor_id = ? AND item_id = ? AND event_type = ? AND status = ?`,
			doctorQueueItem.DoctorID,
			doctorQueueItem.ItemID,
			doctorQueueItem.EventType,
			doctorQueueItem.Status)
	}
	return err
}

func insertItemIntoDoctorQueue(tx tsql.Tx, dqi *DoctorQueueItem, dedupe bool) error {
	if err := dqi.Validate(); err != nil {
		return err
	}

	// delete and reinsert on dedupe to pop the item to the back of the queue if dedupe set to true.
	if dedupe {
		if err := deleteItemFromDoctorQueue(tx, dqi); err != nil {
			return err
		}
	}

	_, err := tx.Exec(`
		INSERT INTO doctor_queue (
			doctor_id, patient_id, item_id, event_type, status,
			description, short_description, action_url, tags)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		dqi.DoctorID,
		dqi.PatientID,
		dqi.ItemID,
		dqi.EventType,
		dqi.Status,
		dqi.Description,
		dqi.ShortDescription,
		dqi.ActionURL.String(),
		strings.Join(dqi.Tags, tagSeparator))

	return err
}

func replaceItemInDoctorQueue(tx tsql.Tx, dqi *DoctorQueueItem, currentState string) error {
	// check if there is an item to replace. If not, then do nothing.
	var id int64
	if err := tx.QueryRow(`
		SELECT id
		FROM doctor_queue
		WHERE status = ? AND doctor_id = ? AND event_type = ? AND item_id = ?`,
		currentState, dqi.DoctorID, dqi.EventType, dqi.ItemID).Scan(&id); err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	_, err := tx.Exec(`
		DELETE FROM doctor_queue
		WHERE status = ? AND doctor_id = ? AND event_type = ? AND item_id = ?`,
		currentState, dqi.DoctorID, dqi.EventType, dqi.ItemID)
	if err != nil {
		return err
	}

	return insertItemIntoDoctorQueue(tx, dqi, false)
}

func (d *dataService) MarkPatientVisitAsOngoingInDoctorQueue(doctorID, patientVisitID int64) error {
	_, err := d.db.Exec(`
		UPDATE doctor_queue SET status = ? WHERE event_type = ? AND item_id = ? AND doctor_id = ?`,
		StatusOngoing,
		DQEventTypePatientVisit,
		patientVisitID,
		doctorID)
	return err
}

// CompleteVisitOnTreatmentPlanGeneration updates the doctor queue upon the generation of a treatment plan to create a completed item as well as
// clear out any submitted visit by the patient pertaining to the case.
func (d *dataService) CompleteVisitOnTreatmentPlanGeneration(
	doctorID, patientVisitID, treatmentPlanID int64,
	updates []*DoctorQueueUpdate) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// get list of possible patient visits that could be in the doctor's queue in this case
	openStates := common.OpenPatientVisitStates()
	vals := []interface{}{treatmentPlanID}
	vals = dbutil.AppendStringsToInterfaceSlice(vals, openStates)
	rows, err := tx.Query(`
		SELECT patient_visit.id
		FROM patient_visit
		INNER JOIN treatment_plan on treatment_plan.patient_case_id = patient_visit.patient_case_id
		WHERE treatment_plan.id = ?
		AND patient_visit.status not in (`+dbutil.MySQLArgs(len(openStates))+`)`, vals...)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var visitIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			tx.Rollback()
			return err
		}

		visitIDs = append(visitIDs, id)
	}

	if err := rows.Err(); err != nil {
		tx.Rollback()
		return err
	}

	if len(visitIDs) > 0 {
		vals := []interface{}{DQItemStatusOngoing, doctorID, DQEventTypePatientVisit}
		vals = dbutil.AppendInt64sToInterfaceSlice(vals, visitIDs)

		_, err = tx.Exec(`
		DELETE FROM doctor_queue
		WHERE status = ? AND doctor_id = ? AND event_type = ?
		AND item_id in (`+dbutil.MySQLArgs(len(visitIDs))+`)`, vals...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := updateDoctorQueue(tx, updates); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *dataService) GetPendingItemsInDoctorQueue(doctorID int64) ([]*DoctorQueueItem, error) {
	rows, err := d.db.Query(fmt.Sprintf(`
		SELECT id, event_type, item_id, enqueue_date, status, doctor_id, patient_id, description, short_description, action_url, tags
		FROM doctor_queue
		WHERE doctor_id = ? AND status IN (%s)
		ORDER BY enqueue_date`, dbutil.MySQLArgs(2)), doctorID, StatusPending, StatusOngoing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return populateDoctorQueueFromRows(rows)
}

func (d *dataService) GetPendingItemsInCCQueues() ([]*DoctorQueueItem, error) {
	rows, err := d.db.Query(`
		SELECT id, event_type, item_id, enqueue_date, status, doctor_id, patient_id,
			description, short_description, action_url, tags
		FROM doctor_queue
		WHERE doctor_id IN (
			SELECT d.id
			FROM doctor d
			INNER JOIN account a ON a.id = d.account_id
			WHERE a.role_type_id = ?
		) AND status IN (?, ?)
		ORDER BY enqueue_date`, d.roleTypeMapping[RoleCC], StatusPending, StatusOngoing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return populateDoctorQueueFromRows(rows)
}

func (d *dataService) GetCompletedItemsInDoctorQueue(doctorID int64) ([]*DoctorQueueItem, error) {
	rows, err := d.db.Query(fmt.Sprintf(`
		SELECT id, event_type, item_id, enqueue_date, status, doctor_id, patient_id, description, short_description, action_url, tags
		FROM doctor_queue
		WHERE doctor_id = ? AND status NOT IN (%s)
		ORDER BY enqueue_date DESC`, dbutil.MySQLArgs(2)), doctorID, StatusPending, StatusOngoing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return populateDoctorQueueFromRows(rows)
}

func (d *dataService) GetPendingItemsForClinic() ([]*DoctorQueueItem, error) {
	// get all the items in in the unassigned queue
	unclaimedQueueItems, err := d.GetAllItemsInUnclaimedQueue()
	if err != nil {
		return nil, err
	}

	// now get all pending items in the doctor queue
	rows, err := d.db.Query(`
		SELECT id, event_type, item_id, enqueue_date, status, doctor_id, patient_id, description, short_description, action_url, tags
		FROM doctor_queue
		WHERE status IN (`+dbutil.MySQLArgs(2)+`)
		ORDER BY enqueue_date`, StatusPending, StatusOngoing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queueItems, err := populateDoctorQueueFromRows(rows)
	if err != nil {
		return nil, err
	}

	queueItems = append(queueItems, unclaimedQueueItems...)

	// sort the items in ascending order so as to return a wholistic view of the items that are pending
	sort.Sort(byTimestamp(queueItems))

	return queueItems, nil
}

func (d *dataService) GetCompletedItemsForClinic() ([]*DoctorQueueItem, error) {
	rows, err := d.db.Query(`
		SELECT id, event_type, item_id, enqueue_date, status, doctor_id, patient_id, description, short_description, action_url, tags
		FROM doctor_queue
		WHERE status NOT IN (`+dbutil.MySQLArgs(2)+`)
		ORDER BY enqueue_date desc`, StatusOngoing, StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return populateDoctorQueueFromRows(rows)
}

func populateDoctorQueueFromRows(rows *sql.Rows) ([]*DoctorQueueItem, error) {
	var doctorQueue []*DoctorQueueItem
	for rows.Next() {
		var queueItem DoctorQueueItem
		var actionURL string
		var tags sql.NullString
		err := rows.Scan(
			&queueItem.ID,
			&queueItem.EventType,
			&queueItem.ItemID,
			&queueItem.EnqueueDate,
			&queueItem.Status,
			&queueItem.DoctorID,
			&queueItem.PatientID,
			&queueItem.Description,
			&queueItem.ShortDescription,
			&actionURL,
			&tags)
		if err != nil {
			return nil, err
		}
		queueItem.QueueType = DQTDoctorQueue

		if actionURL != "" {
			aURL, err := app_url.ParseSpruceAction(actionURL)
			if err != nil {
				golog.Errorf("Unable to parse actionURL: %s", err.Error())
			} else {
				queueItem.ActionURL = &aURL
			}
		}
		if tags.String != "" {
			queueItem.Tags = strings.Split(tags.String, tagSeparator)
		} else {
			queueItem.Tags = []string{}
		}

		doctorQueue = append(doctorQueue, &queueItem)
	}
	return doctorQueue, rows.Err()
}

func (d *dataService) GetMedicationDispenseUnits(languageID int64) ([]int64, []string, error) {
	rows, err := d.db.Query(`select dispense_unit.id, ltext from dispense_unit inner join localized_text on app_text_id = dispense_unit_text_id where language_id=?`, languageID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var dispenseUnitIDs []int64
	var dispenseUnits []string
	for rows.Next() {
		var dipenseUnitID int64
		var dispenseUnit string
		if err := rows.Scan(&dipenseUnitID, &dispenseUnit); err != nil {
			return nil, nil, err
		}
		dispenseUnits = append(dispenseUnits, dispenseUnit)
		dispenseUnitIDs = append(dispenseUnitIDs, dipenseUnitID)
	}
	return dispenseUnitIDs, dispenseUnits, rows.Err()
}

func (d *dataService) AddTreatmentTemplates(doctorTreatmentTemplates []*common.DoctorTreatmentTemplate, doctorID, treatmentPlanID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	for _, doctorTreatmentTemplate := range doctorTreatmentTemplates {
		var treatmentIDInPatientTreatmentPlan int64
		if treatmentPlanID != 0 {
			treatmentIDInPatientTreatmentPlan = doctorTreatmentTemplate.Treatment.ID.Int64()
		}

		treatmentType := treatmentRX
		if doctorTreatmentTemplate.Treatment.OTC {
			treatmentType = treatmentOTC
		}

		columnsAndData := map[string]interface{}{
			"drug_internal_name":    doctorTreatmentTemplate.Treatment.DrugInternalName,
			"dosage_strength":       doctorTreatmentTemplate.Treatment.DosageStrength,
			"type":                  treatmentType,
			"dispense_value":        doctorTreatmentTemplate.Treatment.DispenseValue,
			"dispense_unit_id":      doctorTreatmentTemplate.Treatment.DispenseUnitID.Int64(),
			"refills":               doctorTreatmentTemplate.Treatment.NumberRefills.Int64Value,
			"substitutions_allowed": doctorTreatmentTemplate.Treatment.SubstitutionsAllowed,
			"patient_instructions":  doctorTreatmentTemplate.Treatment.PatientInstructions,
			"pharmacy_notes":        doctorTreatmentTemplate.Treatment.PharmacyNotes,
			"status":                common.TStatusCreated.String(),
			"doctor_id":             doctorID,
			"name":                  doctorTreatmentTemplate.Name,
		}

		if doctorTreatmentTemplate.Treatment.DaysSupply.IsValid {
			columnsAndData["days_supply"] = doctorTreatmentTemplate.Treatment.DaysSupply.Int64Value
		}

		if err := d.includeDrugNameComponentIfNonZero(doctorTreatmentTemplate.Treatment.GenericDrugName, drugNameTable, "generic_drug_name_id", columnsAndData, tx); err != nil {
			tx.Rollback()
			return err
		}

		if err := d.includeDrugNameComponentIfNonZero(doctorTreatmentTemplate.Treatment.DrugName, drugNameTable, "drug_name_id", columnsAndData, tx); err != nil {
			tx.Rollback()
			return err
		}

		if err := d.includeDrugNameComponentIfNonZero(doctorTreatmentTemplate.Treatment.DrugForm, drugFormTable, "drug_form_id", columnsAndData, tx); err != nil {
			tx.Rollback()
			return err
		}

		if err := d.includeDrugNameComponentIfNonZero(doctorTreatmentTemplate.Treatment.DrugRoute, drugRouteTable, "drug_route_id", columnsAndData, tx); err != nil {
			tx.Rollback()
			return err
		}

		columns, values := getKeysAndValuesFromMap(columnsAndData)
		for i, c := range columns {
			columns[i] = dbutil.EscapeMySQLName(c)
		}
		res, err := tx.Exec(fmt.Sprintf(`INSERT INTO dr_treatment_template (%s) VALUES (%s)`,
			strings.Join(columns, ","), dbutil.MySQLArgs(len(values))), values...)
		if err != nil {
			tx.Rollback()
			return err
		}

		drTreatmentTemplateID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}

		// update the treatment object with the information
		doctorTreatmentTemplate.ID = encoding.DeprecatedNewObjectID(drTreatmentTemplateID)

		// add drug db ids to the table
		for drugDbTag, drugDBID := range doctorTreatmentTemplate.Treatment.DrugDBIDs {
			_, err := tx.Exec(`insert into dr_treatment_template_drug_db_id (drug_db_id_tag, drug_db_id, dr_treatment_template_id) values (?, ?, ?)`, drugDbTag, drugDBID, drTreatmentTemplateID)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// mark the fact that the treatment was added as a favorite from a patient's treatment
		// and so the selection needs to be maintained
		if treatmentIDInPatientTreatmentPlan != 0 {

			// delete any pre-existing favorite treatment that is already linked against this treatment in the patient visit,
			// because that means that the client has an out-of-sync list for some reason, and we should treat
			// what the client has as the source of truth. Otherwise, we will have two favorite treatments that are craeted
			// both of which are mapped against the exist same treatment_id
			// this should rarely happen; but what this will do is help ensure that a treatment within a patient visit can only be favorited
			// once and only once.
			var preExistingDoctorFavoriteTreatmentID int64
			err = tx.QueryRow(`select dr_treatment_template_id from treatment_dr_template_selection where treatment_id = ? `, treatmentIDInPatientTreatmentPlan).Scan(&preExistingDoctorFavoriteTreatmentID)
			if err != nil && err != sql.ErrNoRows {
				tx.Rollback()
				return err
			}

			if preExistingDoctorFavoriteTreatmentID != 0 {
				// go ahead and delete the selection
				_, err = tx.Exec(`delete from treatment_dr_template_selection where treatment_id = ?`, treatmentIDInPatientTreatmentPlan)
				if err != nil {
					tx.Rollback()
					return err
				}

				// also, go ahead and mark this particular favorited treatment as deleted
				_, err = tx.Exec(`update dr_treatment_template set status = ? where id = ?`, common.TStatusDeleted.String(), preExistingDoctorFavoriteTreatmentID)
				if err != nil {
					tx.Rollback()
					return err
				}
			}

			_, err = tx.Exec(`insert into treatment_dr_template_selection (treatment_id, dr_treatment_template_id) values (?,?)`, treatmentIDInPatientTreatmentPlan, drTreatmentTemplateID)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

	}

	return tx.Commit()
}

func (d *dataService) DeleteTreatmentTemplates(doctorTreatmentTemplates []*common.DoctorTreatmentTemplate, doctorID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	for _, doctorTreatmentTemplate := range doctorTreatmentTemplates {
		_, err = tx.Exec(`update dr_treatment_template set status=? where id = ? and doctor_id = ?`, common.TStatusDeleted.String(), doctorTreatmentTemplate.ID.Int64(), doctorID)
		if err != nil {
			tx.Rollback()
			return err
		}

		// delete all previous selections for this favorited treatment
		_, err = tx.Exec(`delete from treatment_dr_template_selection where dr_treatment_template_id = ?`, doctorTreatmentTemplate.ID.Int64())
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (d *dataService) GetTreatmentTemplates(doctorID int64) ([]*common.DoctorTreatmentTemplate, error) {
	rows, err := d.db.Query(`
		SELECT dtt.id, dtt.name, drug_internal_name, dosage_strength, type,
			dispense_value, dispense_unit_id, ltext, refills, substitutions_allowed,
			days_supply, COALESCE(pharmacy_notes, ''), patient_instructions, creation_date,
			status, COALESCE(dn.name, ''), COALESCE(dr.name, ''), COALESCE(df.name, ''),
			COALESCE(dgn.name, '')
		FROM dr_treatment_template dtt
		INNER JOIN dispense_unit ON dtt.dispense_unit_id = dispense_unit.id
		INNER JOIN localized_text ON localized_text.app_text_id = dispense_unit.dispense_unit_text_id
		LEFT JOIN drug_name dn ON dn.id = drug_name_id
		LEFT JOIN drug_route dr ON dr.id = drug_route_id
		LEFT JOIN drug_form df ON df.id = drug_form_id
		LEFT JOIN drug_name dgn ON dgn.id = generic_drug_name_id
		WHERE status = ? AND doctor_id = ? AND localized_text.language_id = ?`,
		common.TStatusCreated.String(), doctorID, LanguageIDEnglish)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var treatmentTemplates []*common.DoctorTreatmentTemplate
	for rows.Next() {
		dtt := &common.DoctorTreatmentTemplate{
			Treatment: &common.Treatment{},
		}
		var treatmentType string
		err = rows.Scan(
			&dtt.ID, &dtt.Name, &dtt.Treatment.DrugInternalName, &dtt.Treatment.DosageStrength, &treatmentType,
			&dtt.Treatment.DispenseValue, &dtt.Treatment.DispenseUnitID, &dtt.Treatment.DispenseUnitDescription,
			&dtt.Treatment.NumberRefills, &dtt.Treatment.SubstitutionsAllowed, &dtt.Treatment.DaysSupply,
			&dtt.Treatment.PharmacyNotes, &dtt.Treatment.PatientInstructions, &dtt.Treatment.CreationDate,
			&dtt.Treatment.Status, &dtt.Treatment.DrugName, &dtt.Treatment.DrugRoute, &dtt.Treatment.DrugForm,
			&dtt.Treatment.GenericDrugName)
		if err != nil {
			return nil, err
		}

		dtt.Treatment.OTC = treatmentType == treatmentOTC

		err = d.fillInDrugDBIdsForTreatment(dtt.Treatment, dtt.ID.Int64(), "dr_treatment_template")
		if err != nil {
			return nil, err
		}

		treatmentTemplates = append(treatmentTemplates, dtt)
	}
	return treatmentTemplates, rows.Err()
}

func (d *dataService) SetTreatmentPlanNote(doctorID, treatmentPlanID int64, note string) error {
	// Use NULL for empty note
	msg := sql.NullString{
		String: note,
		Valid:  note != "",
	}
	_, err := d.db.Exec(`UPDATE treatment_plan SET note = ? WHERE id = ? AND doctor_id = ?`,
		msg, treatmentPlanID, doctorID)
	return err
}

func (d *dataService) GetTreatmentPlanNote(treatmentPlanID int64) (string, error) {
	var note sql.NullString
	row := d.db.QueryRow(`SELECT note FROM treatment_plan WHERE id = ?`, treatmentPlanID)
	err := row.Scan(&note)
	if err == sql.ErrNoRows {
		err = ErrNotFound("note")
	}
	return note.String, err
}

func (d *dataService) getIDForNameFromTable(tableName, drugComponentName string) (int64, error) {
	var id int64
	err := d.db.QueryRow(`SELECT id FROM `+dbutil.EscapeMySQLName(tableName)+` WHERE name = ?`, drugComponentName).Scan(&id)
	return id, err
}

func (d *dataService) getOrInsertNameInTable(db db, tableName, drugComponentName string) (int64, error) {
	id, err := d.getIDForNameFromTable(tableName, drugComponentName)
	if err == nil {
		return id, nil
	} else if err != sql.ErrNoRows {
		return 0, err
	}
	res, err := db.Exec(`INSERT INTO `+dbutil.EscapeMySQLName(tableName)+` (name) VALUES (?)`, drugComponentName)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// DoctorUpdate represents the mutable aspects of the doctor record
type DoctorUpdate struct {
	ShortTitle          *string
	LongTitle           *string
	ShortDisplayName    *string
	LongDisplayName     *string
	NPI                 *string
	DEA                 *string
	LargeThumbnailID    *string
	HeroImageID         *string
	DosespotClinicianID *int64
}

func (d *dataService) UpdateDoctor(doctorID int64, update *DoctorUpdate) error {
	args := dbutil.MySQLVarArgs()

	if update.ShortTitle != nil {
		args.Append("short_title", *update.ShortTitle)
	}
	if update.LongTitle != nil {
		args.Append("long_title", *update.LongTitle)
	}
	if update.ShortDisplayName != nil {
		args.Append("short_display_name", *update.ShortDisplayName)
	}
	if update.LongDisplayName != nil {
		args.Append("long_display_name", *update.LongDisplayName)
	}
	if update.NPI != nil {
		args.Append("npi_number", *update.NPI)
	}
	if update.DEA != nil {
		args.Append("dea_number", *update.DEA)
	}
	if update.HeroImageID != nil {
		args.Append("hero_image_id", *update.HeroImageID)
	}
	if update.LargeThumbnailID != nil {
		args.Append("large_thumbnail_id", *update.LargeThumbnailID)
	}
	if update.DosespotClinicianID != nil {
		args.Append("clinician_id", *update.DosespotClinicianID)
	}

	if args.IsEmpty() {
		return nil
	}

	_, err := d.db.Exec(`UPDATE doctor SET `+args.Columns()+` WHERE id = ?`, append(args.Values(), doctorID)...)
	return err
}

func (d *dataService) DoctorAttributes(doctorID int64, names []string) (map[string]string, error) {
	var rows *sql.Rows
	var err error
	if len(names) == 0 {
		rows, err = d.db.Query(`SELECT name, value FROM doctor_attribute WHERE doctor_id = ?`, doctorID)
	} else {
		rows, err = d.db.Query(`SELECT name, value FROM doctor_attribute WHERE doctor_id = ? AND name IN (`+dbutil.MySQLArgs(len(names))+`)`,
			dbutil.AppendStringsToInterfaceSlice([]interface{}{doctorID}, names)...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	attr := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		attr[name] = value
	}
	return attr, rows.Err()
}

func (d *dataService) UpdateDoctorAttributes(doctorID int64, attributes map[string]string) error {
	if len(attributes) == 0 {
		return nil
	}
	var toDelete []interface{}
	inserts := dbutil.MySQLMultiInsert(len(attributes))
	for name, value := range attributes {
		if value == "" {
			toDelete = append(toDelete, name)
		} else {
			inserts.Append(doctorID, name, value)
		}
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	if len(toDelete) != 0 {
		_, err := tx.Exec(`DELETE FROM doctor_attribute WHERE name IN (`+dbutil.MySQLArgs(len(toDelete))+`) AND doctor_id = ?`,
			append(toDelete, doctorID)...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if !inserts.IsEmpty() {
		_, err := tx.Exec(`REPLACE INTO doctor_attribute (doctor_id, name, value) VALUES `+inserts.Query(), inserts.Values()...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *dataService) AddMedicalLicenses(licenses []*common.MedicalLicense) error {
	return d.addMedicalLicenses(d.db, licenses)
}

func (d *dataService) addMedicalLicenses(db db, licenses []*common.MedicalLicense) error {
	if len(licenses) == 0 {
		return nil
	}
	inserts := dbutil.MySQLMultiInsert(len(licenses))
	for _, l := range licenses {
		if l.State == "" || l.Number == "" || l.Status == "" {
			return errors.New("api: license is missing state, number, or status")
		}
		inserts.Append(l.DoctorID, l.State, l.Number, l.Status.String(), l.Expiration)
	}
	_, err := db.Exec(`
		REPLACE INTO doctor_medical_license
			(doctor_id, state, license_number, status, expiration_date)
		VALUES `+inserts.Query(), inserts.Values()...)
	return err
}

func (d *dataService) UpdateMedicalLicenses(doctorID int64, licenses []*common.MedicalLicense) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM doctor_medical_license WHERE doctor_id = ?`, doctorID); err != nil {
		tx.Rollback()
		return err
	}

	if err := d.addMedicalLicenses(tx, licenses); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *dataService) MedicalLicenses(doctorID int64) ([]*common.MedicalLicense, error) {
	rows, err := d.db.Query(`
		SELECT id, state, license_number, status, expiration_date
		FROM doctor_medical_license
		WHERE doctor_id = ?
		ORDER BY state`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []*common.MedicalLicense
	for rows.Next() {
		l := &common.MedicalLicense{DoctorID: doctorID}
		if err := rows.Scan(&l.ID, &l.State, &l.Number, &l.Status, &l.Expiration); err != nil {
			return nil, err
		}
		licenses = append(licenses, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return licenses, nil
}

func (d *dataService) CareProviderProfile(accountID int64) (*common.CareProviderProfile, error) {
	row := d.db.QueryRow(`
		SELECT full_name, why_spruce, qualifications, undergraduate_school, graduate_school,
			medical_school, residency, fellowship, experience, creation_date, modified_date
		FROM care_provider_profile
		WHERE account_id = ?`, accountID)

	profile := common.CareProviderProfile{
		AccountID: accountID,
	}
	// If there's no profile then return an empty struct. There's no need for the
	// caller to care if the profile is empty or doesn't exist.
	if err := row.Scan(
		&profile.FullName, &profile.WhySpruce, &profile.Qualifications, &profile.UndergraduateSchool,
		&profile.GraduateSchool, &profile.MedicalSchool, &profile.Residency, &profile.Fellowship,
		&profile.Experience, &profile.Created, &profile.Modified,
	); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &profile, nil
}

func (d *dataService) UpdateCareProviderProfile(accountID int64, profile *common.CareProviderProfile) error {
	_, err := d.db.Exec(`
		REPLACE INTO care_provider_profile (
			account_id, full_name, why_spruce, qualifications, undergraduate_school,
			graduate_school, medical_school, residency, fellowship, experience
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		accountID, profile.FullName, profile.WhySpruce, profile.Qualifications,
		profile.UndergraduateSchool, profile.GraduateSchool, profile.MedicalSchool,
		profile.Residency, profile.Fellowship, profile.Experience)
	return err
}

func (d *dataService) GetOldestTreatmentPlanInStatuses(max int, statuses []common.TreatmentPlanStatus) ([]*TreatmentPlanAge, error) {
	var whereClause string
	var params []interface{}

	if len(statuses) > 0 {
		whereClause = `WHERE status in (` + dbutil.MySQLArgs(len(statuses)) + `)`
		for _, tpStatus := range statuses {
			params = append(params, tpStatus.String())
		}
	}
	params = append(params, max)

	rows, err := d.db.Query(`
		SELECT id, last_modified_date
		FROM treatment_plan
		`+whereClause+`
		ORDER BY last_modified_date LIMIT ?`, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tpAges []*TreatmentPlanAge
	for rows.Next() {
		var tpAge TreatmentPlanAge
		var lastModifiedDate time.Time
		if err := rows.Scan(
			&tpAge.ID,
			&lastModifiedDate); err != nil {
			return nil, err
		}
		tpAge.Age = time.Since(lastModifiedDate)
		tpAges = append(tpAges, &tpAge)
	}

	return tpAges, rows.Err()
}

func (d *dataService) ListTreatmentPlanResourceGuides(tpID int64) ([]*common.ResourceGuide, error) {
	rows, err := d.db.Query(`
		SELECT id, section_id, ordinal, title, photo_url
		FROM treatment_plan_resource_guide
		INNER JOIN resource_guide rg ON rg.id = resource_guide_id
		WHERE treatment_plan_id = ?`,
		tpID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()

	var guides []*common.ResourceGuide
	for rows.Next() {
		g := &common.ResourceGuide{}
		if err := rows.Scan(&g.ID, &g.SectionID, &g.Ordinal, &g.Title, &g.PhotoURL); err != nil {
			return nil, errors.Trace(err)
		}
		guides = append(guides, g)
	}

	return guides, errors.Trace(rows.Err())
}

func (d *dataService) AddResourceGuidesToTreatmentPlan(tpID int64, guideIDs []int64) error {
	if len(guideIDs) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return errors.Trace(err)
	}

	if err := addResourceGuidesToTreatmentPlan(tx, tpID, guideIDs); err != nil {
		tx.Rollback()
		return errors.Trace(err)
	}

	return errors.Trace(tx.Commit())
}

func addResourceGuidesToTreatmentPlan(tx tsql.Tx, tpID int64, guideIDs []int64) error {
	// TODO: optimize this into a single query. not critical though since
	// the number of queries should be very low (1 or 2 maybe)
	stmt, err := tx.Prepare(`
		REPLACE INTO treatment_plan_resource_guide
			(treatment_plan_id, resource_guide_id)
		VALUES (?, ?)`)
	if err != nil {
		return errors.Trace(err)
	}
	defer stmt.Close()
	for _, id := range guideIDs {
		if _, err := stmt.Exec(tpID, id); err != nil {
			return errors.Trace(err)
		}
	}

	return errors.Trace(err)
}

func (d *dataService) RemoveResourceGuidesFromTreatmentPlan(tpID int64, guideIDs []int64) error {
	if len(guideIDs) == 0 {
		return nil
	}
	// Optimize for the common case (and currently only case)
	if len(guideIDs) == 1 {
		_, err := d.db.Exec(`
			DELETE FROM treatment_plan_resource_guide
			WHERE treatment_plan_id = ?
				AND resource_guide_id = ?`, tpID, guideIDs[0])
		return errors.Trace(err)
	}
	vals := make([]interface{}, 1, len(guideIDs)+1)
	vals[0] = tpID
	vals = dbutil.AppendInt64sToInterfaceSlice(vals, guideIDs)
	_, err := d.db.Exec(`
		DELETE FROM treatment_plan_resource_guide
		WHERE treatment_plan_id = ?
			AND resource_guide_id IN (`+dbutil.MySQLArgs(len(guideIDs))+`)`,
		vals...)
	return errors.Trace(err)
}

func (d *dataService) PracticeModel(doctorID, stateID int64) (*common.PracticeModel, error) {
	pm := &common.PracticeModel{}
	err := d.db.QueryRow(`
		SELECT doctor_id, spruce_pc, practice_extension, state_id
		FROM practice_model WHERE doctor_id = ? AND state_id = ?`, doctorID, stateID).Scan(&pm.DoctorID, &pm.IsSprucePC, &pm.HasPracticeExtension, &pm.StateID)
	if err == sql.ErrNoRows {
		// TODO: Think up a better way tp attach context to ErrNotFound
		return nil, errors.Trace(ErrNotFound(fmt.Sprintf(`practice_model(doctor_id:%d, state_id:%d)`, doctorID, stateID)))
	}
	return pm, errors.Trace(err)
}

// TODO: This should map state ID's to practice models not abbreviations in the near future
func (d *dataService) PracticeModels(doctorID int64) (map[string]*common.PracticeModel, error) {
	rows, err := d.db.Query(`
		SELECT doctor_id, spruce_pc, practice_extension, state_id, state.abbreviation
		FROM practice_model 
		JOIN state ON practice_model.state_id = state.id 
		WHERE doctor_id = ?`, doctorID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()

	pms := make(map[string]*common.PracticeModel)
	for rows.Next() {
		var pm common.PracticeModel
		var stateAbbreviation string
		if err := rows.Scan(&pm.DoctorID, &pm.IsSprucePC, &pm.HasPracticeExtension, &pm.StateID, &stateAbbreviation); err != nil {
			return nil, errors.Trace(err)
		}
		pms[stateAbbreviation] = &pm
	}

	return pms, errors.Trace(rows.Err())
}

func (d *dataService) UpdatePracticeModel(doctorID, stateID int64, pmu *common.PracticeModelUpdate) (int64, error) {
	varArgs := dbutil.MySQLVarArgs()
	if pmu.IsSprucePC != nil {
		varArgs.Append(`spruce_pc`, *pmu.IsSprucePC)
	}
	if pmu.HasPracticeExtension != nil {
		varArgs.Append(`practice_extension`, *pmu.HasPracticeExtension)
	}
	res, err := d.db.Exec(`
		UPDATE practice_model SET `+varArgs.Columns()+` WHERE doctor_id = ? AND state_id = ?`, append(varArgs.Values(), doctorID, stateID)...)
	if err != nil {
		return 0, errors.Trace(err)
	}
	aff, err := res.RowsAffected()
	return aff, errors.Trace(err)
}

func (d *dataService) HasPracticeExtensionInAnyState(doctorID int64) (bool, error) {
	var id int64
	if err := d.db.QueryRow(`SELECT doctor_id FROM practice_model WHERE practice_extension = true AND doctor_id = ? LIMIT 1`, doctorID).Scan(&id); err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, errors.Trace(err)
	}
	return true, nil
}

func (d *dataService) initializePracticeModelInAllStates(doctorID int64, db db) error {
	_, err := db.Exec(
		`INSERT INTO practice_model (doctor_id, state_id)
			SELECT ?, id FROM state 
			WHERE id NOT IN (SELECT state_id FROM practice_model WHERE doctor_id = ?)`, doctorID, doctorID)
	return errors.Trace(err)
}

func (d *dataService) InitializePracticeModelInAllStates(doctorID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Trace(err)
	}

	if err := d.initializePracticeModelInAllStates(doctorID, tx); err != nil {
		tx.Rollback()
		return errors.Trace(err)
	}

	return errors.Trace(tx.Commit())
}

func (d *dataService) UpsertPracticeModelInAllStates(doctorID int64, aspmu *common.AllStatesPracticeModelUpdate) (int64, error) {
	if err := d.InitializePracticeModelInAllStates(doctorID); err != nil {
		return 0, errors.Trace(err)
	}

	varArgs := dbutil.MySQLVarArgs()
	if aspmu.HasPracticeExtension != nil {
		varArgs.Append(`practice_extension`, *aspmu.HasPracticeExtension)
	}
	res, err := d.db.Exec(`UPDATE practice_model SET `+varArgs.Columns()+` WHERE doctor_id = ?`, append(varArgs.Values(), doctorID)...)
	if err != nil {
		return 0, errors.Trace(err)
	}
	aff, err := res.RowsAffected()
	return aff, errors.Trace(err)
}

func (d *dataService) DoctorIDsEligibleInState(careProvidingStateID int64) ([]int64, error) {
	rows, err := d.db.Query(
		`SELECT provider_id FROM care_provider_state_elligibility
		WHERE care_providing_state_id = ?`, careProvidingStateID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Trace(err)
		}
		ids = append(ids, id)
	}
	return ids, errors.Trace(rows.Err())
}
