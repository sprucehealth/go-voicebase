package handlers

import (
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
)

type isAuthenticatedHandler struct {
	authAPI api.AuthAPI
}

// NewIsAuthenticatedHandler returns an initialized instance of isAuthenticatedHandler
func NewIsAuthenticatedHandler(authAPI api.AuthAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(apiservice.NoAuthorizationRequired(
		&isAuthenticatedHandler{
			authAPI: authAPI,
		}), httputil.Get)
}

func (i *isAuthenticatedHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := apiservice.MustCtxAccount(ctx)
	go func() {
		// asyncrhonously update the last opened date for this account
		if err := i.authAPI.UpdateLastOpenedDate(account.ID); err != nil {
			golog.Errorf("Unable to update last opened date for account: %s", err)
		}
	}()
	apiservice.WriteJSONSuccess(w)
}
