package test_integration

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/app_worker"
	"carefront/common"
	"carefront/encoding"
	"carefront/libs/aws/sqs"
	"carefront/libs/erx"
	"carefront/libs/pharmacy"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/samuel/go-metrics/metrics"
)

func getTestRefillRequest(refillRequestQueueItemId, erxPatientId, prescriptionId, clinicianId, pharmacyId int64) *common.RefillRequestItem {
	return &common.RefillRequestItem{
		RxRequestQueueItemId:      refillRequestQueueItemId,
		ReferenceNumber:           "TestReferenceNumber",
		PharmacyRxReferenceNumber: "TestRxReferenceNumber",
		ErxPatientId:              erxPatientId,
		PatientAddedForRequest:    false,
		ClinicianId:               clinicianId,
		RequestedPrescription: &common.Treatment{
			DrugDBIds: map[string]string{
				erx.LexiDrugSynId:     "1234",
				erx.LexiGenProductId:  "12345",
				erx.LexiSynonymTypeId: "123556",
				erx.NDC:               "2415",
			},
			DosageStrength:       "10 mg",
			DispenseValue:        5,
			OTC:                  false,
			SubstitutionsAllowed: true,
			ERx: &common.ERxData{
				DoseSpotClinicianId: clinicianId,
				PrescriptionId:      encoding.NewObjectId(prescriptionId),
				ErxPharmacyId:       pharmacyId,
			},
		},
		DispensedPrescription: &common.Treatment{
			DrugDBIds: map[string]string{
				"drug_db_id_1": "1234",
				"drug_db_id_2": "12345",
			},
			DrugName:                "Teting (This - Drug)",
			DosageStrength:          "10 mg",
			DispenseValue:           5,
			DispenseUnitDescription: "Tablet",
			NumberRefills: encoding.NullInt64{
				IsValid:    true,
				Int64Value: 5,
			},
			SubstitutionsAllowed: false,
			DaysSupply: encoding.NullInt64{
				IsValid:    true,
				Int64Value: 10,
			}, PatientInstructions: "Take once daily",
			OTC: false,
			ERx: &common.ERxData{
				PrescriptionId:      encoding.NewObjectId(5504),
				PrescriptionStatus:  "Requested",
				ErxPharmacyId:       pharmacyId,
				DoseSpotClinicianId: clinicianId,
			},
		},
	}
}

func getTestPreferredPharmacyAndTreatment() (*common.Treatment, *pharmacy.PharmacyData) {
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

	// create the preferred pharmacy for the patient
	pharmacySelection := &pharmacy.PharmacyData{
		SourceId:     "12345",
		Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		AddressLine1: "12345 Marin Street",
		City:         "San Francisco",
		State:        "CA",
		Phone:        "12345667",
	}
	return treatment1, pharmacySelection
}

// Test treatment in treatment plan that has an error after being in the sent state
func TestTreatmentInErrorAfterSentState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// setup test
	doctorId := getDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id %s", err)
	}

	// get treatment ready for doctor to add for patient
	// while creating treatment plan
	prescriptionIdToReturn := int64(1235)
	treatment1, pharmacySelection := getTestPreferredPharmacyAndTreatment()

	stubErxAPI := &erx.StubErxService{
		PrescriptionIdsToReturn: []int64{prescriptionIdToReturn},
		PrescriptionIdToPrescriptionStatuses: map[int64][]common.StatusEvent{
			prescriptionIdToReturn: []common.StatusEvent{common.StatusEvent{
				Status: api.ERX_STATUS_SENT,
			},
			},
		},
	}

	erxStatusQueue := &common.SQSQueue{}
	erxStatusQueue.QueueService = &sqs.StubSQS{}
	erxStatusQueue.QueueUrl = "local-erx"

	submitDoctorVisitReviewHandler := &apiservice.DoctorSubmitPatientVisitReviewHandler{
		DataApi:        testData.DataApi,
		ERxApi:         stubErxAPI,
		ErxStatusQueue: erxStatusQueue,
		ERxRouting:     true,
	}
	ts := httptest.NewServer(submitDoctorVisitReviewHandler)
	defer ts.Close()

	// sign up a patient and get them to submit a patient visit
	signupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patient := signupResponse.Patient
	patientId := patient.PatientId.Int64()
	patientVisitResponse := CreatePatientVisitForPatient(patient.PatientId.Int64(), testData, t)
	SubmitPatientVisitForPatient(patientId, patientVisitResponse.PatientVisitId, testData, t)
	err = testData.DataApi.UpdatePatientPharmacy(patient.PatientId.Int64(), pharmacySelection)
	if err != nil {
		t.Fatal("Unable to update patient pharmacy: " + err.Error())
	}

	// get the doctor to add a treatment to the patient visit that we can track the status of
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)
	treatment1.PatientVisitId = encoding.NewObjectId(patientVisitResponse.PatientVisitId)
	treatmentResponse := addAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(),
		patientVisitResponse.PatientVisitId, t)
	params := url.Values{}
	params.Set("patient_visit_id", strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	resp, err := authPost(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to submit doctor visit review %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d for status code but got %d", http.StatusOK, resp.StatusCode)
	}

	// once the treatment has been submitted, track the status of the submitted treatment to move it to the sent state
	app_worker.ConsumeMessageFromQueue(testData.DataApi, stubErxAPI, erxStatusQueue, metrics.NewBiasedHistogram(), metrics.NewCounter(), metrics.NewCounter())

	// expected state of the treatment here is sent
	statusEvents, err := testData.DataApi.GetPrescriptionStatusEventsForTreatment(treatmentResponse.Treatments[0].Id.Int64())
	if err != nil {
		t.Fatalf("Unable to get status events for treatments: %s", err)
	} else if len(statusEvents) != 2 {
		t.Fatalf("Expected 2 status events instead got %d", len(statusEvents))
	} else if statusEvents[0].Status != api.ERX_STATUS_SENT {
		t.Fatalf("Expected status to be %s instead it was %s", api.ERX_STATUS_SENT, statusEvents[0].Status)
	}

	// now stub the erx api to return a "free-standing" transmission error detail for this treatment
	stubErxAPI.TransmissionErrorsForPrescriptionIds = []int64{prescriptionIdToReturn}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxAPI, metrics.NewCounter(), metrics.NewCounter())

	// there should now be 3 status events for this treatment given that
	// the rx error checker caught the missed transition from sending -> sent -> error
	statusEvents, err = testData.DataApi.GetPrescriptionStatusEventsForTreatment(treatmentResponse.Treatments[0].Id.Int64())
	if err != nil {
		t.Fatalf("Unable to get status events for treatment: %s", err)
	} else if len(statusEvents) != 3 {
		t.Fatalf("Expected 3 status events instead got %d", len(statusEvents))
	} else if statusEvents[0].Status != api.ERX_STATUS_ERROR && statusEvents[1].Status != api.ERX_STATUS_SENT {
		t.Fatalf("Expected a transition from sent -> error, instead got %s -> %s", statusEvents[1].Status, statusEvents[0].Status)
	}

	// there should also be a pending item in the doctor's queue for the errored transmission
	pendingItems, err := testData.DataApi.GetPendingItemsInDoctorQueue(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor queue: %s", err)
	} else if len(pendingItems) != 1 {
		t.Fatalf("Expected 1 item in the pending tab of doctor queue instead got %d", len(pendingItems))
	} else if pendingItems[0].EventType != api.EVENT_TYPE_TRANSMISSION_ERROR {
		t.Fatalf("Expected a %s event type in the doctor queue instead got %s", api.EVENT_TYPE_TRANSMISSION_ERROR, pendingItems[0].EventType)
	}
}

// Test treatment in treatment plan that has an error after being in the sending state
func TestTreatmentInErrorAfterSendingState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// setup test
	doctorId := getDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id %s", err)
	}

	// get treatment ready for doctor to add for patient
	// while creating treatment plan
	prescriptionIdToReturn := int64(1235)
	treatment1, pharmacySelection := getTestPreferredPharmacyAndTreatment()

	stubErxAPI := &erx.StubErxService{
		PrescriptionIdsToReturn: []int64{prescriptionIdToReturn},
		PrescriptionIdToPrescriptionStatuses: map[int64][]common.StatusEvent{
			prescriptionIdToReturn: []common.StatusEvent{common.StatusEvent{
				Status: api.ERX_STATUS_SENT,
			},
			},
		},
	}

	erxStatusQueue := &common.SQSQueue{}
	erxStatusQueue.QueueService = &sqs.StubSQS{}
	erxStatusQueue.QueueUrl = "local-erx"

	submitDoctorVisitReviewHandler := &apiservice.DoctorSubmitPatientVisitReviewHandler{
		DataApi:        testData.DataApi,
		ERxApi:         stubErxAPI,
		ErxStatusQueue: erxStatusQueue,
		ERxRouting:     true,
	}
	ts := httptest.NewServer(submitDoctorVisitReviewHandler)
	defer ts.Close()

	// sign up a patient and get them to submit a patient visit
	signupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patient := signupResponse.Patient
	patientId := patient.PatientId.Int64()
	patientVisitResponse := CreatePatientVisitForPatient(patient.PatientId.Int64(), testData, t)
	SubmitPatientVisitForPatient(patientId, patientVisitResponse.PatientVisitId, testData, t)
	err = testData.DataApi.UpdatePatientPharmacy(patient.PatientId.Int64(), pharmacySelection)
	if err != nil {
		t.Fatal("Unable to update patient pharmacy: " + err.Error())
	}

	// get the doctor to add a treatment to the patient visit that we can track the status of
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)
	treatment1.PatientVisitId = encoding.NewObjectId(patientVisitResponse.PatientVisitId)
	treatmentResponse := addAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(),
		patientVisitResponse.PatientVisitId, t)
	params := url.Values{}
	params.Set("patient_visit_id", strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	resp, err := authPost(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to submit doctor visit review %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d for status code but got %d", http.StatusOK, resp.StatusCode)
	}

	// now stub the erx api to return a "free-standing" transmission error detail for this treatment
	stubErxAPI.TransmissionErrorsForPrescriptionIds = []int64{prescriptionIdToReturn}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxAPI, metrics.NewCounter(), metrics.NewCounter())

	// there should now be 2 status events for this treatment given that
	// the rx error checker caught the transition from sending  -> error
	statusEvents, err := testData.DataApi.GetPrescriptionStatusEventsForTreatment(treatmentResponse.Treatments[0].Id.Int64())
	if err != nil {
		t.Fatalf("Unable to get status events for treatment: %s", err)
	} else if len(statusEvents) != 2 {
		t.Fatalf("Expected 3 status events instead got %d", len(statusEvents))
	} else if statusEvents[0].Status != api.ERX_STATUS_ERROR && statusEvents[1].Status != api.ERX_STATUS_SENDING {
		t.Fatalf("Expected a transition from sent -> error, instead got %s -> %s", statusEvents[1].Status, statusEvents[0].Status)
	}

	// there should also be a pending item in the doctor's queue for the errored transmission
	pendingItems, err := testData.DataApi.GetPendingItemsInDoctorQueue(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor queue: %s", err)
	} else if len(pendingItems) != 1 {
		t.Fatalf("Expected 1 item in the pending tab of doctor queue instead got %d", len(pendingItems))
	} else if pendingItems[0].EventType != api.EVENT_TYPE_TRANSMISSION_ERROR {
		t.Fatalf("Expected a %s event type in the doctor queue instead got %s", api.EVENT_TYPE_TRANSMISSION_ERROR, pendingItems[0].EventType)
	}
}

// Test treatment in treatment plan that has an error after being in the sent state
func TestTreatmentInErrorAfterErorState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// setup test
	doctorId := getDoctorIdOfCurrentPrimaryDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor from id %s", err)
	}

	// get treatment ready for doctor to add for patient
	// while creating treatment plan
	prescriptionIdToReturn := int64(1235)
	treatment1, pharmacySelection := getTestPreferredPharmacyAndTreatment()

	stubErxAPI := &erx.StubErxService{
		PrescriptionIdsToReturn: []int64{prescriptionIdToReturn},
		PrescriptionIdToPrescriptionStatuses: map[int64][]common.StatusEvent{
			prescriptionIdToReturn: []common.StatusEvent{common.StatusEvent{
				Status:        api.ERX_STATUS_ERROR,
				StatusDetails: "test error",
			},
			},
		},
	}

	erxStatusQueue := &common.SQSQueue{}
	erxStatusQueue.QueueService = &sqs.StubSQS{}
	erxStatusQueue.QueueUrl = "local-erx"

	submitDoctorVisitReviewHandler := &apiservice.DoctorSubmitPatientVisitReviewHandler{
		DataApi:        testData.DataApi,
		ERxApi:         stubErxAPI,
		ErxStatusQueue: erxStatusQueue,
		ERxRouting:     true,
	}
	ts := httptest.NewServer(submitDoctorVisitReviewHandler)
	defer ts.Close()

	// sign up a patient and get them to submit a patient visit
	signupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	patient := signupResponse.Patient
	patientId := patient.PatientId.Int64()
	patientVisitResponse := CreatePatientVisitForPatient(patient.PatientId.Int64(), testData, t)
	SubmitPatientVisitForPatient(patientId, patientVisitResponse.PatientVisitId, testData, t)
	err = testData.DataApi.UpdatePatientPharmacy(patient.PatientId.Int64(), pharmacySelection)
	if err != nil {
		t.Fatal("Unable to update patient pharmacy: " + err.Error())
	}

	// get the doctor to add a treatment to the patient visit that we can track the status of
	StartReviewingPatientVisit(patientVisitResponse.PatientVisitId, doctor, testData, t)
	treatment1.PatientVisitId = encoding.NewObjectId(patientVisitResponse.PatientVisitId)
	treatmentResponse := addAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountId.Int64(),
		patientVisitResponse.PatientVisitId, t)
	params := url.Values{}
	params.Set("patient_visit_id", strconv.FormatInt(patientVisitResponse.PatientVisitId, 10))
	resp, err := authPost(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), doctor.AccountId.Int64())
	if err != nil {
		t.Fatalf("Unable to submit doctor visit review %s", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d for status code but got %d", http.StatusOK, resp.StatusCode)
	}

	// once the treatment has been submitted, track the status of the submitted treatment to move it to the sent state
	app_worker.ConsumeMessageFromQueue(testData.DataApi, stubErxAPI, erxStatusQueue, metrics.NewBiasedHistogram(), metrics.NewCounter(), metrics.NewCounter())

	// expected state of the treatment here is sent
	statusEvents, err := testData.DataApi.GetPrescriptionStatusEventsForTreatment(treatmentResponse.Treatments[0].Id.Int64())
	if err != nil {
		t.Fatalf("Unable to get status events for treatments: %s", err)
	} else if len(statusEvents) != 2 {
		t.Fatalf("Expected 2 status events instead got %d", len(statusEvents))
	} else if statusEvents[0].Status != api.ERX_STATUS_ERROR {
		t.Fatalf("Expected status to be %s instead it was %s", api.ERX_STATUS_SENT, statusEvents[0].Status)
	}

	pendingItems, err := testData.DataApi.GetPendingItemsInDoctorQueue(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor queue: %s", err)
	} else if len(pendingItems) != 1 {
		t.Fatalf("Expected 1 item in the pending tab of doctor queue instead got %d", len(pendingItems))
	} else if pendingItems[0].EventType != api.EVENT_TYPE_TRANSMISSION_ERROR {
		t.Fatalf("Expected a %s event type in the doctor queue instead got %s", api.EVENT_TYPE_TRANSMISSION_ERROR, pendingItems[0].EventType)
	}

	// now stub the erx api to return a "free-standing" transmission error detail for this treatment
	stubErxAPI.TransmissionErrorsForPrescriptionIds = []int64{prescriptionIdToReturn}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxAPI, metrics.NewCounter(), metrics.NewCounter())

	// there should now be 3 status events for this treatment given that
	// the rx error checker caught the missed transition from sending -> sent -> error
	statusEvents, err = testData.DataApi.GetPrescriptionStatusEventsForTreatment(treatmentResponse.Treatments[0].Id.Int64())
	if err != nil {
		t.Fatalf("Unable to get status events for treatment: %s", err)
	} else if len(statusEvents) != 2 {
		t.Fatalf("Expected 3 status events instead got %d", len(statusEvents))
	} else if statusEvents[0].Status != api.ERX_STATUS_ERROR && statusEvents[1].Status != api.ERX_STATUS_ERROR {
		t.Fatalf("Expected a transition from sent -> error, instead got %s -> %s", statusEvents[1].Status, statusEvents[0].Status)
	}

	// there should also be a pending item in the doctor's queue for the errored transmission
	pendingItems, err = testData.DataApi.GetPendingItemsInDoctorQueue(doctorId)
	if err != nil {
		t.Fatalf("Unable to get doctor queue: %s", err)
	} else if len(pendingItems) != 1 {
		t.Fatalf("Expected 1 item in the pending tab of doctor queue instead got %d", len(pendingItems))
	} else if pendingItems[0].EventType != api.EVENT_TYPE_TRANSMISSION_ERROR {
		t.Fatalf("Expected a %s event type in the doctor queue instead got %s", api.EVENT_TYPE_TRANSMISSION_ERROR, pendingItems[0].EventType)
	}
}

func TestRefillRequestInErrorAfterSentState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctor := createDoctorWithClinicianId(testData, t)
	// patientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	erxPatientId := int64(123556)
	pharmacyId := int64(12345)
	prescriptionIdForRequestedPrescription := int64(12314)
	refillRequestQueueItemId := int64(12421415)
	approvedRefillRequestPrescriptionId := int64(124424242)

	// add pharmacy to database so that it can be linked to treatment that is added
	pharmacyToReturn := &pharmacy.PharmacyData{
		SourceId:     strconv.FormatInt(pharmacyId, 10),
		Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		Name:         "Walgreens",
		AddressLine1: "116 New Montgomery",
		City:         "San Francisco",
		State:        "CA",
		Postal:       "94115",
	}
	if err := testData.DataApi.AddPharmacy(pharmacyToReturn); err != nil {
		t.Fatalf("Unable to add pharmacy to database: %s", err)
	}

	// get stub erx api to return pharmacy details
	patientToReturn := &common.Patient{
		FirstName:    "Test",
		LastName:     "TestLastName",
		Dob:          encoding.Dob{Year: 1987, Month: 1, Day: 22},
		Email:        "test@test.com",
		Gender:       "male",
		ZipCode:      "90210",
		ERxPatientId: encoding.NewObjectId(erxPatientId),
	}

	refillRequestItem := getTestRefillRequest(refillRequestQueueItemId, erxPatientId, prescriptionIdForRequestedPrescription, doctor.DoseSpotClinicianId, pharmacyId)
	stubErxApi := &erx.StubErxService{
		PharmacyDetailsToReturn:      pharmacyToReturn,
		PatientDetailsToReturn:       patientToReturn,
		RefillRxRequestQueueToReturn: []*common.RefillRequestItem{refillRequestItem},
		RefillRequestPrescriptionIds: map[int64]int64{
			refillRequestQueueItemId: approvedRefillRequestPrescriptionId,
		}, PrescriptionIdToPrescriptionStatuses: map[int64][]common.StatusEvent{
			approvedRefillRequestPrescriptionId: []common.StatusEvent{common.StatusEvent{
				Status: api.ERX_STATUS_SENT,
			},
			},
		},
	}

	// consume the refill request to store the refill request into our system
	app_worker.PerformRefillRecquestCheckCycle(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter(), "test")

	// now lets go ahead and attempt to approve this refill request
	refillRequestStatuses, err := testData.DataApi.GetPendingRefillRequestStatusEventsForClinic()
	if err != nil {
		t.Fatal("Unable to get pending refill requests from clinic: " + err.Error())
	}

	refillRequest, err := testData.DataApi.GetRefillRequestFromId(refillRequestStatuses[0].ItemId)
	if err != nil {
		t.Fatalf("Unable to get refill request from database: %s", err)
	}

	erxStatusQueue := &common.SQSQueue{}
	erxStatusQueue.QueueService = &sqs.StubSQS{}
	erxStatusQueue.QueueUrl = "local-erx"

	approveRefillRequest(refillRequest, doctor.AccountId.Int64(), "this is a test", testData.DataApi, stubErxApi, erxStatusQueue, t)

	// now that the refill request has been approved there should be an item in the message queue to check the status of the
	// prescription that was created as a result of the approval. Let's get this prescription to transition from approved -> sent
	app_worker.ConsumeMessageFromQueue(testData.DataApi, stubErxApi, erxStatusQueue, metrics.NewBiasedHistogram(), metrics.NewCounter(), metrics.NewCounter())

	// now lets get it to transition into the ERROR state
	stubErxApi.TransmissionErrorsForPrescriptionIds = []int64{approvedRefillRequestPrescriptionId}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter())

	refillStatusEvents, err := testData.DataApi.GetRefillStatusEventsForRefillRequest(refillRequest.Id)
	if err != nil {
		t.Fatalf("Unable to get refill status events: %s", err)
	} else if len(refillStatusEvents) != 4 {
		t.Fatalf("Expected 4 refill status events instead got %d", len(refillStatusEvents))
	} else if refillStatusEvents[0].Status != api.RX_REFILL_STATUS_ERROR && refillStatusEvents[1].Status != api.RX_REFILL_STATUS_SENT {
		t.Fatalf("Expected the refill request prescription to transition from SENT -> ERROR but instead it was %s -> %s ", refillRequestStatuses[1].Status, refillRequestStatuses[0].Status)
	}
}

func TestRefillRequestInErrorAfterSendingState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctor := createDoctorWithClinicianId(testData, t)
	// patientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	erxPatientId := int64(123556)
	pharmacyId := int64(12345)
	prescriptionIdForRequestedPrescription := int64(12314)
	refillRequestQueueItemId := int64(12421415)
	approvedRefillRequestPrescriptionId := int64(124424242)

	// add pharmacy to database so that it can be linked to treatment that is added
	pharmacyToReturn := &pharmacy.PharmacyData{
		SourceId:     strconv.FormatInt(pharmacyId, 10),
		Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		Name:         "Walgreens",
		AddressLine1: "116 New Montgomery",
		City:         "San Francisco",
		State:        "CA",
		Postal:       "94115",
	}
	if err := testData.DataApi.AddPharmacy(pharmacyToReturn); err != nil {
		t.Fatalf("Unable to add pharmacy to database: %s", err)
	}

	// get stub erx api to return pharmacy details
	patientToReturn := &common.Patient{
		FirstName:    "Test",
		LastName:     "TestLastName",
		Dob:          encoding.Dob{Year: 1987, Month: 1, Day: 22},
		Email:        "test@test.com",
		Gender:       "male",
		ZipCode:      "90210",
		ERxPatientId: encoding.NewObjectId(erxPatientId),
	}

	refillRequestItem := getTestRefillRequest(refillRequestQueueItemId, erxPatientId, prescriptionIdForRequestedPrescription, doctor.DoseSpotClinicianId, pharmacyId)
	stubErxApi := &erx.StubErxService{
		PharmacyDetailsToReturn:      pharmacyToReturn,
		PatientDetailsToReturn:       patientToReturn,
		RefillRxRequestQueueToReturn: []*common.RefillRequestItem{refillRequestItem},
		RefillRequestPrescriptionIds: map[int64]int64{
			refillRequestQueueItemId: approvedRefillRequestPrescriptionId,
		}, PrescriptionIdToPrescriptionStatuses: map[int64][]common.StatusEvent{
			approvedRefillRequestPrescriptionId: []common.StatusEvent{common.StatusEvent{
				Status: api.ERX_STATUS_SENT,
			},
			},
		},
	}

	// consume the refill request to store the refill request into our system
	app_worker.PerformRefillRecquestCheckCycle(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter(), "test")

	// now lets go ahead and attempt to approve this refill request
	refillRequestStatuses, err := testData.DataApi.GetPendingRefillRequestStatusEventsForClinic()
	if err != nil {
		t.Fatal("Unable to get pending refill requests from clinic: " + err.Error())
	}

	refillRequest, err := testData.DataApi.GetRefillRequestFromId(refillRequestStatuses[0].ItemId)
	if err != nil {
		t.Fatalf("Unable to get refill request from database: %s", err)
	}

	erxStatusQueue := &common.SQSQueue{}
	erxStatusQueue.QueueService = &sqs.StubSQS{}
	erxStatusQueue.QueueUrl = "local-erx"

	approveRefillRequest(refillRequest, doctor.AccountId.Int64(), "this is a test", testData.DataApi, stubErxApi, erxStatusQueue, t)

	// now lets get it to transition into the ERROR state
	stubErxApi.TransmissionErrorsForPrescriptionIds = []int64{approvedRefillRequestPrescriptionId}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter())

	refillStatusEvents, err := testData.DataApi.GetRefillStatusEventsForRefillRequest(refillRequest.Id)
	if err != nil {
		t.Fatalf("Unable to get refill status events: %s", err)
	} else if len(refillStatusEvents) != 3 {
		t.Fatalf("Expected 3 refill status events instead got %d", len(refillStatusEvents))
	} else if refillStatusEvents[0].Status != api.RX_REFILL_STATUS_ERROR && refillStatusEvents[1].Status != api.RX_REFILL_STATUS_APPROVED {
		t.Fatalf("Expected the refill request prescription to transition from SENT -> ERROR but instead it was %s -> %s ", refillRequestStatuses[1].Status, refillRequestStatuses[0].Status)
	}
}

func TestRefillRequestInErrorAfterErrorState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	doctor := createDoctorWithClinicianId(testData, t)
	// patientResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	erxPatientId := int64(123556)
	pharmacyId := int64(12345)
	prescriptionIdForRequestedPrescription := int64(12314)
	refillRequestQueueItemId := int64(12421415)
	approvedRefillRequestPrescriptionId := int64(124424242)

	// add pharmacy to database so that it can be linked to treatment that is added
	pharmacyToReturn := &pharmacy.PharmacyData{
		SourceId:     strconv.FormatInt(pharmacyId, 10),
		Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		Name:         "Walgreens",
		AddressLine1: "116 New Montgomery",
		City:         "San Francisco",
		State:        "CA",
		Postal:       "94115",
	}
	if err := testData.DataApi.AddPharmacy(pharmacyToReturn); err != nil {
		t.Fatalf("Unable to add pharmacy to database: %s", err)
	}

	// get stub erx api to return pharmacy details
	patientToReturn := &common.Patient{
		FirstName:    "Test",
		LastName:     "TestLastName",
		Dob:          encoding.Dob{Year: 1987, Month: 1, Day: 22},
		Email:        "test@test.com",
		Gender:       "male",
		ZipCode:      "90210",
		ERxPatientId: encoding.NewObjectId(erxPatientId),
	}

	refillRequestItem := getTestRefillRequest(refillRequestQueueItemId, erxPatientId, prescriptionIdForRequestedPrescription, doctor.DoseSpotClinicianId, pharmacyId)
	stubErxApi := &erx.StubErxService{
		PharmacyDetailsToReturn:      pharmacyToReturn,
		PatientDetailsToReturn:       patientToReturn,
		RefillRxRequestQueueToReturn: []*common.RefillRequestItem{refillRequestItem},
		RefillRequestPrescriptionIds: map[int64]int64{
			refillRequestQueueItemId: approvedRefillRequestPrescriptionId,
		}, PrescriptionIdToPrescriptionStatuses: map[int64][]common.StatusEvent{
			approvedRefillRequestPrescriptionId: []common.StatusEvent{common.StatusEvent{
				Status:        api.ERX_STATUS_ERROR,
				StatusDetails: "Error state",
			},
			},
		},
	}

	// consume the refill request to store the refill request into our system
	app_worker.PerformRefillRecquestCheckCycle(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter(), "test")

	// now lets go ahead and attempt to approve this refill request
	refillRequestStatuses, err := testData.DataApi.GetPendingRefillRequestStatusEventsForClinic()
	if err != nil {
		t.Fatal("Unable to get pending refill requests from clinic: " + err.Error())
	}

	refillRequest, err := testData.DataApi.GetRefillRequestFromId(refillRequestStatuses[0].ItemId)
	if err != nil {
		t.Fatalf("Unable to get refill request from database: %s", err)
	}

	erxStatusQueue := &common.SQSQueue{}
	erxStatusQueue.QueueService = &sqs.StubSQS{}
	erxStatusQueue.QueueUrl = "local-erx"

	approveRefillRequest(refillRequest, doctor.AccountId.Int64(), "this is a test", testData.DataApi, stubErxApi, erxStatusQueue, t)

	// now that the refill request has been approved there should be an item in the message queue to check the status of the
	// prescription that was created as a result of the approval. Let's get this prescription to transition from approved -> sent
	app_worker.ConsumeMessageFromQueue(testData.DataApi, stubErxApi, erxStatusQueue, metrics.NewBiasedHistogram(), metrics.NewCounter(), metrics.NewCounter())

	// now lets get it to transition into the ERROR state
	stubErxApi.TransmissionErrorsForPrescriptionIds = []int64{approvedRefillRequestPrescriptionId}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter())

	refillStatusEvents, err := testData.DataApi.GetRefillStatusEventsForRefillRequest(refillRequest.Id)
	if err != nil {
		t.Fatalf("Unable to get refill status events: %s", err)
	} else if len(refillStatusEvents) != 3 {
		t.Fatalf("Expected 3 refill status events instead got %d", len(refillStatusEvents))
	} else if refillStatusEvents[0].Status != api.RX_REFILL_STATUS_ERROR && refillStatusEvents[1].Status != api.RX_REFILL_STATUS_APPROVED {
		t.Fatalf("Expected the refill request prescription to transition from APPROVED -> ERROR but instead it was %s -> %s ", refillRequestStatuses[1].Status, refillRequestStatuses[0].Status)
	}
}

// Test unlinked dntf treatment that has an error after being in sent state
func TestUnlinkedDNTFTreatmentSentToErrorState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	unlinkedTreatment := setUpDeniedRefillRequestWithDNTF(t, testData, common.StatusEvent{Status: api.ERX_STATUS_SENT}, false)

	// lets go ahead and setup the stubErxApi to throw a transmission error for this treatment now
	stubErxApi := &erx.StubErxService{
		TransmissionErrorsForPrescriptionIds: []int64{unlinkedTreatment.ERx.PrescriptionId.Int64()},
	}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter())

	unlinkedTreatment, err := testData.DataApi.GetUnlinkedDNTFTreatment(unlinkedTreatment.Id.Int64())
	if err != nil {
		t.Fatalf(err.Error())
	} else if len(unlinkedTreatment.ERx.RxHistory) != 4 {
		t.Fatalf("Expected 4 items in the rx history of an unlinked dntf treatment")
	} else if unlinkedTreatment.ERx.RxHistory[0].Status != api.ERX_STATUS_ERROR || unlinkedTreatment.ERx.RxHistory[1].Status != api.ERX_STATUS_SENT {
		t.Fatalf("Expected rx history to go from Sent -> Error instead it was frmo %s -> %s", unlinkedTreatment.ERx.RxHistory[1].Status, unlinkedTreatment.ERx.RxHistory[0].Status)
	}
}

func TestUnlinkedDNTFTreatmentSendingToErrorState(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}

	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	unlinkedTreatment := setUpDeniedRefillRequestWithDNTF(t, testData, common.StatusEvent{Status: api.ERX_STATUS_ERROR}, false)

	// lets go ahead and setup the stubErxApi to throw a transmission error for this treatment now
	stubErxApi := &erx.StubErxService{
		TransmissionErrorsForPrescriptionIds: []int64{unlinkedTreatment.ERx.PrescriptionId.Int64()},
	}
	app_worker.PerformRxErrorCheck(testData.DataApi, stubErxApi, metrics.NewCounter(), metrics.NewCounter())

	unlinkedTreatment, err := testData.DataApi.GetUnlinkedDNTFTreatment(unlinkedTreatment.Id.Int64())
	if err != nil {
		t.Fatalf(err.Error())
	} else if len(unlinkedTreatment.ERx.RxHistory) != 3 {
		t.Fatalf("Expected 4 items in the rx history of an unlinked dntf treatment")
	} else if unlinkedTreatment.ERx.RxHistory[0].Status != api.ERX_STATUS_ERROR || unlinkedTreatment.ERx.RxHistory[1].Status != api.ERX_STATUS_SENDING {
		t.Fatalf("Expected rx history to go from Sending -> Error instead it was from %s -> %s", unlinkedTreatment.ERx.RxHistory[1].Status, unlinkedTreatment.ERx.RxHistory[0].Status)
	}
}
