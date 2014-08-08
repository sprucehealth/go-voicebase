package test_doctor_queue

import (
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/messages"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestConversationItemsInDoctorQueue(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)
	doctorID := test_integration.GetDoctorIdOfCurrentDoctor(testData, t)
	doctor, err := testData.DataApi.GetDoctorFromId(doctorID)
	test.OK(t, err)

	visit, treatmentPlan := test_integration.CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	patient, err := testData.DataApi.GetPatientFromPatientVisitId(visit.PatientVisitId)
	test.OK(t, err)
	test_integration.SubmitPatientVisitBackToPatient(treatmentPlan.Id.Int64(), doctor, testData, t)

	caseID, err := testData.DataApi.GetPatientCaseIdFromPatientVisitId(visit.PatientVisitId)
	test.OK(t, err)

	test_integration.PostCaseMessage(t, testData, patient.AccountId.Int64(), &messages.PostMessageRequest{
		CaseID:  caseID,
		Message: "foo",
	})

	// ensure that an item is inserted into the doctor queue
	pendingItems, err := testData.DataApi.GetPendingItemsInDoctorQueue(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor queue: %s", err)
	} else if len(pendingItems) != 1 {
		t.Fatalf("Expected 1 item in the pending items but got %d instead", len(pendingItems))
	} else if pendingItems[0].EventType != api.DQEventTypeCaseMessage {
		t.Fatalf("Expected item type to be %s instead it was %s", api.DQEventTypeCaseMessage, pendingItems[0].EventType)
	} else if pendingItems[0].Status != api.DQItemStatusPending {
		t.Fatalf("Expected item to have status %s instead it has %s", api.DQItemStatusPending, pendingItems[0].Status)
	}

	// Reply
	test_integration.PostCaseMessage(t, testData, doctor.AccountId.Int64(), &messages.PostMessageRequest{
		CaseID:  caseID,
		Message: "bar",
	})

	// ensure that the item is marked as completed for the doctor
	pendingItems, err = testData.DataApi.GetPendingItemsInDoctorQueue(doctorID)
	if err != nil {
		t.Fatalf("Unable to get doctor queue: %s", err)
	} else if len(pendingItems) != 0 {
		t.Fatalf("Expected no item in the pending items but got %d instead", len(pendingItems))
	}

	completedItems, err := testData.DataApi.GetCompletedItemsInDoctorQueue(doctorID)
	if err != nil {
		t.Fatalf("Unable to get completed items in the doctor queue: %s", err)
	} else if len(completedItems) != 2 { // one for message, one for treatment plan
		for _, item := range completedItems {
			t.Logf("%+v", item)
		}
		t.Fatalf("Expected 2 items in the completed tab instead got %d", len(completedItems))
	} else if completedItems[1].EventType != api.DQEventTypeCaseMessage {
		t.Fatalf("Expected item of type %s instead got %s", api.DQEventTypeCaseMessage, completedItems[0].EventType)
	} else if completedItems[1].Status != api.DQItemStatusReplied {
		t.Fatalf("Expecte item to have status %s instead it has %s", api.DQItemStatusReplied, completedItems[0].Status)
	}
}
