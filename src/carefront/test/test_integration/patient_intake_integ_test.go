package test_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"strconv"
	"testing"

	"carefront/api"
	"carefront/apiservice"
	"carefront/patient_visit"
)

type AnswerIntakeHandler struct {
	DataApi   api.DataAPI
	accountId int64
}

func getQuestionWithTagAndExpectedType(questionTag, questionType string, t *testing.T, testData TestData) int64 {
	questionInfo, err := testData.DataApi.GetQuestionInfo(questionTag, 1)
	if err != nil {
		t.Fatalf("Unable to query for question q_reason_visit from database: %s", err.Error())
	}

	// need to ensure that the question we are trying to get the information for is a single select
	// question type
	if questionInfo.QuestionType != questionType {
		t.Fatalf("Expected q_reason_visit to be '%s' instead of '%s'", questionType, questionInfo.QuestionType)
	}

	return questionInfo.QuestionId
}

func getAnswerWithTagAndExpectedType(answerTag, answerType string, questionId int64, testData TestData, t *testing.T) int64 {

	potentialAnswers, err := testData.DataApi.GetAnswerInfo(questionId, 1)
	if err != nil {
		t.Fatal("Unable to get answers for question with id " + strconv.FormatInt(questionId, 10))
	}

	expectedAnswerTag := answerTag
	var potentialAnswerId int64
	var potentialAnswerType string
	for _, potentialAnswer := range potentialAnswers {
		if potentialAnswer.AnswerTag == expectedAnswerTag {
			potentialAnswerId = potentialAnswer.AnswerId
			potentialAnswerType = potentialAnswer.AnswerType
		}
	}

	if potentialAnswerId == 0 {
		t.Fatal("Unable to find the answer for the question with intended answer tag " + expectedAnswerTag)
	}

	if potentialAnswerType != answerType {
		t.Fatalf("Potential answer found does not have matching type. Expected %s, Found %s ", answerType, potentialAnswerType)
	}

	return potentialAnswerId
}

func submitPatientAnswerForVisit(PatientId int64, testData TestData, patientIntakeRequestData string, t *testing.T) {
	answerIntakeHandler := patient_visit.NewAnswerIntakeHandler(testData.DataApi)
	patient, err := testData.DataApi.GetPatientFromId(PatientId)
	if err != nil {
		t.Fatal("Unable to get patient information given the patient id when trying to enter patient intake: " + err.Error())
	}

	ts := httptest.NewServer(answerIntakeHandler)
	defer ts.Close()

	resp, err := AuthPost(ts.URL, "application/json", bytes.NewBufferString(patientIntakeRequestData), patient.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to get the patient visit id")
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatalf("Failed to get body: %+v", err)
		resp.Body.Close()
	} else {
		CheckSuccessfulStatusCode(resp, "Unable to submit a single select answer for patient: "+string(body), t)
	}
}

func TestSingleSelectIntake(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// signup a random test patient for which to answer questions
	patientSignedUpResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)

	// now lets go ahead and try and answer the question about the reason for visit given that it is
	// single select
	questionId := getQuestionWithTagAndExpectedType("q_onset_acne", "q_type_single_select", t, testData)
	potentialAnswerId := getAnswerWithTagAndExpectedType("a_onset_six_months", "a_type_multiple_choice", questionId, testData, t)

	// lets go ahead and populate a response for the question
	patientIntakeRequestData := fmt.Sprintf(`{"patient_visit_id": "%d", "questions": [{"potential_answers": [{"potential_answer_id": "%d" } ], "question_id": "%d" }] }`, patientVisitResponse.PatientVisitId, potentialAnswerId, questionId)

	// now, lets go ahead and answer the question for the patient
	submitPatientAnswerForVisit(patientSignedUpResponse.Patient.PatientId.Int64(), testData, patientIntakeRequestData, t)

	// now, get the patient visit again to ensure that a patient answer was registered for the intended question
	patientVisitResponse = GetPatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)

	// lets go through the questions to find the one for which the patient answer should be present
	for _, section := range patientVisitResponse.ClientLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				if question.QuestionId == questionId {
					if question.Answers == nil || len(question.Answers) == 0 {
						t.Fatalf("Expected patient answer for question with id %d, but got none", questionId)
					}
					for _, answer := range GetAnswerIntakesFromAnswers(question.Answers, t) {
						if answer.PotentialAnswerId.Int64() == potentialAnswerId {
							return
						}
					}
				}
			}
		}
	}

	t.Fatalf("While a patient answer exists for question with id %d, unable to find the expected potential answer with id %d", questionId, potentialAnswerId)
}

func TestMultipleChoiceIntake(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// signup a random test patient for which to answer questions
	patientSignedUpResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)

	// now lets go ahead and try and answer the question about the reason for visit given that it is
	// single select
	questionId := getQuestionWithTagAndExpectedType("q_acne_prev_treatment_types", "q_type_multiple_choice", t, testData)
	potentialAnswers, err := testData.DataApi.GetAnswerInfo(questionId, 1)
	if err != nil {
		t.Fatal("Unable to get answers for question with id " + strconv.FormatInt(questionId, 10))
	}

	answerIntakeRequestBody := apiservice.AnswerIntakeRequestBody{}
	answerIntakeRequestBody.PatientVisitId = patientVisitResponse.PatientVisitId

	answerToQuestionItem := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem.QuestionId = questionId
	for _, potentialAnswer := range potentialAnswers {
		if potentialAnswer.AnswerTag == "a_otc_prev_treatment_type" || potentialAnswer.AnswerTag == "a_prescription_prev_treatment_type" {
			answerToQuestionItem.AnswerIntakes = append(answerToQuestionItem.AnswerIntakes, &apiservice.AnswerItem{PotentialAnswerId: potentialAnswer.AnswerId})
		}
	}
	answerIntakeRequestBody.Questions = []*apiservice.AnswerToQuestionItem{answerToQuestionItem}

	requestData, err := json.Marshal(&answerIntakeRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body")
	}
	submitPatientAnswerForVisit(patientSignedUpResponse.Patient.PatientId.Int64(), testData, string(requestData), t)
	// now, get the patient visit again to ensure that a patient answer was registered for the intended question
	patientVisitResponse = GetPatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)

	// lets go through the questions to find the one for which the patient answer should be present
	for _, section := range patientVisitResponse.ClientLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				if question.QuestionId == questionId {
					if question.Answers == nil || len(question.Answers) == 0 {
						t.Fatalf("Expected patient answer for question with id %d, but got none", questionId)
					}
					for _, answer := range GetAnswerIntakesFromAnswers(question.Answers, t) {
						answerNotFound := true
						for _, questionItem := range answerIntakeRequestBody.Questions {
							for _, answerIntake := range questionItem.AnswerIntakes {
								if answerIntake.PotentialAnswerId == answer.PotentialAnswerId.Int64() {
									answerNotFound = false
								}
							}
						}
						if answerNotFound {
							t.Fatal("Expected answer not found in patient answer for patient visit when testing for answering of multiple choice questions.")
						}
					}
				}
			}
		}
	}
}

func TestSingleEntryIntake(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// signup a random test patient for which to answer questions
	patientSignedUpResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)

	questionId := getQuestionWithTagAndExpectedType("q_other_skin_condition_entry", "q_type_single_entry", t, testData)
	potentialAnswerId := getAnswerWithTagAndExpectedType("a_other_skin_condition_entry", "a_type_single_entry", questionId, testData, t)
	answerIntakeRequestBody := apiservice.AnswerIntakeRequestBody{}
	answerIntakeRequestBody.PatientVisitId = patientVisitResponse.PatientVisitId

	answerToQuestionItem := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem.QuestionId = questionId
	answerToQuestionItem.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: potentialAnswerId, AnswerText: "testAnswer"}}
	answerIntakeRequestBody.Questions = []*apiservice.AnswerToQuestionItem{answerToQuestionItem}
	requestData, err := json.Marshal(&answerIntakeRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body")
	}
	submitPatientAnswerForVisit(patientSignedUpResponse.Patient.PatientId.Int64(), testData, string(requestData), t)
	// now, get the patient visit again to ensure that a patient answer was registered for the intended question
	patientVisitResponse = GetPatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)

	// lets go through the questions to find the one for which the patient answer should be present
	for _, section := range patientVisitResponse.ClientLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				if question.QuestionId == questionId {
					if question.Answers == nil || len(question.Answers) == 0 {
						t.Fatalf("Expected patient answer for question with id %d, but got none", questionId)
					}
					for _, answer := range GetAnswerIntakesFromAnswers(question.Answers, t) {
						if answer.PotentialAnswerId.Int64() == potentialAnswerId && answer.AnswerText == "testAnswer" {
							return
						}
					}
				}
			}
		}
	}
	t.Fatalf("While an answer for the expected question exists, unable to find the expected answer with id %d for single entry intake test", potentialAnswerId)
}

func submitFreeTextResponseForPatient(patientVisitResponse *patient_visit.PatientVisitResponse, PatientId int64, freeTextResponse string, testData TestData, t *testing.T) {
	// now lets go ahead and try and answer the question about the reason for visit given that it is
	// single select
	questionId := getQuestionWithTagAndExpectedType("q_changes_acne_worse", "q_type_free_text", t, testData)
	answerIntakeRequestBody := apiservice.AnswerIntakeRequestBody{}
	answerIntakeRequestBody.PatientVisitId = patientVisitResponse.PatientVisitId

	answerToQuestionItem := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem.QuestionId = questionId
	answerToQuestionItem.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{AnswerText: freeTextResponse}}

	answerIntakeRequestBody.Questions = []*apiservice.AnswerToQuestionItem{answerToQuestionItem}
	requestData, err := json.Marshal(&answerIntakeRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body")
	}
	submitPatientAnswerForVisit(PatientId, testData, string(requestData), t)
	// now, get the patient visit again to ensure that a patient answer was registered for the intended question
	patientVisitResponse = GetPatientVisitForPatient(PatientId, testData, t)

	// lets go through the questions to find the one for which the patient answer should be present
	for _, section := range patientVisitResponse.ClientLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				if question.QuestionId == questionId {
					if question.Answers == nil || len(question.Answers) == 0 {
						t.Fatalf("Expected patient answer for question with id %d, but got none", questionId)
					}
					for _, answer := range GetAnswerIntakesFromAnswers(question.Answers, t) {
						if answer.AnswerText == freeTextResponse {
							return
						}
					}
				}
			}
		}
	}

	t.Fatalf("While an answer for the expected question exists, unable to find the expected answer with free text %s for free text intake test", freeTextResponse)
}

func TestFreeTextEntryIntake(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// signup a random test patient for which to answer questions
	patientSignedUpResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedUpResponse.Patient.PatientId.Int64(), testData, t)
	freeTextResponse := "This is a free text response that should be accepted as a response for free text."
	submitFreeTextResponseForPatient(patientVisitResponse, patientSignedUpResponse.Patient.PatientId.Int64(), freeTextResponse, testData, t)

	// submit another free text response to update teh response to this questiuon to ensure that what is returned is this response
	// for this questions
	updatedFreeTextResponse := "This is an updated free text response"
	submitFreeTextResponseForPatient(patientVisitResponse, patientSignedUpResponse.Patient.PatientId.Int64(), updatedFreeTextResponse, testData, t)
}

func addSubAnswerToAnswerIntake(answerIntake *apiservice.AnswerItem, subAnswerQuestionId, subAnswerPotentialAnswerId int64) {
	subQuestionAnswerIntake := &apiservice.SubQuestionAnswerIntake{}
	subQuestionAnswerIntake.QuestionId = subAnswerQuestionId
	subQuestionAnswerIntake.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: subAnswerPotentialAnswerId}}
	if answerIntake.SubQuestionAnswerIntakes == nil {
		answerIntake.SubQuestionAnswerIntakes = make([]*apiservice.SubQuestionAnswerIntake, 0)
	}
	answerIntake.SubQuestionAnswerIntakes = append(answerIntake.SubQuestionAnswerIntakes, subQuestionAnswerIntake)
}
