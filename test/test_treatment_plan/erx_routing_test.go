package test_treatment_plan

import (
	"testing"

	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/pharmacy"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestERXRouting_RXStarted(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close(t)
	testData.Config.ERxRouting = true
	testData.StartAPIServer(t)

	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	_, tp := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// add a treatment
	treatment1 := &common.Treatment{
		DrugInternalName: "Drug1 (Route1 - Form1)",
		TreatmentPlanID:  tp.ID,
		DosageStrength:   "Strength1",
		DispenseValue:    1,
		DispenseUnitID:   encoding.DeprecatedNewObjectID(26),
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
		DrugDBIDs: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	treatmentsResponse := test_integration.AddAndGetTreatmentsForPatientVisit(
		testData, []*common.Treatment{treatment1}, doctor.AccountID.Int64(), tp.ID.Int64(), t)

	// submit the treatment plan back to the patient
	test_integration.SubmitPatientVisitBackToPatient(tp.ID.Int64(), doctor, testData, t)

	// now lets act as though we are in a state where the patient and the treatments have all the erx information needed
	patient, err := testData.DataAPI.GetPatientFromID(tp.PatientID)
	test.OK(t, err)

	erxPatientID := int64(100)
	treatmentPrescriptionID := int64(105)
	err = testData.DataAPI.UpdatePatient(patient.ID, &api.PatientUpdate{
		ERxID: &erxPatientID,
	}, false)
	test.OK(t, err)

	treatmentsResponse.TreatmentList.Treatments[0].ERx = &common.ERxData{
		PrescriptionID: encoding.DeprecatedNewObjectID(treatmentPrescriptionID),
	}

	pharmacySelection := &pharmacy.PharmacyData{
		SourceID:     12345,
		Source:       pharmacy.PharmacySourceSurescripts,
		AddressLine1: "12345 Marin Street",
		City:         "San Francisco",
		State:        "CA",
		Phone:        "12345667",
	}

	err = testData.DataAPI.UpdatePatientPharmacy(patient.ID, pharmacySelection)
	test.OK(t, err)

	err = testData.DataAPI.StartRXRoutingForTreatmentsAndTreatmentPlan(treatmentsResponse.TreatmentList.Treatments, pharmacySelection, tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)

	// at this point the treatment plan is in the rx started state
	// lets go ahead and call the worker to complete the rest of the steps to ensure that its successfully able to activate the treatment plan
	// after routing the prescriptions
	stubERxAPI := testData.Config.ERxAPI.(*erx.StubErxService)
	stubERxAPI.PrescriptionIDToPrescriptionStatuses = map[int64][]common.StatusEvent{
		treatmentPrescriptionID: []common.StatusEvent{{Status: api.ERXStatusEntered}},
	}
	consumed, err := doctor_treatment_plan.NewWorker(
		testData.DataAPI, stubERxAPI, testData.Config.Dispatcher, testData.Config.ERxRoutingQueue,
		testData.Config.ERxStatusQueue, 0, metrics.NewRegistry(),
	).Do()
	test.OK(t, err)
	t.Logf("Doctor treatment plan worker consume message: %t", consumed)

	// at this point the treatment plan should be activated
	treatmentPlan, err := testData.DataAPI.GetAbridgedTreatmentPlan(tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)
	test.Equals(t, common.TPStatusActive, treatmentPlan.Status)

	// there should also be a case message for the patient
	caseMessages, err := testData.DataAPI.ListCaseMessages(treatmentPlan.PatientCaseID.Int64(), 0)
	test.OK(t, err)
	test.Equals(t, 1, len(caseMessages))
}

func TestERXRouting_RXSent(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close(t)
	testData.Config.ERxRouting = true
	testData.StartAPIServer(t)

	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	_, tp := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// add a treatment
	treatment1 := &common.Treatment{
		DrugInternalName: "Drug1 (Route1 - Form1)",
		TreatmentPlanID:  tp.ID,
		DosageStrength:   "Strength1",
		DispenseValue:    1,
		DispenseUnitID:   encoding.DeprecatedNewObjectID(26),
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
		DrugDBIDs: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	treatmentsResponse := test_integration.AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountID.Int64(), tp.ID.Int64(), t)

	// submit the treatment plan back to the patient
	test_integration.SubmitPatientVisitBackToPatient(tp.ID.Int64(), doctor, testData, t)

	// now lets act as though we are in a state where the patient and the treatments have all the erx information needed
	patient, err := testData.DataAPI.GetPatientFromID(tp.PatientID)
	test.OK(t, err)

	erxPatientID := int64(100)
	treatmentPrescriptionID := int64(105)
	err = testData.DataAPI.UpdatePatient(patient.ID, &api.PatientUpdate{
		ERxID: &erxPatientID,
	}, false)
	test.OK(t, err)

	treatmentsResponse.TreatmentList.Treatments[0].ERx = &common.ERxData{
		PrescriptionID: encoding.DeprecatedNewObjectID(treatmentPrescriptionID),
	}

	pharmacySelection := &pharmacy.PharmacyData{
		SourceID:     12345,
		Source:       pharmacy.PharmacySourceSurescripts,
		AddressLine1: "12345 Marin Street",
		City:         "San Francisco",
		State:        "CA",
		Phone:        "12345667",
	}

	err = testData.DataAPI.UpdatePatientPharmacy(patient.ID, pharmacySelection)
	test.OK(t, err)

	err = testData.DataAPI.StartRXRoutingForTreatmentsAndTreatmentPlan(treatmentsResponse.TreatmentList.Treatments, pharmacySelection, tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)

	// at this point the treatment plan is in the rx started state
	// lets go ahead and call the worker to complete the rest of the steps to ensure that its successfully able to activate the treatment plan
	// after routing the prescriptions
	stubERxAPI := testData.Config.ERxAPI.(*erx.StubErxService)
	stubERxAPI.PrescriptionIDToPrescriptionStatuses = map[int64][]common.StatusEvent{
		treatmentPrescriptionID: []common.StatusEvent{{Status: api.ERXStatusSent}},
	}
	consumed, err := doctor_treatment_plan.NewWorker(
		testData.DataAPI, stubERxAPI, testData.Config.Dispatcher, testData.Config.ERxRoutingQueue,
		testData.Config.ERxStatusQueue, 0, metrics.NewRegistry(),
	).Do()
	test.OK(t, err)
	t.Logf("Doctor treatment plan worker consume message: %t", consumed)

	// at this point the treatment plan should be activated
	treatmentPlan, err := testData.DataAPI.GetAbridgedTreatmentPlan(tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)
	test.Equals(t, common.TPStatusActive, treatmentPlan.Status)
	test.Equals(t, true, treatmentPlan.SentDate != nil)

	// there should also be a case message for the patient
	caseMessages, err := testData.DataAPI.ListCaseMessages(treatmentPlan.PatientCaseID.Int64(), 0)
	test.OK(t, err)
	test.Equals(t, 1, len(caseMessages))
}

func TestERxRouting_CaseMessageExistsAlready(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close(t)
	testData.Config.ERxRouting = true
	testData.StartAPIServer(t)

	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	_, tp := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// add a treatment
	treatment1 := &common.Treatment{
		DrugInternalName: "Drug1 (Route1 - Form1)",
		TreatmentPlanID:  tp.ID,
		DosageStrength:   "Strength1",
		DispenseValue:    1,
		DispenseUnitID:   encoding.DeprecatedNewObjectID(26),
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
		DrugDBIDs: map[string]string{
			"drug_db_id_1": "12315",
			"drug_db_id_2": "124",
		},
	}

	treatmentsResponse := test_integration.AddAndGetTreatmentsForPatientVisit(testData, []*common.Treatment{treatment1}, doctor.AccountID.Int64(), tp.ID.Int64(), t)

	// submit the treatment plan back to the patient
	test_integration.SubmitPatientVisitBackToPatient(tp.ID.Int64(), doctor, testData, t)

	// now lets act as though we are in a state where the patient and the treatments have all the erx information needed
	patient, err := testData.DataAPI.GetPatientFromID(tp.PatientID)
	test.OK(t, err)

	erxPatientID := int64(100)
	err = testData.DataAPI.UpdatePatient(patient.ID, &api.PatientUpdate{
		ERxID: &erxPatientID,
	}, false)
	test.OK(t, err)

	treatmentsResponse.TreatmentList.Treatments[0].ERx = &common.ERxData{
		PrescriptionID: encoding.DeprecatedNewObjectID(105),
	}

	pharmacySelection := &pharmacy.PharmacyData{
		SourceID:     12345,
		Source:       pharmacy.PharmacySourceSurescripts,
		AddressLine1: "12345 Marin Street",
		City:         "San Francisco",
		State:        "CA",
		Phone:        "12345667",
	}

	err = testData.DataAPI.UpdatePatientPharmacy(patient.ID, pharmacySelection)
	test.OK(t, err)

	err = testData.DataAPI.StartRXRoutingForTreatmentsAndTreatmentPlan(treatmentsResponse.TreatmentList.Treatments, pharmacySelection, tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)

	// now lets go ahead and activate the treatment plan as well as send the case message for the patient
	err = testData.DataAPI.ActivateTreatmentPlan(tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)

	caseMessage := &common.CaseMessage{
		CaseID:   tp.PatientCaseID.Int64(),
		PersonID: doctor.PersonID,
		Body:     "foo",
		Attachments: []*common.CaseMessageAttachment{
			{
				ItemType: common.AttachmentTypeTreatmentPlan,
				ItemID:   tp.ID.Int64(),
			},
		},
	}
	_, err = testData.DataAPI.CreateCaseMessage(caseMessage)
	test.OK(t, err)

	// now lets go ahead and get the worker to consume the message
	consumed, err := doctor_treatment_plan.NewWorker(
		testData.DataAPI, testData.Config.ERxAPI, testData.Config.Dispatcher, testData.Config.ERxRoutingQueue,
		testData.Config.ERxStatusQueue, 0, metrics.NewRegistry(),
	).Do()
	test.OK(t, err)
	t.Logf("Doctor treatment plan worker consume message: %t", consumed)

	// at this point the treatment plan should be activated
	treatmentPlan, err := testData.DataAPI.GetAbridgedTreatmentPlan(tp.ID.Int64(), doctor.ID.Int64())
	test.OK(t, err)
	test.Equals(t, common.TPStatusActive, treatmentPlan.Status)

	// there should also be just a single case message for the patient
	caseMessages, err := testData.DataAPI.ListCaseMessages(treatmentPlan.PatientCaseID.Int64(), 0)
	test.OK(t, err)
	test.Equals(t, 1, len(caseMessages))
}
