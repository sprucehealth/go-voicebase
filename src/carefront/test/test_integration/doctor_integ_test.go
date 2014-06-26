package test_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"carefront/api"
	"carefront/apiservice"
	"carefront/doctor_treatment_plan"
	"carefront/patient_visit"
)

func TestDoctorRegistration(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	SignupRandomTestDoctor(t, testData)
}

func TestDoctorAuthentication(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	_, email, password := SignupRandomTestDoctor(t, testData)

	doctorAuthHandler := &apiservice.DoctorAuthenticationHandler{AuthApi: testData.AuthApi, DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorAuthHandler)
	defer ts.Close()
	requestBody := bytes.NewBufferString("email=")
	requestBody.WriteString(email)
	requestBody.WriteString("&password=")
	requestBody.WriteString(password)
	res, err := testData.AuthPost(ts.URL, "application/x-www-form-urlencoded", requestBody, 0)
	if err != nil {
		t.Fatal("Unable to authenticate doctor " + err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Unable to read body of response: " + err.Error())
	}
	CheckSuccessfulStatusCode(res, fmt.Sprintf("Unable to make success request to authenticate doctor. Here's the code returned %d and here's the body of the request %s", res.StatusCode, body), t)

	authenticatedDoctorResponse := &apiservice.DoctorAuthenticationResponse{}
	err = json.Unmarshal(body, authenticatedDoctorResponse)
	if err != nil {
		t.Fatal("Unable to parse response from patient authenticated")
	}

	if authenticatedDoctorResponse.Token == "" || authenticatedDoctorResponse.Doctor == nil {
		t.Fatal("Doctor not authenticated as expected")
	}
}

func TestDoctorDrugSearch(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor information from id: " + err.Error())
	}

	erx := setupErxAPI(t)

	// ensure that the autcoomplete api returns results
	autocompleteHandler := &apiservice.AutocompleteHandler{DataApi: testData.DataApi, ERxApi: erx, Role: api.DOCTOR_ROLE}
	ts := httptest.NewServer(autocompleteHandler)
	defer ts.Close()

	resp, err := testData.AuthGet(ts.URL+"?query=pro", doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a successful query to the autocomplete API")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to parse the body of the response: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make a successful query to the autocomplete api for the doctor: "+string(body), t)
	autocompleteResponse := &apiservice.AutocompleteResponse{}
	err = json.Unmarshal(body, autocompleteResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal the response from the autocomplete call into a json object as expected: " + err.Error())
	}

	if autocompleteResponse.Suggestions == nil || len(autocompleteResponse.Suggestions) == 0 {
		t.Fatal("Expected suggestions from the autocomplete api but got none")
	}

	for _, suggestion := range autocompleteResponse.Suggestions {
		if suggestion.Title == "" || suggestion.Subtitle == "" || suggestion.DrugInternalName == "" {
			t.Fatalf("Suggestion structure not filled in with data as expected. %q", suggestion)
		}
	}
}

func TestDoctorDiagnosisOfPatientVisit_Unsuitable(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit but don't pick a treatment plan yet.
	patientSignedupResponse := SignupRandomTestPatient(t, testData)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), testData, t)
	patient, err := testData.DataApi.GetPatientFromId(patientSignedupResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatal("Unable to get patient from id: " + err.Error())
	}
	answerIntakeRequestBody := prepareAnswersForQuestionsInPatientVisit(patientVisitResponse, t)
	SubmitAnswersIntakeForPatient(patient.PatientId.Int64(), patient.AccountId.Int64(), answerIntakeRequestBody, testData, t)
	SubmitPatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)

	answerIntakeRequestBody = &apiservice.AnswerIntakeRequestBody{}
	answerIntakeRequestBody.PatientVisitId = patientVisitResponse.PatientVisitId

	var diagnosisQuestionId int64
	if qi, err := testData.DataApi.GetQuestionInfo("q_acne_diagnosis", 1); err != nil {
		t.Fatalf("Unable to get the questionIds for the question tags requested for the doctor to diagnose patient visit: %s", err.Error())
	} else {
		diagnosisQuestionId = qi.QuestionId
	}

	answerItemList, err := testData.DataApi.GetAnswerInfoForTags([]string{"a_doctor_acne_not_suitable_spruce"}, api.EN_LANGUAGE_ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	diagnosePatientHandler := patient_visit.NewDiagnosePatientHandler(testData.DataApi, testData.AuthApi, "")
	ts := httptest.NewServer(diagnosePatientHandler)
	defer ts.Close()

	answerToQuestionItem := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem.QuestionId = diagnosisQuestionId
	answerToQuestionItem.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: answerItemList[0].AnswerId}}

	answerIntakeRequestBody.Questions = []*apiservice.AnswerToQuestionItem{answerToQuestionItem}

	requestData, err := json.Marshal(answerIntakeRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body")
	}

	resp, err := testData.AuthPost(ts.URL, "application/json", bytes.NewBuffer(requestData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to successfully submit the diagnosis of a patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected to get a %d response but got %d", http.StatusOK, resp.StatusCode)
	}

	// the patient visit should have its state set to TRIAGED
	patientVisit, err := testData.DataApi.GetPatientVisitFromId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatal(err.Error())
	} else if patientVisit.Status != api.CASE_STATUS_TRIAGED {
		t.Fatalf("Expected status to be %s but it was %s instead", api.CASE_STATUS_TRIAGED, patientVisit.Status)
	}

	// ensure that there is no longer a pending item in the doctor queue
	pendingItems, err := testData.DataApi.GetPendingItemsInDoctorQueue(doctorId)
	if err != nil {
		t.Fatalf(err.Error())
	} else if len(pendingItems) != 0 {
		t.Fatalf("Expected no pending items instead got %d", len(pendingItems))
	}

}

func TestDoctorDiagnosisOfPatientVisit(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit but don't pick a treatment plan yet.
	patientSignedupResponse := SignupRandomTestPatient(t, testData)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), testData, t)
	patient, err := testData.DataApi.GetPatientFromId(patientSignedupResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatal("Unable to get patient from id: " + err.Error())
	}
	answerIntakeRequestBody := prepareAnswersForQuestionsInPatientVisit(patientVisitResponse, t)
	SubmitAnswersIntakeForPatient(patient.PatientId.Int64(), patient.AccountId.Int64(), answerIntakeRequestBody, testData, t)
	SubmitPatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)

	// doctor now attempts to diagnose patient visit
	diagnosePatientHandler := patient_visit.NewDiagnosePatientHandler(testData.DataApi, testData.AuthApi, "")
	ts := httptest.NewServer(diagnosePatientHandler)
	defer ts.Close()

	requestParams := bytes.NewBufferString("?patient_visit_id=")
	requestParams.WriteString(strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	diagnosisResponse := patient_visit.GetDiagnosisResponse{}

	resp, err := testData.AuthGet(ts.URL+requestParams.String(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Something went wrong when trying to get diagnoses layout for doctor to diagnose patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected response code 200 instead got %d", resp.StatusCode)
	} else if err = json.NewDecoder(resp.Body).Decode(&diagnosisResponse); err != nil {
		t.Fatal("Unable to unmarshal response for diagnosis of patient visit: " + err.Error())
	} else if diagnosisResponse.DiagnosisLayout == nil || diagnosisResponse.DiagnosisLayout.PatientVisitId != patientVisitResponse.PatientVisitId {
		t.Fatal("Diagnosis response not as expected")
	} else {
		// no doctor answers should exist yet
		for _, section := range diagnosisResponse.DiagnosisLayout.InfoIntakeLayout.Sections {
			for _, question := range section.Questions {
				if len(question.Answers) > 0 {
					t.Fatalf("Expected no answers to exist yet given that diagnosis has not taken place yet answers exist!")
				}
			}
		}
	}

	// Now, actually diagnose the patient visit and check the response to ensure that the doctor diagnosis was returned in the response
	// prepapre a response for the doctor
	SubmitPatientVisitDiagnosis(patientVisitResponse.PatientVisitId, doctor, testData, t)

	// now lets pick a tretament plan and then try to get the diagnosis summary again
	PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, nil, testData, t)

	// now lets pick a different treatment plan and ensure that the diagnosis summary gets linked to this new
	// treatment plan.
	PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, nil, testData, t)

	// lets attempt to diagnose the patient again
	SubmitPatientVisitDiagnosis(patientVisitResponse.PatientVisitId, doctor, testData, t)
}

func TestDoctorSubmissionOfPatientVisitReview(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	patientSignedupResponse := SignupRandomTestPatient(t, testData)

	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	// get patient to start a visit
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), testData, t)

	// submit answers to questions in patient visit
	patient, err := testData.DataApi.GetPatientFromId(patientSignedupResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatal("Unable to get patient from id: " + err.Error())
	}

	answerIntakeRequestBody := prepareAnswersForQuestionsInPatientVisit(patientVisitResponse, t)
	SubmitAnswersIntakeForPatient(patient.PatientId.Int64(), patient.AccountId.Int64(), answerIntakeRequestBody, testData, t)

	// get patient to submit the visit
	SubmitPatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor object from id: " + err.Error())
	}

	jsonData, err := json.Marshal(&doctor_treatment_plan.TreatmentPlanRequestData{})
	if err != nil {
		t.Fatal(err)
	}

	// attempt to submit the treatment plan here. It should fail
	doctorTreatmentPlanHandler := doctor_treatment_plan.NewDoctorTreatmentPlanHandler(testData.DataApi, nil, nil, false)
	ts := httptest.NewServer(doctorTreatmentPlanHandler)
	defer ts.Close()

	resp, err := testData.AuthPut(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a call to submit the patient visit review : " + err.Error())
	} else if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code to be %d but got %d instead. The call should have failed because the patient visit is not being REVIEWED by the doctor yet. ", http.StatusBadRequest, resp.StatusCode)
	}

	// get the doctor to start reviewing the patient visit
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)
	responseData := PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, nil, testData, t)

	caseID, err := testData.DataApi.GetPatientCaseIdFromPatientVisitId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatal(err)
	}

	// Shouldn't be any messages yet
	if msgs, err := testData.DataApi.ListCaseMessages(caseID); err != nil {
		t.Fatal(err)
	} else if len(msgs) != 0 {
		t.Fatalf("Expected no doctor message but got %d", len(msgs))
	}

	jsonData, err = json.Marshal(doctor_treatment_plan.TreatmentPlanRequestData{
		TreatmentPlanId: responseData.TreatmentPlan.Id,
		Message:         "Foo",
	})

	// attempt to submit the patient visit review here. It should work
	resp, err = testData.AuthPut(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful call to submit patient visit review")
	} else if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Expected %d but got %d: %s", http.StatusOK, resp.StatusCode, string(b))
	}

	patientVisit, err := testData.DataApi.GetPatientVisitFromId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatal("Unable to get patient visit given id: " + err.Error())
	}

	if patientVisit.Status != api.CASE_STATUS_TREATED {
		t.Fatalf("Expected the status to be %s but status is %s", api.CASE_STATUS_TREATED, patientVisit.Status)
	}

	// Shouldn't be any messages yet
	if msgs, err := testData.DataApi.ListCaseMessages(caseID); err != nil {
		t.Fatal(err)
	} else if len(msgs) != 1 {
		t.Fatalf("Expected 1 doctor message but got %d", len(msgs))
	}
}
