package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"

	"github.com/sprucehealth/backend/cmd/svc/restapi/feedback"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/httputil"

	"golang.org/x/net/context"
)

type ratingLevelFeedbackConfigHandler struct {
	feedbackClient feedback.DAL
}

type ratingLevelFeedbackConfigData struct {
	Configs map[string]string `json:"configs"`
}

func newRatingLevelFeedbackConfigHandler(feedbackClient feedback.DAL) httputil.ContextHandler {
	return httputil.SupportedMethods(&ratingLevelFeedbackConfigHandler{
		feedbackClient: feedbackClient,
	}, httputil.Get, httputil.Put)
}

func (f *ratingLevelFeedbackConfigHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httputil.Get:
		f.get(ctx, w, r)
	case httputil.Put:
		f.put(ctx, w, r)
	}
}

func (f *ratingLevelFeedbackConfigHandler) get(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := www.MustCtxAccount(ctx)
	audit.LogAction(account.ID, "AdminAPI", "GetRatingConfigs", nil)

	configs, err := f.feedbackClient.RatingConfigs()
	if err != nil {
		www.InternalServerError(w, r, err)
		return
	}

	c := make(map[string]string)
	for rtg, cfg := range configs {
		c[strconv.Itoa(rtg)] = cfg
	}

	httputil.JSONResponse(w, http.StatusOK, &ratingLevelFeedbackConfigData{
		Configs: c,
	})
}

func (f *ratingLevelFeedbackConfigHandler) put(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := www.MustCtxAccount(ctx)
	audit.LogAction(account.ID, "AdminAPI", "PutRatingConfigs", nil)

	var rd ratingLevelFeedbackConfigData
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		www.InternalServerError(w, r, err)
		return
	}

	config := make(map[int]string)
	for rtg, c := range rd.Configs {
		rInt, err := strconv.Atoi(rtg)
		if err != nil {
			www.BadRequestError(w, r, err)
			return
		}
		config[rInt] = c
	}

	if err := f.feedbackClient.UpsertRatingConfigs(config); err != nil {
		www.APIBadRequestError(w, r, errors.Cause(err).Error())
		return
	}

	httputil.JSONResponse(w, http.StatusOK, map[string]interface{}{})
}