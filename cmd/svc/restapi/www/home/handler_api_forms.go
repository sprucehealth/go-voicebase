package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/mux"
)

type formsAPIHandler struct {
	dataAPI api.DataAPI
}

func newFormsAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&formsAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Post)
}

func (h *formsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	formName := mux.Vars(r.Context())["form"]
	formType := common.Forms[formName]
	if formType == nil {
		golog.Warningf("Form %s not found", formName)
		www.APINotFound(w, r)
		return
	}
	form, ok := reflect.New(formType).Interface().(api.Form)
	if !ok {
		www.APIInternalError(w, r, fmt.Errorf("Form type %s does not conform to the Form interface", formName))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	requestID := httputil.RequestID(r.Context())
	if err := h.dataAPI.RecordForm(form, "home", requestID); err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, true)
}
