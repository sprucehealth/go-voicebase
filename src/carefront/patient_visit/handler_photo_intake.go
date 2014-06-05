package patient_visit

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/info_intake"
	"encoding/json"
	"net/http"
)

type PhotoAnswerIntakeHandler struct {
	DataApi api.DataAPI
}

type PhotoAnswerIntakeResponse struct {
	Result string `json:"result"`
}

type PhotoAnswerIntakeQuestionItem struct {
	QuestionId    int64                        `json:"question_id,string"`
	PhotoSections []*common.PhotoIntakeSection `json:"answered_photo_sections"`
}

type PhotoAnswerIntakeRequestData struct {
	PhotoQuestions []*PhotoAnswerIntakeQuestionItem `json:"photo_questions"`
	PatientVisitId int64                            `json:"patient_visit_id,string"`
}

func NewPhotoAnswerIntakeHandler(dataApi api.DataAPI) *PhotoAnswerIntakeHandler {
	return &PhotoAnswerIntakeHandler{
		DataApi: dataApi,
	}
}

func (p *PhotoAnswerIntakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_POST {
		http.NotFound(w, r)
		return
	}

	var requestData PhotoAnswerIntakeRequestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	patientId, err := p.DataApi.GetPatientIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	patientIdFromPatientVisitId, err := p.DataApi.GetPatientIdFromPatientVisitId(requestData.PatientVisitId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	} else if patientIdFromPatientVisitId != patientId {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "patient id retrieved from the patient_visit_id does not match patient id retrieved from auth token")
		return
	}

	for _, photoIntake := range requestData.PhotoQuestions {
		// ensure that intake is for the right question type
		questionType, err := p.DataApi.GetQuestionType(photoIntake.QuestionId)
		if err != nil {
			apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
			return
		} else if questionType != info_intake.QUESTION_TYPE_PHOTO_SECTION {
			apiservice.WriteDeveloperError(w, http.StatusBadRequest, "only photo section question types acceptable for intake via this endpoint")
			return
		}

		if err := p.DataApi.StorePhotoSectionsForQuestion(photoIntake.QuestionId, patientId, requestData.PatientVisitId, photoIntake.PhotoSections); err != nil {
			apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, apiservice.SuccessfulGenericJSONResponse())
}
