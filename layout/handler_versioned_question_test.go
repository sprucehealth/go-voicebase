package layout

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/responses"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_handler"
)

type mockedDataAPI_versionedQuestionHandler struct {
	api.DataAPI
	vq    *common.VersionedQuestion
	vqTag *common.VersionedQuestion
}

func (m mockedDataAPI_versionedQuestionHandler) VersionedQuestionFromID(ID int64) (*common.VersionedQuestion, error) {
	return m.vq, nil
}

func (m mockedDataAPI_versionedQuestionHandler) VersionedQuestion(questionTag string, languageID, version int64) (*common.VersionedQuestion, error) {
	return m.vqTag, nil
}

func TestQuestionHandlerDoctorCannotQuery(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request", nil)
	test.OK(t, err)
	careTeamHandler := NewVersionedQuestionHandler(mockedDataAPI_versionedQuestionHandler{DataAPI: &api.DataService{}})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.DOCTOR_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteError(apiservice.NewAccessForbiddenError(), expectedWriter, r)
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

func TestQuestionHandlerPatientCannotQuery(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request", nil)
	test.OK(t, err)
	careTeamHandler := NewVersionedQuestionHandler(mockedDataAPI_versionedQuestionHandler{DataAPI: &api.DataService{}})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.PATIENT_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteError(apiservice.NewAccessForbiddenError(), expectedWriter, r)
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

func TestQuestionHandlerRequiresParams(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request", nil)
	test.OK(t, err)
	careTeamHandler := NewVersionedQuestionHandler(mockedDataAPI_versionedQuestionHandler{DataAPI: &api.DataService{}})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.ADMIN_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteError(apiservice.NewValidationError("insufficent parameters supplied to form complete query"), expectedWriter, r)
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

func TestQuestionHandlerRequiresCompleteTagQuery(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?tag=my_tag&version=1", nil)
	test.OK(t, err)
	careTeamHandler := NewVersionedQuestionHandler(mockedDataAPI_versionedQuestionHandler{DataAPI: &api.DataService{}})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.ADMIN_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteError(apiservice.NewValidationError("insufficent parameters supplied to form complete query"), expectedWriter, r)
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

func TestQuestionHandlerCanQueryByID(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?id=1", nil)
	test.OK(t, err)
	dbmodel := buildDummyVersionedQuestion("dummy")
	careTeamHandler := NewVersionedQuestionHandler(mockedDataAPI_versionedQuestionHandler{DataAPI: &api.DataService{}, vq: dbmodel})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.ADMIN_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, &versionedQuestionGETResponse{VersionedQuestion: responses.NewVersionedQuestionFromDBModel(dbmodel)})
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

func TestQuestionHandlerCanQueryByTagSet(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?tag=my_tag&version=1&language_id=1", nil)
	test.OK(t, err)
	dbmodel := buildDummyVersionedQuestion("dummy2")
	careTeamHandler := NewVersionedQuestionHandler(mockedDataAPI_versionedQuestionHandler{DataAPI: &api.DataService{}, vqTag: dbmodel})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.ADMIN_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, &versionedQuestionGETResponse{VersionedQuestion: responses.NewVersionedQuestionFromDBModel(dbmodel)})
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

func buildDummyVersionedQuestion(questionText string) *common.VersionedQuestion {
	return &common.VersionedQuestion{
		ID:             1,
		QuestionTypeID: 1,
		QuestionTag:    questionText,
		ParentQuestionID: sql.NullInt64{
			Int64: 1,
			Valid: true,
		},
		Required: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		FormattedFieldTags: sql.NullString{
			String: questionText,
			Valid:  true,
		},
		ToAlert: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		TextHasTokens: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		LanguageID: 1,
		Version:    1,
		QuestionText: sql.NullString{
			String: questionText,
			Valid:  true,
		},
		SubtextText: sql.NullString{
			String: questionText,
			Valid:  true,
		},
		SummaryText: sql.NullString{
			String: questionText,
			Valid:  true,
		},
		AlertText: sql.NullString{
			String: questionText,
			Valid:  true,
		},
		QuestionType: questionText,
	}
}
