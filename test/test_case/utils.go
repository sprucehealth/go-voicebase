package test_case

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/sprucehealth/backend/patient_case"
	"github.com/sprucehealth/backend/test/test_integration"
)

func DismissCaseNotification(notificationId, patientAccountId int64, testData *test_integration.TestData, t *testing.T) {
	handler := patient_case.NewDismissNotificationHandler(testData.DataApi)
	patientServer := httptest.NewServer(handler)
	defer patientServer.Close()

	requestData := map[string]interface{}{
		"notification_id": strconv.FormatInt(notificationId, 10),
	}

	jsonData, err := json.Marshal(&requestData)
	if err != nil {
		t.Fatal(err)
	}

	res, err := testData.AuthPut(patientServer.URL, "application/json", bytes.NewReader(jsonData), patientAccountId)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	} else if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d but got %d", http.StatusOK, res.StatusCode)
	}
}

type testTreatmentPlanData struct {
}

func (t *testTreatmentPlanData) TypeName() string {
	return patient_case.CNTreatmentPlan
}

type testMessageData struct {
}

func (t *testMessageData) TypeName() string {
	return patient_case.CNMessage
}

type testVisitSubmittedNotification struct {
}

func (t *testVisitSubmittedNotification) TypeName() string {
	return patient_case.CNVisitSubmitted
}

func getNotificationTypes() map[string]reflect.Type {
	testNotifyTypes := make(map[string]reflect.Type)
	testNotifyTypes[patient_case.CNMessage] = reflect.TypeOf(reflect.Indirect(reflect.ValueOf(&testMessageData{})).Interface())
	testNotifyTypes[patient_case.CNTreatmentPlan] = reflect.TypeOf(reflect.Indirect(reflect.ValueOf(&testTreatmentPlanData{})).Interface())
	testNotifyTypes[patient_case.CNVisitSubmitted] = reflect.TypeOf(reflect.Indirect(reflect.ValueOf(&testVisitSubmittedNotification{})).Interface())
	return testNotifyTypes
}
