package test_patient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sprucehealth/backend/apiservice/router"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestAccount_PCP(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	pr := test_integration.SignupRandomTestPatient(t, testData)
	// add a pcp for the patient
	pcp := &common.PCP{
		PhysicianName: "Dr. Test Test",
		PracticeName:  "Practice Name",
		PhoneNumber:   "734-846-5522",
		Email:         "test@test.com",
		FaxNumber:     "734-846-5522",
	}

	jsonData, err := json.Marshal(&map[string]interface{}{"pcp": pcp})
	test.OK(t, err)
	res, err := testData.AuthPut(testData.APIServer.URL+router.PatientPCPURLPath, "application/json", bytes.NewReader(jsonData), pr.Patient.AccountId.Int64())
	test.OK(t, err)
	defer res.Body.Close()
	test.Equals(t, http.StatusOK, res.StatusCode)

	// lets retrieve the pcp to ensure its the same
	var responseData struct {
		PCP *common.PCP `json:"pcp"`
	}

	res, err = testData.AuthGet(testData.APIServer.URL+router.PatientPCPURLPath, pr.Patient.AccountId.Int64())
	test.OK(t, err)
	defer res.Body.Close()
	test.Equals(t, http.StatusOK, res.StatusCode)
	err = json.NewDecoder(res.Body).Decode(&responseData)
	test.OK(t, err)
	test.Equals(t, *responseData.PCP, *pcp)
}

func TestAccount_EmergencyContacts(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	pr := test_integration.SignupRandomTestPatient(t, testData)

	emergencyContacts := []*common.EmergencyContact{
		&common.EmergencyContact{
			ID:           1,
			FullName:     "Test 1",
			Relationship: "Test's Brother",
			PhoneNumber:  "734-846-5522",
		},
		&common.EmergencyContact{
			ID:           2,
			FullName:     "Test 2",
			Relationship: "Test's Sister",
			PhoneNumber:  "734-846-5523",
		},
	}

	jsonData, err := json.Marshal(&map[string]interface{}{"emergency_contacts": emergencyContacts})
	test.OK(t, err)
	res, err := testData.AuthPut(testData.APIServer.URL+router.PatientEmergencyContactsURLPath, "application/json", bytes.NewReader(jsonData), pr.Patient.AccountId.Int64())
	test.OK(t, err)
	defer res.Body.Close()
	test.Equals(t, http.StatusOK, res.StatusCode)

	// get the emergency contacts to ensure that it saved
	var responseData struct {
		EmergencyContacts []*common.EmergencyContact `json:"emergency_contacts"`
	}
	res, err = testData.AuthGet(testData.APIServer.URL+router.PatientEmergencyContactsURLPath, pr.Patient.AccountId.Int64())
	test.OK(t, err)
	defer res.Body.Close()
	test.Equals(t, http.StatusOK, res.StatusCode)
	err = json.NewDecoder(res.Body).Decode(&responseData)
	test.OK(t, err)
	test.Equals(t, 2, len(responseData.EmergencyContacts))
	test.Equals(t, *emergencyContacts[0], *responseData.EmergencyContacts[0])
	test.Equals(t, *emergencyContacts[1], *responseData.EmergencyContacts[1])

}
