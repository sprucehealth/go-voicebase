package test_diagnosis

import (
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/diagnosis"
	"github.com/sprucehealth/backend/diagnosis/icd10"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestDiagnosisSet(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	// add a couple diagnosis for testing purposes
	code1 := "T1.0"
	d1 := &icd10.Diagnosis{
		Code:        code1,
		Description: "Test1.0",
		Billable:    true,
	}

	code2 := "T2.0"
	d2 := &icd10.Diagnosis{
		Code:        code2,
		Description: "Test2.0",
		Billable:    true,
	}
	err := icd10.SetDiagnoses(testData.DB, map[string]*icd10.Diagnosis{
		d1.Code: d1,
		d2.Code: d2,
	})
	test.OK(t, err)

	admin := test_integration.CreateRandomAdmin(t, testData)

	test_integration.UploadDetailsLayoutForDiagnosis(`
	{
	"diagnosis_layouts" : [
	{
		"code" : "T1.0",
		"layout_version" : "1.0.0",
		"questions" : [
		{
			"question" : "q_acne_severity",
			"additional_fields": {
      			 "style": "brief_title"
      		}
		},
		{
			"question" : "q_acne_type"
		}]
	}
	]
	}`, admin.AccountID.Int64(), testData, t)

	// lets get the questionID and answerIDs of the questions
	questionInfos, err := testData.DataAPI.GetQuestionInfoForTags([]string{"q_acne_severity", "q_acne_type"}, api.EN_LANGUAGE_ID)
	test.OK(t, err)

	answerInfos, err := testData.DataAPI.GetAnswerInfoForTags([]string{"a_doctor_acne_severity_moderate", "a_acne_comedonal"}, api.EN_LANGUAGE_ID)
	test.OK(t, err)

	var codeID1, codeID2 int64
	err = testData.DB.QueryRow(`SELECT id FROM diagnosis_code WHERE code = ?`, code1).Scan(&codeID1)
	test.OK(t, err)
	err = testData.DB.QueryRow(`SELECT id FROM diagnosis_code WHERE code = ?`, code2).Scan(&codeID2)
	test.OK(t, err)

	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	doctorClient := test_integration.DoctorClient(testData, t, dr.DoctorID)

	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	pv, _ := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	// create diagnosis set including both diagnosis codes
	intakeData := &apiservice.IntakeData{
		Questions: []*apiservice.QuestionAnswerItem{
			&apiservice.QuestionAnswerItem{
				QuestionID: questionInfos[0].QuestionID,
				AnswerIntakes: []*apiservice.AnswerItem{
					&apiservice.AnswerItem{
						PotentialAnswerID: answerInfos[0].AnswerID,
					},
				},
			},
			&apiservice.QuestionAnswerItem{
				QuestionID: questionInfos[1].QuestionID,
				AnswerIntakes: []*apiservice.AnswerItem{
					&apiservice.AnswerItem{
						PotentialAnswerID: answerInfos[1].AnswerID,
					},
				},
			},
		},
	}

	note := "testing w/ this note"
	err = doctorClient.CreateDiagnosisSet(&diagnosis.DiagnosisListRequestData{
		VisitID:      pv.PatientVisitID,
		InternalNote: note,
		Diagnoses: []*diagnosis.DiagnosisInputItem{
			&diagnosis.DiagnosisInputItem{
				CodeID: codeID1,
				LayoutVersion: &common.Version{
					Major: 1,
					Minor: 0,
					Patch: 0,
				},
				Answers: intakeData,
			},
			&diagnosis.DiagnosisInputItem{
				CodeID: codeID2,
			},
		},
	})
	test.OK(t, err)

	// get the diagnosis to test that it was set as expected
	diagnosisListResponse, err := doctorClient.ListDiagnosis(pv.PatientVisitID)
	test.OK(t, err)
	test.Equals(t, 2, len(diagnosisListResponse.Diagnoses))
	test.Equals(t, 2, len(diagnosisListResponse.Diagnoses[0].Questions))
	test.Equals(t, note, diagnosisListResponse.Notes)
	test.Equals(t, false, diagnosisListResponse.CaseManagement.Unsuitable)
	test.Equals(t, d1.Description, diagnosisListResponse.Diagnoses[0].Title)
	test.Equals(t, codeID1, diagnosisListResponse.Diagnoses[0].CodeID)
	test.Equals(t, codeID2, diagnosisListResponse.Diagnoses[1].CodeID)
	test.Equals(t, true, diagnosisListResponse.Diagnoses[0].HasDetails)
	test.Equals(t, false, diagnosisListResponse.Diagnoses[1].HasDetails)
	test.Equals(t, "1.0.0", diagnosisListResponse.Diagnoses[0].LayoutVersion.String())
	test.Equals(t, "1.0.0", diagnosisListResponse.Diagnoses[0].LatestLayoutVersion.String())
	test.Equals(t, true, diagnosisListResponse.Diagnoses[0].Answers.Equals(intakeData))

	// lets update the diagnosis set to remove one code and the note as well
	note = "updated note"
	err = doctorClient.CreateDiagnosisSet(&diagnosis.DiagnosisListRequestData{
		VisitID:      pv.PatientVisitID,
		InternalNote: note,
		Diagnoses: []*diagnosis.DiagnosisInputItem{
			&diagnosis.DiagnosisInputItem{
				CodeID: codeID1,
				LayoutVersion: &common.Version{
					Major: 1,
					Minor: 0,
					Patch: 0,
				},
				Answers: intakeData,
			},
		},
	})
	test.OK(t, err)

	diagnosisListResponse, err = doctorClient.ListDiagnosis(pv.PatientVisitID)
	test.OK(t, err)
	test.Equals(t, 1, len(diagnosisListResponse.Diagnoses))
	test.Equals(t, codeID1, diagnosisListResponse.Diagnoses[0].CodeID)
	test.Equals(t, note, diagnosisListResponse.Notes)

	// now lets update the layout for code1 and ensure that the latest layout version is updated
	// to indicate the new one
	test_integration.UploadDetailsLayoutForDiagnosis(`
	{
	"diagnosis_layouts" : [
	{
		"code" : "T1.0",
		"layout_version" : "1.1.0",
		"questions" : [
		{
			"question" : "q_acne_severity",
			"additional_fields": {
      			 "style": "brief_title"
      		}
		}]
	}
	]
	}`, admin.AccountID.Int64(), testData, t)

	diagnosisListResponse, err = doctorClient.ListDiagnosis(pv.PatientVisitID)
	test.OK(t, err)
	test.Equals(t, "1.1.0", diagnosisListResponse.Diagnoses[0].LatestLayoutVersion.String())
	test.Equals(t, "1.0.0", diagnosisListResponse.Diagnoses[0].LayoutVersion.String())

}

func TestDiagnosisSet_MarkUnsuitable(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	// add a couple diagnosis for testing purposes
	code1 := "T1.0"
	d1 := &icd10.Diagnosis{
		Code:        code1,
		Description: "Test1.0",
		Billable:    true,
	}

	code2 := "T2.0"
	d2 := &icd10.Diagnosis{
		Code:        code2,
		Description: "Test2.0",
		Billable:    true,
	}
	err := icd10.SetDiagnoses(testData.DB, map[string]*icd10.Diagnosis{
		d1.Code: d1,
		d2.Code: d2,
	})
	test.OK(t, err)

	var codeID1, codeID2 int64
	err = testData.DB.QueryRow(`SELECT id FROM diagnosis_code WHERE code = ?`, code1).Scan(&codeID1)
	test.OK(t, err)
	err = testData.DB.QueryRow(`SELECT id FROM diagnosis_code WHERE code = ?`, code2).Scan(&codeID2)
	test.OK(t, err)

	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	doctorClient := test_integration.DoctorClient(testData, t, dr.DoctorID)

	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	pv, _ := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)

	note := "testing w/ this note"
	unsuitableReason := "deal with it"
	err = doctorClient.CreateDiagnosisSet(&diagnosis.DiagnosisListRequestData{
		VisitID:      pv.PatientVisitID,
		InternalNote: note,
		Diagnoses: []*diagnosis.DiagnosisInputItem{
			&diagnosis.DiagnosisInputItem{
				CodeID: codeID1,
			},
			&diagnosis.DiagnosisInputItem{
				CodeID: codeID2,
			},
		},
		CaseManagement: diagnosis.CaseManagementItem{
			Unsuitable: true,
			Reason:     unsuitableReason,
		},
	})
	test.OK(t, err)

	// ensure that the case is marked as being triaged out
	visit, err := testData.DataAPI.GetPatientVisitFromID(pv.PatientVisitID)
	test.OK(t, err)
	test.Equals(t, common.PVStatusTriaged, visit.Status)

	diagnosisListResponse, err := doctorClient.ListDiagnosis(pv.PatientVisitID)
	test.OK(t, err)
	test.Equals(t, true, diagnosisListResponse.CaseManagement.Unsuitable)
	test.Equals(t, unsuitableReason, diagnosisListResponse.CaseManagement.Reason)

}
