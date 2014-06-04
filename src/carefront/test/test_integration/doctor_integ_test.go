package test_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/patient_file"
	"carefront/visit"
)

func TestDoctorRegistration(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	SignupRandomTestDoctor(t, testData.DataApi, testData.AuthApi)
}

func TestDoctorAuthentication(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	_, email, password := SignupRandomTestDoctor(t, testData.DataApi, testData.AuthApi)

	doctorAuthHandler := &apiservice.DoctorAuthenticationHandler{AuthApi: testData.AuthApi, DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorAuthHandler)
	defer ts.Close()
	requestBody := bytes.NewBufferString("email=")
	requestBody.WriteString(email)
	requestBody.WriteString("&password=")
	requestBody.WriteString(password)
	res, err := AuthPost(ts.URL, "application/x-www-form-urlencoded", requestBody, 0)
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

	resp, err := AuthGet(ts.URL+"?query=pro", doctor.AccountId.Int64())
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
	patientSignedupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
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
		diagnosisQuestionId = qi.Id
	}

	answerItemList, err := testData.DataApi.GetAnswerInfoForTags([]string{"a_doctor_acne_not_suitable_spruce"}, api.EN_LANGUAGE_ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	diagnosePatientHandler := visit.NewDiagnosePatientHandler(testData.DataApi, testData.AuthApi, "")
	ts := httptest.NewServer(diagnosePatientHandler)
	defer ts.Close()

	answerToQuestionItem := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem.QuestionId = diagnosisQuestionId
	answerToQuestionItem.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: answerItemList[0].PotentialAnswerId}}

	answerIntakeRequestBody.Questions = []*apiservice.AnswerToQuestionItem{answerToQuestionItem}

	requestData, err := json.Marshal(answerIntakeRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body")
	}

	resp, err := AuthPost(ts.URL, "application/json", bytes.NewBuffer(requestData), doctor.AccountId.Int64())
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
	patientSignedupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
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
	diagnosePatientHandler := visit.NewDiagnosePatientHandler(testData.DataApi, testData.AuthApi, "")
	ts := httptest.NewServer(diagnosePatientHandler)
	defer ts.Close()

	requestParams := bytes.NewBufferString("?patient_visit_id=")
	requestParams.WriteString(strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	diagnosisResponse := visit.GetDiagnosisResponse{}

	resp, err := AuthGet(ts.URL+requestParams.String(), doctor.AccountId.Int64())
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
	diagnosisQuestionId, severityQuestionId, acneTypeQuestionId := SubmitPatientVisitDiagnosis(patientVisitResponse.PatientVisitId, doctor, testData, t)

	// now, get diagnosis layout again and check to ensure that the doctor successfully diagnosed the patient with the expected answers
	resp, err = AuthGet(ts.URL+requestParams.String(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of request to get diagnosis layout after submitting diagnosis: " + err.Error())
	}

	err = json.Unmarshal(body, &diagnosisResponse)
	if err != nil {
		t.Fatal("Unable to marshal response for diagnosis of patient visit after doctor submitted diagnosis: " + err.Error())
	}

	if diagnosisResponse.DiagnosisLayout == nil || diagnosisResponse.DiagnosisLayout.PatientVisitId != patientVisitResponse.PatientVisitId {
		t.Fatal("Diagnosis response not as expected after doctor submitted diagnosis")
	}

	for _, section := range diagnosisResponse.DiagnosisLayout.InfoIntakeLayout.Sections {
		for _, question := range section.Questions {

			for _, doctorResponse := range question.Answers {
				switch doctorResponse.QuestionId.Int64() {
				case diagnosisQuestionId:
					if doctorResponse.PotentialAnswerId.Int64() != 102 {
						t.Fatalf("Doctor response to question id %d expectd to have id %d but has id %d", doctorResponse.QuestionId.Int64(), 102, doctorResponse.PotentialAnswerId.Int64())
					}
				case severityQuestionId:
					if doctorResponse.PotentialAnswerId.Int64() != 107 {
						t.Fatalf("Doctor response to question id %d expectd to have id %d but has id %d", doctorResponse.QuestionId.Int64(), 107, doctorResponse.PotentialAnswerId.Int64())
					}

				case acneTypeQuestionId:
					if doctorResponse.PotentialAnswerId.Int64() != 109 && doctorResponse.PotentialAnswerId.Int64() != 114 && doctorResponse.PotentialAnswerId.Int64() != 113 {
						t.Fatalf("Doctor response to question id %d expectd to have any of ids %s but instead has id %d", doctorResponse.QuestionId.Int64(), "(109,114,113)", doctorResponse.PotentialAnswerId.Int64())
					}

				}
			}
		}
	}

	// check if the diagnosis summary exists for the patient visit
	// at this point NO diagnosis summary should exist because the doctor has not picked a treatment plan yet.
	// given that the diagnosis summary gets associated with a treatment plan, the diagnosis summary is added only after
	// a treatment plan is picked
	diagnosisSummaryHandler := &apiservice.DiagnosisSummaryHandler{DataApi: testData.DataApi}
	ts = httptest.NewServer(diagnosisSummaryHandler)
	defer ts.Close()
	getDiagnosisSummaryResponse := &common.DiagnosisSummary{}
	resp, err = AuthGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to get diagnosis summary for patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	} else if err := json.NewDecoder(resp.Body).Decode(&getDiagnosisSummaryResponse); err != nil {
		t.Fatal("Unable to unmarshal response into json object : " + err.Error())
	} else if getDiagnosisSummaryResponse.Summary != "" {
		t.Fatal("Expected no diagnosis summary to exist given that the treatment plan has not been picked. However, still got back diagnosis summary")
	}

	// now lets pick a tretament plan and then try to get the diagnosis summary again
	PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, nil, testData, t)
	resp, err = AuthGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to get diagnosis summary for patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	} else if err := json.NewDecoder(resp.Body).Decode(&getDiagnosisSummaryResponse); err != nil {
		t.Fatal("Unable to unmarshal response into json object : " + err.Error())
	} else if getDiagnosisSummaryResponse.Summary == "" {
		t.Fatal("Expected diagnosis summary to exist but it doesnt")
	}

	// now lets pick a different treatment plan and ensure that the diagnosis summary gets linked to this new
	// treatment plan.
	PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, nil, testData, t)
	resp, err = AuthGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to get diagnosis summary for patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	} else if err := json.NewDecoder(resp.Body).Decode(&getDiagnosisSummaryResponse); err != nil {
		t.Fatal("Unable to unmarshal response into json object : " + err.Error())
	} else if getDiagnosisSummaryResponse.Summary == "" {
		t.Fatal("Expected diagnosis summary to exist but it doesnt")
	}

	// now lets try and manually update the summary
	updatedSummary := "This is the new value that the diagnosis summary should be updated to"
	params := url.Values{}
	params.Set("patient_visit_id", strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	params.Set("summary", updatedSummary)
	resp, err = AuthPut(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to make call to update diagnosis summary %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to make successfull call to update diagnosis summary")
	}

	// lets get the diagnosis summary again to compare
	resp, err = AuthGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to get diagnosis summary for patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	} else if err := json.NewDecoder(resp.Body).Decode(&getDiagnosisSummaryResponse); err != nil {
		t.Fatal("Unable to unmarshal response into json object : " + err.Error())
	} else if getDiagnosisSummaryResponse.Summary != updatedSummary {
		t.Fatalf("Expected diagnosis summary %s instead got %s", updatedSummary, getDiagnosisSummaryResponse.Summary)
	}

	// lets attempt to diagnose the patient again
	SubmitPatientVisitDiagnosis(patientVisitResponse.PatientVisitId, doctor, testData, t)

	// now get the diagnosis summary again to ensure that it did not change
	resp, err = AuthGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to get diagnosis summary for patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	} else if err := json.NewDecoder(resp.Body).Decode(&getDiagnosisSummaryResponse); err != nil {
		t.Fatal("Unable to unmarshal response into json object : " + err.Error())
	} else if getDiagnosisSummaryResponse.Summary != updatedSummary {
		t.Fatalf("Expected diagnosis summary %s instead got %s", updatedSummary, getDiagnosisSummaryResponse.Summary)
	}

}

func TestDoctorSubmissionOfPatientVisitReview(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	patientSignedupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)

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

	// attempt to submit the patient visit review here. It should fail
	doctorSubmitPatientVisitReviewHandler := &apiservice.DoctorSubmitPatientVisitReviewHandler{DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorSubmitPatientVisitReviewHandler)
	defer ts.Close()

	resp, err := AuthPost(ts.URL, "application/x-www-form-urlencoded", bytes.NewBufferString("patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10)), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a call to submit the patient visit review : " + err.Error())
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to parse the response body for the call to submit patient visit review: " + err.Error())
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code to be %d but got %d instead. The call should have failed because the patient visit is not being REVIEWED by the doctor yet. ", http.StatusBadRequest, resp.StatusCode)
	}

	// get the doctor to start reviewing the patient visit
	doctorPatientVisitReviewHandler := patient_file.NewDoctorPatientVisitReviewHandler(testData.DataApi, nil, testData.CloudStorageService, testData.CloudStorageService)

	ts2 := httptest.NewServer(doctorPatientVisitReviewHandler)
	defer ts2.Close()

	resp, err = AuthGet(ts2.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to get the doctor to start reviewing the patient visit: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make a successful call for doctor to start reviewing patient visti", t)

	// attempt to submit the patient visit review here. It should work
	resp, err = AuthPost(ts.URL, "application/x-www-form-urlencoded", bytes.NewBufferString("patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10)), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful call to submit patient visit review")
	}

	CheckSuccessfulStatusCode(resp, "Unable to make successful call to submit patient visit review", t)

	patientVisit, err := testData.DataApi.GetPatientVisitFromId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatal("Unable to get patient visit given id: " + err.Error())
	}

	if patientVisit.Status != api.CASE_STATUS_TREATED {
		t.Fatalf("Expected the status to be %s but status is %s", api.CASE_STATUS_TREATED, patientVisit.Status)
	}
}

func TestDoctorAddingOfFollowUpForPatientVisit(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	patientVisitResponse, treatmentPlan := SignupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)

	// lets add a follow up time for 1 week from now
	doctorFollowupHandler := apiservice.NewPatientVisitFollowUpHandler(testData.DataApi)
	ts := httptest.NewServer(doctorFollowupHandler)
	defer ts.Close()

	requestBody := fmt.Sprintf("patient_visit_id=%d&follow_up_unit=week&follow_up_value=1", patientVisitResponse.PatientVisitId)
	resp, err := AuthPost(ts.URL, "application/x-www-form-urlencoded", bytes.NewBufferString(requestBody), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful call to add follow up time for patient visit: " + err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read the response body: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make successful call to add follow up for patient visit: "+string(body), t)

	// lets get the follow up time back
	followup, err := testData.DataApi.GetFollowUpTimeForTreatmentPlan(treatmentPlan.Id.Int64())
	if err != nil {
		t.Fatalf(err.Error())
	}

	oneWeekFromNow := time.Now().Add(7 * 24 * 60 * time.Minute)
	year, month, day := oneWeekFromNow.Date()
	year1, month1, day1 := followup.FollowUpTime.Date()

	if year != year1 || month1 != month || math.Abs(float64(day1-day)) > 2 {
		t.Fatalf("Expected date to follow up time returned to be around %d/%d/%d, but got %d/%d/%d instead", year, month, day, year1, month1, day1)
	}
}
