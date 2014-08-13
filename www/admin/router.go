package admin

import (
	"log"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/payment/stripe"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/third_party/github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/www"
)

func SetupRoutes(r *mux.Router, dataAPI api.DataAPI, authAPI api.AuthAPI, stripeCli *stripe.StripeService, signer *common.Signer, stores map[string]storage.Store, templateLoader *www.TemplateLoader, metricsRegistry metrics.Registry) {
	if stores["onboarding"] == nil {
		log.Fatal("onboarding storage not configured")
	}

	templateLoader.MustLoadTemplate("admin/base.html", "base.html", nil)

	adminRoles := []string{api.ADMIN_ROLE}
	authFilter := www.AuthRequiredFilter(authAPI, adminRoles, nil)

	r.Handle(`/admin`, authFilter(http.RedirectHandler("/admin/doctor", http.StatusSeeOther))).Name("admin")
	r.Handle(`/admin/doctor`, authFilter(NewDoctorSearchHandler(r, dataAPI, templateLoader))).Name("admin-doctor-search")
	r.Handle(`/admin/doctor/{id:[0-9]+}`, authFilter(NewDoctorHandler(r, dataAPI, templateLoader))).Name("admin-doctor")
	r.Handle(`/admin/doctor/{id:[0-9]+}/dl/{attr:[A-Za-z0-9_\-]+}`, authFilter(NewDoctorAttrDownloadHandler(r, dataAPI, stores["onboarding"]))).Name("admin-doctor-attr-download")
	r.Handle(`/admin/doctor/onboard`, authFilter(NewDoctorOnboardHandler(r, dataAPI, signer))).Name("admin-doctor-onboard")
	r.Handle(`/admin/resourceguide`, authFilter(NewResourceGuideListHandler(r, dataAPI, templateLoader))).Name("admin-resourceguide-list")
	r.Handle(`/admin/resourceguide/{id:[0-9]+}`, authFilter(NewResourceGuideHandler(r, dataAPI, templateLoader))).Name("admin-resourceguide")
	r.Handle(`/admin/rxguide`, authFilter(NewRXGuideListHandler(r, dataAPI, templateLoader))).Name("admin-rxguide-list")
	r.Handle(`/admin/rxguide/{ndc:[a-zA-Z0-9]+}`, authFilter(NewRXGuideHandler(r, dataAPI, templateLoader))).Name("admin-rxguide")

	apiAuthFailHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		www.JSONResponse(w, r, http.StatusForbidden, &www.APIError{Message: "Access not allowed"})
	})
	apiAuthFilter := www.AuthRequiredFilter(authAPI, adminRoles, apiAuthFailHandler)

	r.Handle(`/admin/api/doctor/{id:[0-9]+}/licenses`, apiAuthFilter(NewMedicalLicenseAPIHandler(dataAPI)))
	r.Handle(`/admin/api/doctor/{id:[0-9]+}/profile`, apiAuthFilter(NewDoctorProfileAPIHandler(dataAPI)))
}
