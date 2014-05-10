package integration

import (
	"bytes"
	patientApiService "carefront/patient"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"carefront/address"
	"carefront/api"
	"carefront/apiservice"

	_ "github.com/go-sql-driver/mysql"
)

func TestPatientRegistration(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)
	SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
}

func TestPatientCareProvidingEllgibility(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	stubAddressValidationService := address.StubAddressValidationService{
		CityStateToReturn: address.CityState{
			City:              "San Francisco",
			State:             "California",
			StateAbbreviation: "CA",
		},
	}

	checkElligibilityHandler := &apiservice.CheckCareProvidingElligibilityHandler{DataApi: testData.DataApi, AddressValidationApi: stubAddressValidationService}
	ts := httptest.NewServer(checkElligibilityHandler)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "?zip_code=94115")

	if err != nil {
		t.Fatal("Unable to successfuly check care providing elligiblity for patient " + err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to successfuly read the body of the response")
	}

	CheckSuccessfulStatusCode(resp, "Unable to make a successful call to check for care providing elligibility: "+string(body), t)

	stubAddressValidationService.CityStateToReturn = address.CityState{
		City:              "Aventura",
		State:             "Florida",
		StateAbbreviation: "FL",
	}

	checkElligibilityHandler.AddressValidationApi = stubAddressValidationService

	resp, err = AuthGet(ts.URL+"?zip_code=33180", 0)
	if err != nil {
		t.Fatal("Unable to successfuly check care providing elligibility for patient" + err.Error())
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read the response from the body for patient care providing elligibility check: " + err.Error())
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected the status code to be 403, but got a %d instead", resp.StatusCode)
	}
}

func TestPatientVisitCreation(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	signedupPatientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patientVisitResponse := CreatePatientVisitForPatient(signedupPatientResponse.Patient.PatientId.Int64(), testData, t)

	if patientVisitResponse.PatientVisitId == 0 {
		t.Fatal("Patient Visit Id not set when it should be.")
	}

	if patientVisitResponse.ClientLayout == nil {
		t.Fatal("The questions for patient intake should be returned as part of the patient visit")
	}

	// checking to ensure that the care team was created
	careTeam, err := testData.DataApi.GetCareTeamForPatient(signedupPatientResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatal("Unable to get care team for patient visit: " + err.Error())
	}

	if !(careTeam == nil || careTeam.PatientId == signedupPatientResponse.Patient.PatientId.Int64()) {
		t.Fatal("Unable to get patient visit id for care team")
	}

	// ensuring that we have a primary doctor assigned to the case
	primaryDoctorFound := false
	for _, assignment := range careTeam.Assignments {
		if assignment.ProviderRole == "DOCTOR" {
			primaryDoctorFound = true
		}
	}

	if primaryDoctorFound == false {
		t.Fatal("Primary doctor not found for patient visit")
	}

	// getting the patient visit again as we should get back the same patient visit id
	// since this patient visit has not been completed
	anotherPatientVisitResponse := getPatientVisitForPatient(signedupPatientResponse.Patient.PatientId.Int64(), testData, t)
	if anotherPatientVisitResponse.PatientVisitId != patientVisitResponse.PatientVisitId {
		t.Fatal("The patient visit id for subsequent calls should be the same so long as we have not closed/submitted the case")
	}
}

func TestPatientVisitSubmission(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	signedupPatientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patientVisitResponse := CreatePatientVisitForPatient(signedupPatientResponse.Patient.PatientId.Int64(), testData, t)

	SubmitPatientVisitForPatient(signedupPatientResponse.Patient.PatientId.Int64(), patientVisitResponse.PatientVisitId, testData, t)

	// try submitting the exact same patient visit again, and it should come back with a 403 given that the case has already been submitted

	patientVisitHandler := apiservice.NewPatientVisitHandler(testData.DataApi, testData.AuthApi,
		testData.CloudStorageService, testData.CloudStorageService)
	patient, err := testData.DataApi.GetPatientFromId(signedupPatientResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatal("Unable to get patient information given the patient id: " + err.Error())
	}

	ts := httptest.NewServer(patientVisitHandler)
	defer ts.Close()
	buffer := bytes.NewBufferString("patient_visit_id=")
	buffer.WriteString(strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))

	resp, err := AuthPut(ts.URL, "application/x-www-form-urlencoded", buffer, patient.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to get the patient visit id")
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected a bad request 403 to be returned when attempting to submit an already submitted patient visit, but instead got %d", resp.StatusCode)
	}
}

func TestPatientAutocompleteForDrugs(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	signedupPatientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)

	autocompleteHandler := apiservice.AutocompleteHandler{
		DataApi: testData.DataApi,
		ERxApi:  setupErxAPI(t),
		Role:    api.PATIENT_ROLE,
	}

	params := url.Values{}
	params.Set("query", "Lipi")

	ts := httptest.NewServer(&autocompleteHandler)
	defer ts.Close()

	resp, err := AuthGet(ts.URL+"?"+params.Encode(), signedupPatientResponse.Patient.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unsuccessful get request to autocomplete api: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to successfully do a drug search from patient side: %s", err)
	}

	autoCompleteResponse := apiservice.AutocompleteResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&autoCompleteResponse); err != nil {
		t.Fatalf("Unable to decode response body into json: %s", err)
	}

	if len(autoCompleteResponse.Suggestions) == 0 {
		t.Fatalf("Expected suggestions to be returned from the autocomplete api instead got 0")
	}
}

func TestPatientInformationUpdate(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	signedupPatientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)

	// attempt to update all expected fields
	expectedFirstName := "howard"
	expectedLastName := "plower"
	expectedPhone := "1234567890"
	expectedGender := "other"
	expectedDob := "1900-01-01"
	params := url.Values{}
	params.Set("first_name", expectedFirstName)
	params.Set("last_name", expectedLastName)
	params.Set("phone", expectedPhone)
	params.Set("gender", expectedGender)
	params.Set("dob", expectedDob)

	patientUpdateHandler := patientApiService.NewUpdateHandler(testData.DataApi)
	ts := httptest.NewServer(patientUpdateHandler)
	defer ts.Close()

	resp, err := AuthPut(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), signedupPatientResponse.Patient.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to update patient information: %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d but got %d", http.StatusOK, resp.StatusCode)
	}

	patient, err := testData.DataApi.GetPatientFromId(signedupPatientResponse.Patient.PatientId.Int64())
	if err != nil {
		t.Fatalf("unable to get patient from id : %s", err)
	}

	if patient.FirstName != expectedFirstName {
		t.Fatalf("Expected first name %s but got %s", expectedFirstName, patient.FirstName)
	} else if patient.LastName != expectedLastName {
		t.Fatalf("Expected last name %s but got %s", expectedLastName, patient.LastName)
	} else if patient.Gender != expectedGender {
		t.Fatalf("Expected gender %s but got %s", expectedGender, patient.Gender)
	} else if patient.Dob.String() != expectedDob {
		t.Fatalf("Expected dob %s but got %s", expectedDob, patient.Dob.String())
	} else if len(patient.PhoneNumbers) != 1 {
		t.Fatalf("Expected 1 phone number to exist instead got %d", len(patient.PhoneNumbers))
	} else if patient.PhoneNumbers[0].Phone != expectedPhone {
		t.Fatalf("Expected phone %s but got %s", expectedPhone, patient.PhoneNumbers[0].Phone)
	}

	// now attempt to update email or zipcode and it should return a bad request
	params.Set("zipcode", "21345")
	params.Set("email", "test@test.com")
	resp, err = AuthPut(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), patient.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to update patient information: %s", err)
	} else if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected %d response code but got %d instead", http.StatusBadRequest, resp.StatusCode)
	}
}
