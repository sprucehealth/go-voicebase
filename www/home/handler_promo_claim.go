package home

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/sprucehealth/backend/email"

	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/cost/promotions"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/www"
)

type promoContext struct {
	Code           string
	Email          string
	State          string
	StateName      string
	Platform       string
	States         []*common.State
	Promo          *promotions.PromotionDisplayInfo
	Claimed        bool
	InState        bool
	Message        string
	SuccessMessage string
	Android        bool
	NoVideo        bool
	Errors         map[string]string
}

type promoClaimHandler struct {
	dataAPI         api.DataAPI
	authAPI         api.AuthAPI
	analyticsLogger analytics.Logger
	template        *template.Template
}

func newPromoClaimHandler(dataAPI api.DataAPI, authAPI api.AuthAPI, analyticsLogger analytics.Logger, templateLoader *www.TemplateLoader) http.Handler {
	return httputil.SupportedMethods(&promoClaimHandler{
		dataAPI:         dataAPI,
		authAPI:         authAPI,
		analyticsLogger: analyticsLogger,
		template:        templateLoader.MustLoadTemplate("promotions/claim.html", "promotions/base.html", nil),
	}, []string{"GET", "POST"})
}

func (h *promoClaimHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &promoContext{
		Email:  r.FormValue("email"),
		State:  r.FormValue("state"),
		Errors: make(map[string]string),
	}

	var err error
	ctx.Code = mux.Vars(r)["code"]
	ctx.Promo, err = promotions.LookupPromoCode(ctx.Code, h.dataAPI, h.analyticsLogger)
	if err != promotions.InvalidCode && err != nil {
		www.InternalServerError(w, r, err)
		return
	}
	if ctx.Promo == nil {
		ctx.Message = "Sorry, that promotion is no longer valid."
	} else {
		ctx.States, err = h.dataAPI.ListStates()
		if err != nil {
			www.InternalServerError(w, r, err)
			return
		}
		for _, s := range ctx.States {
			if s.Abbreviation == ctx.State {
				ctx.StateName = s.Name
				break
			}
		}

		if r.Method == "POST" {
			if ctx.Email == "" || !email.IsValidEmail(ctx.Email) {
				ctx.Errors["email"] = "Please enter a valid email address."
			}
			if ctx.State == "" {
				ctx.Errors["state"] = "Please select a state."
			}

			if len(ctx.Errors) == 0 {
				ctx.SuccessMessage, err = promotions.AssociatePromoCode(ctx.Email, ctx.State, ctx.Code, h.dataAPI, h.authAPI, h.analyticsLogger, nil)
				if err != nil {
					www.InternalServerError(w, r, err)
					return
				}
				ctx.Android = !strings.Contains(r.UserAgent(), "iPhone")
				ctx.Claimed = true
				inState, err := h.dataAPI.IsEligibleToServePatientsInState(ctx.State, api.HEALTH_CONDITION_ACNE_ID)
				if err != nil {
					www.InternalServerError(w, r, err)
					return
				}
				if !inState {
					p := url.Values{
						"email":     []string{ctx.Email},
						"state":     []string{ctx.State},
						"stateName": []string{ctx.StateName},
					}
					http.Redirect(w, r, "/r/"+ctx.Code+"/notify/state?"+p.Encode(), http.StatusSeeOther)
					return
				}
			}
		}
	}

	www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
		Environment: environment.GetCurrent(),
		Title:       template.HTML("Claim a Promotion"),
		SubContext: &homeContext{
			NoBaseHeader: true,
			SubContext:   ctx,
		},
	})
}
