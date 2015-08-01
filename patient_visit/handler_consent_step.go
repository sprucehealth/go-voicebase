package patient_visit

import (
	"encoding/json"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/ptr"
)

type reachedConsentStepHandler struct {
	dataAPI api.DataAPI
}

type reachedConsentStepPostRequest struct {
	VisitID int64 `json:"patient_visit_id,string"`
}

// NewReachedConsentStep returns a new handler that is called by the app when
// a patient younger than 18 reaches the end of their visit and needs parental
// consent to continue further.
func NewReachedConsentStep(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(
				&reachedConsentStepHandler{
					dataAPI: dataAPI,
				}),
			api.RolePatient), httputil.Post)
}

func (h *reachedConsentStepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req reachedConsentStepPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiservice.WriteBadRequestError(err, w, r)
		return
	}

	patientID, err := h.dataAPI.GetPatientIDFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}
	visit, err := h.dataAPI.GetPatientVisitFromID(req.VisitID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// Verify the visit is owned by the patient making the request
	if patientID != visit.PatientID.Int64() {
		apiservice.WriteAccessNotAllowedError(w, r)
		return
	}

	switch visit.Status {
	case common.PVStatusOpen:
		// Only open visits can transition to pending consent
	case common.PVStatusPendingParentalConsent:
		apiservice.WriteJSONSuccess(w)
		return
	default:
		apiservice.WriteValidationError("The visit is not open", w, r)
		return
	}
	_, err = h.dataAPI.UpdatePatientVisit(visit.ID.Int64(), &api.PatientVisitUpdate{
		Status:         ptr.String(common.PVStatusPendingParentalConsent),
		RequiredStatus: ptr.String(common.PVStatusOpen),
	})
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}
	apiservice.WriteJSONSuccess(w)
}