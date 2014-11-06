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

		personID, err = dataAPI.GetPersonIdByRole(api.MA_ROLE, doctorID)
		if err != nil {
			return 0, 0, err
		}

	default:
		return 0, 0, errors.New("Unknown role " + ctx.Role)
	}

	return personID, doctorID, nil
}

func CreateMessageAndAttachments(msg *common.CaseMessage, attachments []*Attachment, personID, doctorID int64, role string, dataAPI api.DataAPI) error {

	if attachments != nil {
		// Validate all attachments
		for _, att := range attachments {
			switch att.Type {
			default:
				return apiservice.NewError("Unknown attachment type "+att.Type, http.StatusBadRequest)
			case common.AttachmentTypeTreatmentPlan:
				// Make sure the treatment plan is a part of the same case
				if role != api.DOCTOR_ROLE {
					return apiservice.NewError("Only a doctor is allowed to attac a treatment plan", http.StatusBadRequest)
				}
				tp, err := dataAPI.GetAbridgedTreatmentPlan(att.ID, doctorID)
				if err != nil {
					return err
				}
				if tp.PatientCaseId.Int64() != msg.CaseID {
					return apiservice.NewError("Treatment plan does not belong to the case", http.StatusBadRequest)
				}
				if tp.DoctorId.Int64() != doctorID {
					return apiservice.NewError("Treatment plan not created by the requesting doctor", http.StatusBadRequest)
				}
			case common.AttachmentTypeVisit:
				// Make sure the visit is part of the same case
				if role != api.DOCTOR_ROLE {
					return apiservice.NewError("Only a doctor is allowed to attach a visit", http.StatusBadRequest)
				}
				visit, err := dataAPI.GetPatientVisitFromId(att.ID)
				if err != nil {
					return err
				}
				if visit.PatientCaseId.Int64() != msg.CaseID {
					return apiservice.NewError("visit does not belong to the case", http.StatusBadRequest)
				}
			case common.AttachmentTypePhoto, common.AttachmentTypeAudio:
				// Make sure media is uploaded by the same person and is unclaimed
				media, err := dataAPI.GetMedia(att.ID)
				if err != nil {
					return err
				}
				if media.UploaderID != personID || media.ClaimerType != "" {
					return apiservice.NewError("Invalid attachment", http.StatusBadRequest)
				}
			}
			msg.Attachments = append(msg.Attachments, &common.CaseMessageAttachment{
				ItemType: att.Type,
				ItemID:   att.ID,
				Title:    att.Title,
			})
		}
	}

	msgID, err := dataAPI.CreateCaseMessage(msg)
	if err != nil {
		return err
	}

	msg.ID = msgID
	return nil

}
