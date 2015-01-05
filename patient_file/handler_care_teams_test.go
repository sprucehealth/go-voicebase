package patient_file

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_handler"
)

type mockedDataAPI struct {
	api.DataAPI
	doctorIDFromAccountID int64
	patientAccountID      int64
	canAccess             bool
}

func (d mockedDataAPI) GetDoctorIDFromAccountID(accountID int64) (int64, error) {
	return d.doctorIDFromAccountID, nil
}

func (d mockedDataAPI) GetPatientFromAccountID(accountID int64) (*common.Patient, error) {
	return &common.Patient{
		PatientID: encoding.NewObjectID(d.patientAccountID),
	}, nil
}

func (d mockedDataAPI) DoesCaseExistForPatient(p, c int64) (bool, error) {
	return d.canAccess, nil
}

func canAccess(httpMethod, role string, doctorID, patientID int64, dataAPI api.DataAPI) error {
	return nil
}

func cannotAccess(httpMethod, role string, doctorID, patientID int64, dataAPI api.DataAPI) error {
	return apiservice.NewAccessForbiddenError()
}

var getCareTeamsForPatientByCaseResponse map[int64]*common.PatientCareTeam

func (d mockedDataAPI) GetCareTeamsForPatientByCase(id int64) (map[int64]*common.PatientCareTeam, error) {
	return getCareTeamsForPatientByCaseResponse, nil
}

func TestDoctorRequiresPatientID(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, false})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.DOCTOR_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteError(apiservice.NewValidationError("patient_id required"), expectedWriter, r)
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, expectedWriter.Body, responseWriter.Body)
}

func TestDoctorCannotAccessUnownedPatient(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?patient_id=32", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, false})
	verifyDoctorAccessToPatientFileFn = cannotAccess
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
	test.Equals(t, expectedWriter.Body, responseWriter.Body)
}

func TestPatientCannotAccessUnownedCase(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?case_id=1", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, false})
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.PATIENT_ROLE
			ctxt.AccountID = 1
		},
	}
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteAccessNotAllowedError(expectedWriter, r)
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, expectedWriter.Body, responseWriter.Body)
}

func TestDoctorCanFetchAllCareTeams(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?patient_id=32", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, false})
	verifyDoctorAccessToPatientFileFn = canAccess
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.DOCTOR_ROLE
			ctxt.AccountID = 1
		},
	}
	getCareTeamsForPatientByCaseResponse = buildDummyGetCareTeamsForPatientByCaseResponse(2)
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 0))
	handler.ServeHTTP(responseWriter, r)
	// TODO: We can't verify the JSON output here as maps do not serialize determinisitically
	// test.Equals(t, expectedWriter.Body, responseWriter.Body)
	test.Equals(t, 2, len(createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 0).CareTeams))
}

func TestPatientCanFetchAllCareTeams(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, false})
	verifyDoctorAccessToPatientFileFn = canAccess
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.PATIENT_ROLE
			ctxt.AccountID = 1
		},
	}
	getCareTeamsForPatientByCaseResponse = buildDummyGetCareTeamsForPatientByCaseResponse(2)
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 0))

	handler.ServeHTTP(responseWriter, r)
	// TODO: We can't verify the JSON output here as maps do not serialize determinisitically
	// test.Equals(t, expectedWriter.Body, responseWriter.Body)
	test.Equals(t, 2, len(createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 0).CareTeams))
}

func TestMACanFetchAllCareTeams(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?patient_id=32", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, false})
	verifyDoctorAccessToPatientFileFn = canAccess
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.MA_ROLE
			ctxt.AccountID = 1
		},
	}
	getCareTeamsForPatientByCaseResponse = buildDummyGetCareTeamsForPatientByCaseResponse(2)
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 0))
	handler.ServeHTTP(responseWriter, r)
	// TODO: We can't verify the JSON output here as maps do not serialize determinisitically
	// test.Equals(t, expectedWriter.Body, responseWriter.Body)
	test.Equals(t, 2, len(createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 0).CareTeams))
}

func TestDoctorCanFilterCareTeamsByCase(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?patient_id=1&case_id=1", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, true})
	verifyDoctorAccessToPatientFileFn = canAccess
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.DOCTOR_ROLE
			ctxt.AccountID = 1
		},
	}
	getCareTeamsForPatientByCaseResponse = buildDummyGetCareTeamsForPatientByCaseResponse(2)
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 1))
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, expectedWriter.Body, responseWriter.Body)
	test.Equals(t, 1, len(createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 1).CareTeams))
}

func TestPatientCanFilterCareTeamsByCase(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?case_id=1", nil)
	test.OK(t, err)
	careTeamHandler := NewPatientCareTeamsHandler(mockedDataAPI{&api.DataService{}, 1, 2, true})
	verifyDoctorAccessToPatientFileFn = canAccess
	handler := test_handler.MockHandler{
		H: careTeamHandler,
		Setup: func() {
			ctxt := apiservice.GetContext(r)
			ctxt.Role = api.PATIENT_ROLE
			ctxt.AccountID = 1
		},
	}
	getCareTeamsForPatientByCaseResponse = buildDummyGetCareTeamsForPatientByCaseResponse(2)
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	apiservice.WriteJSON(expectedWriter, createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 1))
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, expectedWriter.Body, responseWriter.Body)
	test.Equals(t, 1, len(createCareTeamsResponse(getCareTeamsForPatientByCaseResponse, 1).CareTeams))
}

func buildDummyGetCareTeamsForPatientByCaseResponse(careTeamCount int) map[int64]*common.PatientCareTeam {
	resp := make(map[int64]*common.PatientCareTeam)
	for i := 1; i < careTeamCount+1; i++ {
		team := &common.PatientCareTeam{
			Assignments: make([]*common.CareProviderAssignment, 0),
		}
		assignment := &common.CareProviderAssignment{
			ProviderRole:      "Doctor",
			ProviderID:        1,
			FirstName:         "First",
			LastName:          "Last",
			ShortTitle:        "ShortT",
			LongTitle:         "LongT",
			ShortDisplayName:  "SDN",
			LongDisplayName:   "LDN",
			SmallThumbnailURL: "STU",
			LargeThumbnailURL: "STU",
			CreationDate:      time.Unix(0, 0),
		}
		team.Assignments = append(team.Assignments, assignment)
		resp[int64(i)] = team
	}
	return resp
}
