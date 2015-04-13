package patient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
)

type mockDataAPI_UpdateHandler struct {
	api.DataAPI
	state string

	updateAttempted *api.PatientUpdate
}

func (m *mockDataAPI_UpdateHandler) GetPatientIDFromAccountID(id int64) (int64, error) {
	return 1, nil
}
func (m *mockDataAPI_UpdateHandler) UpdatePatient(id int64, update *api.PatientUpdate, updateFromDoctor bool) error {
	m.updateAttempted = update
	return nil
}
func (m *mockDataAPI_UpdateHandler) State(state string) (string, string, error) {
	return state, state, nil
}

type mockAddressValidator_UpdateHandler struct {
	cityState *address.CityState
}

func (m mockAddressValidator_UpdateHandler) ZipcodeLookup(zipcode string) (*address.CityState, error) {
	return m.cityState, nil
}

func TestPatientUpdate(t *testing.T) {
	testPatientUpdate("POST", t)
	testPatientUpdate("PUT", t)
}

func testPatientUpdate(method string, t *testing.T) {
	m := &mockDataAPI_UpdateHandler{}
	ma := &mockAddressValidator_UpdateHandler{
		cityState: &address.CityState{
			City:              "San Francisco",
			State:             "California",
			StateAbbreviation: "CA",
		},
	}
	h := NewUpdateHandler(m, ma)

	u := &UpdateRequest{
		PhoneNumbers: []PhoneNumber{
			{
				Type:   "Cell",
				Number: "2060000000",
			},
		},
		Address: &common.Address{
			AddressLine1: "line1",
			City:         "city",
			State:        "state",
			ZipCode:      "21493",
		},
	}

	jsonData, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err.Error())
	}

	r, err := http.NewRequest(method, "api.spruce.local", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 but got %d", w.Code)
	} else if len(m.updateAttempted.PhoneNumbers) != 1 {
		t.Fatalf("Expected 1 phone number but got %d", len(m.updateAttempted.PhoneNumbers))
	} else if m.updateAttempted.Address == nil {
		t.Fatalf("Expected address update but got none")
	}
}
