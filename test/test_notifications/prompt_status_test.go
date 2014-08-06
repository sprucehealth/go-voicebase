package test_notifications

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sprucehealth/backend/apiservice/router"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/test/test_integration"
)

// Test prompt status on login and signup
func TestPromptStatus_Signup(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	testData.StartAPIServer(t)
	pr := test_integration.SignupRandomTestPatient(t, testData)
	patient := pr.Patient

	if patient.PromptStatus != common.Unprompted {
		t.Fatalf("Expected prompt status %s but got %s", common.Unprompted, patient.PromptStatus)
	}
}

func TestPromptStatus_Login(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	testData.StartAPIServer(t)
	pr := test_integration.SignupRandomTestPatient(t, testData)
	patient := pr.Patient

	// this method would be called when trying to login so checking directly with data service layer
	patient, err := testData.DataApi.GetPatientFromAccountId(patient.AccountId.Int64())
	if err != nil {
		t.Fatalf(err.Error())
	}

	if patient.PromptStatus != common.Unprompted {
		t.Fatalf("Expected prompt status %s but got %s", common.Unprompted, patient.PromptStatus)
	}
}

// Test prompt status after being set
func TestPromptStatus_OnModify(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	testData.StartAPIServer(t)
	pr := test_integration.SignupRandomTestPatient(t, testData)
	patient := pr.Patient

	params := url.Values{}
	params.Set("prompt_status", "DECLINED")

	res, err := testData.AuthPut(testData.APIServer.URL+router.NotificationPromptStatusURLPath, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), patient.AccountId.Int64())
	if err != nil {
		t.Fatal(err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d instead got %d", http.StatusOK, res.StatusCode)
	}

	patient, err = testData.DataApi.GetPatientFromId(patient.PatientId.Int64())
	if err != nil {
		t.Fatal(err.Error())
	}

	if patient.PromptStatus != common.Declined {
		t.Fatalf("Expected prompt status %s instead got %s", common.Declined, patient.PromptStatus)
	}
}

// Test prompt status for doctor
func TestPromptStatus_DoctorSignup(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	testData.StartAPIServer(t)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if doctor.PromptStatus != common.Unprompted {
		t.Fatalf("Expected prompt status for doctor to be %s instead it was %s", common.Unprompted, doctor.PromptStatus)
	}
}

func TestPromptStatus_DoctorOnModify(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	testData.StartAPIServer(t)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf(err.Error())
	}

	params := url.Values{}
	params.Set("prompt_status", "DECLINED")

	res, err := testData.AuthPut(testData.APIServer.URL+router.NotificationPromptStatusURLPath, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal(err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d instead got %d", http.StatusOK, res.StatusCode)
	}

	doctor, err = testData.DataApi.GetDoctorFromId(doctor.DoctorId.Int64())
	if err != nil {
		t.Fatal(err.Error())
	}

	if doctor.PromptStatus != common.Declined {
		t.Fatalf("Expected prompt status %s instead got %s", common.Declined, doctor.PromptStatus)
	}
}
