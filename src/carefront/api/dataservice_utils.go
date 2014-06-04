package api

import (
	"carefront/common"
	"carefront/encoding"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	STATUS_ACTIVE                      = "ACTIVE"
	STATUS_CREATED                     = "CREATED"
	STATUS_CREATING                    = "CREATING"
	STATUS_DELETING                    = "DELETING"
	STATUS_UPDATING                    = "UPDATING"
	STATUS_DELETED                     = "DELETED"
	STATUS_INACTIVE                    = "INACTIVE"
	STATUS_PENDING                     = "PENDING"
	STATUS_ONGOING                     = "ONGOING"
	STATUS_UNCOMMITTED                 = "UNCOMMITTED"
	STATUS_COMMITTED                   = "COMMITTED"
	STATUS_DRAFT                       = "DRAFT"
	ERX_STATUS_SENDING                 = "Sending"
	ERX_STATUS_SENT                    = "eRxSent"
	ERX_STATUS_ERROR                   = "Error"
	ERX_STATUS_SEND_ERROR              = "Send_Error"
	ERX_STATUS_DELETED                 = "Deleted"
	ERX_STATUS_RESOLVED                = "Resolved"
	ERX_STATUS_NEW_RX_FROM_DNTF        = "NewRxFromDNTF"
	treatmentOTC                       = "OTC"
	treatmentRX                        = "RX"
	RX_REFILL_STATUS_SENT              = "RefillRxSent"
	RX_REFILL_STATUS_DELETED           = "RefillRxDeleted"
	RX_REFILL_STATUS_ERROR             = "RefillRxError"
	RX_REFILL_STATUS_ERROR_RESOLVED    = "RefillRxErrorResolved"
	RX_REFILL_STATUS_REQUESTED         = "RefillRxRequested"
	RX_REFILL_STATUS_APPROVED          = "RefillRxApproved"
	RX_REFILL_STATUS_DENIED            = "RefillRxDenied"
	RX_REFILL_DNTF_REASON_CODE         = "DeniedNewRx"
	drDrugSupplementalInstructionTable = "dr_drug_supplemental_instruction"
	drRegimenStepTable                 = "dr_regimen_step"
	drAdvicePointTable                 = "dr_advice_point"
	drugNameTable                      = "drug_name"
	drugFormTable                      = "drug_form"
	drugRouteTable                     = "drug_route"
	doctorPhoneType                    = "MAIN"
	treatmentTable                     = "treatment"
	pharmacyDispensedTreatmentTable    = "pharmacy_dispensed_treatment"
	requestedTreatmentTable            = "requested_treatment"
	unlinkedDntfTreatmentTable         = "unlinked_dntf_treatment"
	addressUsa                         = "USA"
	PENDING_TASK_PATIENT_CARD          = "PATIENT_CARD"
)

type DataService struct {
	db              *sql.DB
	roleTypeMapping map[string]int64
}

func NewDataService(DB *sql.DB) (*DataService, error) {
	dataService := &DataService{db: DB}
	dataService.roleTypeMapping = make(map[string]int64)

	// get the role type mapping into memory for quick access
	rows, err := dataService.db.Query(`select id, role_type_tag from role_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var roleTypeTag string
		if err := rows.Scan(&id, &roleTypeTag); err != nil {
			return nil, err
		}
		dataService.roleTypeMapping[roleTypeTag] = id
	}

	return dataService, rows.Err()
}

func infoIdsFromMap(m map[int64]*common.AnswerIntake) []int64 {
	infoIds := make([]int64, 0, len(m))
	for key := range m {
		infoIds = append(infoIds, key)
	}
	return infoIds
}

func createKeysArrayFromMap(m map[int64]bool) []int64 {
	keys := make([]int64, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func createValuesArrayFromMap(m map[int64]int64) []int64 {
	values := make([]int64, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func enumerateItemsIntoString(ids []int64) string {
	if ids == nil || len(ids) == 0 {
		return ""
	}
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(idsStr, ",")
}

func getKeysAndValuesFromMap(m map[string]interface{}) ([]string, []interface{}) {
	values := make([]interface{}, 0)
	keys := make([]string, 0)
	for key, value := range m {
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

func nReplacements(n int) string {
	if n == 0 {
		return ""
	}

	result := make([]byte, 2*n-1)
	for i := 0; i < len(result)-1; i += 2 {
		result[i] = '?'
		result[i+1] = ','
	}
	result[len(result)-1] = '?'
	return string(result)
}

func appendStringsToInterfaceSlice(interfaceSlice []interface{}, strSlice []string) []interface{} {
	for _, strItem := range strSlice {
		interfaceSlice = append(interfaceSlice, strItem)
	}
	return interfaceSlice
}

func appendInt64sToInterfaceSlice(interfaceSlice []interface{}, int64Slice []int64) []interface{} {
	for _, int64Item := range int64Slice {
		interfaceSlice = append(interfaceSlice, int64Item)
	}
	return interfaceSlice
}

type treatmentType int64

const (
	treatmentForPatientType treatmentType = iota
	pharmacyDispensedTreatmentType
	refillRequestTreatmentType
	unlinkedDNTFTreatmentType
	doctorFavoriteTreatmentType
)

var possibleTreatmentTables = map[treatmentType]string{
	treatmentForPatientType:        "treatment",
	pharmacyDispensedTreatmentType: "pharmacy_dispensed_treatment",
	refillRequestTreatmentType:     "requested_treatment",
	unlinkedDNTFTreatmentType:      "unlinked_dntf_treatment",
	doctorFavoriteTreatmentType:    "dr_favorite_treatment",
}

func (d *DataService) addTreatment(tType treatmentType, treatment *common.Treatment, params map[string]interface{}, tx *sql.Tx) error {
	medicationType := treatmentRX
	if treatment.OTC {
		medicationType = treatmentOTC
	}

	// collecting data for fields that are common to all types of treatment
	columnsAndData := map[string]interface{}{
		"drug_internal_name":    treatment.DrugInternalName,
		"dosage_strength":       treatment.DosageStrength,
		"type":                  medicationType,
		"dispense_value":        treatment.DispenseValue.Float64(),
		"refills":               treatment.NumberRefills.Int64Value,
		"substitutions_allowed": treatment.SubstitutionsAllowed,
		"patient_instructions":  treatment.PatientInstructions,
		"pharmacy_notes":        treatment.PharmacyNotes,
		"status":                treatment.Status,
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

	// add any treatment type specific information to the table
	switch tType {
	case treatmentForPatientType:
		columnsAndData["status"] = STATUS_CREATED
		columnsAndData["dispense_unit_id"] = treatment.DispenseUnitId.Int64()
		if treatment.TreatmentPlanId.Int64() != 0 {
			columnsAndData["treatment_plan_id"] = treatment.TreatmentPlanId.Int64()
		}
	case doctorFavoriteTreatmentType:
		columnsAndData["status"] = STATUS_ACTIVE
		columnsAndData["dispense_unit_id"] = treatment.DispenseUnitId.Int64()
		drFavoriteTreatmentId, ok := params["dr_favorite_treatment_plan_id"]
		if !ok {
			return errors.New("Expected dr_favorite_treatment_planid to be present in the params but it wasnt")
		}
		columnsAndData["dr_favorite_treatment_plan_id"] = drFavoriteTreatmentId
	case pharmacyDispensedTreatmentType:
		columnsAndData["doctor_id"] = treatment.Doctor.DoctorId.Int64()
		columnsAndData["erx_id"] = treatment.ERx.PrescriptionId.Int64()
		columnsAndData["erx_sent_date"] = treatment.ERx.ErxSentDate
		columnsAndData["erx_last_filled_date"] = treatment.ERx.ErxLastDateFilled
		columnsAndData["pharmacy_id"] = treatment.ERx.PharmacyLocalId.Int64()
		columnsAndData["dispense_unit"] = treatment.DispenseUnitDescription
		requestedTreatment, ok := params["requested_treatment"].(*common.Treatment)
		if !ok {
			return errors.New("Expected requested_treatment to be present in the params for adding a pharmacy_dispensed_treatment")
		}
		columnsAndData["requested_treatment_id"] = requestedTreatment.Id.Int64()

	case refillRequestTreatmentType:
		columnsAndData["doctor_id"] = treatment.Doctor.DoctorId.Int64()
		columnsAndData["erx_id"] = treatment.ERx.PrescriptionId.Int64()
		columnsAndData["erx_sent_date"] = treatment.ERx.ErxSentDate
		columnsAndData["erx_last_filled_date"] = treatment.ERx.ErxLastDateFilled
		columnsAndData["pharmacy_id"] = treatment.ERx.PharmacyLocalId.Int64()
		columnsAndData["dispense_unit"] = treatment.DispenseUnitDescription
		if treatment.OriginatingTreatmentId != 0 {
			columnsAndData["originating_treatment_id"] = treatment.OriginatingTreatmentId
		}

	case unlinkedDNTFTreatmentType:
		columnsAndData["doctor_id"] = treatment.DoctorId.Int64()
		columnsAndData["patient_id"] = treatment.PatientId.Int64()
		columnsAndData["dispense_unit_id"] = treatment.DispenseUnitId.Int64()

	default:
		return errors.New("Unexpected type of treatment trying to be added to a table")
	}

	columns, values := getKeysAndValuesFromMap(columnsAndData)
	res, err := tx.Exec(fmt.Sprintf(`insert into %s (%s) values (%s)`, possibleTreatmentTables[tType], strings.Join(columns, ","), nReplacements(len(values))), values...)
	if err != nil {
		return err
	}

	treatmentId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// update the treatment object with the information
	treatment.Id = encoding.NewObjectId(treatmentId)

	// add drug db ids to the table
	for drugDbTag, drugDbId := range treatment.DrugDBIds {
		_, err := tx.Exec(fmt.Sprintf(`insert into %s_drug_db_id (drug_db_id_tag, drug_db_id, %s_id) values (?, ?, ?)`, possibleTreatmentTables[tType], possibleTreatmentTables[tType]), drugDbTag, drugDbId, treatmentId)
		if err != nil {
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
