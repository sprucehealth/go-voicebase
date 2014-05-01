package integration

import (
	"bytes"
	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/encoding"
	"carefront/libs/erx"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestFavoriteTreatmentPlan(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctorId := getDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	patientVisitResponse, _ := signupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)

	favoriteTreatmentPlan := createFavoriteTreatmentPlan(patientVisitResponse.PatientVisitId, testData, doctor, t)

	originalRegimenPlan := favoriteTreatmentPlan.RegimenPlan
	originalAdvice := favoriteTreatmentPlan.Advice

	// now lets go ahead and update the favorite treatment plan

	updatedName := "Updating name"
	favoriteTreatmentPlan.Name = updatedName
	favoriteTreatmentPlan.RegimenPlan.RegimenSections = []*common.RegimenSection{favoriteTreatmentPlan.RegimenPlan.RegimenSections[0]}
	favoriteTreatmentPlan.Advice.SelectedAdvicePoints = []*common.DoctorInstructionItem{favoriteTreatmentPlan.Advice.SelectedAdvicePoints[0]}

	requestData := &apiservice.DoctorFavoriteTreatmentPlansRequestData{}
	requestData.FavoriteTreatmentPlan = favoriteTreatmentPlan
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Unable to marshal json data: %s", err)
	}

	ts := httptest.NewServer(&apiservice.DoctorFavoriteTreatmentPlansHandler{
		DataApi: testData.DataApi,
	})
	defer ts.Close()

	responseData := &apiservice.DoctorFavoriteTreatmentPlansResponseData{}
	resp, err := authPut(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to make call to update favorite treatment plan %s", err)
	} else if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to decode response body into json object %s", err)
	} else if responseData.FavoriteTreatmentPlan == nil {
		t.Fatalf("Expected 1 favorite treatment plan to be returned instead got back %d", len(responseData.FavoriteTreatmentPlans))
	} else if len(responseData.FavoriteTreatmentPlan.RegimenPlan.RegimenSections) != 1 {
		t.Fatalf("Expected 1 section in the regimen plan instead got %d", len(responseData.FavoriteTreatmentPlan.RegimenPlan.RegimenSections))
	} else if len(responseData.FavoriteTreatmentPlan.Advice.SelectedAdvicePoints) != 1 {
		t.Fatalf("Expected 1 section in the advice instead got %d", len(responseData.FavoriteTreatmentPlan.Advice.SelectedAdvicePoints))
	} else if responseData.FavoriteTreatmentPlan.Name != updatedName {
		t.Fatalf("Expected name of favorite treatment plan to be %s instead got %s", updatedName, responseData.FavoriteTreatmentPlan.Name)
	}

	CheckSuccessfulStatusCode(resp, "unable to make call to update favorite treatment plan", t)

	// lets go ahead and add another favorited treatment
	favoriteTreatmentPlan2 := &common.FavoriteTreatmentPlan{
		Name: "Test Treatment Plan #2",
		Treatments: []*common.Treatment{&common.Treatment{
			DrugDBIds: map[string]string{
				erx.LexiDrugSynId:     "1234",
				erx.LexiGenProductId:  "12345",
				erx.LexiSynonymTypeId: "123556",
				erx.NDC:               "2415",
			},
			DrugName:                "Teting (This - Drug)",
			DosageStrength:          "10 mg",
			DispenseValue:           5,
			DispenseUnitDescription: "Tablet",
			DispenseUnitId:          encoding.NewObjectId(19),
			NumberRefills: encoding.NullInt64{
				IsValid:    true,
				Int64Value: 5,
			},
			SubstitutionsAllowed: false,
			DaysSupply: encoding.NullInt64{
				IsValid:    true,
				Int64Value: 5,
			},
			PatientInstructions: "Take once daily",
			OTC:                 false,
		},
		},
		RegimenPlan: originalRegimenPlan,
		Advice:      originalAdvice,
	}

	requestData.FavoriteTreatmentPlan = favoriteTreatmentPlan2
	jsonData, err = json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Unable to marshal favorited treatment plan %s", err)
	}

	resp, err = authPost(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to add another favorite treatment plan %s", err)
	}

	CheckSuccessfulStatusCode(resp, "unable to add another favorite treatment plan", t)

	resp, err = authGet(ts.URL, doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unabke to get list of favorite treatment plans %s", err)
	} else if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into a list of favorite treatment plans %s", err)
	} else if len(responseData.FavoriteTreatmentPlans) != 2 {
		t.Fatalf("Expected 2 favorite treatment plans instead got %d", len(responseData.FavoriteTreatmentPlans))
	} else if len(responseData.FavoriteTreatmentPlans[0].RegimenPlan.RegimenSections) != 1 {
		t.Fatalf("Expected favorite treatment plan to have 1 regimen section")
	} else if len(responseData.FavoriteTreatmentPlans[0].Advice.SelectedAdvicePoints) != 1 {
		t.Fatalf("Expected favorite treatment plan to have 1 advice point")
	} else if len(responseData.FavoriteTreatmentPlans[1].RegimenPlan.RegimenSections) != 1 {
		t.Fatalf("Expected favorite treatment plan to have 2 regimen sections")
	} else if len(responseData.FavoriteTreatmentPlans[1].Advice.SelectedAdvicePoints) != 1 {
		t.Fatalf("Expected favorite treatment plan to have 2 advice points")
	}

	CheckSuccessfulStatusCode(resp, "Unable to get list of favorite treatment plans for doctor", t)

	// lets go ahead and delete favorite treatment plan
	params := url.Values{}
	params.Set("favorite_treatment_plan_id", strconv.FormatInt(responseData.FavoriteTreatmentPlans[0].Id.Int64(), 10))
	resp, err = authDelete(ts.URL+"?"+params.Encode(), "application/x-www-form-urlencoded", nil, doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to delete favorite treatment plan %s", err)
	}

	CheckSuccessfulStatusCode(resp, "Unable to delete favorite treatment plan", t)
}

func TestFavoriteTreatmentPlan_PickingAFavoriteTreatmentPlan(t *testing.T) {
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctorId := getDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	patientVisitResponse, _ := signupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)

	// create a favorite treatment plan
	favoriteTreamentPlan := createFavoriteTreatmentPlan(patientVisitResponse.PatientVisitId, testData, doctor, t)

	// lets attempt to get the treatment plan for the patient visit
	ts := httptest.NewServer(&apiservice.DoctorTreatmentPlanHandler{
		DataApi: testData.DataApi,
	})
	defer ts.Close()

	responseData := &apiservice.DoctorTreatmentPlanResponse{}
	if resp, err := authGet(ts.URL+"?patient_visit_id="+strconv.FormatInt(patientVisitResponse.PatientVisitId, 10), doctor.AccountId.Int64()); err != nil {
		t.Fatalf("Unable to make call to get treatment plan for patient visit")
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d response for getting treatment plan instead got %d", http.StatusOK, resp.StatusCode)
	} else if json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into struct %s", err)
	} else if responseData.TreatmentPlan == nil {
		t.Fatalf("Expected treatment plan to exist")
	} else if responseData.TreatmentPlan.TreatmentList != nil && len(responseData.TreatmentPlan.TreatmentList.Treatments) != 0 {
		t.Fatalf("Expected there to exist no treatments in treatment plan")
	} else if responseData.TreatmentPlan.RegimenPlan != nil && len(responseData.TreatmentPlan.RegimenPlan.RegimenSections) != 0 {
		t.Fatalf("Expected regimen to not exist for treatment plan instead we have %d regimen sections", len(responseData.TreatmentPlan.RegimenPlan.RegimenSections))
	} else if len(responseData.TreatmentPlan.RegimenPlan.AllRegimenSteps) == 0 {
		t.Fatalf("Expected regimen steps to exist given that they were created to create the treatment plan")
	} else if responseData.TreatmentPlan.Advice != nil && len(responseData.TreatmentPlan.Advice.SelectedAdvicePoints) != 0 {
		t.Fatalf("Expected there to exist no advice points for treatment plan")
	} else if len(responseData.TreatmentPlan.Advice.AllAdvicePoints) == 0 {
		t.Fatalf("Expected there to exist advice points given that some were created when creating the favorite treatment plan")
	}

	// now lets attempt to pick the added favorite treatment plan and compare the two again
	// this time the treatment plan should be populated with data from the favorite treatment plan
	responseData = pickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, favoriteTreamentPlan, testData, t)
	if responseData.TreatmentPlan == nil {
		t.Fatalf("Expected treatment plan to exist")
	} else if responseData.TreatmentPlan.TreatmentList != nil && len(responseData.TreatmentPlan.TreatmentList.Treatments) != 1 {
		t.Fatalf("Expected there to exist no treatments in treatment plan")
	} else if responseData.TreatmentPlan.TreatmentList.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Status should indicate UNCOMMITTED for treatment section when the doctor has not committed the section")
	} else if responseData.TreatmentPlan.RegimenPlan != nil && len(responseData.TreatmentPlan.RegimenPlan.RegimenSections) != 2 {
		t.Fatalf("Expected regimen to not exist for treatment plan instead we have %d regimen sections", len(responseData.TreatmentPlan.RegimenPlan.RegimenSections))
	} else if len(responseData.TreatmentPlan.RegimenPlan.AllRegimenSteps) != 2 {
		t.Fatalf("Expected there to exist 2 regimen steps in the master list instead got %d", len(responseData.TreatmentPlan.RegimenPlan.AllRegimenSteps))
	} else if responseData.TreatmentPlan.RegimenPlan.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Status should indicate UNCOMMITTED for regimen plan when the doctor has not committed the section")
	} else if responseData.TreatmentPlan.Advice != nil && len(responseData.TreatmentPlan.Advice.SelectedAdvicePoints) != 2 {
		t.Fatalf("Expected there to exist no advice points for treatment plan")
	} else if len(responseData.TreatmentPlan.Advice.AllAdvicePoints) != 2 {
		t.Fatalf("Expected there to exist 2 advice points in the master list instead got %d", len(responseData.TreatmentPlan.Advice.AllAdvicePoints))
	} else if responseData.TreatmentPlan.Advice.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Status should indicate UNCOMMITTED for advice when the doctor has not committed the section")
	}

	var count int64
	if err := testData.DB.QueryRow(`select count(*) from treatment_plan where patient_visit_id = ?`, patientVisitResponse.PatientVisitId).Scan(&count); err != nil {
		t.Fatalf("Unable to query database to get number of treatment plans for patient visit: %s", err)
	} else if count != 1 {
		t.Fatalf("Expected 1 treatment plan for patient visit instead got %d", count)
	}

}

func createFavoriteTreatmentPlan(patientVisitId int64, testData TestData, doctor *common.Doctor, t *testing.T) *common.FavoriteTreatmentPlan {

	// lets submit a regimen plan for this patient
	// reason we do this is because the regimen steps have to exist before treatment plan can be favorited,
	// and the only way we can create regimen steps today is in the context of a patient visit
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.PatientVisitId = encoding.NewObjectId(patientVisitId)

	regimenStep1 := &common.DoctorInstructionItem{}
	regimenStep1.Text = "Regimen Step 1"
	regimenStep1.State = common.STATE_ADDED

	regimenStep2 := &common.DoctorInstructionItem{}
	regimenStep2.Text = "Regimen Step 2"
	regimenStep2.State = common.STATE_ADDED

	regimenSection := &common.RegimenSection{}
	regimenSection.RegimenName = "morning"
	regimenSection.RegimenSteps = []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
		Text:  regimenStep1.Text,
		State: common.STATE_ADDED,
	},
	}

	regimenSection2 := &common.RegimenSection{}
	regimenSection2.RegimenName = "night"
	regimenSection2.RegimenSteps = []*common.DoctorInstructionItem{&common.DoctorInstructionItem{
		Text:  regimenStep2.Text,
		State: common.STATE_ADDED,
	},
	}

	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
	regimenPlanResponse := createRegimenPlanForPatientVisit(regimenPlanRequest, testData, doctor, t)
	validateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// lets submit advice for this patient
	// lets go ahead and add a couple of advice points
	// reason we do this is because the advice steps have to exist before treatment plan can be favorited,
	// and the only way we can create advice steps today is in the context of a patient visit
	advicePoint1 := &common.DoctorInstructionItem{Text: "Advice point 1", State: common.STATE_ADDED}
	advicePoint2 := &common.DoctorInstructionItem{Text: "Advice point 2", State: common.STATE_ADDED}

	// lets go ahead and create a request for this patient visit
	doctorAdviceRequest := &common.Advice{}
	doctorAdviceRequest.AllAdvicePoints = []*common.DoctorInstructionItem{advicePoint1, advicePoint2}
	doctorAdviceRequest.PatientVisitId = encoding.NewObjectId(patientVisitId)

	doctorAdviceResponse := updateAdvicePointsForPatientVisit(doctorAdviceRequest, testData, doctor, t)
	validateAdviceRequestAgainstResponse(doctorAdviceRequest, doctorAdviceResponse, t)

	// prepare the regimen steps and the advice points to be added into the sections
	// after the global list for each has been updated to include items.
	// the reason this is important is because favorite treatment plans require items to exist that are linked
	// from the master list
	regimenSection.RegimenSteps[0].ParentId = regimenPlanResponse.AllRegimenSteps[0].Id
	regimenSection2.RegimenSteps[0].ParentId = regimenPlanResponse.AllRegimenSteps[1].Id
	advicePoint1 = &common.DoctorInstructionItem{
		Text:     advicePoint1.Text,
		ParentId: doctorAdviceResponse.AllAdvicePoints[0].Id,
	}
	advicePoint2 = &common.DoctorInstructionItem{
		Text:     advicePoint2.Text,
		ParentId: doctorAdviceResponse.AllAdvicePoints[1].Id,
	}

	// lets add a favorite treatment plan for doctor
	favoriteTreatmentPlan := &common.FavoriteTreatmentPlan{
		Name: "Test Treatment Plan",
		Treatments: []*common.Treatment{&common.Treatment{
			DrugDBIds: map[string]string{
				erx.LexiDrugSynId:     "1234",
				erx.LexiGenProductId:  "12345",
				erx.LexiSynonymTypeId: "123556",
				erx.NDC:               "2415",
			},
			DrugName:                "Teting (This - Drug)",
			DosageStrength:          "10 mg",
			DispenseValue:           5,
			DispenseUnitDescription: "Tablet",
			DispenseUnitId:          encoding.NewObjectId(19),
			NumberRefills: encoding.NullInt64{
				IsValid:    true,
				Int64Value: 5,
			},
			SubstitutionsAllowed: false,
			DaysSupply: encoding.NullInt64{
				IsValid:    true,
				Int64Value: 5,
			},
			PatientInstructions: "Take once daily",
			OTC:                 false,
		},
		},
		RegimenPlan: &common.RegimenPlan{
			AllRegimenSteps: regimenPlanResponse.AllRegimenSteps,
			RegimenSections: []*common.RegimenSection{regimenSection, regimenSection2},
		},
		Advice: &common.Advice{
			AllAdvicePoints:      doctorAdviceResponse.AllAdvicePoints,
			SelectedAdvicePoints: []*common.DoctorInstructionItem{advicePoint1, advicePoint2},
		},
	}

	ts := httptest.NewServer(&apiservice.DoctorFavoriteTreatmentPlansHandler{
		DataApi: testData.DataApi,
	})
	defer ts.Close()

	requestData := &apiservice.DoctorFavoriteTreatmentPlansRequestData{
		FavoriteTreatmentPlan: favoriteTreatmentPlan,
	}
	jsonData, err := json.Marshal(&requestData)
	if err != nil {
		t.Fatalf("Unable to marshal json %s", err)
	}

	resp, err := authPost(ts.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to add favorite treatment plan: %s", err)
	}

	responseData := &apiservice.DoctorFavoriteTreatmentPlansResponseData{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into json %s", responseData)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 response for adding a favorite treatment plan but got %d instead", resp.StatusCode)
	} else if responseData.FavoriteTreatmentPlan == nil {
		t.Fatalf("Expected to get back the treatment plan added but got none")
	} else if responseData.FavoriteTreatmentPlan.RegimenPlan == nil || len(responseData.FavoriteTreatmentPlan.RegimenPlan.RegimenSections) != 2 {
		t.Fatalf("Expected to have a regimen plan or 2 items in the regimen section")
	} else if len(responseData.FavoriteTreatmentPlan.Advice.SelectedAdvicePoints) != 2 {
		t.Fatalf("Expected 2 items in the advice list")
	}

	return responseData.FavoriteTreatmentPlan
}
