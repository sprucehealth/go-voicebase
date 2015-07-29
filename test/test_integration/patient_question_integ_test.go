package test_integration

import (
	"testing"

	"github.com/sprucehealth/backend/info_intake"
	"github.com/sprucehealth/backend/test"
)

func TestNoPotentialAnswerForQuestionTypes(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close(t)
	testData.StartAPIServer(t)

	// no free text question type should have potential answers associated with it
	rows, err := testData.DB.Query(`SELECT question.id FROM question WHERE question_type IN ('q_type_free_text', 'q_type_autocomplete')`)
	if err != nil {
		t.Fatal("Unable to query database for a list of question ids : " + err.Error())
	}
	defer rows.Close()

	var questionIDs []int64
	for rows.Next() {
		var id int64
		test.OK(t, rows.Scan(&id))
		questionIDs = append(questionIDs, id)
	}
	test.OK(t, rows.Err())

	// for each of these question ids, there should be no potential responses
	for _, questionID := range questionIDs {
		answerInfos, err := testData.DataAPI.GetAnswerInfo(questionID, 1)
		if err != nil {
			t.Fatal("Error when trying to get answer for question (which should return no answers) : " + err.Error())
		}
		if !(answerInfos == nil || len(answerInfos) == 0) {
			t.Fatal("No potential answers should be returned for these questions")
		}
	}
}

// This test is to ensure that additional fields are set for the
// autocomplete question type, as they should be for the client to
// be able to show additional pieces of content in the question
func TestAdditionalFieldsInAutocompleteQuestion(t *testing.T) {
	testData := SetupTest(t)
	defer testData.Close(t)
	testData.StartAPIServer(t)

	// signup a random test patient for which to answer questions
	patientSignedUpResponse := SignupRandomTestPatientWithPharmacyAndAddress(t, testData)
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedUpResponse.Patient.ID.Int64(), testData, t)

	// lets go through the questions to find the one for which the patient answer should be present
	for _, section := range patientVisitResponse.ClientLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				if question.QuestionType == info_intake.QuestionTypeAutocomplete {
					if question.AdditionalFields == nil || len(question.AdditionalFields) == 0 {
						t.Fatal("Expected additional fields to be set for the autocomplete question type")
					}
				}
			}
		}
	}
}
