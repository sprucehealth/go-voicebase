package test_treatment_plan

import (
	"bytes"
	"carefront/common"
	"carefront/doctor_treatment_plan"
	"carefront/patient_visit"
	"carefront/test/test_integration"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRegimenForPatientVisit(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// attempt to get the regimen plan or a patient visit
	regimenPlan := test_integration.GetRegimenPlanForTreatmentPlan(testData, doctor, treatmentPlan.Id.Int64(), t)

	if len(regimenPlan.AllRegimenSteps) > 0 {
		t.Fatal("There should be no regimen steps given that none have been created yet")
	}

	if len(regimenPlan.RegimenSections) > 0 {
		t.Fatal("There should be no regimen sections for the patient visit given that none have been created yet")
	}

	// adding new regimen steps to the doctor but not to the patient visit
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id

	regimenStep1 := &common.DoctorInstructionItem{}
	regimenStep1.Text = "Regimen Step 1"
	regimenStep1.State = common.STATE_ADDED

	regimenStep2 := &common.DoctorInstructionItem{}
	regimenStep2.Text = "Regimen Step 2"
	regimenStep2.State = common.STATE_ADDED

	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	if len(regimenPlanResponse.RegimenSections) > 0 {
		t.Fatal("Regimen section should not exist even though regimen steps were created by doctor")
	}

	// make the response the request since the response always returns the updated view of the system
	regimenPlanRequest = regimenPlanResponse

	// now lets add a couple regimen steps to a regimen section
	regimenSection := &common.RegimenSection{}
	regimenSection.RegimenName = "morning"
	regimenSection.RegimenSteps = []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
		ParentId: regimenPlanRequest.AllRegimenSteps[0].Id,
		Text:     regimenPlanRequest.AllRegimenSteps[0].Text,
	},
	}

	regimenSection2 := &common.RegimenSection{}
	regimenSection2.RegimenName = "night"
	regimenSection2.RegimenSteps = []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
		ParentId: regimenPlanRequest.AllRegimenSteps[1].Id,
		Text:     regimenPlanRequest.AllRegimenSteps[1].Text,
	},
	}

	regimenPlanRequest.RegimenSections = []*common.RegimenSection{regimenSection, regimenSection2}
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanResponse, regimenPlanResponse, t)

	if len(regimenPlanResponse.RegimenSections) != 2 {
		t.Fatalf("Expected the number of regimen sections to be 2 but there are %d instead", len(regimenPlanResponse.RegimenSections))
	} else if !regimenPlanResponse.RegimenSections[0].RegimenSteps[0].ParentId.IsValid {
		t.Fatalf("Expected the regimen step to have a parent id but it doesnt")
	} else if !regimenPlanResponse.RegimenSections[0].RegimenSteps[0].ParentId.IsValid {
		t.Fatalf("Expected the regimen step to have a parent id but it doesnt")
	}

	// now remove a section from the request
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.RegimenSections = []*common.RegimenSection{regimenPlanRequest.RegimenSections[0]}

	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	if len(regimenPlanResponse.RegimenSections) != 1 {
		t.Fatalf("Expected the number of regimen sections to be 2 but there are %d instead", len(regimenPlanResponse.RegimenSections))
	}

	// lets update a regimen step in the section
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.AllRegimenSteps[0].Text = "UPDATED 1"
	regimenPlanRequest.AllRegimenSteps[0].State = common.STATE_MODIFIED
	regimenPlanRequest.RegimenSections[0].RegimenSteps[0].Text = "UPDATED 1"
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// lets delete a regimen step
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{regimenPlanRequest.AllRegimenSteps[0]}
	regimenPlanRequest.RegimenSections = []*common.RegimenSection{}
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)
	if len(regimenPlanResponse.AllRegimenSteps) != 1 {
		t.Fatal("Should only have 1 regimen step given that we just deleted one from the list")
	}

	// lets attempt to remove the regimen step, but keep it in the regimen section.
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{}
	regimenPlanRequest.RegimenSections = []*common.RegimenSection{regimenSection}
	doctorRegimenHandler := doctor_treatment_plan.NewRegimenHandler(testData.DataApi)
	ts := httptest.NewServer(doctorRegimenHandler)
	defer ts.Close()

	requestBody, err := json.Marshal(regimenPlanRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding regimen steps: " + err.Error())
	}

	resp, err := testData.AuthPost(ts.URL, "application/json", bytes.NewBuffer(requestBody), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful request to create regimen for patient visit")
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Expected to get a bad request for when the regimen step does not exist in the regimen sections")
	}

	// get patient to start a visit

	_, treatmentPlan = test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	regimenPlan = test_integration.GetRegimenPlanForTreatmentPlan(testData, doctor, treatmentPlan.Id.Int64(), t)
	if len(regimenPlan.RegimenSections) > 0 {
		t.Fatal("There should not be any regimen sections for a new patient visit")
	}

	if len(regimenPlan.AllRegimenSteps) != 0 {
		t.Fatal("There should be no regimen steps existing globally for this doctor")
	}
}

func TestRegimenForPatientVisit_AddOnlyToPatientVisit(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// add regimen steps only to section and not to master list
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id

	regimenStep1 := &common.DoctorInstructionItem{}
	regimenStep1.Text = "Regimen Step 1"
	regimenStep1.State = common.STATE_ADDED

	regimenStep2 := &common.DoctorInstructionItem{}
	regimenStep2.Text = "Regimen Step 2"
	regimenStep2.State = common.STATE_ADDED

	regimenSection := &common.RegimenSection{}
	regimenSection.RegimenName = "morning"
	regimenSection.RegimenSteps = []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
		Text: regimenStep1.Text,
	},
	}

	regimenSection2 := &common.RegimenSection{}
	regimenSection2.RegimenName = "night"
	regimenSection2.RegimenSteps = []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
		Text: regimenStep2.Text,
	},
	}

	regimenPlanRequest.RegimenSections = []*common.RegimenSection{regimenSection, regimenSection2}
	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	if len(regimenPlanResponse.RegimenSections) != 2 {
		t.Fatalf("Expected 2 regimen sections but got %d", len(regimenPlanResponse.RegimenSections))
	} else if regimenPlanRequest.RegimenSections[0].RegimenSteps[0].ParentId.IsValid {
		t.Fatal("Expected parent id to not exist for regimen step but it does")
	} else if regimenPlanRequest.RegimenSections[1].RegimenSteps[0].ParentId.IsValid {
		t.Fatal("Expected parent id to not exist for regimen step but it does")
	}

}

func TestRegimenForPatientVisit_AddingMultipleItemsWithSameText(t *testing.T) {

	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// add multiple items with the exact same text and ensure that they all get assigned new ids
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id
	regimenPlanRequest.AllRegimenSteps = make([]*common.DoctorInstructionItem, 0)

	for i := 0; i < 5; i++ {
		regimenPlanRequest.AllRegimenSteps = append(regimenPlanRequest.AllRegimenSteps, &common.DoctorInstructionItem{
			Text:  "Regimen Step",
			State: common.STATE_ADDED,
		})

		regimenPlanRequest.RegimenSections = append(regimenPlanRequest.RegimenSections, &common.RegimenSection{
			RegimenName: "test " + strconv.Itoa(i),
			RegimenSteps: []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
				Text:  "Regimen Step",
				State: common.STATE_ADDED,
			},
			},
		})
	}

	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

}

// The purpose of this test is to ensure that we do not let the client specify text for
// items in the regimen sections that does not match up to what is indicated in the global list, if the
// linkage exists in the global list.
func TestRegimenForPatientVisit_ErrorTextDifferentForLinkedItem(t *testing.T) {

	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// add multiple items with the exact same text and ensure that they all get assigned new ids
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id
	regimenPlanRequest.AllRegimenSteps = make([]*common.DoctorInstructionItem, 0)

	for i := 0; i < 5; i++ {
		regimenPlanRequest.AllRegimenSteps = append(regimenPlanRequest.AllRegimenSteps, &common.DoctorInstructionItem{
			Text:  "Regimen Step",
			State: common.STATE_ADDED,
		})

		regimenPlanRequest.RegimenSections = append(regimenPlanRequest.RegimenSections, &common.RegimenSection{
			RegimenName: "test " + strconv.Itoa(i),
			RegimenSteps: []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
				Text:  "Regimen Step",
				State: common.STATE_ADDED,
			},
			},
		})
	}

	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// all steps in the response should have a parent id
	for i := 0; i < 5; i++ {
		parentId := regimenPlanResponse.RegimenSections[i].RegimenSteps[0].ParentId
		if !parentId.IsValid || parentId.Int64() == 0 {
			t.Fatalf("Expected parentId to exist")
		}
	}

	regimenPlanRequest = regimenPlanResponse

	// lets go ahead and update each item in the list
	for i := 0; i < 5; i++ {
		regimenPlanRequest.AllRegimenSteps[i].Text = "Updated Regimen Step"
		regimenPlanRequest.AllRegimenSteps[i].State = common.STATE_MODIFIED

		// text cannot be different given that the parent id maps to an item in the global list so this should error out
		regimenPlanRequest.RegimenSections[i].RegimenSteps[0].Text = "Updated Regimen Step " + strconv.Itoa(i)
		regimenPlanRequest.RegimenSections[i].RegimenSteps[0].State = common.STATE_MODIFIED
	}

	doctorRegimenHandler := doctor_treatment_plan.NewRegimenHandler(testData.DataApi)
	ts := httptest.NewServer(doctorRegimenHandler)
	defer ts.Close()

	requestBody, err := json.Marshal(regimenPlanRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding regimen steps: " + err.Error())
	}

	resp, err := testData.AuthPost(ts.URL, "application/json", bytes.NewBuffer(requestBody), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful request to create regimen for patient visit")
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected to get a bad request for when the regimen step's text is different than what its linked to instead got %d", resp.StatusCode)
	}

}

func TestRegimenForPatientVisit_UpdatingMultipleItemsWithSameText(t *testing.T) {

	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// add multiple items with the exact same text and ensure that they all get assigned new ids
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id
	regimenPlanRequest.AllRegimenSteps = make([]*common.DoctorInstructionItem, 0)

	for i := 0; i < 5; i++ {
		regimenPlanRequest.AllRegimenSteps = append(regimenPlanRequest.AllRegimenSteps, &common.DoctorInstructionItem{
			Text:  "Regimen Step",
			State: common.STATE_ADDED,
		})

		regimenPlanRequest.RegimenSections = append(regimenPlanRequest.RegimenSections, &common.RegimenSection{
			RegimenName: "test " + strconv.Itoa(i),
			RegimenSteps: []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
				Text:  "Regimen Step",
				State: common.STATE_ADDED,
			},
			},
		})
	}

	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	regimenPlanRequest = regimenPlanResponse

	// lets go ahead and update each item in the list
	for i := 0; i < 5; i++ {
		regimenPlanRequest.AllRegimenSteps[i].Text = "Updated Regimen Step"
		regimenPlanRequest.AllRegimenSteps[i].State = common.STATE_MODIFIED

		regimenPlanRequest.RegimenSections[i].RegimenSteps[0].Text = "Updated Regimen Step"
		regimenPlanRequest.RegimenSections[i].RegimenSteps[0].State = common.STATE_MODIFIED
	}

	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)
}

func TestRegimenForPatientVisit_UpdatingItemLinkedToDeletedItem(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// add multiple items with the exact same text and ensure that they all get assigned new ids
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id
	regimenPlanRequest.AllRegimenSteps = make([]*common.DoctorInstructionItem, 0)

	for i := 0; i < 5; i++ {
		regimenPlanRequest.AllRegimenSteps = append(regimenPlanRequest.AllRegimenSteps, &common.DoctorInstructionItem{
			Text:  "Regimen Step",
			State: common.STATE_ADDED,
		})

		regimenPlanRequest.RegimenSections = append(regimenPlanRequest.RegimenSections, &common.RegimenSection{
			RegimenName: "test " + strconv.Itoa(i),
			RegimenSteps: []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
				Text:  "Regimen Step",
				State: common.STATE_ADDED,
			},
			},
		})
	}

	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// now lets update the global set of regimen steps in the context of another patient's visit
	_, treatmentPlan2 := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	regimenPlanResponse = test_integration.GetRegimenPlanForTreatmentPlan(testData, doctor, treatmentPlan2.Id.Int64(), t)

	// lets go ahead and delete one of the items from the regimen step
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.TreatmentPlanId = treatmentPlan2.Id
	regimenPlanRequest.AllRegimenSteps = regimenPlanRequest.AllRegimenSteps[0:4]

	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	if len(regimenPlanResponse.AllRegimenSteps) != 4 {
		t.Fatalf("Expected there to exist 4 items in the global regimen steps after deleting one of them instead got %d items ", len(regimenPlanResponse.AllRegimenSteps))
	}

	// now, lets go back to the previous patient and attempt to get the regimen plan
	regimenPlanResponse = test_integration.GetRegimenPlanForTreatmentPlan(testData, doctor, treatmentPlan.Id.Int64(), t)
	if len(regimenPlanResponse.AllRegimenSteps) != 4 && len(regimenPlanResponse.RegimenSections) != 5 {
		t.Fatalf("Expected 4 items in the global regimen steps and 5 items in the regimen sections instead got %d in global regimen list and %d items in the regimen sections", len(regimenPlanRequest.AllRegimenSteps), len(regimenPlanRequest.RegimenSections))
	}

	// now lets go ahead and try and modify the item in the regimen section
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.RegimenSections[4].RegimenSteps[0].State = common.STATE_MODIFIED
	regimenPlanRequest.TreatmentPlanId = treatmentPlan2.Id
	updatedText := "Updating text for an item linked to deleted item"
	regimenPlanRequest.RegimenSections[4].RegimenSteps[0].Text = updatedText

	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	if len(regimenPlanResponse.AllRegimenSteps) != 4 && len(regimenPlanResponse.RegimenSections) != 5 {
		t.Fatalf("Expected 4 items in the global regimen steps and 5 items in the regimen sections instead got %d in global regimen list and %d items in the regimen sections", len(regimenPlanRequest.AllRegimenSteps), len(regimenPlanRequest.RegimenSections))
	}

	if regimenPlanResponse.RegimenSections[4].RegimenSteps[0].Text != updatedText {
		t.Fatalf("Exepcted text to have updated for item linked to deleted item but it didn't")
	}

	// now lets go ahead and echo back the response to the server to ensure that it takes the list
	// as it modified back without any issue. This is essentially to ensure that it passes the validation
	// of text being modified for an item that is no longer active in the master list
	regimenPlanRequest = regimenPlanResponse
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	// now lets go ahead and remove the item from the regimen section
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.RegimenSections = regimenPlanRequest.RegimenSections[:4]
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)
}

// The purpose of this test is to ensure that when regimen steps are updated,
// we are keeping track of the original step that has been modified via a source_id
func TestRegimenForPatientVisit_TrackingSourceId(t *testing.T) {

	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)

	_, treatmentPlan, doctor := setupTestForRegimenCreation(t, testData)

	// adding new regimen steps to the doctor but not to the patient visit
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id

	regimenStep1 := &common.DoctorInstructionItem{}
	regimenStep1.Text = "Regimen Step 1"
	regimenStep1.State = common.STATE_ADDED

	regimenStep2 := &common.DoctorInstructionItem{}
	regimenStep2.Text = "Regimen Step 2"
	regimenStep2.State = common.STATE_ADDED

	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
	regimenPlanResponse := test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	if len(regimenPlanResponse.RegimenSections) > 0 {
		t.Fatal("Regimen section should not exist even though regimen steps were created by doctor")
	}

	// keep track of the source ids of both steps
	sourceId1 := regimenPlanResponse.AllRegimenSteps[0].Id.Int64()
	sourceId2 := regimenPlanResponse.AllRegimenSteps[1].Id.Int64()

	// lets update both steps
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.AllRegimenSteps[0].State = common.STATE_MODIFIED
	regimenPlanRequest.AllRegimenSteps[0].Text = "Updated step 1"
	regimenPlanRequest.AllRegimenSteps[1].State = common.STATE_MODIFIED
	regimenPlanRequest.AllRegimenSteps[1].Text = "Updated step 2"
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// the source id of the two returned steps should match the source id of the original steps
	var updatedItemSourceId1, updatedItemSourceId2 sql.NullInt64
	if err := testData.DB.QueryRow(`select source_id from dr_regimen_step where id=?`, regimenPlanResponse.AllRegimenSteps[0].Id.Int64()).Scan(&updatedItemSourceId1); err != nil {
		t.Fatalf("Expected the query to get source_id to succeed instead it failed: %s", err)
	}

	if updatedItemSourceId1.Int64 != sourceId1 {
		t.Fatalf("Expected the sourceId retrieved from the updated item (%d) to match the id of the original item (%d)", updatedItemSourceId1.Int64, sourceId1)
	}

	if err := testData.DB.QueryRow(`select source_id from dr_regimen_step where id=?`, regimenPlanResponse.AllRegimenSteps[1].Id.Int64()).Scan(&updatedItemSourceId2); err != nil {
		t.Fatalf("Expected the query to get source_id to succeed instead it failed: %s", err)
	}

	if updatedItemSourceId2.Int64 != sourceId2 {
		t.Fatalf("Expected the sourceId retrieved from the updated item (%d) to match the id of the original item (%d)", updatedItemSourceId2.Int64, sourceId2)
	}

	// lets update again and the source id should still match
	regimenPlanRequest = regimenPlanResponse
	regimenPlanRequest.AllRegimenSteps[0].State = common.STATE_MODIFIED
	regimenPlanRequest.AllRegimenSteps[0].Text = "Updated again step 1"
	regimenPlanRequest.AllRegimenSteps[1].State = common.STATE_MODIFIED
	regimenPlanRequest.AllRegimenSteps[1].Text = "Updated again step 2"
	regimenPlanResponse = test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	test_integration.ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// the source id of the two returned steps should match the source id of the original steps
	if err := testData.DB.QueryRow(`select source_id from dr_regimen_step where id=?`, regimenPlanResponse.AllRegimenSteps[0].Id.Int64()).Scan(&updatedItemSourceId1); err != nil {
		t.Fatalf("Expected the query to get source_id to succeed instead it failed: %s", err)
	}

	if updatedItemSourceId1.Int64 != sourceId1 {
		t.Fatalf("Expected the sourceId retrieved from the updated item (%d) to match the id of the original item (%d)", updatedItemSourceId1.Int64, sourceId1)
	}

	if err := testData.DB.QueryRow(`select source_id from dr_regimen_step where id=?`, regimenPlanResponse.AllRegimenSteps[1].Id.Int64()).Scan(&updatedItemSourceId2); err != nil {
		t.Fatalf("Expected the query to get source_id to succeed instead it failed: %s", err)
	}

	if updatedItemSourceId2.Int64 != sourceId2 {
		t.Fatalf("Expected the sourceId retrieved from the updated item (%d) to match the id of the original item (%d)", updatedItemSourceId2.Int64, sourceId2)
	}

}

func setupTestForRegimenCreation(t *testing.T, testData *test_integration.TestData) (*patient_visit.PatientVisitResponse, *common.DoctorTreatmentPlan, *common.Doctor) {
	// get the current primary doctor
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	return patientVisitResponse, treatmentPlan, doctor
}
