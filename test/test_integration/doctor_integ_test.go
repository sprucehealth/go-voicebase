package test_integration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice/router"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/doctor"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/misc/handlers"
	"github.com/sprucehealth/backend/patient_visit"
	"github.com/sprucehealth/backend/test"
)

func TestDoctorRegistration(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	SignupRandomTestDoctor(t, testData)
}

func TestDoctorAuthentication(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	_, email, password := SignupRandomTestDoctor(t, testData)

	requestBody := bytes.NewBufferString("email=")
	requestBody.WriteString(email)
	requestBody.WriteString("&password=")
	requestBody.WriteString(password)
	res, err := testData.AuthPost(testData.APIServer.URL+router.DoctorAuthenticateURLPath, "application/x-www-form-urlencoded", requestBody, 0)
	if err != nil {
		t.Fatal("Unable to authenticate doctor " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Unable to read body of response: " + err.Error())
	}
	test.Equals(t, http.StatusOK, res.StatusCode)

	authenticatedDoctorResponse := &doctor.AuthenticationResponse{}
	err = json.Unmarshal(body, authenticatedDoctorResponse)
	if err != nil {
		t.Fatal("Unable to parse response from patient authenticated: " + err.Error())
	}

	if authenticatedDoctorResponse.Token == "" || authenticatedDoctorResponse.Doctor == nil {
		t.Fatal("Doctor not authenticated as expected")
	}
}

func TestDoctorTwoFactorAuthentication(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	dres, email, password := SignupRandomTestDoctor(t, testData)
	doc, err := testData.DataApi.GetDoctorFromId(dres.DoctorId)
	if err != nil {
		t.Fatal(err)
	}

	// Enable two factor auth for the account

	if err := testData.AuthApi.UpdateAccount(doc.AccountId.Int64(), nil, api.BoolPtr(true)); err != nil {
		t.Fatal(err)
	}

	// First sign in for a device should return a two factor required response

	authReq := &doctor.AuthenticationRequestData{Email: email, Password: password}
	authRes := &doctor.AuthenticationResponse{}
	httpRes, err := testData.AuthPostJSON(testData.APIServer.URL+router.DoctorAuthenticateURLPath, 0, authReq, authRes)
	if err != nil {
		t.Fatal(err)
	}
	test.Equals(t, http.StatusOK, httpRes.StatusCode)

	if !authRes.TwoFactorRequired {
		t.Fatal("Expected two_factor_required to be true")
	}
	if authRes.TwoFactorToken == "" {
		t.Fatal("Two factor token not returned")
	}
	if authRes.Doctor != nil {
		t.Error("Doctor should not be set when two factor is required")
	}
	if authRes.Token != "" {
		t.Error("Token should not be set when two factor is required")
	}

	if len(testData.SMSAPI.Sent) == 0 {
		t.Fatal("Two factor SMS not sent")
	}
	t.Logf("%+v", testData.SMSAPI.Sent[0])
	testData.SMSAPI.Sent = nil

	// Test sending new two factor code

	tfReq := &doctor.TwoFactorRequest{Token: authRes.TwoFactorToken, Resend: true}
	tfRes := &doctor.AuthenticationResponse{}
	httpRes, err = testData.AuthPostJSON(testData.APIServer.URL+router.DoctorAuthenticateTwoFactorURLPath, 0, tfReq, tfRes)
	if err != nil {
		t.Fatal(err)
	}
	test.Equals(t, http.StatusOK, httpRes.StatusCode)

	if tfRes.Doctor != nil {
		t.Error("Doctor should not be set on resend")
	}
	if tfRes.Token != "" {
		t.Error("Token should not be set on resend")
	}

	if len(testData.SMSAPI.Sent) == 0 {
		t.Fatal("SMS resend failed")
	}
	sms := testData.SMSAPI.Sent[0]
	code := regexp.MustCompile(`\d+`).FindString(sms.Text)
	if code == "" {
		t.Fatal("Didn't find code in SMS")
	}

	// Test successful two factor request

	tfReq = &doctor.TwoFactorRequest{Token: authRes.TwoFactorToken, Code: code}
	tfRes = &doctor.AuthenticationResponse{}
	httpRes, err = testData.AuthPostJSON(testData.APIServer.URL+router.DoctorAuthenticateTwoFactorURLPath, 0, tfReq, tfRes)
	if err != nil {
		t.Fatal(err)
	}
	test.Equals(t, http.StatusOK, httpRes.StatusCode)

	if tfRes.Token == "" {
		t.Errorf("Token not provided on successful 2fa")
	}
	if tfRes.Doctor == nil {
		t.Errorf("Doctor not provided on successful 2fa")
	}
	if tfRes.TwoFactorRequired {
		t.Errorf("two_factor_required should not be true on successful 2fa")
	}

	// After a device is verified, subsequent auth requests should not require 2fa

	authReq = &doctor.AuthenticationRequestData{Email: email, Password: password}
	authRes = &doctor.AuthenticationResponse{}
	httpRes, err = testData.AuthPostJSON(testData.APIServer.URL+router.DoctorAuthenticateURLPath, 0, authReq, authRes)
	if err != nil {
		t.Fatal(err)
	}
	test.Equals(t, http.StatusOK, httpRes.StatusCode)

	if authRes.TwoFactorRequired {
		t.Errorf("two_factor_required should not be set")
	}
	if authRes.Token == "" {
		t.Errorf("Token not provided")
	}
	if authRes.Doctor == nil {
		t.Errorf("Doctor not provided")
	}
}

func TestDoctorDrugSearch(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()

	// use a real dosespot service before instantiating the server
	testData.Config.ERxAPI = testData.ERxApi
	testData.StartAPIServer(t)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor information from id: " + err.Error())
	}

	resp, err := testData.AuthGet(testData.APIServer.URL+router.AutocompleteURLPath+"?query=ben", doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a successful query to the autocomplete API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 but got %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to parse the body of the response: " + err.Error())
	}

	autocompleteResponse := &handlers.AutocompleteResponse{}
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
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)

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
	answerIntakeRequestBody := PrepareAnswersForQuestionsInPatientVisit(patientVisitResponse, t)
	SubmitAnswersIntakeForPatient(patient.PatientId.Int64(), patient.AccountId.Int64(), answerIntakeRequestBody, testData, t)
	SubmitPatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)

	answerIntakeRequestBody = PrepareAnswersForDiagnosingAsUnsuitableForSpruce(testData, t, patientVisitResponse.PatientVisitId)
	SubmitPatientVisitDiagnosisWithIntake(patientVisitResponse.PatientVisitId, doctor.AccountId.Int64(), answerIntakeRequestBody, testData, t)

	// the patient visit should have its state set to TRIAGED
	patientVisit, err := testData.DataApi.GetPatientVisitFromId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatal(err.Error())
	} else if patientVisit.Status != common.PVStatusTriaged {
		t.Fatalf("Expected status to be %s but it was %s instead", common.PVStatusTriaged, patientVisit.Status)
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
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)

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
	answerIntakeRequestBody := PrepareAnswersForQuestionsInPatientVisit(patientVisitResponse, t)
	SubmitAnswersIntakeForPatient(patient.PatientId.Int64(), patient.AccountId.Int64(), answerIntakeRequestBody, testData, t)
	SubmitPatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)

	// doctor now attempts to diagnose patient visit
	requestParams := bytes.NewBufferString("?patient_visit_id=")
	requestParams.WriteString(strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	diagnosisResponse := patient_visit.GetDiagnosisResponse{}

	resp, err := testData.AuthGet(testData.APIServer.URL+router.DoctorVisitDiagnosisURLPath+requestParams.String(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Something went wrong when trying to get diagnoses layout for doctor to diagnose patient visit: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	patientSignedupResponse := SignupRandomTestPatient(t, testData)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)

	// get patient to start a visit
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), testData, t)

	// submit answers to questions in patient visit
	patient, err := testData.DataApi.GetPatientFromId(patientSignedupResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatal("Unable to get patient from id: " + err.Error())
	}

	answerIntakeRequestBody := PrepareAnswersForQuestionsInPatientVisit(patientVisitResponse, t)
	SubmitAnswersIntakeForPatient(patient.PatientId.Int64(), patient.AccountId.Int64(), answerIntakeRequestBody, testData, t)

	// get patient to submit the visit
	SubmitPatientVisitForPatient(patientSignedupResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor object from id: " + err.Error())
	}

	jsonData, err := json.Marshal(&doctor_treatment_plan.TreatmentPlanRequestData{})
	test.OK(t, err)

	// attempt to submit the treatment plan here. It should fail

	resp, err := testData.AuthPut(testData.APIServer.URL+router.DoctorTreatmentPlansURLPath, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a call to submit the patient visit review : " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code to be %d but got %d instead. The call should have failed because the patient visit is not being REVIEWED by the doctor yet. ", http.StatusBadRequest, resp.StatusCode)
	}

	// get the doctor to start reviewing the patient visit
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)
	responseData := PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, nil, testData, t)

	caseID, err := testData.DataApi.GetPatientCaseIdFromPatientVisitId(patientVisitResponse.PatientVisitId)
	test.OK(t, err)

	// Shouldn't be any messages yet
	if msgs, err := testData.DataApi.ListCaseMessages(caseID, api.PATIENT_ROLE); err != nil {
		t.Fatal(err)
	} else if len(msgs) != 0 {
		t.Fatalf("Expected no doctor message but got %d", len(msgs))
	}

	jsonData, err = json.Marshal(doctor_treatment_plan.TreatmentPlanRequestData{
		TreatmentPlanId: responseData.TreatmentPlan.Id.Int64(),
		Message:         "Foo",
	})

	// attempt to submit the patient visit review here. It should work
	resp, err = testData.AuthPut(testData.APIServer.URL+router.DoctorTreatmentPlansURLPath, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful call to submit patient visit review")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Expected %d but got %d: %s", http.StatusOK, resp.StatusCode, string(b))
	}

	patientVisit, err := testData.DataApi.GetPatientVisitFromId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatal("Unable to get patient visit given id: " + err.Error())
	}

	if patientVisit.Status != common.PVStatusTreated {
		t.Fatalf("Expected the status to be %s but status is %s", common.PVStatusTreated, patientVisit.Status)
	}

	// Shouldn't be any messages yet
	if msgs, err := testData.DataApi.ListCaseMessages(caseID, api.PATIENT_ROLE); err != nil {
		t.Fatal(err)
	} else if len(msgs) != 1 {
		t.Fatalf("Expected 1 doctor message but got %d", len(msgs))
	}
}
