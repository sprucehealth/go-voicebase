package notify

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
)

type promptStatusHandler struct {
	dataApi api.DataAPI
}

func NewPromptStatusHandler(dataApi api.DataAPI) http.Handler {
	return &promptStatusHandler{
		dataApi: dataApi,
	}
}

type promptStatusRequestData struct {
	PromptStatus string `schema:"prompt_status"`
}

func (p *promptStatusHandler) IsAuthorized(r *http.Request) (bool, error) {
	if r.Method != apiservice.HTTP_PUT {
		return false, apiservice.NewResourceNotFoundError("", r)
	}

	return true, nil
}

func (p *promptStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	rData := &promptStatusRequestData{}
	if err := apiservice.DecodeRequestData(rData, r); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	pStatus, err := common.GetPushPromptStatus(rData.PromptStatus)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := p.dataApi.SetPushPromptStatus(apiservice.GetContext(r).AccountId, pStatus); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, apiservice.SuccessfulGenericJSONResponse())
}
