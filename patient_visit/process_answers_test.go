package patient_visit

import (
	"encoding/json"
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/info_intake"
	"github.com/sprucehealth/backend/patient"
	"github.com/sprucehealth/backend/tagging/model"
	"github.com/sprucehealth/backend/tagging/response"
	"github.com/sprucehealth/backend/test"
)

type mockDataAPI_processPatientAnswers struct {
	api.DataAPI
	layoutVersion *api.LayoutVersion
	answers       map[int64][]common.Answer
	maAssignment  *common.CareProviderAssignment
	patient       *common.Patient
	doctor        *common.Doctor
	cases         []*common.PatientCase
	templates     []*common.ScheduledMessageTemplate

	messageScheduled *common.ScheduledMessage
}

func (d *mockDataAPI_processPatientAnswers) GetPatientLayout(layoutVersionID, languageID int64) (*api.LayoutVersion, error) {
	return d.layoutVersion, nil
}
func (d *mockDataAPI_processPatientAnswers) AnswersForQuestions(questionIDs []int64, info api.IntakeInfo) (map[int64][]common.Answer, error) {
	return d.answers, nil
}
func (d *mockDataAPI_processPatientAnswers) GetActiveCareTeamMemberForCase(role string, patientCaseID int64) (*common.CareProviderAssignment, error) {
	return d.maAssignment, nil
}
func (d *mockDataAPI_processPatientAnswers) GetPatientFromID(id int64) (*common.Patient, error) {
	return d.patient, nil
}
func (d *mockDataAPI_processPatientAnswers) GetDoctorFromID(id int64) (*common.Doctor, error) {
	return d.doctor, nil
}
func (d *mockDataAPI_processPatientAnswers) GetCasesForPatient(patientID int64, states []string) ([]*common.PatientCase, error) {
	return d.cases, nil
}
func (d *mockDataAPI_processPatientAnswers) AddAlertsForVisit(visitID int64, alerts []*common.Alert) error {
	return nil
}
func (d *mockDataAPI_processPatientAnswers) ScheduledMessageTemplates(eventType string) ([]*common.ScheduledMessageTemplate, error) {
	return d.templates, nil
}
func (d *mockDataAPI_processPatientAnswers) CreateScheduledMessage(msg *common.ScheduledMessage) (int64, error) {
	d.messageScheduled = msg
	return 0, nil
}

type mockTaggingClient_processPatientAnswers struct {
	TagsCreated map[int64][]string
	TagsDeleted map[int64][]string
}

func (t *mockTaggingClient_processPatientAnswers) CaseAssociations(ms []*model.TagMembership, start, end int64) ([]*response.TagAssociation, error) {
	return nil, nil
}
func (t *mockTaggingClient_processPatientAnswers) CaseTagMemberships(caseID int64) (map[string]*model.TagMembership, error) {
	return nil, nil
}
func (t *mockTaggingClient_processPatientAnswers) DeleteTag(id int64) (int64, error) {
	return 0, nil
}
func (t *mockTaggingClient_processPatientAnswers) DeleteTagCaseAssociation(text string, caseID int64) error {
	if t.TagsDeleted == nil {
		t.TagsDeleted = make(map[int64][]string)
	}
	t.TagsDeleted[caseID] = append(t.TagsDeleted[caseID], text)
	return nil
}
func (t *mockTaggingClient_processPatientAnswers) DeleteTagCaseMembership(tagID, caseID int64) error {
	return nil
}
func (t *mockTaggingClient_processPatientAnswers) InsertTagAssociation(tag *model.Tag, membership *model.TagMembership) (int64, error) {
	if t.TagsCreated == nil {
		t.TagsCreated = make(map[int64][]string)
	}
	t.TagsCreated[*membership.CaseID] = append(t.TagsCreated[*membership.CaseID], tag.Text)
	return 0, nil
}
func (t *mockTaggingClient_processPatientAnswers) TagMembershipQuery(query string, pastTrigger bool) ([]*model.TagMembership, error) {
	return nil, nil
}
func (t *mockTaggingClient_processPatientAnswers) Tag(tagText string) (*response.Tag, error) {
	return nil, nil
}
func (t *mockTaggingClient_processPatientAnswers) Tags(tagText []string, common bool) ([]*response.Tag, error) {
	return nil, nil
}
func (t *mockTaggingClient_processPatientAnswers) InsertTagSavedSearch(ss *model.TagSavedSearch) (int64, error) {
	return 0, nil
}
func (t *mockTaggingClient_processPatientAnswers) DeleteTagSavedSearch(ssID int64) (int64, error) {
	return 0, nil
}
func (t *mockTaggingClient_processPatientAnswers) InsertTag(tag *model.Tag) (int64, error) {
	return 0, nil
}
func (t *mockTaggingClient_processPatientAnswers) TagSavedSearchs() ([]*model.TagSavedSearch, error) {
	return nil, nil
}
func (t *mockTaggingClient_processPatientAnswers) UpdateTag(tag *model.TagUpdate) error {
	return nil
}
func (t *mockTaggingClient_processPatientAnswers) UpdateTagCaseMembership(membership *model.TagMembershipUpdate) error {
	return nil
}

// TestProcessAnswers_InsuredScheduledMessage is to ensure that a message
// gets scheduled to be automatically sent to the patient when the patient answers that
// they are insured.
func TestProcessAnswers_InsuredScheduledMessage(t *testing.T) {
	testProcessAnswersForInsurance(t, insuredPatientEvent, "adgkag")
}

// TestProcessAnswers_UninsuredScheduledMessage is to ensure that a message
// gets scheduled to be automatically sent to the patient when the patient answers that
// they are not insured.
func TestProcessAnswers_UninsuredScheduledMessage(t *testing.T) {
	testProcessAnswersForInsurance(t, uninsuredPatientEvent, noInsuranceAnswerTags[0])
	testProcessAnswersForInsurance(t, uninsuredPatientEvent, noInsuranceAnswerTags[1])
}

func testProcessAnswersForInsurance(t *testing.T, event string, answerTag string) {

	layoutData := &info_intake.InfoIntakeLayout{
		Sections: []*info_intake.Section{
			{
				Screens: []*info_intake.Screen{
					{
						Questions: []*info_intake.Question{
							{
								QuestionID:  10,
								QuestionTag: insuranceCoverageQuestionTag,
								PotentialAnswers: []*info_intake.PotentialAnswer{
									{
										AnswerTag: answerTag,
										AnswerID:  5,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(layoutData)
	if err != nil {
		t.Fatalf(err.Error())
	}

	m := &mockDataAPI_processPatientAnswers{
		layoutVersion: &api.LayoutVersion{
			Layout: jsonData,
		},
		answers: map[int64][]common.Answer{
			10: []common.Answer{
				&common.AnswerIntake{
					PotentialAnswerID: encoding.NewObjectID(5),
				},
			},
		},
		maAssignment: &common.CareProviderAssignment{
			ProviderRole: api.RoleCC,
		},
		doctor:  &common.Doctor{},
		patient: &common.Patient{},
		templates: []*common.ScheduledMessageTemplate{
			{
				Message: "testing",
			},
		},
		cases: []*common.PatientCase{
			{
				ID: encoding.NewObjectID(1),
			},
		},
	}

	caseID := encoding.NewObjectID(1)
	ev := &patient.VisitSubmittedEvent{
		Visit: &common.PatientVisit{
			PatientCaseID: caseID,
		},
		PatientCaseID: caseID.Int64(),
	}

	taggingClient := &mockTaggingClient_processPatientAnswers{}
	processPatientAnswers(m, "api.spruce.local", ev, taggingClient)

	test.Assert(t, len(taggingClient.TagsCreated) == 1, "Expected only 1 tag to have been created")
	test.Assert(t, len(taggingClient.TagsDeleted) == 1, "Expected only 1 tag to have been deleted")
	if answerTag == noInsuranceAnswerTags[0] || answerTag == noInsuranceAnswerTags[1] {
		test.Equals(t, []string{"Uninsured"}, taggingClient.TagsCreated[caseID.Int64()])
		test.Equals(t, []string{"Insured"}, taggingClient.TagsDeleted[caseID.Int64()])
	} else {
		test.Equals(t, []string{"Insured"}, taggingClient.TagsCreated[caseID.Int64()])
		test.Equals(t, []string{"Uninsured"}, taggingClient.TagsDeleted[caseID.Int64()])
	}

	if m.messageScheduled == nil {
		t.Fatal("Expected message to be scheduled but it wasnt")
	} else if m.messageScheduled.Event != event {
		t.Fatalf("Expected scheduled message to be for event %s but it was for event %s", event, m.messageScheduled.Event)
	} else if m.messageScheduled.Status != common.SMScheduled {
		t.Fatalf("Expected scheduled message to have status %s but instaed it had status %s", common.SMScheduled, m.messageScheduled.Status.String())
	} else if m.messageScheduled.Scheduled.IsZero() {
		t.Fatalf("Expected message to be scheduled for some time in the future")
	}
}

// TestProcessAnswers_SecondCase is a test to ensure that no automated message
// gets scheduled for the patient if this is the patient's second case for which they are submitting a visit.
func TestProcessAnswers_SecondCase(t *testing.T) {

	layoutData := &info_intake.InfoIntakeLayout{
		Sections: []*info_intake.Section{
			{
				Screens: []*info_intake.Screen{
					{
						Questions: []*info_intake.Question{
							{
								QuestionID:  10,
								QuestionTag: insuranceCoverageQuestionTag,
								PotentialAnswers: []*info_intake.PotentialAnswer{
									{
										AnswerTag: noInsuranceAnswerTags[0],
										AnswerID:  5,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(layoutData)
	if err != nil {
		t.Fatalf(err.Error())
	}

	m := &mockDataAPI_processPatientAnswers{
		layoutVersion: &api.LayoutVersion{
			Layout: jsonData,
		},
		answers: map[int64][]common.Answer{
			10: []common.Answer{
				&common.AnswerIntake{
					PotentialAnswerID: encoding.NewObjectID(5),
				},
			},
		},
		maAssignment: &common.CareProviderAssignment{
			ProviderRole: api.RoleCC,
		},
		doctor:  &common.Doctor{},
		patient: &common.Patient{},
		templates: []*common.ScheduledMessageTemplate{
			{
				Message: "testing",
			},
		},
		cases: []*common.PatientCase{
			{
				ID: encoding.NewObjectID(2),
			},
			{
				ID: encoding.NewObjectID(1),
			},
		},
	}

	caseID := encoding.NewObjectID(1)
	ev := &patient.VisitSubmittedEvent{
		Visit: &common.PatientVisit{
			PatientCaseID: caseID,
		},
		PatientCaseID: caseID.Int64(),
	}

	taggingClient := &mockTaggingClient_processPatientAnswers{}
	processPatientAnswers(m, "api.spruce.local", ev, taggingClient)

	test.Assert(t, len(taggingClient.TagsCreated) == 1, "Expected only 1 tag to have been created")
	test.Assert(t, len(taggingClient.TagsDeleted) == 1, "Expected only 1 tag to have been deleted")
	test.Equals(t, []string{"Uninsured"}, taggingClient.TagsCreated[caseID.Int64()])
	test.Equals(t, []string{"Insured"}, taggingClient.TagsDeleted[caseID.Int64()])

	if m.messageScheduled != nil {
		t.Fatal("Expected no message to get scheduled for a subsequent visit")
	}
}

// TestProcessAnswers_FollowupVisit is a test to ensure that no automated message
// gets scheduled for the patient if this is the patient's followup visit in their first case.
func TestProcessAnswers_FollowupVisit(t *testing.T) {
	layoutData := &info_intake.InfoIntakeLayout{
		Sections: []*info_intake.Section{
			{
				Screens: []*info_intake.Screen{
					{
						Questions: []*info_intake.Question{
							{
								QuestionID:  10,
								QuestionTag: insuranceCoverageQuestionTag,
								PotentialAnswers: []*info_intake.PotentialAnswer{
									{
										AnswerTag: noInsuranceAnswerTags[0],
										AnswerID:  5,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(layoutData)
	if err != nil {
		t.Fatalf(err.Error())
	}

	m := &mockDataAPI_processPatientAnswers{
		layoutVersion: &api.LayoutVersion{
			Layout: jsonData,
		},
		answers: map[int64][]common.Answer{
			10: []common.Answer{
				&common.AnswerIntake{
					PotentialAnswerID: encoding.NewObjectID(5),
				},
			},
		},
		maAssignment: &common.CareProviderAssignment{
			ProviderRole: api.RoleCC,
		},
		doctor:  &common.Doctor{},
		patient: &common.Patient{},
		templates: []*common.ScheduledMessageTemplate{
			{
				Message: "testing",
			},
		},
		cases: []*common.PatientCase{
			{
				ID: encoding.NewObjectID(2),
			},
		},
	}

	caseID := encoding.NewObjectID(1)
	ev := &patient.VisitSubmittedEvent{
		Visit: &common.PatientVisit{
			PatientCaseID: caseID,
		},
		PatientCaseID: caseID.Int64(),
	}

	taggingClient := &mockTaggingClient_processPatientAnswers{}
	processPatientAnswers(m, "api.spruce.local", ev, taggingClient)

	test.Assert(t, len(taggingClient.TagsCreated) == 1, "Expected only 1 tag to have been created")
	test.Assert(t, len(taggingClient.TagsDeleted) == 1, "Expected only 1 tag to have been deleted")
	test.Equals(t, []string{"Uninsured"}, taggingClient.TagsCreated[caseID.Int64()])
	test.Equals(t, []string{"Insured"}, taggingClient.TagsDeleted[caseID.Int64()])

	if m.messageScheduled != nil {
		t.Fatal("Expected no message to get scheduled for a subsequent visit")
	}
}
