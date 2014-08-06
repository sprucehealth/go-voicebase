package messages

import (
	"errors"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
)

func validateAccess(dataAPI api.DataAPI, r *http.Request, patientCase *common.PatientCase) (personID, doctorID int64, err error) {
	ctx := apiservice.GetContext(r)
	switch ctx.Role {
	case api.DOCTOR_ROLE:
		doctorID, err = dataAPI.GetDoctorIdFromAccountId(ctx.AccountId)
		if err != nil {
			return 0, 0, err
		}

		if err := apiservice.ValidateAccessToPatientCase(r.Method, ctx.Role, doctorID, patientCase.PatientId.Int64(), patientCase.Id.Int64(), dataAPI); err != nil {
			return 0, 0, err
		}

		personID, err = dataAPI.GetPersonIdByRole(api.DOCTOR_ROLE, doctorID)
		if err != nil {
			return 0, 0, err
		}
	case api.PATIENT_ROLE:
		patientID, err := dataAPI.GetPatientIdFromAccountId(ctx.AccountId)
		if err != nil {
			return 0, 0, err
		}
		if patientCase.PatientId.Int64() != patientID {
			return 0, 0, apiservice.NewValidationError("Not authorized", r)
		}
		personID, err = dataAPI.GetPersonIdByRole(api.PATIENT_ROLE, patientID)
		if err != nil {
			return 0, 0, err
		}
	case api.MA_ROLE:
		// For messaging, we let the MA POST as well as GET from the message thread given
		// they will be an active participant in the thread.
		doctorID, err = dataAPI.GetDoctorIdFromAccountId(ctx.AccountId)
		if err != nil {
			return 0, 0, err
		}

		personID, err = dataAPI.GetPersonIdByRole(api.DOCTOR_ROLE, doctorID)
		if err != nil {
			return 0, 0, err
		}

	default:
		return 0, 0, errors.New("Unknown role " + ctx.Role)
	}

	return personID, doctorID, nil
}
