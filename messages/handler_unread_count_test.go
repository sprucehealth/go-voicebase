package messages

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/test"
)

type unreadCountDataAPI struct {
	api.DataAPI
	pcase *common.PatientCase
}

func (d *unreadCountDataAPI) GetPatientCaseFromID(id int64) (*common.PatientCase, error) {
	return d.pcase, nil
}

func (d *unreadCountDataAPI) GetPatientIDFromAccountID(accountID int64) (int64, error) {
	return 2, nil
}

func (d *unreadCountDataAPI) GetPersonIDByRole(roleType string, roleID int64) (int64, error) {
	return 3, nil
}

func (d *unreadCountDataAPI) UnreadMessageCount(caseID, personID int64) (int, error) {
	if caseID != 1 {
		return 0, errors.New("case ID does not match")
	}
	if personID != 3 {
		return 0, errors.New("person ID does not match")
	}
	return 4, nil
}

func TestUnreadCountHandler(t *testing.T) {
	dataAPI := &unreadCountDataAPI{
		pcase: &common.PatientCase{
			ID:        encoding.NewObjectID(1),
			PatientID: encoding.NewObjectID(2),
		},
	}
	hand := NewUnreadCountHandler(dataAPI)

	r, err := http.NewRequest("GET", "/?case_id=1", nil)
	test.OK(t, err)
	ctx := apiservice.GetContext(r)
	ctx.Role = api.RolePatient
	w := httptest.NewRecorder()
	hand.ServeHTTP(w, r)
	test.Equals(t, w.Code, http.StatusOK)
	test.Equals(t, "{\"unread_count\":4}\n", w.Body.String())
}