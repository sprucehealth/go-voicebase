package test_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/doctor_treatment_plan"
	"carefront/encoding"
	"carefront/libs/erx"
	"carefront/patient_file"
	"carefront/patient_visit"
)

func SignupRandomTestDoctor(t *testing.T, testData *TestData) (signedupDoctorResponse *apiservice.DoctorSignedupResponse, email, password string) {
	authHandler := &apiservice.SignupDoctorHandler{AuthApi: testData.AuthApi, DataApi: testData.DataApi}
	ts := httptest.NewServer(authHandler)
	defer ts.Close()

	email = strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.com"
	password = "12345"
	params := &url.Values{}
	params.Set("first_name", "Test")
	params.Set("last_name", "Test")
	params.Set("email", email)
	params.Set("password", password)
	params.Set("dob", "1987-11-08")
	params.Set("gender", "male")
	params.Set("clinician_id", os.Getenv("DOSESPOT_USER_ID"))
	params.Set("phone", "123451616")
	params.Set("address_line_1", "12345 Main street")
	params.Set("address_line_2", "apt 11415")
	params.Set("city", "san francisco")
	params.Set("state", "ca")
	params.Set("zip_code", "94115")

	res, err := testData.AuthPost(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), 0)
	if err != nil {
		t.Fatal("Unable to make post request for registering patient: " + err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Unable to read body of response: " + err.Error())
	}
	CheckSuccessfulStatusCode(res, fmt.Sprintf("Unable to make success request to signup patient. Here's the code returned %d and here's the body of the request %s", res.StatusCode, body), t)

	signedupDoctorResponse = &apiservice.DoctorSignedupResponse{}
	err = json.Unmarshal(body, signedupDoctorResponse)
	if err != nil {
		t.Fatal("Unable to parse response from patient signed up")
	}
	return signedupDoctorResponse, email, password
}

func setupErxAPI(t *testing.T) *erx.DoseSpotService {
	clinicKey := os.Getenv("DOSESPOT_CLINIC_KEY")
	clinicId, _ := strconv.ParseInt(os.Getenv("DOSESPOT_CLINIC_ID"), 10, 64)
	userId, _ := strconv.ParseInt(os.Getenv("DOSESPOT_USER_ID"), 10, 64)

	if clinicKey == "" {
		t.Log("WARNING: skipping doctor drug search test since the dosespot ids are not present as environment variables")
		t.SkipNow()
	}

	erx := erx.NewDoseSpotService(clinicId, userId, clinicKey, nil)
	return erx
}

func SubmitPatientVisitDiagnosis(PatientVisitId int64, doctor *common.Doctor, testData *TestData, t *testing.T) {
	answerIntakeRequestBody := &apiservice.AnswerIntakeRequestBody{}
	answerIntakeRequestBody.PatientVisitId = PatientVisitId
	var diagnosisQuestionId, severityQuestionId, acneTypeQuestionId int64
	if qi, err := testData.DataApi.GetQuestionInfo("q_acne_diagnosis", 1); err != nil {
		t.Fatalf("Unable to get the questionIds for the question tags requested for the doctor to diagnose patient visit: %s", err.Error())
	} else {
		diagnosisQuestionId = qi.QuestionId
	}
	if qi, err := testData.DataApi.GetQuestionInfo("q_acne_severity", 1); err != nil {
		t.Fatalf("Unable to get the questionIds for the question tags requested for the doctor to diagnose patient visit: %s", err.Error())
	} else {
		severityQuestionId = qi.QuestionId
	}
	if qi, err := testData.DataApi.GetQuestionInfo("q_acne_type", 1); err != nil {
		t.Fatalf("Unable to get the questionIds for the question tags requested for the doctor to diagnose patient visit: %s", err.Error())
	} else {
		acneTypeQuestionId = qi.QuestionId
	}

	diagnosePatientHandler := patient_visit.NewDiagnosePatientHandler(testData.DataApi, testData.AuthApi, "")
	ts := httptest.NewServer(diagnosePatientHandler)
	defer ts.Close()

	answerInfo, err := testData.DataApi.GetAnswerInfoForTags([]string{"a_doctor_acne_vulgaris"}, api.EN_LANGUAGE_ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	answerToQuestionItem := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem.QuestionId = diagnosisQuestionId
	answerToQuestionItem.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: answerInfo[0].AnswerId}}

	answerInfo, err = testData.DataApi.GetAnswerInfoForTags([]string{"a_doctor_acne_severity_severity"}, api.EN_LANGUAGE_ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	answerToQuestionItem2 := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem2.QuestionId = severityQuestionId
	answerToQuestionItem2.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: answerInfo[0].AnswerId}}

	answerInfo, err = testData.DataApi.GetAnswerInfoForTags([]string{"a_acne_inflammatory"}, api.EN_LANGUAGE_ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	answerToQuestionItem3 := &apiservice.AnswerToQuestionItem{}
	answerToQuestionItem3.QuestionId = acneTypeQuestionId
	answerToQuestionItem3.AnswerIntakes = []*apiservice.AnswerItem{&apiservice.AnswerItem{PotentialAnswerId: answerInfo[0].AnswerId}}

	answerIntakeRequestBody.Questions = []*apiservice.AnswerToQuestionItem{answerToQuestionItem, answerToQuestionItem2, answerToQuestionItem3}

	requestData, err := json.Marshal(answerIntakeRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body")
	}

	resp, err := testData.AuthPost(ts.URL, "application/json", bytes.NewBuffer(requestData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to successfully submit the diagnosis of a patient visit: " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatal(err.Error())
	}

	// now, get diagnosis layout again and check to ensure that the doctor successfully diagnosed the patient with the expected answers
	diagnosisLayout, err := patient_visit.GetDiagnosisLayout(testData.DataApi, PatientVisitId, doctor.DoctorId.Int64())
	if err != nil {
		t.Fatal(err.Error())
	}

	if diagnosisLayout == nil {
		t.Fatal("Diagnosis response not as expected after doctor submitted diagnosis")
	}

	for _, section := range diagnosisLayout.InfoIntakeLayout.Sections {
		for _, question := range section.Questions {

			for _, response := range GetAnswerIntakesFromAnswers(question.Answers, t) {
				switch response.QuestionId.Int64() {
				case diagnosisQuestionId:
					if response.PotentialAnswerId.Int64() != answerToQuestionItem.AnswerIntakes[0].PotentialAnswerId {
						t.Fatalf("Doctor response to question id %d expectd to have id %d but has id %d", response.QuestionId.Int64(), 102, response.PotentialAnswerId.Int64())
					}
				case severityQuestionId:
					if response.PotentialAnswerId.Int64() != answerToQuestionItem2.AnswerIntakes[0].PotentialAnswerId {
						t.Fatalf("Doctor response to question id %d expectd to have id %d but has id %d", response.QuestionId.Int64(), 107, response.PotentialAnswerId.Int64())
					}

				case acneTypeQuestionId:
					if response.PotentialAnswerId.Int64() != answerToQuestionItem3.AnswerIntakes[0].PotentialAnswerId {
						t.Fatalf("Doctor response to question id %d expectd to have any of ids %s but instead has id %d", response.QuestionId.Int64(), "(109,114,113)", response.PotentialAnswerId.Int64())
					}

				}
			}
		}
	}

	return
}

func StartReviewingPatientVisit(patientVisitId int64, doctor *common.Doctor, testData *TestData, t *testing.T) {
	doctorPatientVisitReviewHandler := patient_file.NewDoctorPatientVisitReviewHandler(testData.DataApi)

	ts := httptest.NewServer(doctorPatientVisitReviewHandler)
	defer ts.Close()

	resp, err := testData.AuthGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitId, 10), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to get patient visit review for patient: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make successful call to get patient visit review: ", t)
}

func PickATreatmentPlan(parent *common.TreatmentPlanParent, contentSource *common.TreatmentPlanContentSource, doctor *common.Doctor, testData *TestData, t *testing.T) *doctor_treatment_plan.DoctorTreatmentPlanResponse {
	doctorPickTreatmentPlanHandler := doctor_treatment_plan.NewDoctorTreatmentPlanHandler(testData.DataApi, nil, nil, false)

	ts := httptest.NewServer(doctorPickTreatmentPlanHandler)
	defer ts.Close()

	requestData := doctor_treatment_plan.PickTreatmentPlanRequestData{
		TPParent: parent,
	}

	if contentSource != nil {
		requestData.TPContentSource = contentSource
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := testData.AuthPost(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to pick a treatment plan for the patient visit doctor %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected successful picking up of treatment plan instead got %d", resp.StatusCode)
	}

	responseData := &doctor_treatment_plan.DoctorTreatmentPlanResponse{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into response json: %s", err)
	}

	return responseData
}

func PickATreatmentPlanForPatientVisit(patientVisitId int64, doctor *common.Doctor, favoriteTreatmentPlan *common.FavoriteTreatmentPlan, testData *TestData, t *testing.T) *doctor_treatment_plan.DoctorTreatmentPlanResponse {
	doctorPickTreatmentPlanHandler := doctor_treatment_plan.NewDoctorTreatmentPlanHandler(testData.DataApi, nil, nil, false)

	ts := httptest.NewServer(doctorPickTreatmentPlanHandler)
	defer ts.Close()

	requestData := doctor_treatment_plan.PickTreatmentPlanRequestData{
		TPParent: &common.TreatmentPlanParent{
			ParentId:   encoding.NewObjectId(patientVisitId),
			ParentType: common.TPParentTypePatientVisit,
		},
	}

	if favoriteTreatmentPlan != nil {
		requestData.TPContentSource = &common.TreatmentPlanContentSource{
			ContentSourceType: common.TPContentSourceTypeFTP,
			ContentSourceId:   favoriteTreatmentPlan.Id,
		}
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := testData.AuthPost(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to pick a treatment plan for the patient visit doctor %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected successful picking up of treatment plan instead got %d", resp.StatusCode)
	}

	responseData := &doctor_treatment_plan.DoctorTreatmentPlanResponse{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into response json: %s", err)
	}

	return responseData
}

func SubmitPatientVisitBackToPatient(treatmentPlanId int64, doctor *common.Doctor, testData *TestData, t *testing.T) {
	doctorTreatmentPlanHandler := doctor_treatment_plan.NewDoctorTreatmentPlanHandler(testData.DataApi, nil, nil, false)
	ts := httptest.NewServer(doctorTreatmentPlanHandler)
	defer ts.Close()

	requestData := &doctor_treatment_plan.TreatmentPlanRequestData{
		TreatmentPlanId: encoding.NewObjectId(treatmentPlanId),
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := testData.AuthPut(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make call to close patient visit " + err.Error())
	}
	CheckSuccessfulStatusCode(resp, "Unable to make successful call to close the patient visit", t)
}
