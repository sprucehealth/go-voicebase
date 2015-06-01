package test_integration

import (
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/messages"
	"github.com/sprucehealth/backend/test"
)

func TestPersonCreation(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	// Make sure a person row is inserted when creating a patient

	pr := SignupRandomTestPatientWithPharmacyAndAddress(t, testData)
	patientID := pr.Patient.ID.Int64()
	if pid, err := testData.DataAPI.GetPersonIDByRole(api.RolePatient, patientID); err != nil {
		t.Fatalf("Failed to get person for role %s/%d: %s", api.RolePatient, patientID, err.Error())
	} else if pid <= 0 {
		t.Fatalf("Invalid patient ID %d", pid)
	}

	// Make sure a person row is inserted when creating a doctor

	dr, _, _ := SignupRandomTestDoctor(t, testData)
	doctorID := dr.DoctorID
	if pid, err := testData.DataAPI.GetPersonIDByRole(api.RoleDoctor, doctorID); err != nil {
		t.Fatalf("Failed to get person for role %s/%d: %s", api.RoleDoctor, doctorID, err.Error())
	} else if pid <= 0 {
		t.Fatalf("Invalid patient ID %d", pid)
	}
}

func TestCaseMessages(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	doctorID := GetDoctorIDOfCurrentDoctor(testData, t)
	doctor, err := testData.DataAPI.GetDoctorFromID(doctorID)
	test.OK(t, err)
	doctorPersonID, err := testData.DataAPI.GetPersonIDByRole(api.RoleDoctor, doctorID)
	test.OK(t, err)

	visit, treatmentPlan := CreateRandomPatientVisitAndPickTP(t, testData, doctor)
	patient, err := testData.DataAPI.GetPatientFromPatientVisitID(visit.PatientVisitID)
	test.OK(t, err)
	patientPersonID, err := testData.DataAPI.GetPersonIDByRole(api.RolePatient, patient.ID.Int64())
	test.OK(t, err)

	doctorCli := DoctorClient(testData, t, doctorID)
	patientCli := PatientClient(testData, t, patient.ID.Int64())

	test.OK(t, doctorCli.UpdateTreatmentPlanNote(treatmentPlan.ID.Int64(), "foo"))
	test.OK(t, doctorCli.SubmitTreatmentPlan(treatmentPlan.ID.Int64()))

	caseID, err := testData.DataAPI.GetPatientCaseIDFromPatientVisitID(visit.PatientVisitID)
	test.OK(t, err)

	photoID, _ := UploadPhoto(t, testData, doctor.AccountID.Int64())

	audioID, _ := uploadMedia(t, testData, doctor.AccountID.Int64())
	attachments := []*messages.Attachment{
		&messages.Attachment{
			Type: common.AttachmentTypePhoto,
			ID:   photoID,
		},
		&messages.Attachment{
			Type: common.AttachmentTypeAudio,
			ID:   audioID,
		},
	}

	_, err = doctorCli.PostCaseMessage(caseID, "foo", attachments)
	test.OK(t, err)

	msgs, err := testData.DataAPI.ListCaseMessages(caseID, api.RoleDoctor)
	if err != nil {
		t.Fatal(err)
	} else if len(msgs) != 2 { // one we just posted and one for the treatment plan submission
		t.Fatalf("Expected 2 message. Got %d", len(msgs))
	}

	m := msgs[len(msgs)-1]
	if len(m.Attachments) != 2 {
		t.Fatalf("Expected 2 attachment. Got %d", len(m.Attachments))
	}
	a := m.Attachments[0]
	if a.ItemType != common.AttachmentTypePhoto || a.ItemID != photoID {
		t.Fatalf("Wrong attachment type or ID")

	}
	photo, err := testData.DataAPI.GetMedia(photoID)
	test.OK(t, err)
	ok, err := testData.DataAPI.MediaHasClaim(photo.ID, common.ClaimerTypeConversationMessage, m.ID)
	test.OK(t, err)
	test.Equals(t, true, ok)

	b := m.Attachments[1]
	if b.ItemType != common.AttachmentTypeAudio || b.ItemID != audioID {
		t.Fatalf("Wrong attachment type or ID")
	}
	media, err := testData.DataAPI.GetMedia(audioID)
	if err != nil {
		t.Fatal(err)
	}
	ok, err = testData.DataAPI.MediaHasClaim(media.ID, common.ClaimerTypeConversationMessage, m.ID)
	test.OK(t, err)
	test.Equals(t, true, ok)

	if participants, err := testData.DataAPI.CaseMessageParticipants(caseID, false); err != nil {
		t.Fatal(err)
	} else if len(participants) != 1 {
		t.Fatalf("Expected 1 participant. Got %d", len(participants))
	} else if participants[doctorPersonID] == nil {
		t.Fatalf("Participant does not match")
	} else if participants[doctorPersonID].Unread {
		t.Fatalf("Expected conversation to be read")
	}

	// Reply from patient
	_, err = patientCli.PostCaseMessage(caseID, "bar", nil)
	test.OK(t, err)

	if msgs, err = testData.DataAPI.ListCaseMessages(caseID, api.RolePatient); err != nil {
		t.Fatal(err)
	} else if len(msgs) != 3 {
		t.Fatalf("Expected 3 messages. Got %d", len(msgs))
	}

	if participants, err := testData.DataAPI.CaseMessageParticipants(caseID, false); err != nil {
		t.Fatal(err)
	} else if len(participants) != 2 {
		t.Fatalf("Expected 2 participants. Got %d", len(participants))
	} else if participants[doctorPersonID] == nil {
		t.Fatalf("Participant does not exist")
	} else if !participants[doctorPersonID].Unread {
		t.Fatalf("Expected doctor's conversation to be unread")
	} else if participants[patientPersonID] == nil {
		t.Fatalf("Participant does not exist")
	} else if participants[patientPersonID].Unread {
		t.Fatalf("Expected patient's conversation to be read")
	}

	if err := testData.DataAPI.MarkCaseMessagesAsRead(caseID, doctorPersonID); err != nil {
		t.Fatal(err)
	}

	if participants, err := testData.DataAPI.CaseMessageParticipants(caseID, false); err != nil {
		t.Fatal(err)
	} else if participants[doctorPersonID].Unread {
		t.Fatalf("Expected doctor's conversation to be read")
	}
}
