package test_treatment_plan

import (
	"bytes"
	"encoding/json"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/test/test_integration"
	"net/http"
	"net/http/httptest"
	"testing"
)

// This test is to ensure that treatment plans can be versioned
// and that the content source and the parent are created as expected
func TestVersionTreatmentPlan_NewTP(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit and doctor to pick treatment plan
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	patientId, err := testData.DataApi.GetPatientIdFromPatientVisitId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// submit treatment plan
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	// now try to start a new treatment plan that is a version of the previous one
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, nil, doctor, testData, t)

	if tpResponse.TreatmentPlan.Id.Int64() == treatmentPlan.Id.Int64() {
		t.Fatal("Expected treatment plan to be different given that it was just versioned")
	}

	currentTreatmentPlan, err := testData.DataApi.GetTreatmentPlan(tpResponse.TreatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatal(err)
	}

	// the first treatment plan should be the parent of this treatment plan
	if currentTreatmentPlan.Parent.ParentType != common.TPParentTypeTreatmentPlan ||
		currentTreatmentPlan.Parent.ParentId.Int64() != treatmentPlan.Id.Int64() {
		t.Fatalf("expected treatment plan id %d to be the parent of treatment plan id %d but it wasnt", treatmentPlan.Id.Int64(), currentTreatmentPlan.Id.Int64())
	}

	// there should be no content source for this treatment plan
	if currentTreatmentPlan.ContentSource != nil {
		t.Fatal("Expected no content source for this treatment plan")
	}

	// there should be no treatments, regimen or advice
	if len(currentTreatmentPlan.TreatmentList.Treatments) > 0 {
		t.Fatalf("Expected no treatments isntead got %d", len(currentTreatmentPlan.TreatmentList.Treatments))
	} else if len(currentTreatmentPlan.RegimenPlan.RegimenSections) > 0 {
		t.Fatalf("Expected no regimen sections instead got %d", len(currentTreatmentPlan.RegimenPlan.RegimenSections))
	} else if len(currentTreatmentPlan.Advice.SelectedAdvicePoints) > 0 {
		t.Fatalf("Expected no advice points instead got %d", len(currentTreatmentPlan.Advice.SelectedAdvicePoints))
	}

	// should get back 1 treatment plan in draft and the other one active
	treatmentPlanResponse := test_integration.GetListOfTreatmentPlansForPatient(patientId, doctor.AccountId.Int64(), testData, t)
	if len(treatmentPlanResponse.DraftTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treamtent plan in draft instead got %d", len(treatmentPlanResponse.DraftTreatmentPlans))
	} else if len(treatmentPlanResponse.ActiveTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treatment plan in active mode instead got %d", len(treatmentPlanResponse.ActiveTreatmentPlans))
	}

	// now go ahead and submit the treatment plan
	test_integration.SubmitPatientVisitBackToPatient(currentTreatmentPlan.Id.Int64(), doctor, testData, t)

	// the new versioned treatment plan should be active and the previous one inactice
	treatmentPlanResponse = test_integration.GetListOfTreatmentPlansForPatient(patientId, doctor.AccountId.Int64(), testData, t)
	if len(treatmentPlanResponse.DraftTreatmentPlans) != 0 {
		t.Fatalf("Expected 0 treamtent plans in draft instead got %d", len(treatmentPlanResponse.DraftTreatmentPlans))
	} else if len(treatmentPlanResponse.ActiveTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treatment plan in active mode instead got %d", len(treatmentPlanResponse.ActiveTreatmentPlans))
	} else if treatmentPlanResponse.ActiveTreatmentPlans[0].Id.Int64() != currentTreatmentPlan.Id.Int64() {
		t.Fatalf("Expected treatment plan id %d instead got %d", currentTreatmentPlan.Id.Int64(), treatmentPlanResponse.ActiveTreatmentPlans[0].Id.Int64())
	} else if len(treatmentPlanResponse.InactiveTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 inactive treatment plan instead got %d", len(treatmentPlanResponse.InactiveTreatmentPlans))
	} else if treatmentPlanResponse.InactiveTreatmentPlans[0].Id.Int64() != treatmentPlan.Id.Int64() {
		t.Fatalf("Expected inactive treatment plan to be %d instead it was %d", treatmentPlan.Id.Int64(), treatmentPlanResponse.InactiveTreatmentPlans[0].Id.Int64())
	}
}

// This test is to ensure that we can start with a previous treatment plan
// when versioning a treatment plan
func TestVersionTreatmentPlan_PrevTP(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit and doctor to pick treatment plan
	_, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// add treatments
	treatment1 := &common.Treatment{
		DrugInternalName: "Advil",
		TreatmentPlanId:  treatmentPlan.Id,
		DosageStrength:   "10 mg",
		DispenseValue:    1,
		DispenseUnitId:   encoding.NewObjectId(26),
		NumberRefills: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 1,
		},
		SubstitutionsAllowed: true,
		DaysSupply: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 1,
		},
		OTC:                 true,
		PharmacyNotes:       "testing pharmacy notes",
		PatientInstructions: "patient instructions",
		DrugDBIds: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}
	test_integration.AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	// add advice
	advicePoint1 := &common.DoctorInstructionItem{Text: "Advice point 1", State: common.STATE_ADDED}
	advicePoint2 := &common.DoctorInstructionItem{Text: "Advice point 2", State: common.STATE_ADDED}
	doctorAdviceRequest := &common.Advice{}
	doctorAdviceRequest.AllAdvicePoints = []*common.DoctorInstructionItem{advicePoint1, advicePoint2}
	doctorAdviceRequest.SelectedAdvicePoints = doctorAdviceRequest.AllAdvicePoints
	doctorAdviceRequest.TreatmentPlanId = treatmentPlan.Id
	test_integration.UpdateAdvicePointsForPatientVisit(doctorAdviceRequest, testData, doctor, t)

	// add regimen steps
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = treatmentPlan.Id
	regimenStep1 := &common.DoctorInstructionItem{}
	regimenStep1.Text = "Regimen Step 1"
	regimenStep1.State = common.STATE_ADDED
	regimenStep2 := &common.DoctorInstructionItem{}
	regimenStep2.Text = "Regimen Step 2"
	regimenStep2.State = common.STATE_ADDED
	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
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
	test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	// submit the treatment plan
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	// now try to start a new treatment plan that is a version of the previous one
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, &common.TreatmentPlanContentSource{
		ContentSourceType: common.TPContentSourceTypeTreatmentPlan,
		ContentSourceId:   treatmentPlan.Id,
	}, doctor, testData, t)

	if tpResponse.TreatmentPlan.Id.Int64() == treatmentPlan.Id.Int64() {
		t.Fatal("Expected treatment plan to be different given that it was just versioned")
	}

	// this treatment plan should have the same contents as the treatment plan picked
	// as the content source
	if err != nil {
		t.Fatal(err)
	} else if len(tpResponse.TreatmentPlan.TreatmentList.Treatments) != 1 {
		t.Fatalf("Expected 1 treatment instead got %d", len(tpResponse.TreatmentPlan.TreatmentList.Treatments))
	} else if tpResponse.TreatmentPlan.TreatmentList.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Expected the treatment list to be uncommitted but it wasnt")
	} else if len(tpResponse.TreatmentPlan.Advice.SelectedAdvicePoints) != 2 {
		t.Fatalf("Expected 2 advice poitns instead got %d", len(tpResponse.TreatmentPlan.Advice.SelectedAdvicePoints))
	} else if tpResponse.TreatmentPlan.Advice.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Expected the advice to be uncommitted but it wasnt")
	} else if len(tpResponse.TreatmentPlan.RegimenPlan.RegimenSections) != 2 {
		t.Fatalf("Expected 2 regimen sections instead got %d", len(tpResponse.TreatmentPlan.RegimenPlan.RegimenSections))
	} else if tpResponse.TreatmentPlan.RegimenPlan.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Expected the regimen plan to be uncommitted but it wasnt")
	}

	// ensure that the content source is the treatment plan
	if tpResponse.TreatmentPlan.ContentSource == nil ||
		tpResponse.TreatmentPlan.ContentSource.ContentSourceType != common.TPContentSourceTypeTreatmentPlan {
		t.Fatalf("Expected the content source to be treatment plan but it wasnt")
	} else if tpResponse.TreatmentPlan.ContentSource.HasDeviated {
		t.Fatal("Didn't expect the treatment plan to deviate from the content source yet")
	}

	// now try to modify the treatment and it should mark the treatment plan as having deviated from the source
	treatment1.DispenseValue = encoding.HighPrecisionFloat64(21151)
	test_integration.AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(), tpResponse.TreatmentPlan.Id.Int64(), t)

	currentTreatmentPlan, err := testData.DataApi.GetAbridgedTreatmentPlan(tpResponse.TreatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatal(err)
	}

	if !currentTreatmentPlan.ContentSource.HasDeviated {
		t.Fatal("Expected the treatment plan to have deviated from the content source but it hasnt")
	}
}

// This test is to ensure that we can create multiple versions of treatment plans
// and submit them with no problem
func TestVersionTreatmentPlan_MultipleRevs(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit and doctor to pick treatment plan
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	patientId, err := testData.DataApi.GetPatientIdFromPatientVisitId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// submit the treatment plan
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	// now try to start a new treatment plan from scratch
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, nil, doctor, testData, t)

	// add treatments
	treatment1 := &common.Treatment{
		DrugInternalName: "Advil",
		TreatmentPlanId:  tpResponse.TreatmentPlan.Id,
		DosageStrength:   "10 mg",
		DispenseValue:    1,
		DispenseUnitId:   encoding.NewObjectId(26),
		NumberRefills: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 1,
		},
		SubstitutionsAllowed: true,
		DaysSupply: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 1,
		},
		OTC:                 true,
		PharmacyNotes:       "testing pharmacy notes",
		PatientInstructions: "patient instructions",
		DrugDBIds: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}
	test_integration.AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(), tpResponse.TreatmentPlan.Id.Int64(), t)

	// add advice
	advicePoint1 := &common.DoctorInstructionItem{Text: "Advice point 1", State: common.STATE_ADDED}
	advicePoint2 := &common.DoctorInstructionItem{Text: "Advice point 2", State: common.STATE_ADDED}
	doctorAdviceRequest := &common.Advice{}
	doctorAdviceRequest.AllAdvicePoints = []*common.DoctorInstructionItem{advicePoint1, advicePoint2}
	doctorAdviceRequest.SelectedAdvicePoints = doctorAdviceRequest.AllAdvicePoints
	doctorAdviceRequest.TreatmentPlanId = tpResponse.TreatmentPlan.Id
	test_integration.UpdateAdvicePointsForPatientVisit(doctorAdviceRequest, testData, doctor, t)

	// add regimen steps
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = tpResponse.TreatmentPlan.Id
	regimenStep1 := &common.DoctorInstructionItem{}
	regimenStep1.Text = "Regimen Step 1"
	regimenStep1.State = common.STATE_ADDED
	regimenStep2 := &common.DoctorInstructionItem{}
	regimenStep2.Text = "Regimen Step 2"
	regimenStep2.State = common.STATE_ADDED
	regimenPlanRequest.AllRegimenSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
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
	test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	// submit the treatment plan
	test_integration.SubmitPatientVisitBackToPatient(tpResponse.TreatmentPlan.Id.Int64(), doctor, testData, t)

	// start yet another treatment plan, this time from the previous treatment plan
	tpResponse2 := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   tpResponse.TreatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, &common.TreatmentPlanContentSource{
		ContentSourceType: common.TPContentSourceTypeTreatmentPlan,
		ContentSourceId:   tpResponse.TreatmentPlan.Id,
	}, doctor, testData, t)

	parentTreatmentPlan, err := testData.DataApi.GetTreatmentPlan(tpResponse.TreatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatal(err)
	}

	if !parentTreatmentPlan.Equals(tpResponse2.TreatmentPlan) {
		t.Fatal("Expected the parent and the newly versioned treatment plan to be equal but they are not")
	}

	treatmentPlanResponse := test_integration.GetListOfTreatmentPlansForPatient(patientId, doctor.AccountId.Int64(), testData, t)
	if len(treatmentPlanResponse.DraftTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treamtent plans in draft instead got %d", len(treatmentPlanResponse.DraftTreatmentPlans))
	} else if len(treatmentPlanResponse.ActiveTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treatment plan in active mode instead got %d", len(treatmentPlanResponse.ActiveTreatmentPlans))
	} else if len(treatmentPlanResponse.InactiveTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 inactive treatment plan instead got %d", len(treatmentPlanResponse.InactiveTreatmentPlans))
	}
}

// This test is to ensure that we don't allow versioning from an inactive treatment plan
func TestVersionTreatmentPlan_PickingFromInactiveTP(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit and doctor to pick treatment plan
	_, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// submit the treatment plan
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	// now try to start a new treatment plan from scratch
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, nil, doctor, testData, t)

	// submit the treatment plan
	test_integration.SubmitPatientVisitBackToPatient(tpResponse.TreatmentPlan.Id.Int64(), doctor, testData, t)

	// attempt to start yet another treatment plan but this time trying to pick from
	// an inactive treatment plan. this should fail
	doctorTretmentPlanHandler := doctor_treatment_plan.NewDoctorTreatmentPlanHandler(testData.DataApi, nil, nil, false)
	doctorServer := httptest.NewServer(doctorTretmentPlanHandler)
	defer doctorServer.Close()

	jsonData, err := json.Marshal(&doctor_treatment_plan.PickTreatmentPlanRequestData{
		TPParent: &common.TreatmentPlanParent{
			ParentId:   treatmentPlan.Id,
			ParentType: common.TPParentTypeTreatmentPlan,
		},
	})

	res, err := testData.AuthPut(doctorServer.URL, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal(err)
	} else if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected %d but got %d", http.StatusBadRequest, res.StatusCode)
	}

}

// This test is to ensure that doctor can pick from a favorite treatment plan to
// version a treatment plan
func TestVersionTreatmentPlan_PickFromFTP(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit and doctor to pick treatment plan
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	favoriteTreatmentPlan := test_integration.CreateFavoriteTreatmentPlan(patientVisitResponse.PatientVisitId, treatmentPlan.Id.Int64(), testData, doctor, t)

	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	// now try to start a new treatment plan from an FTP
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, &common.TreatmentPlanContentSource{
		ContentSourceType: common.TPContentSourceTypeFTP,
		ContentSourceId:   favoriteTreatmentPlan.Id,
	}, doctor, testData, t)

	if !favoriteTreatmentPlan.EqualsDoctorTreatmentPlan(tpResponse.TreatmentPlan) {
		t.Fatal("Expected contents of favorite treatment plan to be the same as that of the treatment plan")
	}
}

// This test is to ensure that the most active treatment plan is shared with the patient
func TestVersionTreatmentPlan_TPForPatient(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)
	patientId, err := testData.DataApi.GetPatientIdFromPatientVisitId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// version treatment plan
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, nil, doctor, testData, t)

	// submit version to make it active
	test_integration.SubmitPatientVisitBackToPatient(tpResponse.TreatmentPlan.Id.Int64(), doctor, testData, t)

	treatmentPlanForPatient, err := testData.DataApi.GetActiveTreatmentPlanForPatient(patientId)
	if err != nil {
		t.Fatal(err)
	}

	if treatmentPlanForPatient.Id.Int64() != tpResponse.TreatmentPlan.Id.Int64() {
		t.Fatal("Expected the latest treatment plan to be the one considered active for patient but it wasnt the case")
	}
}

// This test is to ensure that we don't deviate the treatment plan
// unless the data has actually changed
func TestVersionTreatmentPlan_DeviationFromFTP(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// get patient to start a visit and doctor to pick treatment plan
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	favoriteTreatmentPlan := test_integration.CreateFavoriteTreatmentPlan(patientVisitResponse.PatientVisitId, treatmentPlan.Id.Int64(), testData, doctor, t)

	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	// now try to start a new treatment plan from an FTP
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, &common.TreatmentPlanContentSource{
		ContentSourceType: common.TPContentSourceTypeFTP,
		ContentSourceId:   favoriteTreatmentPlan.Id,
	}, doctor, testData, t)

	// now, submit the exact same treatments to commit it
	test_integration.AddAndGetTreatmentsForPatientVisit(testData, favoriteTreatmentPlan.TreatmentList.Treatments, doctor.AccountId.Int64(), tpResponse.TreatmentPlan.Id.Int64(), t)

	currentTreatmentPlan, err := testData.DataApi.GetAbridgedTreatmentPlan(tpResponse.TreatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatal(err)
	} else if currentTreatmentPlan.ContentSource.HasDeviated {
		t.Fatal("Did not expect treatment plan to deviate from source but it did")
	}

	// submit the exact same regimen
	regimenPlanRequest := &common.RegimenPlan{}
	regimenPlanRequest.TreatmentPlanId = tpResponse.TreatmentPlan.Id
	regimenPlanRequest.RegimenSections = favoriteTreatmentPlan.RegimenPlan.RegimenSections
	test_integration.CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	currentTreatmentPlan, err = testData.DataApi.GetAbridgedTreatmentPlan(tpResponse.TreatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatal(err)
	} else if currentTreatmentPlan.ContentSource.HasDeviated {
		t.Fatal("Did not expect treatment plan to deviate from source but it did")
	}

	// submit the exact same advice
	doctorAdviceRequest := &common.Advice{}
	doctorAdviceRequest.SelectedAdvicePoints = favoriteTreatmentPlan.Advice.SelectedAdvicePoints
	doctorAdviceRequest.TreatmentPlanId = tpResponse.TreatmentPlan.Id
	test_integration.UpdateAdvicePointsForPatientVisit(doctorAdviceRequest, testData, doctor, t)

	currentTreatmentPlan, err = testData.DataApi.GetAbridgedTreatmentPlan(tpResponse.TreatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatal(err)
	} else if currentTreatmentPlan.ContentSource.HasDeviated {
		t.Fatal("Did not expect treatment plan to deviate from source but it did")
	}
}

func TestVersionTreatmentPlan_DeleteOlderDraft(t *testing.T) {
	testData := test_integration.SetupIntegrationTest(t)
	defer test_integration.TearDownIntegrationTest(t, testData)
	doctorId := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}
	patientVisitResponse, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)
	patientId, err := testData.DataApi.GetPatientIdFromPatientVisitId(patientVisitResponse.PatientVisitId)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// attempt to version treatment plan
	tpResponse := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, nil, doctor, testData, t)

	// attempt to version again
	tpResponse2 := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, nil, doctor, testData, t)

	// two treatment plans should be different given that older one should be deleted
	if tpResponse.TreatmentPlan.Id.Int64() == tpResponse2.TreatmentPlan.Id.Int64() {
		t.Fatal("Expected a new treatment plan to be created if the user attempts to pick again")
	}

	// attempt to create FTP under the new versioned treatment plan
	favoriteTreatmentPlan := test_integration.CreateFavoriteTreatmentPlan(patientVisitResponse.PatientVisitId, tpResponse2.TreatmentPlan.Id.Int64(), testData, doctor, t)

	// attempt to start a new TP now with this FTP
	tpResponse3 := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentId:   treatmentPlan.Id,
		ParentType: common.TPParentTypeTreatmentPlan,
	}, &common.TreatmentPlanContentSource{
		ContentSourceType: common.TPContentSourceTypeFTP,
		ContentSourceId:   favoriteTreatmentPlan.Id,
	}, doctor, testData, t)

	if tpResponse3.TreatmentPlan.Id.Int64() == tpResponse2.TreatmentPlan.Id.Int64() {
		t.Fatal("Expected the newly created treatment plan to have a different id than the previous one")
	}

	// there should only exist 1 draft and 1 active treatment plan
	treatmentPlanResponse := test_integration.GetListOfTreatmentPlansForPatient(patientId, doctor.AccountId.Int64(), testData, t)
	if len(treatmentPlanResponse.DraftTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treamtent plans in draft instead got %d", len(treatmentPlanResponse.DraftTreatmentPlans))
	} else if len(treatmentPlanResponse.ActiveTreatmentPlans) != 1 {
		t.Fatalf("Expected 1 treatment plan in active mode instead got %d", len(treatmentPlanResponse.ActiveTreatmentPlans))
	} else if len(treatmentPlanResponse.InactiveTreatmentPlans) != 0 {
		t.Fatalf("Expected 0 inactive treatment plans instead got %d", len(treatmentPlanResponse.InactiveTreatmentPlans))
	}

}
