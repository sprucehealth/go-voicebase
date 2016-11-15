package admin

import (
	"encoding/json"
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/email"
	"github.com/sprucehealth/backend/cmd/svc/restapi/email/campaigns"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/mandrill"
	"github.com/sprucehealth/backend/libs/sig"
)

type emailTestSendHandler struct {
	emailService email.Service
	signer       *sig.Signer
	webDomain    string
}

type emailTestSendRequest struct {
	Type string `json:"type"`
}

type emailTestSendResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func newEmailTestSendHandler(emailService email.Service, signer *sig.Signer, webDomain string) http.Handler {
	return httputil.SupportedMethods(&emailTestSendHandler{
		emailService: emailService,
		signer:       signer,
		webDomain:    webDomain,
	}, httputil.Post)
}

func (h *emailTestSendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req emailTestSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		www.APIBadRequestError(w, r, "Failed to parse request body")
		return
	}

	account := www.MustCtxAccount(r.Context())
	audit.LogAction(account.ID, "AdminAPI", "SendTestEmail", map[string]interface{}{"type": req.Type})

	vars := campaigns.VarsForAccount(account.ID, req.Type, h.signer, h.webDomain)
	if _, err := h.emailService.Send([]int64{account.ID}, req.Type, map[int64][]mandrill.Var{account.ID: vars}, &mandrill.Message{}, 0); err != nil {
		httputil.JSONResponse(w, http.StatusOK, &emailTestSendResponse{Success: false, Error: err.Error()})
		return
	}
	httputil.JSONResponse(w, http.StatusOK, &emailTestSendResponse{Success: true})
}
