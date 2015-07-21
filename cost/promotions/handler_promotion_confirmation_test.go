package promotions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_handler"
)

type mockDataAPIPromotionConfirmationHandler struct {
	api.DataAPI
	lookupPromoCodeParam         string
	lookupPromoCodeErr           error
	lookupPromoCode              *common.PromoCode
	referralProgramParam         int64
	referralProgramErr           error
	referralProgram              *common.ReferralProgram
	getPatientFromAccountIDParam int64
	getPatientFromAccountIDErr   error
	getPatientFromAccountID      *common.Patient
	getDoctorFromAccountIDParam  int64
	getDoctorFromAccountIDErr    error
	getDoctorFromAccountID       *common.Doctor
	promotionParam               int64
	promotionErr                 error
	promotion                    *common.Promotion
}

func (m *mockDataAPIPromotionConfirmationHandler) LookupPromoCode(code string) (*common.PromoCode, error) {
	m.lookupPromoCodeParam = code
	return m.lookupPromoCode, m.lookupPromoCodeErr
}

func (m *mockDataAPIPromotionConfirmationHandler) ReferralProgram(codeID int64, types map[string]reflect.Type) (*common.ReferralProgram, error) {
	m.referralProgramParam = codeID
	return m.referralProgram, m.referralProgramErr
}

func (m *mockDataAPIPromotionConfirmationHandler) GetPatientFromAccountID(accountID int64) (patient *common.Patient, err error) {
	m.getPatientFromAccountIDParam = accountID
	return m.getPatientFromAccountID, m.getPatientFromAccountIDErr
}

func (m *mockDataAPIPromotionConfirmationHandler) GetDoctorFromAccountID(accountID int64) (patient *common.Doctor, err error) {
	m.getDoctorFromAccountIDParam = accountID
	return m.getDoctorFromAccountID, m.getDoctorFromAccountIDErr
}

func (m *mockDataAPIPromotionConfirmationHandler) Promotion(codeID int64, types map[string]reflect.Type) (*common.Promotion, error) {
	m.promotionParam = codeID
	return m.promotion, m.promotionErr
}

func TestPromotionConfirmationHandlerGETRequiresParams(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{DataAPI: &api.DataService{}}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, http.StatusBadRequest, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETNoPromotion(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:            &api.DataService{},
		lookupPromoCodeErr: api.ErrNotFound(`promotion_code`),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, "foo", dataAPI.lookupPromoCodeParam)
	test.Equals(t, http.StatusNotFound, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETCodeLookupErr(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:            &api.DataService{},
		lookupPromoCodeErr: errors.New("Foo"),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, http.StatusInternalServerError, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETPromotionLookupErr(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:         &api.DataService{},
		lookupPromoCode: &common.PromoCode{ID: 1, Code: "foo", IsReferral: false},
		promotionErr:    errors.New("Foo"),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, int64(1), dataAPI.promotionParam)
	test.Equals(t, http.StatusInternalServerError, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETReferralLookupErr(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:            &api.DataService{},
		lookupPromoCode:    &common.PromoCode{ID: 1, Code: "foo", IsReferral: true},
		referralProgramErr: errors.New("Foo"),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, int64(1), dataAPI.referralProgramParam)
	test.Equals(t, http.StatusInternalServerError, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETReferralGetPatientFromAccountIDErr(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:                    &api.DataService{},
		lookupPromoCode:            &common.PromoCode{ID: 1, Code: "foo", IsReferral: true},
		referralProgram:            createReferralProgram(2, "imageURL"),
		getPatientFromAccountIDErr: errors.New("Foo"),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, int64(2), dataAPI.getPatientFromAccountIDParam)
	test.Equals(t, http.StatusInternalServerError, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETReferralPatientNotFoundGetDoctorFromAccountIDErr(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:                    &api.DataService{},
		lookupPromoCode:            &common.PromoCode{ID: 1, Code: "foo", IsReferral: true},
		referralProgram:            createReferralProgram(2, "imageURL"),
		getPatientFromAccountIDErr: api.ErrNotFound(`patient`),
		getDoctorFromAccountIDErr:  errors.New("Foo"),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	responseWriter := httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	test.Equals(t, int64(2), dataAPI.getDoctorFromAccountIDParam)
	test.Equals(t, http.StatusInternalServerError, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETReferralImageProvided(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:                 &api.DataService{},
		lookupPromoCode:         &common.PromoCode{ID: 1, Code: "foo", IsReferral: true},
		referralProgram:         createReferralProgram(2, "imageURL"),
		getPatientFromAccountID: &common.Patient{FirstName: "FirstName"},
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	httputil.JSONResponse(expectedWriter, http.StatusOK, &PromotionConfirmationGETResponse{
		Title:       "Your friend FirstName has given you a free visit.",
		ImageURL:    "imageURL",
		BodyText:    "successMsg",
		ButtonTitle: "Let's Go",
	})
	test.Equals(t, expectedWriter.Body.String(), responseWriter.Body.String())
	test.Equals(t, http.StatusOK, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETReferralDoctorImageNotProvided(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:                    &api.DataService{},
		lookupPromoCode:            &common.PromoCode{ID: 1, Code: "foo", IsReferral: true},
		referralProgram:            createReferralProgram(2, ""),
		getPatientFromAccountIDErr: api.ErrNotFound(`patient`),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	httputil.JSONResponse(expectedWriter, http.StatusOK, &PromotionConfirmationGETResponse{
		Title:       "Welcome to Spruce",
		ImageURL:    DefaultPromotionImageURL,
		BodyText:    "successMsg",
		ButtonTitle: "Let's Go",
	})
	test.Equals(t, expectedWriter.Body.String(), responseWriter.Body.String())
	test.Equals(t, http.StatusOK, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETPromotionImage(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:         &api.DataService{},
		lookupPromoCode: &common.PromoCode{ID: 1, Code: "foo", IsReferral: false},
		promotion:       createPromotion("imageURL", "", nil, 0),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	httputil.JSONResponse(expectedWriter, http.StatusOK, &PromotionConfirmationGETResponse{
		Title:       "displayMsg",
		ImageURL:    "imageURL",
		BodyText:    "successMsg",
		ButtonTitle: "Let's Go",
	})
	test.Equals(t, expectedWriter.Body.String(), responseWriter.Body.String())
	test.Equals(t, http.StatusOK, responseWriter.Code)
}

func TestPromotionConfirmationHandlerGETPromotionNoImage(t *testing.T) {
	r, err := http.NewRequest("GET", "mock.api.request?code=foo", nil)
	test.OK(t, err)
	dataAPI := &mockDataAPIPromotionConfirmationHandler{
		DataAPI:         &api.DataService{},
		lookupPromoCode: &common.PromoCode{ID: 1, Code: "foo", IsReferral: false},
		promotion:       createPromotion("", "", nil, 0),
	}
	promoConfHandler := NewPromotionConfirmationHandler(dataAPI)
	handler := test_handler.MockHandler{
		H: promoConfHandler,
	}

	expectedWriter, responseWriter := httptest.NewRecorder(), httptest.NewRecorder()
	handler.ServeHTTP(responseWriter, r)
	httputil.JSONResponse(expectedWriter, http.StatusOK, &PromotionConfirmationGETResponse{
		Title:       "displayMsg",
		ImageURL:    DefaultPromotionImageURL,
		BodyText:    "successMsg",
		ButtonTitle: "Let's Go",
	})
	test.Equals(t, expectedWriter.Body.String(), responseWriter.Body.String())
	test.Equals(t, http.StatusOK, responseWriter.Code)
}

func createReferralProgram(accountID int64, imageURL string) *common.ReferralProgram {
	rp, _ := NewGiveReferralProgram("title", "description", "group", nil,
		NewPercentOffVisitPromotion(0,
			"group", "displayMsg", "shortMsg", "successMsg", imageURL,
			1, 1, true), nil)
	return &common.ReferralProgram{
		AccountID: accountID,
		Data:      rp,
	}
}

func createPromotion(imageURL, group string, expires *time.Time, value int) *common.Promotion {
	p := NewPercentOffVisitPromotion(value,
		"group", "displayMsg", "shortMsg", "successMsg", imageURL,
		1, 1, true)
	return &common.Promotion{
		Data:    p,
		Expires: expires,
		Group:   group,
	}
}
