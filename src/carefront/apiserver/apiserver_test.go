package main

import (
	"carefront/api"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// FakeResponseWriter for testing purposes
type FakeResponseWriter struct {
	Headers http.Header
	body    []byte
}

// Implementing the ResponseWriter interface
func (f *FakeResponseWriter) Header() http.Header {
	return f.Headers
}

func (f *FakeResponseWriter) Write(response_body []byte) (int, error) {
	// writing status ok since if its gotten this far, it means that its going to
	// be a succesful writing of a response
	f.WriteHeader(http.StatusOK)
	f.body = response_body
	return 0, nil
}

func (f *FakeResponseWriter) WriteHeader(statusCode int) {
	f.Headers.Add("Status", strconv.Itoa(statusCode))
}

func createAndReturnFakeAuthApi() *api.MockAuth {
	return &api.MockAuth{
		Accounts: map[string]api.MockAccount{
			"kajham": api.MockAccount{
				Id:       1,
				Login:    "kajham",
				Password: "12345",
			},
		},
		Tokens: map[string]int64{
			"tokenForKajham": 1,
		},
	}
}

func setupPingHandlerInMux() *AuthServeMux {
	fakeAuthApi := createAndReturnFakeAuthApi()
	pingHandler := PingHandler(0)
	mux := &AuthServeMux{*http.NewServeMux(), fakeAuthApi}
	mux.Handle("/v1/ping", pingHandler)

	return mux
}

func setupAuthHandlerInMux() *AuthServeMux {
	fakeAuthApi := createAndReturnFakeAuthApi()
	authHandler := &AuthenticationHandler{fakeAuthApi}
	mux := &AuthServeMux{*http.NewServeMux(), fakeAuthApi}
	mux.Handle("/v1/authenticate", authHandler)

	return mux

}
func testPing(successfulPing bool, t *testing.T) *FakeResponseWriter {
	// SETUP
	/* 	with a dummy test account that is simulating to be authenticated
	already, which is why there is a corresponding token for it in the map
	*/
	mux := setupPingHandlerInMux()

	// TEST
	req, _ := http.NewRequest("GET", "http://localhost:8080/v1/ping", nil)

	if successfulPing {
		req.Header.Add("Authorization", "token tokenForKajham")
	}

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)
	return responseWriter
}

func TestSuccessfulPing(t *testing.T) {
	responseWriter := testPing(true, t)

	statusCode := responseWriter.Headers.Get("Status")
	responseBody := string(responseWriter.body)
	if (responseBody != Pong) ||
		(statusCode != strconv.Itoa(http.StatusOK)) {
		t.Errorf("Expected %q with status code %q, but got %q with status code %q", Pong, http.StatusOK, responseBody, statusCode)
	}
}

func TestUnauthorizedPing(t *testing.T) {
	responseWriter := testPing(false, t)
	statusCode := responseWriter.Headers.Get("Status")

	if statusCode != strconv.Itoa(http.StatusForbidden) {
		t.Errorf("Expected status code %q, but got status code %q", http.StatusForbidden, statusCode)
	}
}

func TestIncorrectTokenPing(t *testing.T) {
	// SETUP
	mux := setupPingHandlerInMux()

	// TEST
	req, _ := http.NewRequest("GET", "http://localhost:8080/v1/ping", nil)
	req.Header.Add("Authorization", "token incorrectToken")

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)

	statusCode := responseWriter.Headers.Get("Status")
	if statusCode != strconv.Itoa(http.StatusForbidden) {
		t.Errorf("Expected status code %q, but got status code %q", http.StatusForbidden, statusCode)
	}
}

func TestMalformedAuthorizationHeader(t *testing.T) {
	// SETUP
	mux := setupPingHandlerInMux()
	// TEST
	req, _ := http.NewRequest("GET", "http://localhost:8080/v1/ping", nil)
	req.Header.Add("Authorization", "incorrectToken")

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)

	statusCode := responseWriter.Headers.Get("Status")
	if statusCode != strconv.Itoa(http.StatusForbidden) {
		t.Errorf("Expected status code %q, but got status code %q", http.StatusForbidden, statusCode)
	}
}

func TestSuccessfulLogin(t *testing.T) {
	mux := setupAuthHandlerInMux()

	req, _ := http.NewRequest("POST", "http://localhost:8080/v1/authenticate", strings.NewReader("login=kajham&password=12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)

	statusCode := responseWriter.Headers.Get("Status")
	if statusCode != strconv.Itoa(http.StatusOK) {
		t.Errorf("Expected status code %q, but got %q", http.StatusOK, statusCode)
	}
}

func TestUnsuccessfulLoginDueToPassword(t *testing.T) {
	mux := setupAuthHandlerInMux()

	req, _ := http.NewRequest("POST", "http://localhost:8080/v1/authenticate", strings.NewReader("login=kajham&password=ShouldFail"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)

	statusCode := responseWriter.Headers.Get("Status")
	if statusCode != strconv.Itoa(http.StatusForbidden) {
		t.Errorf("Expected status code %q, but got %q", http.StatusForbidden, statusCode)
	}
}

func TestUnsuccessfulLoginDueToUsername(t *testing.T) {
	mux := setupAuthHandlerInMux()

	req, _ := http.NewRequest("POST", "http://localhost:8080/v1/authenticate", strings.NewReader("login=kajaja&password=12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)

	statusCode := responseWriter.Headers.Get("Status")
	if statusCode != strconv.Itoa(http.StatusForbidden) {
		t.Errorf("Expected status code %q, but got %q", http.StatusForbidden, statusCode)
	}
}

func TestUnsuccessfulLoginDueToMissingParams(t *testing.T) {
	mux := setupAuthHandlerInMux()

	req, _ := http.NewRequest("POST", "http://localhost:8080/v1/authenticate", nil)

	responseWriter := &FakeResponseWriter{make(map[string][]string), make([]byte, 20)}
	mux.ServeHTTP(responseWriter, req)

	statusCode := responseWriter.Headers.Get("Status")
	if statusCode != strconv.Itoa(http.StatusForbidden) {
		t.Errorf("Expected status code %q, but got %q", http.StatusForbidden, statusCode)
	}
}
