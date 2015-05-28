package patient_visit

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/info_intake"
	"github.com/sprucehealth/backend/libs/httputil"
)

type photoAnswerIntakeHandler struct {
	dataAPI api.DataAPI
}

type PhotoAnswerIntakeResponse struct {
	Result string `json:"result"`
}

type PhotoAnswerIntakeQuestionItem struct {
	QuestionID    int64                        `json:"question_id,string"`
	PhotoSections []*common.PhotoIntakeSection `json:"answered_photo_sections"`
}

type PhotoAnswerIntakeRequestData struct {
	PhotoQuestions []*PhotoAnswerIntakeQuestionItem `json:"photo_questions"`
	SessionID      string                           `json:"session_id"`
	SessionCounter uint                             `json:"counter"`
	PatientVisitID int64                            `json:"patient_visit_id,string"`
}

func NewPhotoAnswerIntakeHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(apiservice.AuthorizationRequired(
		&photoAnswerIntakeHandler{
			dataAPI: dataAPI,
		}), httputil.Post)
}

func (p *photoAnswerIntakeHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)
	if ctxt.Role != api.RolePatient {
		return false, apiservice.NewAccessForbiddenError()
	}

	return true, nil
}

func (p *photoAnswerIntakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var requestData PhotoAnswerIntakeRequestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		apiservice.WriteBadRequestError(err, w, r)
		return
	}

	patientID, err := p.dataAPI.GetPatientIDFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	patientIDFromPatientVisitID, err := p.dataAPI.GetPatientIDFromPatientVisitID(requestData.PatientVisitID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	} else if patientIDFromPatientVisitID != patientID {
		apiservice.WriteValidationError("patient id retrieved from the patient_visit_id does not match patient id retrieved from auth token", w, r)
		return
	}

	for _, photoIntake := range requestData.PhotoQuestions {
		// ensure that intake is for the right question type
		questionType, err := p.dataAPI.GetQuestionType(photoIntake.QuestionID)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		} else if questionType != info_intake.QuestionTypePhotoSection {
			apiservice.WriteValidationError("only photo section question types acceptable for intake via this endpoint", w, r)
			return
		}

		// get photo slots for the question and ensure that all slot ids in the request
		// belong to this question
		photoSlots, err := p.dataAPI.GetPhotoSlotsInfo(photoIntake.QuestionID, api.LanguageIDEnglish)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		photoSlotIDMapping := make(map[int64]bool)
		for _, photoSlot := range photoSlots {
			photoSlotIDMapping[photoSlot.ID] = true
		}

		for _, photoSection := range photoIntake.PhotoSections {
			for _, photo := range photoSection.Photos {
				if !photoSlotIDMapping[photo.SlotID] {
					apiservice.WriteUserError(w, http.StatusBadRequest, fmt.Sprintf("Slot id %d not associated with photo question id %d: ", photo.SlotID, photoIntake.QuestionID))
					return
				}
			}
		}

		if err := p.dataAPI.StorePhotoSectionsForQuestion(
			photoIntake.QuestionID, patientID, requestData.PatientVisitID,
			requestData.SessionID,
			requestData.SessionCounter,
			photoIntake.PhotoSections,
		); err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	apiservice.WriteJSONSuccess(w)
}
