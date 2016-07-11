package admin

import (
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/feedback"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/httputil"

	"golang.org/x/net/context"
)

type feedbackTemplateListHandler struct {
	feedbackClient feedback.DAL
}

type feedbackTemplateListResponse struct {
	Templates []*feedback.FeedbackTemplateData `json:"templates"`
}

func newFeedbackTemplateListHandler(feedbackClient feedback.DAL) httputil.ContextHandler {
	return httputil.SupportedMethods(&feedbackTemplateListHandler{
		feedbackClient: feedbackClient,
	}, httputil.Get)
}

func (f *feedbackTemplateListHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := www.MustCtxAccount(ctx)
	audit.LogAction(account.ID, "AdminAPI", "ListActiveFeedbackTemplates", nil)

	templates, err := f.feedbackClient.ListActiveTemplates()
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, &feedbackTemplateListResponse{
		Templates: templates,
	})
}