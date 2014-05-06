package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"carefront/api"
	"carefront/apiservice"
	"carefront/info_intake"
	patientApiService "carefront/patient"
	thriftapi "carefront/thrift/api"
)

func signupRandomTestPatient(t *testing.T, dataApi api.DataAPI, authApi thriftapi.Auth) *patientApiService.PatientSignedupResponse {
	authHandler := patientApiService.NewSignupHandler(dataApi, authApi)
	ts := httptest.NewServer(authHandler)
	defer ts.Close()

	requestBody := bytes.NewBufferString("first_name=Test&last_name=Test&email=")
	requestBody.WriteString(strconv.FormatInt(rand.Int63(), 10))
	requestBody.WriteString("@example.com&password=12345&dob=1987-11-08&zip_code=94115&phone=7348465522&gender=male")
	res, err := authPost(ts.URL, "application/x-www-form-urlencoded", requestBody, 0)
	if err != nil {
		t.Fatal("Unable to make post request for registering patient: " + err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Unable to read body of response: " + err.Error())
	}
	CheckSuccessfulStatusCode(res, fmt.Sprintf("Unable to make success request to signup patient. Here's the code returned %d and here's the body of the request %s", res.StatusCode, body), t)

	signedupPatientResponse := &patientApiService.PatientSignedupResponse{}
	err = json.Unmarshal(body, signedupPatientResponse)
	if err != nil {
		t.Fatal("Unable to parse response from patient signed up")
	}
	return signedupPatientResponse
}

func getPatientVisitForPatient(patientId int64, testData TestData, t *testing.T) *apiservice.PatientVisitResponse {
	patientVisitHandler := apiservice.NewPatientVisitHandler(testData.DataApi, testData.AuthApi,
		testData.CloudStorageService, testData.CloudStorageService)
	patient, err := testData.DataApi.GetPatientFromId(patientId)
	if err != nil {
		t.Fatal("Unable to get patient information given the patient id: " + err.Error())
	}

	ts := httptest.NewServer(patientVisitHandler)
	defer ts.Close()

	// register a patient visit for this patient
	resp, err := authGet(ts.URL, patient.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to get the patient visit id")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of the response for the new patient visit call: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unsuccessful call to register new patient visit: "+string(body), t)

	patientVisitResponse := &apiservice.PatientVisitResponse{}
	err = json.Unmarshal(body, patientVisitResponse)
	if err != nil {
		t.Fatal("Unable to unmarshall response body into patient visit response: " + err.Error())
	}

	return patientVisitResponse
}

func createPatientVisitForPatient(patientId int64, testData TestData, t *testing.T) *apiservice.PatientVisitResponse {
	patientVisitHandler := apiservice.NewPatientVisitHandler(testData.DataApi, testData.AuthApi,
		testData.CloudStorageService, testData.CloudStorageService)
	patient, err := testData.DataApi.GetPatientFromId(patientId)
	if err != nil {
		t.Fatal("Unable to get patient information given the patient id: " + err.Error())
	}

	ts := httptest.NewServer(patientVisitHandler)
	defer ts.Close()

	// register a patient visit for this patient
	resp, err := authPost(ts.URL, "application/x-www-form-urlencoded", nil, patient.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to get the patient visit id")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of the response for the new patient visit call: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unsuccessful call to register new patient visit: "+string(body), t)

	patientVisitResponse := &apiservice.PatientVisitResponse{}
	err = json.Unmarshal(body, patientVisitResponse)
	if err != nil {
		t.Fatal("Unable to unmarshall response body into patient visit response: " + err.Error())
	}

	return patientVisitResponse
}

// randomly answering all top level questions in the patient visit, regardless of the condition under which the questions are presented to the user.
// the goal of this is to get all questions answered so as to render the views for the doctor layout, not to test the sanity of the answers the patient inputs.
func prepareAnswersForQuestionsInPatientVisit(patientVisitResponse *apiservice.PatientVisitResponse, t *testing.T) *apiservice.AnswerIntakeRequestBody {
	answerIntakeRequestBody := apiservice.AnswerIntakeRequestBody{}
	answerIntakeRequestBody.PatientVisitId = patientVisitResponse.PatientVisitId
	answerIntakeRequestBody.Questions = make([]*apiservice.AnswerToQuestionItem, 0)
	for _, section := range patientVisitResponse.ClientLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				switch question.QuestionTypes[0] {
				case info_intake.QUESTION_TYPE_SINGLE_SELECT:
					answerIntakeRequestBody.Questions = append(answerIntakeRequestBody.Questions, &apiservice.AnswerToQuestionItem{
						QuestionId: question.QuestionId,
						AnswerIntakes: []*apiservice.AnswerItem{&apiservice.AnswerItem{
							PotentialAnswerId: question.PotentialAnswers[0].AnswerId,
						},
						},
					})
				case info_intake.QUESTION_TYPE_MULTIPLE_CHOICE:
					answerIntakeRequestBody.Questions = append(answerIntakeRequestBody.Questions, &apiservice.AnswerToQuestionItem{
						QuestionId: question.QuestionId,
						AnswerIntakes: []*apiservice.AnswerItem{
							&apiservice.AnswerItem{
								PotentialAnswerId: question.PotentialAnswers[0].AnswerId,
							},
							&apiservice.AnswerItem{
								PotentialAnswerId: question.PotentialAnswers[1].AnswerId,
							},
						},
					})
				case info_intake.QUESTION_TYPE_AUTOCOMPLETE:
					answerIntakeRequestBody.Questions = append(answerIntakeRequestBody.Questions, &apiservice.AnswerToQuestionItem{
						QuestionId: question.QuestionId,
						AnswerIntakes: []*apiservice.AnswerItem{
							&apiservice.AnswerItem{
								AnswerText: "autocomplete 1",
							},
						},
					})
				case info_intake.QUESTION_TYPE_FREE_TEXT:
					answerIntakeRequestBody.Questions = append(answerIntakeRequestBody.Questions, &apiservice.AnswerToQuestionItem{
						QuestionId: question.QuestionId,
						AnswerIntakes: []*apiservice.AnswerItem{
							&apiservice.AnswerItem{
								AnswerText: "This is a test answer",
							},
						},
					})
				}
			}
		}
	}
	return &answerIntakeRequestBody
}

func submitAnswersIntakeForPatient(patientId, patientAccountId int64, answerIntakeRequestBody *apiservice.AnswerIntakeRequestBody, testData TestData, t *testing.T) {
	answerIntakeHandler := &apiservice.AnswerIntakeHandler{
		DataApi: testData.DataApi,
	}

	ts := httptest.NewServer(answerIntakeHandler)
	defer ts.Close()

	jsonData, err := json.Marshal(answerIntakeRequestBody)
	if err != nil {
		t.Fatalf("Unable to marshal answer intake body: %s", err)
	}
	resp, err := authPost(ts.URL, "application/json", bytes.NewReader(jsonData), patientAccountId)
	if err != nil {
		t.Fatalf("Unable to successfully make request to submit answer intake: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to successfuly make call to submit answer intake. Expected 200 but got %d", resp.StatusCode)
	}
}

func submitPatientVisitForPatient(patientId, patientVisitId int64, testData TestData, t *testing.T) {
	patientVisitHandler := apiservice.NewPatientVisitHandler(testData.DataApi, testData.AuthApi,
		testData.CloudStorageService, testData.CloudStorageService)
	patient, err := testData.DataApi.GetPatientFromId(patientId)
	if err != nil {
		t.Fatal("Unable to get patient information given the patient id: " + err.Error())
	}

	ts := httptest.NewServer(patientVisitHandler)
	defer ts.Close()
	buffer := bytes.NewBufferString("patient_visit_id=")
	buffer.WriteString(strconv.FormatInt(patientVisitId, 10))

	resp, err := authPut(ts.URL, "application/x-www-form-urlencoded", buffer, patient.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to get the patient visit id")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of the response for the new patient visit call: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unsuccessful call to register new patient visit: "+string(body), t)

	// get the patient visit information to ensure that the case has been submitted
	patientVisit, err := testData.DataApi.GetPatientVisitFromId(patientVisitId)
	if patientVisit.Status != "SUBMITTED" {
		t.Fatalf("Case status should be submitted after the case was submitted to the doctor, but its not. It is %s instead.", patientVisit.Status)
	}
}
