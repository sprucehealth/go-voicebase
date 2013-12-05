package apiservice

import (
	"encoding/json"
	"net/http"

	"carefront/api"
	"carefront/common"
	"carefront/info_intake"
	thriftapi "carefront/thrift/api"
)

const (
	HEALTH_CONDITION_ACNE_ID = 1
)

type PatientVisitHandler struct {
	DataApi                    api.DataAPI
	AuthApi                    thriftapi.Auth
	LayoutStorageService       api.CloudStorageAPI
	PatientPhotoStorageService api.CloudStorageAPI
	accountId                  int64
}

type PatientVisitResponse struct {
	PatientVisitId int64                        `json:"patient_visit_id,string"`
	ClientLayout   *info_intake.HealthCondition `json:"health_condition,omitempty"`
}

func NewPatientVisitHandler(dataApi api.DataAPI, authApi thriftapi.Auth, layoutStorageService api.CloudStorageAPI, patientPhotoStorageService api.CloudStorageAPI) *PatientVisitHandler {
	return &PatientVisitHandler{dataApi, authApi, layoutStorageService, patientPhotoStorageService, 0}
}

func (s *PatientVisitHandler) AccountIdFromAuthToken(accountId int64) {
	s.accountId = accountId
}

func (s *PatientVisitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.returnNewOrOpenPatientVisit(w, r)
	}
}

func (s *PatientVisitHandler) returnNewOrOpenPatientVisit(w http.ResponseWriter, r *http.Request) {

	patientId, err := s.DataApi.GetPatientIdFromAccountId(s.accountId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patientId from the accountId retreived from the auth token: "+err.Error())
		return
	}

	isNewPatientVisit := false
	var healthCondition *info_intake.HealthCondition
	var layoutVersionId int64
	// check if there is an open patient visit for the given health condition and return
	// that to the patient
	patientVisitId, err := s.DataApi.GetActivePatientVisitIdForHealthCondition(patientId, HEALTH_CONDITION_ACNE_ID)
	if err == api.NoRowsError {
		isNewPatientVisit = true
		// if there isn't one, then pick the current active condition layout to send to the client for the patient to enter information
		healthCondition, layoutVersionId, err = s.getCurrentActiveClientLayoutForHealthCondition(HEALTH_CONDITION_ACNE_ID, api.EN_LANGUAGE_ID)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get current active client digestable layout: "+err.Error())
			return
		}
		patientVisitId, err = s.DataApi.CreateNewPatientVisit(patientId, HEALTH_CONDITION_ACNE_ID, layoutVersionId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to create new patient visit id: "+err.Error())
			return
		}
	} else if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, `unable to retrieve the current active patient 
			visit for the health condition from the patient id: `+err.Error())
		return
	} else {
		// if there is an active patient visit record, then ensure to lookup the layout to send to the patient
		// based on what layout was shown to the patient at the time of opening of the patient visit, NOT the current
		// based on what is the current active layout because that may have potentially changed and we want to ensure
		// to not confuse the patient by changing the question structure under their feet for this particular patient visit
		// in other words, want to show them what they have already seen in terms of a flow.
		healthCondition, layoutVersionId, err = s.getClientLayoutForPatientVisit(patientVisitId, api.EN_LANGUAGE_ID)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get client layout for existing patient visit: "+err.Error())
			return
		}
	}

	// identify sections that are global
	globalSectionIds, err := s.DataApi.GetGlobalSectionIds()
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get global section ids: "+err.Error())
		return
	}

	globalQuestionIds := make([]int64, 0)
	for _, sectionId := range globalSectionIds {
		questionIds := getQuestionIdsInSectionInHealthConditionLayout(healthCondition, sectionId)
		globalQuestionIds = append(globalQuestionIds, questionIds...)
	}

	// get the answers that the patient has previously entered for all sections that are considered global
	globalSectionPatientAnswers, err := s.DataApi.GetPatientAnswersForQuestionsInGlobalSections(globalQuestionIds, patientId)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patient answers for global sections: "+err.Error())
		return
	}
	s.populateHealthConditionWithPatientAnswers(healthCondition, globalSectionPatientAnswers)

	if !isNewPatientVisit {
		// get answers that the patient has previously entered for this particular patient visit
		// and feed the answers into the layout
		sectionIdsForHealthCondition, err := s.DataApi.GetSectionIdsForHealthCondition(HEALTH_CONDITION_ACNE_ID)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get section ids for health condition: "+err.Error())
			return
		}
		questionIdsInAllSections := make([]int64, 0)
		for _, sectionId := range sectionIdsForHealthCondition {
			questionIds := getQuestionIdsInSectionInHealthConditionLayout(healthCondition, sectionId)
			questionIdsInAllSections = append(questionIdsInAllSections, questionIds...)
		}
		patientAnswersForVisit, err := s.DataApi.GetPatientAnswersForQuestionsInPatientVisit(questionIdsInAllSections, patientId, patientVisitId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patient answers for patient visit: "+err.Error())
			return
		}
		s.populateHealthConditionWithPatientAnswers(healthCondition, patientAnswersForVisit)
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, PatientVisitResponse{patientVisitId, healthCondition})
}

func getQuestionIdsInSectionInHealthConditionLayout(healthCondition *info_intake.HealthCondition, sectionId int64) (questionIds []int64) {
	questionIds = make([]int64, 0)
	for _, section := range healthCondition.Sections {
		if section.SectionId == sectionId {
			for _, screen := range section.Screens {
				for _, question := range screen.Questions {
					questionIds = append(questionIds, question.QuestionId)
				}
			}
		}
	}
	return
}

func (s *PatientVisitHandler) populateHealthConditionWithPatientAnswers(healthCondition *info_intake.HealthCondition, patientAnswers map[int64][]*common.PatientAnswer) {
	for _, section := range healthCondition.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				// go through each question to see if there exists a patient answer for it
				if patientAnswers[question.QuestionId] != nil {
					question.PatientAnswers = patientAnswers[question.QuestionId]
					GetSignedUrlsForAnswersInQuestion(question, s.PatientPhotoStorageService)
				}
			}
		}
	}
}

func (s *PatientVisitHandler) getCurrentActiveClientLayoutForHealthCondition(healthConditionId, languageId int64) (healthCondition *info_intake.HealthCondition, layoutVersionId int64, err error) {
	bucket, key, region, layoutVersionId, err := s.DataApi.GetStorageInfoOfCurrentActivePatientLayout(languageId, healthConditionId)
	if err != nil {
		return
	}

	healthCondition, err = s.getHealthConditionObjectAtLocation(bucket, key, region)
	return
}

func (s *PatientVisitHandler) getClientLayoutForPatientVisit(patientVisitId, languageId int64) (healthCondition *info_intake.HealthCondition, layoutVersionId int64, err error) {
	layoutVersionId, err = s.DataApi.GetLayoutVersionIdForPatientVisit(patientVisitId)
	if err != nil {
		return
	}

	bucket, key, region, err := s.DataApi.GetStorageInfoForClientLayout(layoutVersionId, languageId)
	if err != nil {
		return
	}

	healthCondition, err = s.getHealthConditionObjectAtLocation(bucket, key, region)
	return
}

func (s *PatientVisitHandler) getHealthConditionObjectAtLocation(bucket, key, region string) (healthCondition *info_intake.HealthCondition, err error) {

	data, err := s.LayoutStorageService.GetObjectAtLocation(bucket, key, region)
	if err != nil {
		return
	}
	healthCondition = &info_intake.HealthCondition{}
	err = json.Unmarshal(data, healthCondition)
	if err != nil {
		return
	}
	return
}
