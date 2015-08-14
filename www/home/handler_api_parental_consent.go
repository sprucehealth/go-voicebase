package home

import (
	"encoding/json"
	"net/http"

	"github.com/SpruceHealth/schema"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/patient"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

type parentalConsentAPIHandler struct {
	dataAPI    api.DataAPI
	dispatcher dispatch.Publisher
}

type parentalConsentAPIPOSTRequest struct {
	ChildPatientID common.PatientID `json:"child_patient_id"`
	Relationship   string           `json:"relationship"`
}

type parentalConsentAPIPOSTResponse struct{}

type parentalconsentAPIGETRequest struct {
	ChildPatientID common.PatientID `schema:"child_patient_id,required"`
}

type parentalConsentAPIGETResponse struct {
	Children []*childResponse `json:"children"`
}

type childResponse struct {
	ChildPatientID common.PatientID `json:"child_patient_id"`
	ChildFirstName string           `json:"child_first_name"`
	ChildGender    string           `json:"child_gender"`
	Consented      bool             `json:"consented"`
	Relationship   string           `json:"relationship,omitempty"`
}

func (r *parentalConsentAPIPOSTRequest) Validate() (bool, string) {
	if r.Relationship == "" {
		return false, "Relationship required"
	}
	return true, ""
}

func newParentalConsentAPIHAndler(dataAPI api.DataAPI, dispatcher dispatch.Publisher) httputil.ContextHandler {
	return httputil.SupportedMethods(
		www.APIRoleRequiredHandler(&parentalConsentAPIHandler{
			dataAPI:    dataAPI,
			dispatcher: dispatcher,
		}, api.RolePatient), httputil.Post, httputil.Get)
}

func (h *parentalConsentAPIHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httputil.Post:
		h.post(ctx, w, r)
	case httputil.Get:
		h.get(ctx, w, r)
	}
}

func (h *parentalConsentAPIHandler) post(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req parentalConsentAPIPOSTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		www.APIBadRequestError(w, r, err.Error())
		return
	}
	token := parentalConsentCookie(req.ChildPatientID, r)
	if !patient.ValidateParentalConsentToken(h.dataAPI, token, req.ChildPatientID) {
		if !environment.IsDev() {
			www.APIForbidden(w, r)
			return
		}
		// In dev let it work anyway but log it so it's obviousl what's happening
		golog.Errorf("Token is invalid but allowing in dev")
	}
	if ok, reason := req.Validate(); !ok {
		www.APIGeneralError(w, r, "invalid_request", reason)
		return
	}

	account := www.MustCtxAccount(ctx)
	parent, err := h.dataAPI.GetPatientFromAccountID(account.ID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	if parent.IsUnder18() {
		www.APIGeneralError(w, r, "under_age", "A parent or guardian must be 18 or older")
		return
	}

	newConsent, err := h.dataAPI.GrantParentChildConsent(parent.ID, req.ChildPatientID, req.Relationship)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	if newConsent {
		// It's possible this is a second child for the same parent in which case we'll already have identification photos.
		proof, err := h.dataAPI.ParentConsentProof(parent.ID)
		if err != nil {
			if !api.IsErrNotFound(err) {
				www.APIInternalError(w, r, err)
				return
			}
		} else if proof.IsComplete() {
			if err := patient.ParentalConsentCompleted(h.dataAPI, h.dispatcher, parent.ID, req.ChildPatientID); err != nil {
				www.APIInternalError(w, r, err)
				return
			}
		}
	}
	httputil.JSONResponse(w, http.StatusOK, parentalConsentAPIPOSTResponse{})
}

func (h *parentalConsentAPIHandler) get(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req parentalconsentAPIGETRequest
	if err := r.ParseForm(); err != nil {
		www.APIBadRequestError(w, r, "Bad request")
		return
	}
	if err := schema.NewDecoder().Decode(&req, r.Form); err != nil {
		www.APIBadRequestError(w, r, err.Error())
		return
	}

	account := www.MustCtxAccount(ctx)
	parentPatientID, err := h.dataAPI.GetPatientIDFromAccountID(account.ID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	consents, err := h.dataAPI.ParentalConsent(req.ChildPatientID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	// find the consent by the parent
	var consented *common.ParentalConsent
	for _, consent := range consents {
		if consent.ParentPatientID == parentPatientID {
			consented = consent
			break
		}
	}

	// Make sure parent has access to the child. Either a link exists (consent) or the provide token is valid.
	if consented == nil && !patient.ValidateParentalConsentToken(h.dataAPI, parentalConsentCookie(req.ChildPatientID, r), req.ChildPatientID) {
		www.APIForbidden(w, r)
		return
	}

	child, err := h.dataAPI.Patient(req.ChildPatientID, true)
	if err != nil {
		www.APIForbidden(w, r)
		return
	}

	var c bool
	if consented != nil {
		c = consented.Consented
	}

	var relationship string
	if consented != nil {
		relationship = consented.Relationship
	}

	res := &parentalConsentAPIGETResponse{
		Children: []*childResponse{
			{
				ChildPatientID: child.ID,
				ChildFirstName: child.FirstName,
				ChildGender:    child.Gender,
				Consented:      c,
				Relationship:   relationship,
			},
		},
	}
	httputil.JSONResponse(w, http.StatusOK, res)
}
