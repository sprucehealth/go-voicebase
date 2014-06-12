package test_integration

import (
	"bytes"
	"carefront/apiservice"
	"carefront/common"
	"carefront/encoding"
	"carefront/libs/erx"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMedicationStrengthSearch(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from id: " + err.Error())
	}

	erx := setupErxAPI(t)
	medicationStrengthSearchHandler := &apiservice.MedicationStrengthSearchHandler{DataApi: testData.DataApi, ERxApi: erx}
	ts := httptest.NewServer(medicationStrengthSearchHandler)
	defer ts.Close()

	resp, err := AuthGet(ts.URL+"?drug_internal_name="+url.QueryEscape("Benzoyl Peroxide Topical (topical - cream)"), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a successful query to the medication strength api: " + err.Error())
	}

	medicationStrengthResponse := &apiservice.MedicationStrengthSearchResponse{}
	err = json.NewDecoder(resp.Body).Decode(medicationStrengthResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal the response from the medication strength search api into a json object as expected: " + err.Error())
	}
	CheckSuccessfulStatusCode(resp, "Unable to make a successful query to the medication strength api for the doctor: ", t)

	if medicationStrengthResponse.MedicationStrengths == nil || len(medicationStrengthResponse.MedicationStrengths) == 0 {
		t.Fatal("Expected a list of medication strengths from the api but got none")
	}
}

func TestNewTreatmentSelection(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from id: " + err.Error())
	}

	erxApi := setupErxAPI(t)
	newTreatmentHandler := &apiservice.NewTreatmentHandler{DataApi: testData.DataApi, ERxApi: erxApi}
	ts := httptest.NewServer(newTreatmentHandler)
	defer ts.Close()

	resp, err := AuthGet(ts.URL+"?drug_internal_name="+url.QueryEscape("Lisinopril (oral - tablet)")+"&medication_strength="+url.QueryEscape("10 mg"), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a successful query to the medication strength api: " + err.Error())
	}

	newTreatmentResponse := &apiservice.NewTreatmentResponse{}
	err = json.NewDecoder(resp.Body).Decode(newTreatmentResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal the response from the medication strength search api into a json object as expected: " + err.Error())
	}
	CheckSuccessfulStatusCode(resp, "Unable to make a successful query to the medication strength api for the docto", t)

	if newTreatmentResponse.Treatment == nil {
		t.Fatal("Expected medication object to be populated but its not")
	}

	if newTreatmentResponse.Treatment.DrugDBIds == nil || len(newTreatmentResponse.Treatment.DrugDBIds) == 0 {
		t.Fatal("Expected additional drug db ids to be returned from api but none were")
	}

	if newTreatmentResponse.Treatment.DrugDBIds[erx.LexiDrugSynId] == "0" || newTreatmentResponse.Treatment.DrugDBIds[erx.LexiSynonymTypeId] == "0" || newTreatmentResponse.Treatment.DrugDBIds[erx.LexiGenProductId] == "0" {
		t.Fatal("Expected additional drug db ids not set (lexi_drug_syn_id and lexi_synonym_type_id")
	}

	// Let's run a test for an OTC product to ensure that the OTC flag is set as expected
	resp, err = AuthGet(ts.URL+"?drug_internal_name="+url.QueryEscape("Fish Oil (oral - capsule)")+"&medication_strength="+url.QueryEscape("500 mg"), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make a successful query to the medication strength api: " + err.Error())
	}

	newTreatmentResponse = &apiservice.NewTreatmentResponse{}
	err = json.NewDecoder(resp.Body).Decode(newTreatmentResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal the response from the medication strength search api into a json object as expected: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make a successful query to the medication strength api for the doctor for an OTC product: ", t)

	if newTreatmentResponse.Treatment == nil || newTreatmentResponse.Treatment.OTC == false {
		t.Fatal("Expected the medication object to be returned and for the medication returned to be an OTC product")
	}

	// Let's ensure that we are returning a bad request to the doctor if they select a controlled substance
	urlValues := url.Values{}
	urlValues.Set("drug_internal_name", "Testosterone (buccal - film, extended release)")
	urlValues.Set("medication_strength", "30 mg/12 hr")
	resp, err = AuthGet(ts.URL+"?"+urlValues.Encode(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successful call to selected a controlled substance as a medication: " + err.Error())
	}

	if resp.StatusCode != apiservice.HTTP_UNPROCESSABLE_ENTITY {
		t.Fatal("Expected a bad request when attempting to select a controlled substance given that we don't allow routing of controlled substances using our platform")
	}

	// Let's ensure that we are rejecting a drug description that is longer than 105 characters to be routed via eRX.
	urlValues = url.Values{}
	urlValues.Set("drug_internal_name", "Clinimix E Sulfite-Free 2.75% with 10% Dextrose and Electrolytes (intravenous - solution)")
	urlValues.Set("medication_strength", "Amino Acids 2.75% with 10% Dextrose and Electrolytes (Clinimix E Sulfite-Free)")
	resp, err = AuthGet(ts.URL+"?"+urlValues.Encode(), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make successfull call to select a drug whose description is longer than the limit" + err.Error())
	}
	if resp.StatusCode != apiservice.HTTP_UNPROCESSABLE_ENTITY {
		t.Fatal("Expected a bad request when attempting to select a drug with longer than max drug description to ensure that we don't send through eRx but advice doc to call drug in")
	}
}

func TestDispenseUnitIds(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	medicationDispenseUnitsHandler := &apiservice.MedicationDispenseUnitsHandler{DataApi: testData.DataApi}
	ts := httptest.NewServer(medicationDispenseUnitsHandler)
	defer ts.Close()

	resp, err := AuthGet(ts.URL, 0)
	if err != nil {
		t.Fatal("Unable to make a successful query to the medication dispense units api: " + err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to parse the body of the response: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make a successful query to the medication dispense units api for the doctor: "+string(body), t)
	medicationDispenseUnitsResponse := &apiservice.MedicationDispenseUnitsResponse{}
	err = json.Unmarshal(body, medicationDispenseUnitsResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal the response from the medication strength search api into a json object as expected: " + err.Error())
	}

	if medicationDispenseUnitsResponse.DispenseUnits == nil || len(medicationDispenseUnitsResponse.DispenseUnits) == 0 {
		t.Fatal("Expected dispense unit ids to be returned from api but none were")
	}

	for _, dispenseUnitItem := range medicationDispenseUnitsResponse.DispenseUnits {
		if dispenseUnitItem.Id == 0 || dispenseUnitItem.Text == "" {
			t.Fatal("Dispense Unit item was empty when this is not expected")
		}
	}

}

func TestAddTreatments(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	_, treatmentPlan := SignupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)

	// doctor now attempts to add a couple treatments for patient
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

	treatment2 := &common.Treatment{
		DrugInternalName: "Advil 2",
		TreatmentPlanId:  treatmentPlan.Id,
		DosageStrength:   "100 mg",
		DispenseValue:    2,
		DispenseUnitId:   encoding.NewObjectId(27),
		NumberRefills: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 3,
		},
		SubstitutionsAllowed: false,
		DaysSupply:           encoding.NullInt64{}, OTC: false,
		PharmacyNotes:       "testing pharmacy notes 2",
		PatientInstructions: "patient instructions 2",
		DrugDBIds: map[string]string{
			"drug_db_id_3": "12414",
			"drug_db_id_4": "214",
		},
	}

	treatments := []*common.Treatment{treatment1, treatment2}

	getTreatmentsResponse := addAndGetTreatmentsForPatientVisit(testData, treatments, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	for _, treatment := range getTreatmentsResponse.TreatmentList.Treatments {
		switch treatment.DrugInternalName {
		case treatment1.DrugInternalName:
			compareTreatments(treatment, treatment1, t)
		case treatment2.DrugInternalName:
			compareTreatments(treatment, treatment2, t)
		}
	}

	// now lets go ahead and post an update where we have just one treatment for the patient visit which was updated while the other was deleted
	treatments[0].DispenseValue = 10
	treatments = []*common.Treatment{treatments[0]}
	getTreatmentsResponse = addAndGetTreatmentsForPatientVisit(testData, treatments, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	// there should be just one treatment and its name should be the name that we just set
	if len(getTreatmentsResponse.TreatmentList.Treatments) != 1 {
		t.Fatal("Expected just 1 treatment to be returned after update")
	}

	// the dispense value should be set to 10
	if getTreatmentsResponse.TreatmentList.Treatments[0].DispenseValue != 10 {
		t.Fatal("Expected the updated dispense value to be set to 10")
	}

}

func TestTreatmentTemplates(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}
	_, treatmentPlan := SignupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)

	// doctor now attempts to favorite a treatment
	treatment1 := &common.Treatment{
		DrugInternalName: "DrugName (DrugRoute - DrugForm)",
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
		PatientInstructions: "patient insturctions",
		DrugDBIds: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	treatmentTemplate := &common.DoctorTreatmentTemplate{}
	treatmentTemplate.Name = "Favorite Treatment #1"
	treatmentTemplate.Treatment = treatment1

	doctorTreatmentTemplatesHandler := &apiservice.DoctorTreatmentTemplatesHandler{DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorTreatmentTemplatesHandler)
	defer ts.Close()

	treatmentTemplatesRequest := &apiservice.DoctorTreatmentTemplatesRequest{
		TreatmentPlanId:    treatmentPlan.Id,
		TreatmentTemplates: []*common.DoctorTreatmentTemplate{treatmentTemplate},
	}
	data, err := json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err := AuthPost(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d but got %d instead", http.StatusOK, resp.StatusCode)
	}

	treatmentTemplatesResponse := &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.NewDecoder(resp.Body).Decode(treatmentTemplatesResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	}

	if treatmentTemplatesResponse.TreatmentTemplates == nil || len(treatmentTemplatesResponse.TreatmentTemplates) != 1 {
		t.Fatal("Expected 1 favorited treatment in response but got none")
	}

	if treatmentTemplatesResponse.TreatmentTemplates[0].Name != treatmentTemplate.Name {
		t.Fatal("Expected the same favorited treatment to be returned that was added")
	}

	if treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugName != "DrugName" ||
		treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugRoute != "DrugRoute" ||
		treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugForm != "DrugForm" {
		t.Fatalf("Expected the drug internal name to have been broken into its components %s %s %s", treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugName,
			treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugRoute, treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugForm)
	}

	// also ensure that drug db ids is not null or empty
	if len(treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugDBIds) != 2 {
		t.Fatalf("Expected 2 drug db ids to exist instead got %d", len(treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugDBIds))
	}

	treatment2 := &common.Treatment{
		DrugInternalName: "DrugName2",
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

	treatmentTemplate2 := &common.DoctorTreatmentTemplate{}
	treatmentTemplate2.Name = "Treatment Template #2"
	treatmentTemplate2.Treatment = treatment2

	treatmentTemplatesRequest.TreatmentTemplates[0] = treatmentTemplate2
	data, err = json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err = AuthPost(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d but got %d instead", http.StatusOK, resp.StatusCode)
	}

	treatmentTemplatesResponse = &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.NewDecoder(resp.Body).Decode(treatmentTemplatesResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	} else if treatmentTemplatesResponse.TreatmentTemplates == nil || len(treatmentTemplatesResponse.TreatmentTemplates) != 2 {
		t.Fatal("Expected 2 favorited treatments in response")
	} else if treatmentTemplatesResponse.TreatmentTemplates[0].Name != treatmentTemplate.Name {
		t.Fatal("Expected the same favorited treatment to be returned that was added")
	} else if treatmentTemplatesResponse.TreatmentTemplates[1].Name != treatmentTemplate2.Name {
		t.Fatal("Expected the same favorited treatment to be returned that was added")
	}

	// lets go ahead and delete each of the treatments
	treatmentTemplatesRequest.TreatmentTemplates = treatmentTemplatesResponse.TreatmentTemplates
	treatmentTemplatesRequest.TreatmentPlanId = treatmentPlan.Id
	data, err = json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err = AuthDelete(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d but got %d instead", http.StatusOK, resp.StatusCode)
	}

	treatmentTemplatesResponse = &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.NewDecoder(resp.Body).Decode(treatmentTemplatesResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	}

	if len(treatmentTemplatesResponse.TreatmentTemplates) != 0 {
		t.Fatal("Expected 1 favorited treatment after deleting the first one")
	}
}

func TestTreatmentTemplatesInContextOfPatientVisit(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	// create random patient
	_, treatmentPlan := SignupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)

	// doctor now attempts to favorite a treatment
	treatment1 := &common.Treatment{
		DrugInternalName: "DrugName (DrugRoute - DrugForm)",
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
		PatientInstructions: "patient insturctions",
		DrugDBIds: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	treatmentTemplate := &common.DoctorTreatmentTemplate{}
	treatmentTemplate.Name = "Favorite Treatment #1"
	treatmentTemplate.Treatment = treatment1

	doctorFavoriteTreatmentsHandler := &apiservice.DoctorTreatmentTemplatesHandler{DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorFavoriteTreatmentsHandler)
	defer ts.Close()

	treatmentTemplatesRequest := &apiservice.DoctorTreatmentTemplatesRequest{TreatmentTemplates: []*common.DoctorTreatmentTemplate{treatmentTemplate}}
	treatmentTemplatesRequest.TreatmentPlanId = treatmentPlan.Id
	data, err := json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err := AuthPost(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of the post request made to add treatments to patient visit: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unsuccessful call made to add favorite treatment for doctor "+string(body), t)

	treatmentTemplatesResponse := &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.Unmarshal(body, treatmentTemplatesResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	}

	if treatmentTemplatesResponse.TreatmentTemplates == nil || len(treatmentTemplatesResponse.TreatmentTemplates) != 1 {
		t.Fatal("Expected 1 favorited treatment in response but got none")
	}

	if treatmentTemplatesResponse.TreatmentTemplates[0].Name != treatmentTemplate.Name {
		t.Fatal("Expected the same favorited treatment to be returned that was added")
	}

	if treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugName != "DrugName" ||
		treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugRoute != "DrugRoute" ||
		treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugForm != "DrugForm" {
		t.Fatalf("Expected the drug internal name to have been broken into its components %s %s %s", treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugName,
			treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugRoute, treatmentTemplatesResponse.TreatmentTemplates[0].Treatment.DrugForm)
	}

	treatment2 := &common.Treatment{
		DrugInternalName: "DrugName2 (DrugRoute - DrugForm)",
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
		}, OTC: true,
		PharmacyNotes:       "testing pharmacy notes",
		PatientInstructions: "patient instructions",
		DrugDBIds: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	// lets add this as a treatment to the patient visit
	getTreatmentsResponse := addAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment2}, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	if len(getTreatmentsResponse.TreatmentList.Treatments) != 1 {
		t.Fatal("Expected patient visit to have 1 treatment")
	}

	// now, lets favorite a treatment that exists for the patient visit
	treatmentTemplate2 := &common.DoctorTreatmentTemplate{}
	treatmentTemplate2.Name = "Favorite Treatment #2"
	treatmentTemplate2.Treatment = getTreatmentsResponse.TreatmentList.Treatments[0]
	treatmentTemplatesRequest.TreatmentTemplates[0] = treatmentTemplate2
	treatmentTemplatesRequest.TreatmentPlanId = treatmentPlan.Id

	data, err = json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp2, err := AuthPost(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	}

	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {
		t.Fatal("Unable to read from response body: " + err.Error())
	}
	CheckSuccessfulStatusCode(resp2, "Unsuccessful call made to add favorite treatment for doctor ", t)

	treatmentTemplatesResponse = &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.Unmarshal(body, treatmentTemplatesResponse)

	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	}

	if treatmentTemplatesResponse.TreatmentTemplates == nil || len(treatmentTemplatesResponse.TreatmentTemplates) != 2 {
		t.Fatal("Expected 2 favorited treatments in response")
	}

	if treatmentTemplatesResponse.TreatmentTemplates[0].Name != treatmentTemplate.Name {
		t.Fatal("Expected the same favorited treatment to be returned that was added")
	}

	if treatmentTemplatesResponse.TreatmentTemplates[1].Name != treatmentTemplate2.Name {
		t.Fatal("Expected the same favorited treatment to be returned that was added")
	}

	if len(treatmentTemplatesResponse.TreatmentTemplates) == 0 {
		t.Fatal("Expected there to be 1 treatment added to the visit and the doctor")
	}

	if treatmentTemplatesResponse.Treatments[0].DoctorTreatmentTemplateId.Int64() != treatmentTemplatesResponse.TreatmentTemplates[1].Id.Int64() {
		t.Fatal("Expected the favoriteTreatmentId to be set for the treatment and to be set to the right treatment")
	}

	// now, lets go ahead and add a treatment to the patient visit from a favorite treatment
	treatment1.DoctorTreatmentTemplateId = encoding.NewObjectId(treatmentTemplatesResponse.TreatmentTemplates[0].Id.Int64())
	treatment2.DoctorTreatmentTemplateId = encoding.NewObjectId(treatmentTemplatesResponse.TreatmentTemplates[1].Id.Int64())
	getTreatmentsResponse = addAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1, treatment2}, doctor.AccountId.Int64(), treatmentPlan.Id.Int64(), t)

	if len(getTreatmentsResponse.TreatmentList.Treatments) != 2 {
		t.Fatal("There should exist 2 treatments for the patient visit")
	}

	if getTreatmentsResponse.TreatmentList.Treatments[0].DoctorTreatmentTemplateId.Int64() == 0 || getTreatmentsResponse.TreatmentList.Treatments[1].DoctorTreatmentTemplateId.Int64() == 0 {
		t.Fatal("Expected the doctorFavoriteId to be set for both treatments given that they were added from favorites")
	}

	treatmentTemplate.Id = encoding.NewObjectId(getTreatmentsResponse.TreatmentList.Treatments[0].DoctorTreatmentTemplateId.Int64())
	treatmentTemplate.Treatment = getTreatmentsResponse.TreatmentList.Treatments[0]
	treatmentTemplatesRequest.TreatmentTemplates = []*common.DoctorTreatmentTemplate{treatmentTemplate}
	treatmentTemplatesRequest.TreatmentPlanId = treatmentPlan.Id
	// lets delete a favorite that is also a treatment in the patient visit
	data, err = json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err = AuthDelete(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	}

	treatmentTemplatesResponse = &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.NewDecoder(resp.Body).Decode(treatmentTemplatesResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unsuccessful call made to add favorite treatment for doctor ", t)

	if len(treatmentTemplatesResponse.TreatmentTemplates) != 1 {
		t.Fatal("Expected 1 favorited treatment after deleting the first one")
	}

	// ensure that treatments are still returned
	if len(treatmentTemplatesResponse.Treatments) != 2 {
		t.Fatal("Expected there to exist 2 treatments for the patient visit even after deleting one of the treatments")
	}

	if treatmentTemplatesResponse.Treatments[0].DoctorTreatmentTemplateId.Int64() != 0 {
		t.Fatal("Expected the first treatment to no longer be a favorited treatment")
	}
}

func TestTreatmentTemplateWithDrugOutOfMarket(t *testing.T) {

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := GetDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	_, treatmentPlan := SignupAndSubmitPatientVisitForRandomPatient(t, testData, doctor)
	// doctor now attempts to favorite a treatment
	treatment1 := &common.Treatment{
		DrugInternalName: "DrugName (DrugRoute - DrugForm)",
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
		}, OTC: true,
		PharmacyNotes:       "testing pharmacy notes",
		PatientInstructions: "patient insturctions",
		DrugDBIds: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	treatmentTemplate := &common.DoctorTreatmentTemplate{}
	treatmentTemplate.Name = "Favorite Treatment #1"
	treatmentTemplate.Treatment = treatment1

	doctorFavoriteTreatmentsHandler := &apiservice.DoctorTreatmentTemplatesHandler{DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorFavoriteTreatmentsHandler)
	defer ts.Close()

	treatmentTemplatesRequest := &apiservice.DoctorTreatmentTemplatesRequest{TreatmentTemplates: []*common.DoctorTreatmentTemplate{treatmentTemplate}}
	treatmentTemplatesRequest.TreatmentPlanId = treatmentPlan.Id
	data, err := json.Marshal(&treatmentTemplatesRequest)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err := AuthPost(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to make POST request to add treatments to patient visit " + err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of the post request made to add treatments to patient visit: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unsuccessful call made to add favorite treatment for doctor "+string(body), t)

	treatmentTemplatesResponse := &apiservice.DoctorTreatmentTemplatesResponse{}
	err = json.Unmarshal(body, treatmentTemplatesResponse)
	if err != nil {
		t.Fatal("Unable to unmarshal response into object : " + err.Error())
	}

	// lets' attempt to add the favorited treatment to a patient visit. It should fail because the stubErxApi is wired
	// to return no medication to indicate drug is no longer in market
	treatment1.DoctorTreatmentTemplateId = treatmentTemplatesResponse.TreatmentTemplates[0].Id
	stubErxApi := &erx.StubErxService{}
	treatmentRequestBody := apiservice.AddTreatmentsRequestBody{
		TreatmentPlanId: treatmentPlan.Id,
		Treatments:      []*common.Treatment{treatment1},
	}

	treatmentsHandler := &apiservice.TreatmentsHandler{
		ErxApi:  stubErxApi,
		DataApi: testData.DataApi,
	}

	ts = httptest.NewServer(treatmentsHandler)
	defer ts.Close()

	data, err = json.Marshal(&treatmentRequestBody)
	if err != nil {
		t.Fatal("Unable to marshal request body for adding treatments to patient visit")
	}

	resp, err = AuthPost(ts.URL, "application/json", bytes.NewBuffer(data), doctor.AccountId.Int64())
	if err != nil {
		t.Fatal("Unable to add treatments to patient visit: " + err.Error())
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected the call to add treatments to error out with bad request (400) because treatment is out of market, but instead got %d returned", resp.StatusCode)
	}
}

func compareTreatments(treatment *common.Treatment, treatment1 *common.Treatment, t *testing.T) {
	if treatment.DosageStrength != treatment1.DosageStrength || treatment.DispenseValue != treatment1.DispenseValue ||
		treatment.DispenseUnitId.Int64() != treatment1.DispenseUnitId.Int64() || treatment.PatientInstructions != treatment1.PatientInstructions ||
		treatment.PharmacyNotes != treatment1.PharmacyNotes || treatment.NumberRefills != treatment1.NumberRefills ||
		treatment.SubstitutionsAllowed != treatment1.SubstitutionsAllowed || treatment.DaysSupply != treatment1.DaysSupply ||
		treatment.OTC != treatment1.OTC {
		treatmentData, _ := json.MarshalIndent(treatment, "", " ")
		treatment1Data, _ := json.MarshalIndent(treatment1, "", " ")

		t.Fatalf("Treatment returned from the call to get treatments for patient visit not the same as what was added for the patient visit: treatment returned: %s, treatment added: %s", string(treatmentData), string(treatment1Data))
	}

	for drugDBIdTag, drugDBId := range treatment.DrugDBIds {
		if treatment1.DrugDBIds[drugDBIdTag] == "" || treatment1.DrugDBIds[drugDBIdTag] != drugDBId {
			treatmentData, _ := json.MarshalIndent(treatment, "", " ")
			treatment1Data, _ := json.MarshalIndent(treatment1, "", " ")

			t.Fatalf("Treatment returned from the call to get treatments for patient visit not the same as what was added for the patient visit: treatment returned: %s, treatment added: %s", string(treatmentData), string(treatment1Data))
		}
	}
}
