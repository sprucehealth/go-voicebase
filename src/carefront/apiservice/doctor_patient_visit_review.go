package apiservice

import (
	"carefront/api"
	"carefront/common"
	"carefront/info_intake"
	"carefront/libs/pharmacy"
	"encoding/json"
	"net/http"
	"strings"

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

func NewDoctorPatientVisitReviewHandler(dataApi api.DataAPI, layoutStorageService api.CloudStorageAPI, patientPhotoStorageService api.CloudStorageAPI) *DoctorPatientVisitReviewHandler {
	return &DoctorPatientVisitReviewHandler{DataApi: dataApi, LayoutStorageService: layoutStorageService, PatientPhotoStorageService: patientPhotoStorageService}
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

	patient, err := p.DataApi.GetPatientFromId(patientVisit.PatientId.Int64())
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patient info based on account id: "+err.Error())
		return
	}

	bucket, key, region, _, err := p.DataApi.GetStorageInfoOfCurrentActiveDoctorLayout(patientVisit.HealthConditionId.Int64())
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the active layout version for the doctor's view of the patient visit: "+err.Error())
		return
	}

	data, _, err := p.LayoutStorageService.GetObjectAtLocation(bucket, key, region)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get doctor layout for patient visit from s3: "+err.Error())
		return
	}

	patientVisitOverview := &info_intake.PatientVisitOverview{}
	err = json.Unmarshal(data, patientVisitOverview)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to parse doctor layout for patient visit from s3: "+err.Error())
		return
	}

	fillInPatientVisitInfoIntoOverview(patientVisit, patientVisitOverview)
	patientVisitOverview.Patient = patient

	// capitalizing the gender for display purposes. TODO Make how we do this better for v1
	patientVisitOverview.Patient.Gender = strings.Title(patient.Gender)

	p.filterOutGenderSpecificQuestionsAndSubSectionsFromOverview(patientVisitOverview, patient)

	questionIds := getQuestionIdsFromPatientVisitOverview(patientVisitOverview)
	patientAnswersForQuestions, err := p.DataApi.GetAnswersForQuestionsInPatientVisit(api.PATIENT_ROLE, questionIds, patientVisit.PatientId.Int64(), patientVisit.PatientVisitId.Int64())
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get patient answers for questions : "+err.Error())
		return
	}
	p.populatePatientVisitOverviewWithPatientAnswers(patientAnswersForQuestions, patientVisitOverview, patient)
	p.removeQuestionsWithNoAnswersBasedOnFlag(patientVisitOverview, patient)
	WriteJSONToHTTPResponseWriter(w, http.StatusOK, DoctorPatientVisitReviewResponse{
		DoctorLayout:    patientVisitOverview,
		TreatmentPlanId: treatmentPlanId})
}

func (p *DoctorPatientVisitReviewHandler) populatePatientVisitOverviewWithPatientAnswers(patientAnswers map[int64][]*common.AnswerIntake,
	patientVisitOverview *info_intake.PatientVisitOverview,
	patient *common.Patient) {
	// collect all question ids for which to get patient answers
	for _, section := range patientVisitOverview.Sections {
		for _, subSection := range section.SubSections {
			for _, question := range subSection.Questions {
				if question.QuestionId != 0 {
					if patientAnswers[question.QuestionId] != nil {
						question.Answers = patientAnswers[question.QuestionId]
						GetSignedUrlsForAnswersInQuestion(&question.Question, p.PatientPhotoStorageService)
					}
				}
			}
		}
	}
	return
}

func (p *DoctorPatientVisitReviewHandler) filterOutGenderSpecificQuestionsAndSubSectionsFromOverview(patientVisitOverview *info_intake.PatientVisitOverview, patient *common.Patient) {
	for _, section := range patientVisitOverview.Sections {
		filteredSubSections := make([]*info_intake.PatientVisitOverviewSubSection, 0)
		for _, subSection := range section.SubSections {
			if !(subSection.GenderFilter == "" || subSection.GenderFilter == patient.Gender) {
				continue
			}
			filteredQuestions := make([]*info_intake.PatientVisitOverviewQuestion, 0)
			for _, question := range subSection.Questions {
				if question.GenderFilter == "" || question.GenderFilter == patient.Gender {
					filteredQuestions = append(filteredQuestions, question)
				}
			}
			subSection.Questions = filteredQuestions
			filteredSubSections = append(filteredSubSections, subSection)
		}
		section.SubSections = filteredSubSections
	}
}

func (p *DoctorPatientVisitReviewHandler) removeQuestionsWithNoAnswersBasedOnFlag(patientVisitOverview *info_intake.PatientVisitOverview, patient *common.Patient) {
	for _, section := range patientVisitOverview.Sections {
		for _, subSection := range section.SubSections {
			filteredQuestions := make([]*info_intake.PatientVisitOverviewQuestion, 0)
			for _, question := range subSection.Questions {
				if question.RemoveQuestionIfNoAnswer == true {
					if question.Answers != nil && len(question.Answers) > 0 {
						filteredQuestions = append(filteredQuestions, question)
					}
				} else {
					filteredQuestions = append(filteredQuestions, question)
				}
			}
			subSection.Questions = filteredQuestions
		}
	}

}

func fillInPatientVisitInfoIntoOverview(patientVisit *common.PatientVisit, patientVisitOverview *info_intake.PatientVisitOverview) {
	patientVisitOverview.PatientVisitTime = patientVisit.SubmittedDate
	patientVisitOverview.PatientVisitId = patientVisit.PatientVisitId.Int64()
	patientVisitOverview.HealthConditionId = patientVisit.HealthConditionId.Int64()
}

func getQuestionIdsFromPatientVisitOverview(patientVisitOverview *info_intake.PatientVisitOverview) []int64 {
	// collect all question ids for which to get patient answers
	questionIds := make([]int64, 0)
	for _, section := range patientVisitOverview.Sections {
		for _, subSection := range section.SubSections {
			for _, question := range subSection.Questions {
				if question.QuestionId != 0 {
					questionIds = append(questionIds, question.QuestionId)
				}
			}
		}
	}
	return questionIds
}
