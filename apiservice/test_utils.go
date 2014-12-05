package apiservice

import (
	"net/http"

	"github.com/sprucehealth/backend/environment"
)

const (
	VerifyAuthCode = 3198456
	authorization  = "authorization"
	authentication = "authentication"
)

func verifyAuthSetupInTest(
	w http.ResponseWriter,
	r *http.Request,
	h http.Handler,
	action string, response int) bool {

	if environment.IsTest() {
		test := r.FormValue("test")
		if test != "" && test == action {
			WriteJSON(w, map[string]interface{}{
				"result": response,
			})
			return true
		} else if test != "" {
			// bypass the check in the handler if the test parameter
			// value does not match the intended action. This is so that
			// any request handlers deeper in the chain can handle the test
			// probe appropriately
			h.ServeHTTP(w, r)
			return true
		}
	}
	return false
}
