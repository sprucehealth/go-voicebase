package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sprucehealth/backend/common"
)

var (
	promoCodeDoesNotExist = errors.New("Promotion code does not exist")
)

func (d *DataService) LookupPromoCode(code string) (*common.PromoCode, error) {
	var promoCode common.PromoCode
	err := d.db.QueryRow(`SELECT id, code, is_referral FROM promotion_code where code = ?`, code).Scan(&promoCode.ID, &promoCode.Code, &promoCode.IsReferral)
	if err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	return &promoCode, nil
}

func (d *DataService) PromoCodeForPatientExists(patientID, codeID int64) (bool, error) {
	var id int64
	if err := d.db.QueryRow(`SELECT promotion_code_id FROM patient_promotion
		WHERE patient_id = ? AND promotion_code_id = ? LIMIT 1`, patientID, codeID).
		Scan(&id); err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DataService) PromotionCountInGroupForPatient(patientID int64, group string) (int, error) {
	var count int
	if err := d.db.QueryRow(`
		SELECT count(*) 
		FROM patient_promotion
		INNER JOIN promotion_group on promotion_group.id = promotion_group_id
		WHERE promotion_group.name = ?
		AND patient_id = ?`, group, patientID).Scan(&count); err == sql.ErrNoRows {
		return 0, NoRowsError
	} else if err != nil {
		return 0, err
	}

	return count, nil
}

func (d *DataService) PromoCodePrefixes() ([]string, error) {
	rows, err := d.db.Query(`SELECT prefix FROM promo_code_prefix where status = 'ACTIVE'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefixes []string
	for rows.Next() {
		var prefix string
		if err := rows.Scan(&prefix); err != nil {
			return nil, err
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes, rows.Err()
}

func (d *DataService) CreatePromoCodePrefix(prefix string) error {
	_, err := d.db.Exec(`INSERT INTO promo_code_prefix (prefix, status) VALUES (?,?)`, prefix, STATUS_ACTIVE)
	return err
}

func (d *DataService) CreatePromotionGroup(promotionGroup *common.PromotionGroup) (int64, error) {
	res, err := d.db.Exec(`INSERT INTO promotion_group (name, max_allowed_promos) VALUES (?, ?)`, promotionGroup.Name, promotionGroup.MaxAllowedPromos)
	if err != nil {
		return 0, err
	}
	promotionGroup.ID, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return promotionGroup.ID, nil
}

func (d *DataService) PromotionGroup(name string) (*common.PromotionGroup, error) {
	var promotionGroup common.PromotionGroup
	if err := d.db.QueryRow(`SELECT name, max_allowed_promos FROM promotion_group WHERE name = ?`, name).
		Scan(&promotionGroup.Name, &promotionGroup.MaxAllowedPromos); err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	return &promotionGroup, nil
}

func (d *DataService) CreatePromotion(promotion *common.Promotion) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	if err := createPromotion(tx, promotion); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) Promotion(codeID int64, types map[string]reflect.Type) (*common.Promotion, error) {
	var promotion common.Promotion
	var promotionType string
	var data []byte
	err := d.db.QueryRow(`
		SELECT promotion_code.code, promo_type, promo_data, promotion_group.name, expires, created
		FROM promotion  
		INNER JOIN promotion_code on promotion_code.id = promotion_code_id
		INNER JOIN promotion_group on promotion_group.id = promotion_group_id
		WHERE promotion_code_id = ?`, codeID).Scan(
		&promotion.Code,
		&promotionType,
		&data,
		&promotion.Group,
		&promotion.Expires,
		&promotion.Created)
	if err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	promotionDataType, ok := types[promotionType]
	if !ok {
		return nil, fmt.Errorf("Unable to find promotion type: %s", promotionType)
	}

	promotion.Data = reflect.New(promotionDataType).Interface().(common.Typed)
	if err := json.Unmarshal(data, &promotion.Data); err != nil {
		return nil, err
	}

	return &promotion, nil
}

func (d *DataService) CreateReferralProgramTemplate(template *common.ReferralProgramTemplate) (int64, error) {
	jsonData, err := json.Marshal(template.Data)
	if err != nil {
		return 0, err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`UPDATE referral_program_template set status = ? where role_type_id = ?`,
		common.RSInactive.String(), d.roleTypeMapping[template.Role])
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO referral_program_template (role_type_id, referral_type, referral_data, status)
		VALUES (?,?,?,?)
		`, d.roleTypeMapping[template.Role], template.Data.TypeName(), jsonData, template.Status.String())
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	template.ID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return template.ID, tx.Commit()
}

func (d *DataService) ActiveReferralProgramTemplate(role string, types map[string]reflect.Type) (*common.ReferralProgramTemplate, error) {
	var template common.ReferralProgramTemplate
	var referralType string
	var data []byte
	err := d.db.QueryRow(`
		SELECT id, role_type_id, referral_type, referral_data, status
		FROM referral_program_template
		WHERE role_type_id = ? and status = ?`, d.roleTypeMapping[role], common.RSActive.String()).Scan(
		&template.ID,
		&template.RoleTypeID,
		&referralType,
		&data,
		&template.Status)
	if err != nil {
		return nil, err
	}

	referralDataType, ok := types[referralType]
	if !ok {
		return nil, fmt.Errorf("Unable to find referral type: %s", referralType)
	}

	template.Data = reflect.New(referralDataType).Interface().(common.Typed)
	if err := json.Unmarshal(data, &template.Data); err != nil {
		return nil, err
	}

	return &template, nil
}

func (d *DataService) ReferralProgram(codeID int64, types map[string]reflect.Type) (*common.ReferralProgram, error) {
	var referralProgram common.ReferralProgram
	var referralType string
	var referralData []byte
	if err := d.db.QueryRow(`
		SELECT referral_program_template_id, account_id, promotion_code_id, referral_type, referral_data, created, status
		FROM referral_program
		WHERE promotion_code_id = ?`, codeID).Scan(
		&referralProgram.TemplateID,
		&referralProgram.AccountID,
		&referralProgram.CodeID,
		&referralType,
		&referralData,
		&referralProgram.Created,
		&referralProgram.Status); err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	referralDataType, ok := types[referralType]
	if !ok {
		return nil, fmt.Errorf("Unable to find referral type: %s", referralType)
	}

	referralProgram.Data = reflect.New(referralDataType).Interface().(common.Typed)
	if err := json.Unmarshal(referralData, &referralProgram.Data); err != nil {
		return nil, err
	}

	return &referralProgram, nil
}

func (d *DataService) ActiveReferralProgramForAccount(accountID int64, types map[string]reflect.Type) (*common.ReferralProgram, error) {
	var referralProgram common.ReferralProgram
	var referralType string
	var referralData []byte
	if err := d.db.QueryRow(
		`SELECT referral_program_template_id, account_id, promotion_code_id, referral_type, referral_data, created, status
		FROM referral_program
		WHERE account_id = ? AND status = ?`, accountID, common.RSActive.String()).Scan(
		&referralProgram.TemplateID,
		&referralProgram.AccountID,
		&referralProgram.CodeID,
		&referralType,
		&referralData,
		&referralProgram.Created,
		&referralProgram.Status); err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	referralDataType, ok := types[referralType]
	if !ok {
		return nil, fmt.Errorf("Unable to find referral type: %s", referralType)
	}

	referralProgram.Data = reflect.New(referralDataType).Interface().(common.Typed)
	if err := json.Unmarshal(referralData, &referralProgram.Data); err != nil {
		return nil, err
	}

	return &referralProgram, nil
}

func (d *DataService) PendingPromotionsForPatient(patientID int64, types map[string]reflect.Type) ([]*common.PatientPromotion, error) {
	rows, err := d.db.Query(`
			SELECT patient_id, promotion_code_id, promotion_group_id, promo_type, promo_data, expires, created, status
			FROM patient_promotion
			WHERE patient_id = ?
			AND status = ?`, patientID, common.PSPending.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pendingPromotions []*common.PatientPromotion
	for rows.Next() {
		var promotion common.PatientPromotion
		var promotionType string
		var data []byte

		if err := rows.Scan(
			&promotion.PatientID,
			&promotion.CodeID,
			&promotion.GroupID,
			&promotionType,
			&data,
			&promotion.Expires,
			&promotion.Created,
			&promotion.Status); err != nil {
			return nil, err
		}

		promotionDataType, ok := types[promotionType]
		if !ok {
			return nil, fmt.Errorf("Unable to find promotion type: %s", promotionType)
		}

		promotion.Data = reflect.New(promotionDataType).Interface().(common.Typed)
		if err := json.Unmarshal(data, &promotion.Data); err != nil {
			return nil, err
		}

		pendingPromotions = append(pendingPromotions, &promotion)
	}

	return pendingPromotions, rows.Err()
}

func (d *DataService) CreateReferralProgram(referralProgram *common.ReferralProgram) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// make any other referral programs for this particular accountID inactive
	_, err = tx.Exec(`UPDATE referral_program SET status = ? WHERE account_id = ? and status = ? `, common.RSInactive.String(), referralProgram.AccountID, common.RSActive.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	// create the promotion code
	res, err := tx.Exec(`INSERT INTO promotion_code (code, is_referral) values (?,?)`, referralProgram.Code, true)
	if err != nil {
		tx.Rollback()
		return err
	}

	referralProgram.CodeID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	jsonData, err := json.Marshal(referralProgram.Data)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`INSERT INTO referral_program (referral_program_template_id, account_id, promotion_code_id, referral_type, referral_data, status) 
		VALUES (?,?,?,?,?,?)`, referralProgram.TemplateID, referralProgram.AccountID, referralProgram.CodeID, referralProgram.Data.TypeName(), jsonData, referralProgram.Status.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) UpdateReferralProgram(accountID int64, codeID int64, data common.Typed) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		UPDATE referral_program 
		SET referral_data = ? 
		WHERE account_id = ? and promotion_code_id = ?`, jsonData, accountID, codeID)
	if err != nil {
		return err
	}

	return nil
}

func createPromotion(tx *sql.Tx, promotion *common.Promotion) error {
	// create promotion code entry
	res, err := tx.Exec(`INSERT INTO promotion_code (code, is_referral) values (?,?)`, promotion.Code, false)
	if err != nil {
		return err
	}

	promotion.CodeID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	// get the promotionGroupID
	var promotionGroupID int64
	err = tx.QueryRow(`SELECT id from promotion_group where name = ?`, promotion.Group).Scan(&promotionGroupID)
	if err == sql.ErrNoRows {
		return errors.New("Cannot create promotion because the group does not exist")
	} else if err != nil {
		return err
	}

	// encode the data
	jsonData, err := json.Marshal(promotion.Data)
	if err != nil {
		return err
	}

	// create the promotion
	_, err = tx.Exec(`
		INSERT INTO promotion (promotion_code_id, promo_type, promo_data, promotion_group_id, expires)
		VALUES (?,?,?,?,?)`, promotion.CodeID, promotion.Data.TypeName(), jsonData, promotionGroupID, promotion.Expires)
	if err != nil {
		return err
	}

	return nil
}
func (d *DataService) CreatePatientPromotion(patientPromotion *common.PatientPromotion) error {
	// lookup code based on id

	if patientPromotion.CodeID == 0 {
		if err := d.db.QueryRow(`SELECT id from promotion_code where code = ?`, patientPromotion.Code).
			Scan(&patientPromotion.CodeID); err == sql.ErrNoRows {
			return promoCodeDoesNotExist
		} else if err != nil {
			return err
		}
	}

	if patientPromotion.GroupID == 0 {
		if err := d.db.QueryRow(`SELECT id from promotion_group where name = ?`, patientPromotion.Group).
			Scan(&patientPromotion.GroupID); err == sql.ErrNoRows {
			return errors.New("Promotion group does not exist")
		} else if err != nil {
			return err
		}
	}

	jsonData, err := json.Marshal(patientPromotion.Data)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		INSERT INTO patient_promotion (patient_id, promotion_code_id, promotion_group_id, promo_type, promo_data, expires, status)
		VALUES (?,?,?,?,?,?,?)`, patientPromotion.PatientID,
		patientPromotion.CodeID, patientPromotion.GroupID, patientPromotion.Data.TypeName(),
		jsonData, patientPromotion.Expires, patientPromotion.Status.String())

	return err
}

func (d *DataService) UpdatePatientPromotion(patientID, promoCodeID int64, update *PatientPromotionUpdate) error {
	if update == nil {
		return nil
	}

	var cols []string
	var vals []interface{}

	if update.PromotionData != nil {
		jsonData, err := json.Marshal(update.PromotionData)
		if err != nil {
			return err
		}

		cols = append(cols, "promo_data = ?")
		vals = append(vals, jsonData)
	}

	if update.Status != nil {
		cols = append(cols, "status = ?")
		vals = append(vals, update.Status.String())
	}

	if len(cols) == 0 {
		return nil
	}

	vals = append(vals, patientID, promoCodeID)

	_, err := d.db.Exec(fmt.Sprintf(
		`UPDATE patient_promotion SET %s WHERE patient_id = ? AND promotion_code_id = ?`,
		strings.Join(cols, ",")), vals...)
	return err
}

func (d *DataService) UpdateCredit(patientID int64, credit int, description string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	var patientCredit int
	if err := tx.QueryRow(`SELECT credit FROM patient_credit WHERE patient_id = ? FOR UPDATE`, patientID).
		Scan(&patientCredit); err != sql.ErrNoRows && err != nil {
		tx.Rollback()
		return err
	}

	patientCredit += credit
	if patientCredit < 0 {
		tx.Rollback()
		return errors.New("Cannot drop patient credit below 0")
	}

	// add to credit history
	res, err := tx.Exec(`
		INSERT INTO patient_credit_history (patient_id, credit, description)
		VALUES (?,?,?)`, patientID, credit, description)
	if err != nil {
		tx.Rollback()
		return err
	}

	creditHistoryID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO patient_credit (patient_id, credit, last_checked_patient_credit_history_id)
		VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE credit = ?,last_checked_patient_credit_history_id=? `, patientID, patientCredit, creditHistoryID, patientCredit, creditHistoryID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) PatientCredit(patientID int64) (*common.PatientCredit, error) {
	var patientCredit common.PatientCredit
	err := d.db.QueryRow(`SELECT patient_id, credit FROM patient_credit where patient_id = ?`, patientID).
		Scan(&patientCredit.PatientID, &patientCredit.Credit)
	if err == sql.ErrNoRows {
		return nil, NoRowsError
	}

	return &patientCredit, nil
}

func (d *DataService) PendingReferralTrackingForPatient(patientID int64) (*common.ReferralTrackingEntry, error) {
	var entry common.ReferralTrackingEntry

	if err := d.db.QueryRow(`
		SELECT promotion_code_id, claiming_patient_id, referring_account_id, created, status
		FROM patient_referral_tracking
		WHERE status = ? and claiming_patient_id = ?`, common.RTSPending.String(), patientID).Scan(
		&entry.CodeID,
		&entry.ClaimingPatientID,
		&entry.ReferringAccountID,
		&entry.Created,
		&entry.Status); err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (d *DataService) TrackPatientReferral(referralTracking *common.ReferralTrackingEntry) error {
	_, err := d.db.Exec(`
		REPLACE INTO patient_referral_tracking 
		(promotion_code_id, claiming_patient_id, referring_account_id, status) 
		VALUES (?,?,?,?)`, referralTracking.CodeID, referralTracking.ClaimingPatientID, referralTracking.ReferringAccountID, referralTracking.Status.String())
	return err
}

func (d *DataService) UpdatePatientReferral(patientID int64, status common.ReferralTrackingStatus) error {
	_, err := d.db.Exec(`UPDATE patient_referral_tracking SET status = ? WHERE claiming_patient_id = ?`, status.String(), patientID)
	return err
}

func (d *DataService) CreateParkedAccount(parkedAccount *common.ParkedAccount) (int64, error) {
	parkedAccount.Email = normalizeEmail(parkedAccount.Email)
	res, err := d.db.Exec(`INSERT INTO parked_account (email, state, promotion_code_id, patient_created) VALUES (?,?,?,?)`,
		parkedAccount.Email, parkedAccount.State, parkedAccount.CodeID, parkedAccount.PatientCreated)
	if err != nil {
		return 0, err
	}
	parkedAccount.ID, err = res.LastInsertId()
	return parkedAccount.ID, err
}

func (d *DataService) ParkedAccount(email string) (*common.ParkedAccount, error) {
	var parkedAccount common.ParkedAccount
	if err := d.db.QueryRow(`
		SELECT parked_account.id, email, state, promotion_code_id, code, is_referral, patient_created 
		FROM parked_account 
		INNER JOIN promotion_code on promotion_code.id = promotion_code_id
		WHERE email = ?`, email).Scan(
		&parkedAccount.ID,
		&parkedAccount.Email,
		&parkedAccount.State,
		&parkedAccount.CodeID,
		&parkedAccount.Code,
		&parkedAccount.IsReferral,
		&parkedAccount.PatientCreated,
	); err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}

	return &parkedAccount, nil
}

func (d *DataService) MarkParkedAccountAsPatientCreated(id int64) error {
	_, err := d.db.Exec(`UPDATE parked_account set patient_created = 1 WHERE id = ?`, id)
	return err
}
