package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/ratelimit"
	"golang.org/x/net/context"
)

type stubEmailChecker struct {
	ExistingEmails map[string]bool
}

func (ec *stubEmailChecker) AccountForEmail(email string) (*common.Account, error) {
	if ec.ExistingEmails[email] {
		return &common.Account{Email: email}, nil
	}
	return nil, api.ErrLoginDoesNotExist
}

func TestCheckEmailHandler(t *testing.T) {
	ec := &stubEmailChecker{
		ExistingEmails: map[string]bool{
			"used@somewhere.com": true,
		},
	}

	h := NewCheckEmailHandler(ec, ratelimit.NullKeyed{}, metrics.NewRegistry())

	// Test missing email argument
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(context.Background(), rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected code %d got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "{\"available\":false}\n" {
		t.Errorf("Expected unavailable response, got '%s'", rec.Body.String())
	}

	// Test available email
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/?email=unused@somewhere.com", nil)
	h.ServeHTTP(context.Background(), rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected code %d got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "{\"available\":true}\n" {
		t.Errorf("Expected available response, got '%s'", rec.Body.String())
	}

	// Test unavailable email
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/?email=used@somewhere.com", nil)
	h.ServeHTTP(context.Background(), rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected code %d got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "{\"available\":false}\n" {
		t.Errorf("Expected unavailable response, got '%s'", rec.Body.String())
	}
}
