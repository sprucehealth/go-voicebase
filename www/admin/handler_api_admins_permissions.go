package admin

import (
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

type adminsPermissionsAPIHandler struct {
	authAPI api.AuthAPI
}

func newAdminsPermissionsAPIHandler(authAPI api.AuthAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(&adminsPermissionsAPIHandler{
		authAPI: authAPI,
	}, httputil.Get)
}

func (h *adminsPermissionsAPIHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(mux.Vars(ctx)["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	account := www.MustCtxAccount(ctx)
	audit.LogAction(account.ID, "AdminAPI", "GetAdminPermissions", map[string]interface{}{"param_account_id": accountID})

	// Verify account exists and is the correct role
	acc, err := h.authAPI.GetAccount(accountID)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	} else if acc.Role != api.RoleAdmin {
		www.APINotFound(w, r)
		return
	}

	perms, err := h.authAPI.PermissionsForAccount(accountID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, &struct {
		Permissions []string `json:"permissions"`
	}{
		Permissions: perms,
	})
}
