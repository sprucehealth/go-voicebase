package apiservice

import (
	"carefront/api"
	"carefront/common"
	"carefront/info_intake"
	"carefront/libs/pharmacy"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SpruceHealth/mapstructure"
	"github.com/gorilla/schema"
)

type DoctorPatientVisitReviewHandler struct {
	DataApi                    api.DataAPI
	PharmacySearchService      pharmacy.PharmacySearchAPI
	LayoutStorageService       api.CloudStorageAPI
	PatientPhotoStorageService api.CloudStorageAPI
}

type DoctorPatientVisitReviewRequestBody struct {
	PatientVisitId  int64 `schema:"patient_visit_id"`
	TreatmentPlanId int64 `schema:"treatment_plan_id"`
}

type DoctorPatientVisitReviewResponse struct {
	DoctorLayout    *info_intake.PatientVisitOverview `json:"patient_visit_overview,omitempty"`
	TreatmentPlanId int64                             `json:"treatment_plan_id"`
}

func (p *DoctorPatientVisitReviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != HTTP_GET {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	var requestData DoctorPatientVisitReviewRequestBody
	if err := schema.NewDecoder().Decode(&requestData, r.Form); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	patientVisitId := requestData.PatientVisitId
	treatmentPlanId := requestData.TreatmentPlanId
	if err := ensureTreatmentPlanOrPatientVisitIdPresent(p.DataApi, treatmentPlanId, &patientVisitId); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	patientVisit, err := p.DataApi.GetPatientVisitFromId(patientVisitId)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to get patient visit information from database based on provided patient visit id : "+err.Error())
		return
	}

	// ensure that the doctor is authorized to work on this case
	patientVisitReviewData, statusCode, err := ValidateDoctorAccessToPatientVisitAndGetRelevantData(patientVisit.PatientVisitId.Int64(), GetContext(r).AccountId, p.DataApi)
	if err != nil {
		WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	// udpate the status of the case and the item in the doctor's queue
	if patientVisit.Status == api.CASE_STATUS_SUBMITTED {
		treatmentPlanId, err = p.DataApi.StartNewTreatmentPlanForPatientVisit(patientVisit.PatientId.Int64(), patientVisit.PatientVisitId.Int64(), patientVisitReviewData.DoctorId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update the status of the visit to reviewing: "+err.Error())
			return
		}

		if err := p.DataApi.MarkPatientVisitAsOngoingInDoctorQueue(patientVisitReviewData.DoctorId, patientVisit.PatientVisitId.Int64()); err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update the item in the queue for the doctor that speaks to this patient visit: "+err.Error())
			return
		}

		if err := p.DataApi.RecordDoctorAssignmentToPatientVisit(patientVisit.PatientVisitId.Int64(), patientVisitReviewData.DoctorId); err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to assign the patient visit to this doctor: "+err.Error())
			return
		}
	} else {
		treatmentPlanId, err = p.DataApi.GetActiveTreatmentPlanForPatientVisit(patientVisitReviewData.DoctorId, patientVisit.PatientVisitId.Int64())
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get treatment plan id for patient visit: "+err.Error())
			return
		}
	}

	patientAnswersForQuestions, err := p.DataApi.GetAnswersForQuestionsInPatientVisit(patientVisit.PatientId.Int64(), patientVisit.PatientVisitId.Int64())
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patient answers for questions : "+err.Error())
		return
	}

	questionIds := make([]int64, len(patientAnswersForQuestions))
	var i int
	for key, _ := range patientAnswersForQuestions {
		questionIds[i] = key
		i++
	}

	questionInfos, err := p.DataApi.GetQuestionInfoForIds(questionIds, api.EN_LANGUAGE_ID)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get question info for question ids : "+err.Error())
		return
	}

	context, err := populateContextForRenderingLayout(patientAnswersForQuestions, questionInfos, p.DataApi, p.PatientPhotoStorageService)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to populate context for rendering layout: "+err.Error())
		return
	}

	// TODO get the appropriate template to render here
	fileContents, err := ioutil.ReadFile("../carefront/api-response-examples/v1/doctor/visit/review_v2_template.json")
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to open file to render the template: "+err.Error())
		return
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(fileContents, &jsonData)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unbale to unmarshal file contents into map[string]interface{}: "+err.Error())
	}

	sectionList := &DVisitReviewSectionListView{}
	decoderConfig := &mapstructure.DecoderConfig{
		Result:  sectionList,
		TagName: "json",
	}
	decoderConfig.SetRegistry(dVisitReviewViewTypeRegistry.Map())

	d, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to create new decoder: "+err.Error())
		return
	}

	err = d.Decode(jsonData)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to parse template into structure: "+err.Error())
		return
	}

	renderedJsonData, err := sectionList.Render(context)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to render template into expected view layout for doctor visit review: "+err.Error())
		return
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, renderedJsonData)
}

func populateContextForRenderingLayout(patientAnswersForQuestions map[int64][]*common.AnswerIntake, questionInfos []*common.QuestionInfo, dataApi api.DataAPI, photoStorageService api.CloudStorageAPI) (common.ViewContext, error) {
	context := new(common.ViewContext)
	// go through each question
	for _, questionInfo := range questionInfos {
		switch questionInfo.Type {

		case info_intake.QUESTION_TYPE_PHOTO, info_intake.QUESTION_TYPE_MULTIPLE_PHOTO, info_intake.QUESTION_TYPE_SINGLE_PHOTO:
			populatePhotos(patientAnswersForQuestions[questionInfo.Id], context, photoStorageService)

		case info_intake.QUESTION_TYPE_AUTOCOMPLETE:
			populateDataForAnswerWithSubAnswers(patientAnswersForQuestions[questionInfo.Id], questionInfo, context)

		case info_intake.QUESTION_TYPE_MULTIPLE_CHOICE, info_intake.QUESTION_TYPE_SINGLE_SELECT:
			if err := populateCheckedUncheckedData(patientAnswersForQuestions[questionInfo.Id], questionInfo, context, dataApi); err != nil {
				return nil, err
			}

		case info_intake.QUESTION_TYPE_SINGLE_ENTRY, info_intake.QUESTION_TYPE_FREE_TEXT:
			if err := populateDataForSingleEntryAnswers(patientAnswersForQuestions[questionInfo.Id], questionInfo, context); err != nil {
				return nil, err
			}
		}
	}

	return *context, nil
}

func populateCheckedUncheckedData(patientAnswers []*common.AnswerIntake, questionInfo *common.QuestionInfo, context *common.ViewContext, dataApi api.DataAPI) error {
	answerInfos, err := dataApi.GetAnswerInfo(questionInfo.Id, api.EN_LANGUAGE_ID)
	if err != nil {
		return err
	}

	checkedUncheckedItems := make([]CheckedUncheckedData, len(answerInfos))
	for i, answerInfo := range answerInfos {
		answerSelected := false

		for _, patientAnswer := range patientAnswers {
			if patientAnswer.PotentialAnswerId.Int64() == answerInfo.PotentialAnswerId {
				answerSelected = true
			}
		}

		checkedUncheckedItems[i] = CheckedUncheckedData{
			Value:     answerInfo.Answer,
			IsChecked: answerSelected,
		}
	}

	context.Set(fmt.Sprintf("%s:question_summary", questionInfo.QuestionTag), questionInfo.Summary)
	context.Set(fmt.Sprintf("%s:answers", questionInfo.QuestionTag), checkedUncheckedItems)
	return nil
}

func populatePhotos(patientAnswers []*common.AnswerIntake, context *common.ViewContext, photoStorageService api.CloudStorageAPI) {
	var photos []PhotoData
	photoData, ok := context.Get("patient_visit_photos")

	if !ok || photoData == nil {
		photos = make([]PhotoData, 0)
	} else {
		photos = photoData.([]PhotoData)
	}

	for _, answerIntake := range patientAnswers {
		photos = append(photos, PhotoData{
			Title:    answerIntake.PotentialAnswer,
			PhotoUrl: GetSignedUrlForAnswer(answerIntake, photoStorageService),
		})
	}

	context.Set("patient_visit_photos", photos)
}

func populateDataForSingleEntryAnswers(patientAnswers []*common.AnswerIntake, questionInfo *common.QuestionInfo, context *common.ViewContext) error {
	if len(patientAnswers) == 0 {
		return nil
	}

	if len(patientAnswers) > 1 {
		return fmt.Errorf("Expected just one answer for question %s instead we have  %d", questionInfo.QuestionTag, len(patientAnswers))
	}

	answer := patientAnswers[0].AnswerText
	if answer == "" {
		answer = patientAnswers[0].AnswerSummary
	}
	if answer == "" {
		answer = patientAnswers[0].PotentialAnswer
	}

	context.Set(fmt.Sprintf("%s:question_summary", questionInfo.QuestionTag), questionInfo.Summary)
	context.Set(fmt.Sprintf("%s:answer", questionInfo.QuestionTag), answer)
	return nil
}

func populateDataForAnswerWithSubAnswers(patientAnswers []*common.AnswerIntake, questionInfo *common.QuestionInfo, context *common.ViewContext) {
	data := make([]TitleSubtitleSubItemsData, len(patientAnswers))
	for _, patientAnswer := range patientAnswers {

		items := make([]string, len(patientAnswer.SubAnswers))
		for i, subAnswer := range patientAnswer.SubAnswers {
			if subAnswer.AnswerSummary != "" {
				items[i] = subAnswer.AnswerSummary
			} else {
				items[i] = subAnswer.PotentialAnswer
			}
		}

		data = append(data, TitleSubtitleSubItemsData{
			Title:    patientAnswer.AnswerText,
			SubItems: items,
		})
	}
	context.Set(fmt.Sprintf("%s:question_summary", questionInfo.QuestionTag), questionInfo.Summary)
	context.Set(fmt.Sprintf("%s:answers", questionInfo.QuestionTag), data)
}
