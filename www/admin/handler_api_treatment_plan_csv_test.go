package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/erx"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

type mockedDataAPI_handlerTreatmentPlanCSV struct {
	api.DataAPI
}

type mockedERXAPI_handlerTreatmentPlanCSV struct {
	erx.ERxAPI
}

func TestHandlerTreatmentPlanCSVRequiresParams(t *testing.T) {
	r, err := http.NewRequest("PUT", "mock.api.request", strings.NewReader("Foo"))
	r.Header.Set("Content-Type", "multipart/form-data;boundary=---------------------------")
	test.OK(t, err)
	handler := newTreatmentPlanCSVHandler(mockedDataAPI_handlerTreatmentPlanCSV{}, mockedERXAPI_handlerTreatmentPlanCSV{})
	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	www.APIBadRequestError(expectedWriter, r, fmt.Errorf("multipart: NextPart: EOF").Error())
	handler.ServeHTTP(context.Background(), responseWriter, r)
	test.Equals(t, string(expectedWriter.Body.Bytes()), string(responseWriter.Body.Bytes()))
}

type CloseableStringReader struct {
	*strings.Reader
}

func (m CloseableStringReader) Close() error {
	return nil
}

func TestHandlerFileDataRead(t *testing.T) {
	file := CloseableStringReader{Reader: strings.NewReader(`framework_name,Anti-aging,Anti-aging,Skin discoloration`)}
	data, err := csvDataFromFile(file)
	test.OK(t, err)
	test.Equals(t, [][]string{[]string{`framework_name`, `Anti-aging`, `Anti-aging`, `Skin discoloration`}}, data)
}
