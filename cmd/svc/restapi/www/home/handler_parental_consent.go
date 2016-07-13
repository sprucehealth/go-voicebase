package home

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"time"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/mediastore"
	"github.com/sprucehealth/backend/cmd/svc/restapi/patient"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
)

type parentalConsentHandler struct {
	dataAPI         api.DataAPI
	mediaStore      *mediastore.Store
	template        *template.Template
	landingTemplate *template.Template
}

type parentalConsentContext struct {
	Account     *common.Account
	Environment string
	Hydration   *parentalConsentHydration
}

type parentalConsentHydration struct {
	ChildDetails               *patientContext
	ParentalConsent            *consentContext
	IsParentSignedIn           bool
	IdentityVerificationImages *identitiyImageContext
}

type patientContext struct {
	PatientID string `json:"patientID"`
	FirstName string `json:"firstName"`
	Gender    string `json:"gender"`
}

type consentContext struct {
	Consented    bool   `json:"consented"`
	Relationship string `json:"relationship"`
}

type identitiyImageContext struct {
	Types map[string]string `json:"types"`
}

func checkParentalConsentAccessToken(w http.ResponseWriter, r *http.Request, dataAPI api.DataAPI, childPatientID common.PatientID) bool {
	token := r.FormValue("t")
	fromCookie := false
	if token == "" {
		token = parentalConsentCookie(childPatientID, r)
		fromCookie = true
	}
	if token == "" {
		return false
	}
	hasAccess := patient.ValidateParentalConsentToken(dataAPI, token, childPatientID)
	if hasAccess && !fromCookie {
		// Only set the cookie if the token is actually valid
		cookie := newParentalConsentCookie(childPatientID, token, r)
		http.SetCookie(w, cookie)
	}
	return hasAccess
}

func newParentalConsentHandler(dataAPI api.DataAPI, mediaStore *mediastore.Store, templateLoader *www.TemplateLoader) http.Handler {
	return httputil.SupportedMethods(
		&parentalConsentHandler{
			dataAPI:         dataAPI,
			mediaStore:      mediaStore,
			template:        templateLoader.MustLoadTemplate("home/parental-consent.html", "", nil),
			landingTemplate: templateLoader.MustLoadTemplate("home/parental-landing.html", "home/parental-base.html", nil),
		}, httputil.Get)
}

func (h *parentalConsentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The person may not be signed in which is fine. Account will be nil then.
	account, _ := www.CtxAccount(r.Context())

	childPatientID, err := common.ParsePatientID(mux.Vars(r.Context())["childid"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	hasAccess := checkParentalConsentAccessToken(w, r, h.dataAPI, childPatientID)

	var consent *common.ParentalConsent
	var parentPatientID common.PatientID
	idProof := map[string]string{}
	if account != nil {
		if account.Role != api.RolePatient {
			www.RedirectToSignIn(w, r)
			return
		}
		parentPatientID, err = h.dataAPI.GetPatientIDFromAccountID(account.ID)
		if err != nil {
			www.InternalServerError(w, r, err)
			return
		}
		consents, err := h.dataAPI.ParentalConsent(childPatientID)
		if err != nil {
			www.InternalServerError(w, r, err)
			return
		}
		for _, c := range consents {
			if c.ParentPatientID == parentPatientID {
				consent = c
				break
			}
		}
		if len(consents) != 0 && consent == nil {
			// Only allow one parent to start the process. For now just redirect to sign in, but we should have a better state here.
			www.RedirectToSignIn(w, r)
			return
		}
		if consent != nil {
			hasAccess = true
		}
		proof, err := h.dataAPI.ParentConsentProof(parentPatientID)
		if err == nil {
			if proof.SelfiePhotoID != nil {
				idProof["selfie"], err = h.mediaStore.SignedURL(*proof.SelfiePhotoID, time.Hour)
				if err != nil {
					www.InternalServerError(w, r, err)
					return
				}
			}
			if proof.GovernmentIDPhotoID != nil {
				idProof["governmentid"], err = h.mediaStore.SignedURL(*proof.GovernmentIDPhotoID, time.Hour)
				if err != nil {
					www.InternalServerError(w, r, err)
					return
				}
			}
		}
	}
	if consent == nil {
		consent = &common.ParentalConsent{}
	}

	if !hasAccess {
		if !environment.IsDev() {
			www.RedirectToSignIn(w, r)
			return
		}
		// In dev let it work anyway but log it so it's obviousl what's happening
		golog.Errorf("Token is invalid but allowing in dev")
	}

	child, err := h.dataAPI.Patient(childPatientID, true)
	if api.IsErrNotFound(err) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		www.InternalServerError(w, r, err)
		return
	}

	// Already approved so jump straight to medical record
	if child.HasParentalConsent {
		http.Redirect(w, r, fmt.Sprintf("/pc/%s/medrecord", childPatientID), http.StatusSeeOther)
		return
	}

	if page := mux.Vars(r.Context())["page"]; page == "" {
		pronoun := "they"
		possessivePronoun := "their"
		switch child.Gender {
		case "male":
			pronoun = "he"
			possessivePronoun = "his"
		case "female":
			pronoun = "she"
			possessivePronoun = "her"
		}
		www.TemplateResponse(w, http.StatusOK, h.landingTemplate, &struct {
			Environment string
			Title       template.HTML
			SubContext  interface{}
		}{
			Environment: environment.GetCurrent(),
			Title:       "Parental Consent | Spruce",
			SubContext: struct {
				ChildID                int64
				ChildFirstName         string
				ChildPronoun           string
				ChildPossessivePronoun string
			}{
				ChildID:                child.ID.Int64(),
				ChildFirstName:         child.FirstName,
				ChildPronoun:           pronoun,
				ChildPossessivePronoun: possessivePronoun,
			},
		})
		return
	}

	www.TemplateResponse(w, http.StatusOK, h.template, &parentalConsentContext{
		Environment: environment.GetCurrent(),
		Hydration: &parentalConsentHydration{
			ChildDetails: &patientContext{
				PatientID: child.ID.String(),
				FirstName: child.FirstName,
				Gender:    child.Gender,
			},
			IsParentSignedIn: account != nil,
			ParentalConsent: &consentContext{
				Consented:    consent.Consented,
				Relationship: consent.Relationship,
			},
			IdentityVerificationImages: &identitiyImageContext{
				Types: idProof,
			},
		},
	})
}

type parentalLandingHandler struct {
	dataAPI  api.DataAPI
	template *template.Template
	title    string
	ctx      interface{}
}

func newParentalLandingHandler(dataAPI api.DataAPI, templateLoader *www.TemplateLoader, tmpl, title string, ctxFun func() interface{}) http.Handler {
	var ctx interface{}
	if ctxFun != nil {
		ctx = ctxFun()
	}
	return httputil.SupportedMethods(&parentalLandingHandler{
		dataAPI:  dataAPI,
		title:    title,
		template: templateLoader.MustLoadTemplate(tmpl, "home/parental-base.html", nil),
		ctx:      ctx,
	}, httputil.Get)
}

func (h *parentalLandingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	www.TemplateResponse(w, http.StatusOK, h.template, &struct {
		Account     *common.Account
		Environment string
		Title       template.HTML
		SubContext  interface{}
	}{
		Environment: environment.GetCurrent(),
		Title:       template.HTML(html.EscapeString(h.title)),
		SubContext:  h.ctx,
	})
}
