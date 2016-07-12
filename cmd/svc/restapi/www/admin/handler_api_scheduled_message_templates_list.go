package admin

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/httputil"
)

type schedMessageTemplatesListAPIHandler struct {
	dataAPI api.DataAPI
}

func newSchedMessageTemplatesListAPIHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(&schedMessageTemplatesListAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get, httputil.Post)
}

func (h *schedMessageTemplatesListAPIHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := www.MustCtxAccount(ctx)

	if r.Method == "POST" {
		audit.LogAction(account.ID, "AdminAPI", "CreateScheduledMessageTemplate", nil)

		var rep common.ScheduledMessageTemplate
		if err := json.NewDecoder(r.Body).Decode(&rep); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		if err := h.dataAPI.CreateScheduledMessageTemplate(&rep); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, rep.ID)
		return
	}

	audit.LogAction(account.ID, "AdminAPI", "ListScheduledMessageTemplates", nil)

	templates, err := h.dataAPI.ListScheduledMessageTemplates()
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, templates)
}
