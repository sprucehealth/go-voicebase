package apiservice

import (
	"carefront/libs/golog"
	"fmt"
	"net/http"
)

type spruceError struct {
	DeveloperError     string `json:"developer_error,omitempty"`
	UserError          string `json:"user_error,omitempty"`
	DeveloperErrorCode int64  `json:"developer_code,string,omitempty"`
	HTTPStatusCode     int    `json:"-"`
	RequestID          int64  `json:"request_id,string,omitempty"`
}

type NotAuthorizedError string

func (e NotAuthorizedError) Error() string {
	return fmt.Sprintf("not authorized: %s", string(e))
}

func NewValidationError(msg string, r *http.Request) *spruceError {
	return &spruceError{
		UserError:      msg,
		HTTPStatusCode: http.StatusBadRequest,
		RequestID:      GetContext(r).RequestID,
	}
}

func wrapInternalError(err error, code int, r *http.Request) *spruceError {
	return &spruceError{
		DeveloperError: err.Error(),
		UserError:      genericUserErrorMessage,
		RequestID:      GetContext(r).RequestID,
		HTTPStatusCode: code,
	}
}

func (s *spruceError) Error() string {
	var msg string
	if s.DeveloperErrorCode > 0 {
		msg = fmt.Sprintf("RequestID: %d, Error: %s, ErrorCode: %d", s.RequestID, s.DeveloperError, s.DeveloperErrorCode)
	} else {
		msg = fmt.Sprintf("RequestID: %d, Error: %s", s.RequestID, s.DeveloperError)
	}
	return msg
}

var IsDev = false

func WriteError(err error, w http.ResponseWriter, r *http.Request) {
	switch err := err.(type) {
	case *spruceError:
		writeSpruceError(err, w, r)
	case NotAuthorizedError:
		writeSpruceError(&spruceError{
			UserError:      string(err),
			HTTPStatusCode: http.StatusForbidden,
			RequestID:      GetContext(r).RequestID,
		}, w, r)
	default:
		writeSpruceError(wrapInternalError(err, http.StatusInternalServerError, r), w, r)
	}
}

func WriteValidationError(msg string, w http.ResponseWriter, r *http.Request) {
	writeSpruceError(NewValidationError(msg, r), w, r)
}

// WriteBadRequestError is used for errors that occur during parsing of the HTTP request.
func WriteBadRequestError(err error, w http.ResponseWriter, r *http.Request) {
	writeSpruceError(wrapInternalError(err, http.StatusBadRequest, r), w, r)
}

// WriteAccessNotAllowedError is used when the user is authenticated but not
// authorized to access a given resource. Hopefully the user will never see
// this since the client shouldn't present the option to begin with.
func WriteAccessNotAllowedError(w http.ResponseWriter, r *http.Request) {
	writeSpruceError(spruceError{
		UserError:      "Access not permitted",
		RequestID:      GetContext(r).RequestID,
		HTTPStatusCode: http.StatusForbidden,
	}, w, r)
}

func writeSpruceError(err *spruceError, w http.ResponseWriter, r *http.Request) {
	golog.Logf(3, golog.ERR, err.Error())

	// remove the developer error information if we are not dealing with
	// before sending information across the wire
	if !IsDev {
		err.DeveloperError = ""
	}
	WriteJSONToHTTPResponseWriter(w, err.HTTPStatusCode, err)
}
