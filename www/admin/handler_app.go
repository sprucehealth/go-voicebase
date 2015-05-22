package admin

import (
	"html/template"
	"net/http"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/www"
)

type appHandler struct {
	template *template.Template
}

func NewAppHandler(templateLoader *www.TemplateLoader) http.Handler {
	return httputil.SupportedMethods(&appHandler{
		template: templateLoader.MustLoadTemplate("admin/app.html", "admin/base.html", nil),
	}, httputil.Get)
}

func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	account := context.Get(r, www.CKAccount).(*common.Account)
	perms := context.Get(r, www.CKPermissions).(www.Permissions)

	audit.LogAction(account.ID, "Admin", "LoadAdminApp", nil)

	www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
		Title: "Admin",
		SubContext: &struct {
			Account     *common.Account
			Permissions map[string]bool
			Environment string
		}{
			Account:     account,
			Permissions: perms,
			Environment: environment.GetCurrent(),
		},
	})
}
