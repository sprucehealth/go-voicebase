package test_integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice/router"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/erx"
)

func TestFavoriteTreatmentPlan(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	cli := DoctorClient(testData, t, doctorID)

	_, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	favoriteTreatmentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	originalRegimenPlan := favoriteTreatmentPlan.RegimenPlan

	// now lets go ahead and update the favorite treatment plan

	updatedName := "Updating name"
	favoriteTreatmentPlan.Name = updatedName
	favoriteTreatmentPlan.RegimenPlan.Sections = favoriteTreatmentPlan.RegimenPlan.Sections[1:]

	if ftp, err := cli.UpdateFavoriteTreatmentPlan(favoriteTreatmentPlan); err != nil {
		t.Fatal(err)
	} else if len(ftp.RegimenPlan.Sections) != 1 {
		t.Fatalf("Expected 1 section in the regimen plan instead got %d", len(ftp.RegimenPlan.Sections))
	} else if ftp.Name != updatedName {
		t.Fatalf("Expected name of favorite treatment plan to be %s instead got %s", updatedName, ftp.Name)
	}

	// lets go ahead and add another favorited treatment
	favoriteTreatmentPlan2 := &common.FavoriteTreatmentPlan{
		Name: "Test Treatment Plan #2",
		TreatmentList: &common.TreatmentList{
			Treatments: []*common.Treatment{{
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
			}},
		},
		RegimenPlan: originalRegimenPlan,
	}

	if err := cli.CreateFavoriteTreatmentPlan(favoriteTreatmentPlan2); err != nil {
		t.Fatal(err)
	}

	ftps, err := cli.ListFavoriteTreatmentPlans()
	if err != nil {
		t.Fatal(err)
	} else if len(ftps) != 2 {
		t.Fatalf("Expected 2 favorite treatment plans instead got %d", len(ftps))
	} else if len(ftps[0].RegimenPlan.Sections) != 1 {
		t.Fatalf("Expected favorite treatment plan to have 1 regimen section")
	} else if len(ftps[1].RegimenPlan.Sections) != 1 {
		t.Fatalf("Expected favorite treatment plan to have 2 regimen sections")
	}

	// lets go ahead and delete favorite treatment plan
	if err := cli.DeleteFavoriteTreatmentPlan(ftps[0].Id.Int64()); err != nil {
		t.Fatal(err)
	}
}

// This test ensures to check that after deleting a FTP, the TP that was created
// from the FTP has its content source deleted and getting the TP still works
func TestFavoriteTreatmentPlan_DeletingFTP(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}
	cli := DoctorClient(testData, t, doctorID)

	patientVisitResponse, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	favoriteTreatmentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	// lets start a new TP based on FTP
	responseData := PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentType: common.TPParentTypePatientVisit,
		ParentId:   encoding.NewObjectId(patientVisitResponse.PatientVisitId),
	}, &common.TreatmentPlanContentSource{
		Type: common.TPContentSourceTypeFTP,
		ID:   favoriteTreatmentPlan.Id,
	}, doctor, testData, t)

	// ensure that this TP has the FTP as its content source
	if responseData.TreatmentPlan.ContentSource == nil ||
		responseData.TreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		responseData.TreatmentPlan.ContentSource.ID.Int64() != favoriteTreatmentPlan.Id.Int64() {
		t.Fatal("Expected the newly created Treatment plan to have the FTP as its source")
	}

	// now lets go ahead and delete the FTP
	if err := cli.DeleteFavoriteTreatmentPlan(favoriteTreatmentPlan.Id.Int64()); err != nil {
		t.Fatal(err)
	}

	// now if we try to get the TP initially created from the FTP, the content source should not exist
	if tp, err := cli.TreatmentPlan(responseData.TreatmentPlan.Id.Int64(), false); err != nil {
		t.Fatal(err)
	} else if tp.ContentSource != nil {
		t.Fatal("Expected nil content source for treatment plan after deleting FTP from which the TP was started")
	}
}

// This test ensures that even if an FTP is deleted that was picked as content source for a TP that has been activated for a patient,
// the content source gets deleted while TP remains unaltered
func TestFavoriteTreatmentPlan_DeletingFTP_ActiveTP(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}
	cli := DoctorClient(testData, t, doctorID)

	patientVisitResponse, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	favoriteTreatmentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	// lets start a new TP based on FTP
	responseData := PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentType: common.TPParentTypePatientVisit,
		ParentId:   encoding.NewObjectId(patientVisitResponse.PatientVisitId),
	}, &common.TreatmentPlanContentSource{
		Type: common.TPContentSourceTypeFTP,
		ID:   favoriteTreatmentPlan.Id,
	}, doctor, testData, t)

	// ensure that this TP has the FTP as its content source
	if responseData.TreatmentPlan.ContentSource == nil ||
		responseData.TreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		responseData.TreatmentPlan.ContentSource.ID.Int64() != favoriteTreatmentPlan.Id.Int64() {
		t.Fatal("Expected the newly created Treatment plan to have the FTP as its source")
	}

	// submit the treatments for the TP
	AddAndGetTreatmentsForPatientVisit(testData, favoriteTreatmentPlan.TreatmentList.Treatments, doctor.AccountId.Int64(), responseData.TreatmentPlan.Id.Int64(), t)

	// submit regimen for TP
	regimenPlanRequest := &common.RegimenPlan{
		TreatmentPlanID: responseData.TreatmentPlan.Id,
		Sections:        favoriteTreatmentPlan.RegimenPlan.Sections,
	}
	CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	// submit treatment plan to patient
	SubmitPatientVisitBackToPatient(responseData.TreatmentPlan.Id.Int64(), doctor, testData, t)

	// now lets go ahead and delete the FTP
	if err := cli.DeleteFavoriteTreatmentPlan(favoriteTreatmentPlan.Id.Int64()); err != nil {
		t.Fatal(err)
	}

	// now if we try to get the TP initially created from the FTP, the content source should not exist
	if tp, err := cli.TreatmentPlan(responseData.TreatmentPlan.Id.Int64(), false); err != nil {
		t.Fatal(err)
	} else if tp.ContentSource != nil {
		t.Fatal("Expected nil content source for treatment plan after deleting FTP from which the TP was started")
	} else if !tp.IsActive() {
		t.Fatalf("Expected the treatment plan to be active but it wasnt")
	} else if !favoriteTreatmentPlan.EqualsTreatmentPlan(tp) {
		t.Fatal("Even though the FTP was deleted, the contents of the TP and FTP should still match")
	}
}

func TestFavoriteTreatmentPlan_PickingAFavoriteTreatmentPlan(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}
	cli := DoctorClient(testData, t, doctorID)

	patientVisitResponse, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// create a favorite treatment plan
	favoriteTreamentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	if tp, err := cli.TreatmentPlan(treatmentPlan.Id.Int64(), false); err != nil {
		t.Fatal(err)
	} else if tp.TreatmentList != nil && len(tp.TreatmentList.Treatments) != 0 {
		t.Fatalf("Expected there to exist no treatments in treatment plan")
	} else if tp.RegimenPlan != nil && len(tp.RegimenPlan.Sections) != 0 {
		t.Fatalf("Expected regimen to not exist for treatment plan instead we have %d regimen sections", len(tp.RegimenPlan.Sections))
	} else if len(tp.RegimenPlan.AllSteps) == 0 {
		t.Fatalf("Expected regimen steps to exist given that they were created to create the treatment plan")
	}

	// now lets attempt to pick the added favorite treatment plan and compare the two again
	// this time the treatment plan should be populated with data from the favorite treatment plan
	responseData := PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, favoriteTreamentPlan, testData, t)
	if responseData.TreatmentPlan == nil {
		t.Fatalf("Expected treatment plan to exist")
	} else if responseData.TreatmentPlan.TreatmentList != nil && len(responseData.TreatmentPlan.TreatmentList.Treatments) != 1 {
		t.Fatalf("Expected there to exist no treatments in treatment plan")
	} else if responseData.TreatmentPlan.TreatmentList.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Status should indicate UNCOMMITTED for treatment section when the doctor has not committed the section")
	} else if responseData.TreatmentPlan.RegimenPlan != nil && len(responseData.TreatmentPlan.RegimenPlan.Sections) != 2 {
		t.Fatalf("Expected regimen to not exist for treatment plan instead we have %d regimen sections", len(responseData.TreatmentPlan.RegimenPlan.Sections))
	} else if len(responseData.TreatmentPlan.RegimenPlan.AllSteps) != 2 {
		t.Fatalf("Expected there to exist 2 regimen steps in the master list instead got %d", len(responseData.TreatmentPlan.RegimenPlan.AllSteps))
	} else if responseData.TreatmentPlan.RegimenPlan.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Status should indicate UNCOMMITTED for regimen plan when the doctor has not committed the section")
	} else if !favoriteTreamentPlan.EqualsTreatmentPlan(responseData.TreatmentPlan) {
		t.Fatal("Expected the contents of the favorite treatment plan to be the same as that of the treatment plan but its not")
	}

	var count int64
	if err := testData.DB.QueryRow(`select count(*) from treatment_plan inner join treatment_plan_patient_visit_mapping on treatment_plan_id = treatment_plan.id where patient_visit_id = ?`, patientVisitResponse.PatientVisitId).Scan(&count); err != nil {
		t.Fatalf("Unable to query database to get number of treatment plans for patient visit: %s", err)
	} else if count != 1 {
		t.Fatalf("Expected 1 treatment plan for patient visit instead got %d", count)
	}
}

func TestFavoriteTreatmentPlan_CommittedStateForTreatmentPlan(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}
	cli := DoctorClient(testData, t, doctorID)

	patientVisitResponse, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// create a favorite treatment plan
	favoriteTreamentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	// pick this favorite treatment plan for the visit
	responseData := PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, favoriteTreamentPlan, testData, t)
	treatmentPlanId := responseData.TreatmentPlan.Id.Int64()
	// lets attempt to submit regimen section for patient visit
	regimenPlanRequest := &common.RegimenPlan{
		TreatmentPlanID: responseData.TreatmentPlan.Id,
		AllSteps:        favoriteTreamentPlan.RegimenPlan.AllSteps,
		Sections:        favoriteTreamentPlan.RegimenPlan.Sections,
	}
	CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	// now lets attempt to get the treatment plan for the patient visit
	// the regimen plan should indicate that it was committed while the rest of the sections
	// should continue to be in the UNCOMMITTED state
	if tp, err := cli.TreatmentPlan(treatmentPlanId, false); err != nil {
		t.Fatal(err)
	} else if tp.TreatmentList.Status != api.STATUS_UNCOMMITTED {
		t.Fatalf("Expected the status to be UNCOMMITTED for treatments")
	} else if tp.RegimenPlan.Status != api.STATUS_COMMITTED {
		t.Fatalf("Expected regimen status to not be COMMITTED")
	}

	// now lets go ahead and add a treatment to the treatment plan
	AddAndGetTreatmentsForPatientVisit(testData, favoriteTreamentPlan.TreatmentList.Treatments, doctor.AccountId.Int64(), treatmentPlanId, t)

	// now the treatment section should also indicate that it has been committed
	if tp, err := cli.TreatmentPlan(treatmentPlanId, false); err != nil {
		t.Fatal(err)
	} else if tp.TreatmentList.Status != api.STATUS_COMMITTED {
		t.Fatalf("Expected the status to be in the committed state")
	} else if tp.RegimenPlan.Status != api.STATUS_COMMITTED {
		t.Fatalf("Expected regimen status to be in the committed state")
	}
}

func TestFavoriteTreatmentPlan_BreakingMappingOnModify(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	patientVisitResponse, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// create a favorite treatment plan
	favoriteTreamentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	// pick this favorite treatment plan for the visit
	responseData := PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, favoriteTreamentPlan, testData, t)

	// lets attempt to modify and submit regimen section for patient visit
	regimenPlanRequest := &common.RegimenPlan{
		TreatmentPlanID: responseData.TreatmentPlan.Id,
		AllSteps:        favoriteTreamentPlan.RegimenPlan.AllSteps,
		Sections:        favoriteTreamentPlan.RegimenPlan.Sections[:1],
	}
	CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)

	// the regimen plan should indicate that it was committed while the rest of the sections
	// should continue to be in the UNCOMMITTED state
	params := url.Values{}
	params.Set("treatment_plan_id", strconv.FormatInt(responseData.TreatmentPlan.Id.Int64(), 10))
	params.Set("abridged", "true")
	responseData = &doctor_treatment_plan.DoctorTreatmentPlanResponse{}
	resp, err := testData.AuthGet(testData.APIServer.URL+router.DoctorTreatmentPlansURLPath+"?"+params.Encode(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to make call to get treatment plan for patient visit")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d response for getting treatment plan instead got %d", http.StatusOK, resp.StatusCode)
	} else if json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into struct %s", err)
	} else if responseData.TreatmentPlan.ContentSource == nil || responseData.TreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		responseData.TreatmentPlan.ContentSource.ID.Int64() == 0 || !responseData.TreatmentPlan.ContentSource.HasDeviated {
		t.Fatalf("Expected the treatment plan to indicate that it has deviated from the original content source (ftp) but it doesnt do so")
	}

	// lets try modfying treatments on a new treatment plan picked from favorites
	responseData = PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, favoriteTreamentPlan, testData, t)

	// lets make sure linkage exists
	if responseData.TreatmentPlan.ContentSource == nil || responseData.TreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		responseData.TreatmentPlan.ContentSource.ID.Int64() == 0 {
		t.Fatalf("Expected the treatment plan to come from a favorite treatment plan")
	} else if responseData.TreatmentPlan.ContentSource.ID.Int64() != favoriteTreamentPlan.Id.Int64() {
		t.Fatalf("Got a different favorite treatment plan linking to the treatment plan. Expected %d got %d", favoriteTreamentPlan.Id.Int64(), responseData.TreatmentPlan.ContentSource.ID.Int64())
	}

	// modify treatment
	favoriteTreamentPlan.TreatmentList.Treatments[0].DispenseValue = encoding.HighPrecisionFloat64(123.12345)
	AddAndGetTreatmentsForPatientVisit(testData, favoriteTreamentPlan.TreatmentList.Treatments, doctor.AccountId.Int64(), responseData.TreatmentPlan.Id.Int64(), t)

	// linkage should now be broken
	params.Set("treatment_plan_id", strconv.FormatInt(responseData.TreatmentPlan.Id.Int64(), 10))
	resp, err = testData.AuthGet(testData.APIServer.URL+router.DoctorTreatmentPlansURLPath+"?"+params.Encode(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to make call to get treatment plan for patient visit")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d response for getting treatment plan instead got %d", http.StatusOK, resp.StatusCode)
	} else if json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into struct %s", err)
	} else if responseData.TreatmentPlan.ContentSource == nil || responseData.TreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		responseData.TreatmentPlan.ContentSource.ID.Int64() == 0 || !responseData.TreatmentPlan.ContentSource.HasDeviated {
		t.Fatalf("Expected the treatment plan to indicate that it has deviated from the original content source (ftp) but it doesnt do so")
	}

}

// This test is to cover the scenario where if a doctor modifies,say, the treatment section after
// starting from a favorite treatment plan, we ensure that the rest of the sections are still prefilled
// with the contents of the favorite treatment plan
func TestFavoriteTreatmentPlan_BreakingMappingOnModify_PrefillRestOfData(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	patientVisitResponse, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// create a favorite treatment plan
	favoriteTreamentPlan := CreateFavoriteTreatmentPlan(treatmentPlan.Id.Int64(), testData, doctor, t)

	// pick this favorite treatment plan for the visit
	responseData := PickATreatmentPlanForPatientVisit(patientVisitResponse.PatientVisitId, doctor, favoriteTreamentPlan, testData, t)

	// modify treatment
	favoriteTreamentPlan.TreatmentList.Treatments[0].DispenseValue = encoding.HighPrecisionFloat64(123.12345)
	AddAndGetTreatmentsForPatientVisit(testData, favoriteTreamentPlan.TreatmentList.Treatments, doctor.AccountId.Int64(), responseData.TreatmentPlan.Id.Int64(), t)

	params := url.Values{}
	params.Set("treatment_plan_id", strconv.FormatInt(responseData.TreatmentPlan.Id.Int64(), 10))
	responseData = &doctor_treatment_plan.DoctorTreatmentPlanResponse{}
	resp, err := testData.AuthGet(testData.APIServer.URL+router.DoctorTreatmentPlansURLPath+"?"+params.Encode(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to make call to get treatment plan for patient visit")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d response for getting treatment plan instead got %d", http.StatusOK, resp.StatusCode)
	} else if json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into struct %s", err)
	}

	// the treatments should be in the committed state while the regimen and advice should still be prefilled in the uncommitted state
	if responseData.TreatmentPlan.TreatmentList == nil || len(responseData.TreatmentPlan.TreatmentList.Treatments) == 0 || responseData.TreatmentPlan.TreatmentList.Status != api.STATUS_COMMITTED {
		t.Fatal("Expected treatments to exist and be in COMMITTED state")
	} else if responseData.TreatmentPlan.RegimenPlan == nil || len(responseData.TreatmentPlan.RegimenPlan.Sections) == 0 || responseData.TreatmentPlan.RegimenPlan.Status != api.STATUS_UNCOMMITTED {
		t.Fatal("Expected regimen plan to be prefilled with FTP and be in UNCOMMITTED state")
	}
}

// This test ensures that the user can create a favorite treatment plan
// in the context of treatment plan by specifying the treatment plan to associate the
// favorite treatment plan with
func TestFavoriteTreatmentPlan_InContextOfTreatmentPlan(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	_, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	regimenPlanRequest := &common.RegimenPlan{
		TreatmentPlanID: treatmentPlan.Id,
	}

	regimenStep1 := &common.DoctorInstructionItem{
		Text:  "Regimen Step 1",
		State: common.STATE_ADDED,
	}

	regimenStep2 := &common.DoctorInstructionItem{
		Text:  "Regimen Step 2",
		State: common.STATE_ADDED,
	}

	regimenSection := &common.RegimenSection{
		Name: "morning",
		Steps: []*common.DoctorInstructionItem{
			{
				Text:  regimenStep1.Text,
				State: common.STATE_ADDED,
			},
			{
				Text:  regimenStep2.Text,
				State: common.STATE_ADDED,
			},
		},
	}

	regimenSection2 := &common.RegimenSection{
		Name: "night",
		Steps: []*common.DoctorInstructionItem{{
			Text:  regimenStep2.Text,
			State: common.STATE_ADDED,
		}},
	}

	regimenPlanRequest.AllSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
	regimenPlanRequest.Sections = []*common.RegimenSection{regimenSection, regimenSection2}
	regimenPlanResponse := CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	// prepare the regimen steps and the advice points to be added into the sections
	// after the global list for each has been updated to include items.
	// the reason this is important is because favorite treatment plans require items to exist that are linked
	// from the master list
	regimenSection.Steps[0].ParentID = regimenPlanResponse.AllSteps[0].ID
	regimenSection.Steps[1].ParentID = regimenPlanResponse.AllSteps[1].ID
	regimenSection2.Steps[0].ParentID = regimenPlanResponse.AllSteps[1].ID

	treatment1 := &common.Treatment{
		DrugDBIds: map[string]string{
			erx.LexiDrugSynId:     "1234",
			erx.LexiGenProductId:  "12345",
			erx.LexiSynonymTypeId: "123556",
			erx.NDC:               "2415",
		},
		DrugInternalName:        "Teting (This - Drug)",
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
	}

	AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	// lets add a favorite treatment plan for doctor
	favoriteTreatmentPlan := &common.FavoriteTreatmentPlan{
		Name: "Test Treatment Plan",
		TreatmentList: &common.TreatmentList{
			Treatments: []*common.Treatment{treatment1},
		},
		RegimenPlan: &common.RegimenPlan{
			AllSteps: regimenPlanResponse.AllSteps,
			Sections: []*common.RegimenSection{regimenSection, regimenSection2},
		},
	}

	requestData := &doctor_treatment_plan.DoctorFavoriteTreatmentPlansRequestData{
		FavoriteTreatmentPlan: favoriteTreatmentPlan,
		TreatmentPlanID:       treatmentPlan.Id.Int64(),
	}
	jsonData, err := json.Marshal(&requestData)
	if err != nil {
		t.Fatalf("Unable to marshal json %s", err)
	}

	resp, err := testData.AuthPost(testData.APIServer.URL+router.DoctorFTPURLPath, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to add favorite treatment plan: %s", err)
	}

	responseData := &doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into json %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 response for adding a favorite treatment plan but got %d instead", resp.StatusCode)
	} else if responseData.FavoriteTreatmentPlan == nil {
		t.Fatalf("Expected to get back the treatment plan added but got none")
	} else if responseData.FavoriteTreatmentPlan.RegimenPlan == nil || len(responseData.FavoriteTreatmentPlan.RegimenPlan.Sections) != 2 {
		t.Fatalf("Expected to have a regimen plan or 2 items in the regimen section")
	}

	abbreviatedTreatmentPlan, err := testData.DataApi.GetAbridgedTreatmentPlan(treatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatalf("Unable to get abbreviated favorite treatment plan: %s", err)
	} else if abbreviatedTreatmentPlan.ContentSource == nil || abbreviatedTreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		abbreviatedTreatmentPlan.ContentSource.ID.Int64() != responseData.FavoriteTreatmentPlan.Id.Int64() {
		t.Fatalf("Expected the link between treatmenet plan and favorite treatment plan to exist but it doesnt")
	}
}

func TestFavoriteTreatmentPlan_InContextOfTreatmentPlan_EmptyRegimen(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	_, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	regimenPlanRequest := &common.RegimenPlan{
		TreatmentPlanID: treatmentPlan.Id,
	}

	regimenStep1 := &common.DoctorInstructionItem{
		Text:  "Regimen Step 1",
		State: common.STATE_ADDED,
	}

	regimenStep2 := &common.DoctorInstructionItem{
		Text:  "Regimen Step 2",
		State: common.STATE_ADDED,
	}

	regimenPlanRequest.AllSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
	regimenPlanResponse := CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	treatment1 := &common.Treatment{
		DrugDBIds: map[string]string{
			erx.LexiDrugSynId:     "1234",
			erx.LexiGenProductId:  "12345",
			erx.LexiSynonymTypeId: "123556",
			erx.NDC:               "2415",
		},
		DrugInternalName:        "Teting (This - Drug)",
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
	}

	AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	// lets add a favorite treatment plan for doctor
	favoriteTreatmentPlan := &common.FavoriteTreatmentPlan{
		Name: "Test Treatment Plan",
		TreatmentList: &common.TreatmentList{
			Treatments: []*common.Treatment{treatment1},
		},
		RegimenPlan: &common.RegimenPlan{
			AllSteps: regimenPlanResponse.AllSteps,
		},
	}

	requestData := &doctor_treatment_plan.DoctorFavoriteTreatmentPlansRequestData{
		FavoriteTreatmentPlan: favoriteTreatmentPlan,
		TreatmentPlanID:       treatmentPlan.Id.Int64(),
	}
	jsonData, err := json.Marshal(&requestData)
	if err != nil {
		t.Fatalf("Unable to marshal json %s", err)
	}

	resp, err := testData.AuthPost(testData.APIServer.URL+router.DoctorFTPURLPath, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to add favorite treatment plan: %s", err)
	}

	responseData := &doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into json %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 response for adding a favorite treatment plan but got %d instead", resp.StatusCode)
	} else if responseData.FavoriteTreatmentPlan == nil {
		t.Fatalf("Expected to get back the treatment plan added but got none")
	}

	abbreviatedTreatmentPlan, err := testData.DataApi.GetAbridgedTreatmentPlan(treatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatalf("Unable to get abbreviated favorite treatment plan: %s", err)
	} else if abbreviatedTreatmentPlan.ContentSource == nil || abbreviatedTreatmentPlan.ContentSource.Type != common.TPContentSourceTypeFTP ||
		abbreviatedTreatmentPlan.ContentSource.ID.Int64() != responseData.FavoriteTreatmentPlan.Id.Int64() {
		t.Fatalf("Expected the link between treatmenet plan and favorite treatment plan to exist but it doesnt")
	}

}

func TestFavoriteTreatmentPlan_InContextOfTreatmentPlan_TwoDontMatch(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorId := GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id: %s", err)
	}

	_, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	regimenPlanRequest := &common.RegimenPlan{
		TreatmentPlanID: treatmentPlan.Id,
	}

	regimenStep1 := &common.DoctorInstructionItem{
		Text:  "Regimen Step 1",
		State: common.STATE_ADDED,
	}

	regimenStep2 := &common.DoctorInstructionItem{
		Text:  "Regimen Step 2",
		State: common.STATE_ADDED,
	}

	regimenPlanRequest.Sections = []*common.RegimenSection{
		&common.RegimenSection{
			Name: "dgag",
			Steps: []*common.DoctorInstructionItem{
				regimenStep1,
			},
		},
	}
	regimenPlanRequest.AllSteps = []*common.DoctorInstructionItem{regimenStep1, regimenStep2}
	regimenPlanResponse := CreateRegimenPlanForTreatmentPlan(regimenPlanRequest, testData, doctor, t)
	ValidateRegimenRequestAgainstResponse(regimenPlanRequest, regimenPlanResponse, t)

	treatment1 := &common.Treatment{
		DrugDBIds: map[string]string{
			erx.LexiDrugSynId:     "1234",
			erx.LexiGenProductId:  "12345",
			erx.LexiSynonymTypeId: "123556",
			erx.NDC:               "2415",
		},
		DrugInternalName:        "Teting (This - Drug)",
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
	}

	AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	// lets add a favorite treatment plan for doctor
	favoriteTreatmentPlan := &common.FavoriteTreatmentPlan{
		Name: "Test Treatment Plan",
		TreatmentList: &common.TreatmentList{
			Treatments: []*common.Treatment{treatment1},
		},
		RegimenPlan: &common.RegimenPlan{
			AllSteps: regimenPlanResponse.AllSteps,
		},
	}
	requestData := &doctor_treatment_plan.DoctorFavoriteTreatmentPlansRequestData{
		FavoriteTreatmentPlan: favoriteTreatmentPlan,
		TreatmentPlanID:       treatmentPlan.Id.Int64(),
	}
	jsonData, err := json.Marshal(&requestData)
	if err != nil {
		t.Fatalf("Unable to marshal json %s", err)
	}

	resp, err := testData.AuthPost(testData.APIServer.URL+router.DoctorFTPURLPath, "application/json", bytes.NewReader(jsonData), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to add favorite treatment plan: %s", err)
	}

	responseData := &doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		t.Fatalf("Unable to unmarshal response into json %s", err)
	} else if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 response for adding a favorite treatment plan but got %d instead", resp.StatusCode)
	}

	abbreviatedTreatmentPlan, err := testData.DataApi.GetAbridgedTreatmentPlan(treatmentPlan.Id.Int64(), doctorId)
	if err != nil {
		t.Fatalf("Unable to get abbreviated favorite treatment plan: %s", err)
	} else if abbreviatedTreatmentPlan.ContentSource != nil {
		t.Fatalf("Expected the treatment plan to not indicate that it was linked to another doctor's favorite treatment plan")
	}

}
