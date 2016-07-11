package test_demo

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice/apipaths"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/demo"
	"github.com/sprucehealth/backend/cmd/svc/restapi/diagnosis/handlers"
	"github.com/sprucehealth/backend/cmd/svc/restapi/test/test_integration"
	"github.com/sprucehealth/backend/libs/test"
)

func TestTrainingCase(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close(t)
	testData.StartAPIServer(t)

	// Upload the latest versions of the review and intake
	latestIntakeVersion := determineLatestVersionedFile("intake-", t)
	latestReviewVersion := determineLatestVersionedFile("review-", t)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	test_integration.AddFileToMultipartWriter(writer, "intake", latestIntakeVersion, "../data/"+latestIntakeVersion, t)
	test_integration.AddFileToMultipartWriter(writer, "review", latestReviewVersion, "../data/"+latestReviewVersion, t)
	test_integration.AddFieldToMultipartWriter(writer, "patient_app_version", "1.0.0", t)
	test_integration.AddFieldToMultipartWriter(writer, "doctor_app_version", "1.0.0", t)
	test_integration.AddFieldToMultipartWriter(writer, "platform", "iOS", t)
	err := writer.Close()
	test.OK(t, err)
	resp, err := testData.AdminAuthPost(testData.AdminAPIServer.URL+`/admin/api/layout`, writer.FormDataContentType(), body, testData.AdminUser)
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// have the demo worker run ones to create the training cases
	demo.LocalServerURL = testData.APIServer.URL
	w := demo.NewWorker(testData.DataAPI, &test_integration.TestLock{}, "www.spruce.local", "us-east-1")
	w.CacheQAInformation()

	// create training cases
	test.OK(t, w.Do())

	// check for number of pending training cases. It should be greater than 0
	pendingTrainingCases, err := testData.DataAPI.TrainingCaseSetCount(common.TCSStatusPending)
	test.OK(t, err)
	test.Equals(t, true, pendingTrainingCases > 0)

	// lets get a doctor to claim 1 training case set
	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)
	resp, err = testData.AuthPost(testData.APIServer.URL+apipaths.TrainingCasesURLPath, "", nil, doctor.AccountID.Int64())
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)
	doctorCli := test_integration.DoctorClient(testData, t, dr.DoctorID)

	// now the doctor should have non-zero number of pending cases in their inbox
	pendingVisits, err := testData.DataAPI.GetPendingItemsInDoctorQueue(dr.DoctorID)
	test.OK(t, err)
	test.Equals(t, true, len(pendingVisits) > 0)

	// now lets go ahead and try to diagnose one of those cases up until the point of visit submission
	patientVisit, err := testData.DataAPI.GetPatientVisitFromID(pendingVisits[0].ItemID)
	test.OK(t, err)
	test_integration.GrantDoctorAccessToPatientCase(t, testData, doctor, patientVisit.PatientCaseID.Int64())
	test_integration.StartReviewingPatientVisit(patientVisit.ID.Int64(), doctor, testData, t)
	test.OK(t, doctorCli.CreateDiagnosisSet(&handlers.DiagnosisListRequestData{
		VisitID: patientVisit.ID.Int64(),
		Diagnoses: []*handlers.DiagnosisInputItem{
			{
				CodeID: "diag_l730",
			},
		},
	}))
	tp := test_integration.PickATreatmentPlan(&common.TreatmentPlanParent{
		ParentID:   patientVisit.ID,
		ParentType: common.TPParentTypePatientVisit,
	}, nil, doctor, testData, t)
	test_integration.SubmitPatientVisitBackToPatient(tp.TreatmentPlan.ID.Int64(), doctor, testData, t)
}

func determineLatestVersionedFile(prefix string, t *testing.T) string {
	files, err := ioutil.ReadDir("../data/")
	test.OK(t, err)

	var fileNamesToCompare []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) && strings.HasSuffix(f.Name(), ".json") {
			fileNamesToCompare = append(fileNamesToCompare, f.Name())
		}
	}

	if len(fileNamesToCompare) > 0 {
		sort.Strings(fileNamesToCompare)
	} else {
		t.Fatalf("File with prefix %s not found", prefix)
	}

	return fileNamesToCompare[len(fileNamesToCompare)-1]
}