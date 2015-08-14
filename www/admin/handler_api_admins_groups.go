package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

type adminsGroupsAPIHandler struct {
	authAPI api.AuthAPI
}

func newAdminsGroupsAPIHandler(authAPI api.AuthAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(&adminsGroupsAPIHandler{
		authAPI: authAPI,
	}, httputil.Get, httputil.Post)
}

func (h *adminsGroupsAPIHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(mux.Vars(ctx)["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	account := www.MustCtxAccount(ctx)

	if r.Method == httputil.Post {
		// Use a string key because JSON
		var groups map[string]bool
		if err := json.NewDecoder(r.Body).Decode(&groups); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		groupsUpdate := make(map[int64]bool, len(groups))
		for idStr, state := range groups {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				www.APIInternalError(w, r, err)
				return
			}
			groupsUpdate[id] = state
		}

		audit.LogAction(account.ID, "AdminAPI", "UpdateAdminGroups", map[string]interface{}{"param_account_id": accountID, "groups": groups})

		if err := h.authAPI.UpdateGroupsForAccount(accountID, groupsUpdate); err != nil {
			www.InternalServerError(w, r, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, true)
		return
	}

	audit.LogAction(account.ID, "AdminAPI", "GetAdminGroups", map[string]interface{}{"param_account_id": accountID})

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

	groups, err := h.authAPI.GroupsForAccount(accountID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, &struct {
		Groups []*common.AccountGroup `json:"groups"`
	}{
		Groups: groups,
	})
}
