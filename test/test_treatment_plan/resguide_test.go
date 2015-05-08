package test_treatment_plan

import (
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/responses"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestTPResourceGuides(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := test_integration.GetDoctorIDOfCurrentDoctor(testData, t)
	doctor, err := testData.DataAPI.GetDoctorFromID(doctorID)
	test.OK(t, err)
	doctorCli := test_integration.DoctorClient(testData, t, doctorID)

	_, tp0 := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	tp, err := doctorCli.TreatmentPlan(tp0.ID.Int64(), false, doctor_treatment_plan.ResourceGuidesSection)
	test.OK(t, err)
	test.Equals(t, 0, len(tp.ResourceGuides))

	_, guideIDs := test_integration.CreateTestResourceGuides(t, testData)

	test.OK(t, doctorCli.AddResourceGuidesToTreatmentPlan(tp.ID.Int64(), guideIDs))

	tp, err = doctorCli.TreatmentPlan(tp.ID.Int64(), false, doctor_treatment_plan.ResourceGuidesSection)
	test.OK(t, err)
	test.Equals(t, 2, len(tp.ResourceGuides))

	// Should be idempotent
	test.OK(t, doctorCli.AddResourceGuidesToTreatmentPlan(tp.ID.Int64(), guideIDs))

	tp, err = doctorCli.TreatmentPlan(tp.ID.Int64(), false, doctor_treatment_plan.ResourceGuidesSection)
	test.OK(t, err)
	test.Equals(t, 2, len(tp.ResourceGuides))

	test.OK(t, doctorCli.RemoveResourceGuideFromTreatmentPlan(tp.ID.Int64(), guideIDs[1]))

	tp, err = doctorCli.TreatmentPlan(tp.ID.Int64(), false, doctor_treatment_plan.ResourceGuidesSection)
	test.OK(t, err)
	test.Equals(t, 1, len(tp.ResourceGuides))

	test.OK(t, doctorCli.AddResourceGuidesToTreatmentPlan(tp.ID.Int64(), guideIDs))
	test.OK(t, testData.DataAPI.RemoveResourceGuidesFromTreatmentPlan(tp.ID.Int64(), guideIDs))
	tp, err = doctorCli.TreatmentPlan(tp.ID.Int64(), false, doctor_treatment_plan.ResourceGuidesSection)
	test.OK(t, err)
	test.Equals(t, 0, len(tp.ResourceGuides))
}

func TestFTPResourceGuides(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := test_integration.GetDoctorIDOfCurrentDoctor(testData, t)
	doctor, err := testData.DataAPI.GetDoctorFromID(doctorID)
	test.OK(t, err)
	doctorCli := test_integration.DoctorClient(testData, t, doctorID)

	_, guideIDs := test_integration.CreateTestResourceGuides(t, testData)

	// Create a patient treatment plan, and save a draft message
	visit, tp0 := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	test_integration.AddTreatmentsToTreatmentPlan(tp0.ID.Int64(), doctor, t, testData)
	test_integration.AddRegimenPlanForTreatmentPlan(tp0.ID.Int64(), doctor, t, testData)
	test.OK(t, doctorCli.AddResourceGuidesToTreatmentPlan(tp0.ID.Int64(), guideIDs))

	// Refetch the treatment plan to fill in with recent updates
	tp, err := doctorCli.TreatmentPlan(tp0.ID.Int64(), false, doctor_treatment_plan.AllSections)
	test.OK(t, err)
	test.Equals(t, 2, len(tp.ResourceGuides))

	ftp := &responses.FavoriteTreatmentPlan{
		Name:          "Test FTP",
		TreatmentList: tp.TreatmentList,
		RegimenPlan:   tp.RegimenPlan,
		Note:          tp.Note,
	}

	// Test creating ftp when resource guides don't match
	_, err = doctorCli.CreateFavoriteTreatmentPlanFromTreatmentPlan(ftp, tp.ID.Int64())
	test.Equals(t, false, err == nil)

	ftp.ResourceGuides = tp.ResourceGuides
	_, err = doctorCli.CreateFavoriteTreatmentPlanFromTreatmentPlan(ftp, tp.ID.Int64())
	test.OK(t, err)

	ftps, err := doctorCli.ListFavoriteTreatmentPlansForTag(api.AcnePathwayTag)
	test.OK(t, err)
	test.Equals(t, 1, len(ftps))
	test.Equals(t, len(tp.ResourceGuides), len(ftps[0].ResourceGuides))

	// Make sure treatment plan created from an ftp that has resource guides also
	// gets the guides.
	tp, err = doctorCli.PickTreatmentPlanForVisit(visit.PatientVisitID, ftps[0])
	test.OK(t, err)
	test.Equals(t, len(ftps[0].ResourceGuides), len(tp.ResourceGuides))

	// update the note so that we can submit the treatment plan
	err = doctorCli.UpdateTreatmentPlanNote(tp.ID.Int64(), "Some note")
	test.OK(t, err)
	err = doctorCli.SubmitTreatmentPlan(tp.ID.Int64())
	test.OK(t, err)

	// ensure that even after submitting the TP the resource guides are still there
	tp, err = doctorCli.TreatmentPlan(tp.ID.Int64(), false, doctor_treatment_plan.AllSections)
	test.OK(t, err)
	test.Equals(t, len(ftps[0].ResourceGuides), len(tp.ResourceGuides))

	// ensure that we can delete the ftp and tp that has resource guide

	// lets create yet another TP so that we have a draft that we can then try to delete
	tp, err = doctorCli.PickTreatmentPlanForVisit(visit.PatientVisitID, ftps[0])
	test.OK(t, err)

	err = doctorCli.DeleteFavoriteTreatmentPlan(ftps[0].ID.Int64())
	test.OK(t, err)

	err = doctorCli.DeleteTreatmentPlan(tp.ID.Int64())
	test.OK(t, err)
}
