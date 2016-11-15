package doctor_queue

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/app_url"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/tagging"
	"github.com/sprucehealth/backend/cmd/svc/restapi/tagging/query"
	"github.com/sprucehealth/backend/cmd/svc/restapi/tagging/response"
	"github.com/sprucehealth/schema"
)

type patientsFeedHandler struct {
	dataAPI       api.DataAPI
	taggingClient tagging.Client
}

type CaseFeedGETRequest struct {
	Query       string `schema:"query"`
	Start       *int64 `schema:"start"`
	End         *int64 `schema:"end"`
	PastTrigger bool   `schema:"past_trigger"`
}

func (r *CaseFeedGETRequest) IsTagQuery() bool {
	return r.Query != "" || r.PastTrigger
}

type PatientsFeedItem struct {
	ID               string                `json:"id"` // Unique to the content of the item
	PatientFirstName string                `json:"patient_first_name"`
	PatientLastName  string                `json:"patient_last_name"`
	LastVisitTime    int64                 `json:"last_visit_time"` // unix timestamp
	LastVisitDoctor  string                `json:"last_visit_doctor"`
	ActionURL        *app_url.SpruceAction `json:"action_url"`
	Tags             []string              `json:"tags"`
	CaseTags         []*response.Tag       `json:"case_tags"`
}

type PatientsFeedResponse struct {
	Items []*PatientsFeedItem `json:"items"`
}

func NewPatientsFeedHandler(dataAPI api.DataAPI, taggingClient tagging.Client) http.Handler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(
				&patientsFeedHandler{
					dataAPI:       dataAPI,
					taggingClient: taggingClient,
				}), api.RoleDoctor, api.RoleCC),
		httputil.Get)
}

func (h *patientsFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		req, err := h.parseGETRequest(r)
		if err != nil {
			apiservice.WriteBadRequestError(err, w, r)
			return
		}
		h.serveGET(w, r, req)
	}
}

func (h *patientsFeedHandler) parseGETRequest(r *http.Request) (*CaseFeedGETRequest, error) {
	rd := &CaseFeedGETRequest{}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}
	if err := schema.NewDecoder().Decode(rd, r.Form); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	return rd, nil
}

func (h *patientsFeedHandler) serveGET(w http.ResponseWriter, r *http.Request, rd *CaseFeedGETRequest) {
	account := apiservice.MustCtxAccount(r.Context())

	// Query items. MA gets all items. Doctors get only the cases they're involved with.
	var items []*common.PatientCaseFeedItem
	var err error
	var caseIDs []int64
	ops := tagging.TONone
	if rd.PastTrigger {
		ops = tagging.TOPastTrigger
	}
	if account.Role == api.RoleCC {
		//Only CC can access tag search functionality
		if rd.IsTagQuery() {
			memberships, err := h.taggingClient.TagMembershipQuery(rd.Query, ops)
			if query.IsErrBadExpression(err) {
				apiservice.WriteBadRequestError(err, w, r)
				return
			}
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}

			caseIDLookup := make(map[int64]bool)
			caseIDs = make([]int64, 0, len(memberships))
			for _, m := range memberships {
				if !caseIDLookup[*m.CaseID] {
					caseIDs = append(caseIDs, *m.CaseID)
					caseIDLookup[*m.CaseID] = true
				}
			}
		}

		var start, end *time.Time
		if rd.Start != nil {
			t := time.Unix(*rd.Start, 0)
			start = &t
		}
		if rd.End != nil {
			t := time.Unix(*rd.End, 0)
			end = &t
		}
		// Don't lookup any items if we provided query params and we didn't find any memberships
		if (rd.IsTagQuery() && len(caseIDs) > 0) || !rd.IsTagQuery() {
			items, err = h.dataAPI.PatientCaseFeed(caseIDs, start, end)
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}
		}
	} else {
		var doctorID int64
		doctorID, err = h.dataAPI.GetDoctorIDFromAccountID(account.ID)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
		items, err = h.dataAPI.PatientCaseFeedForDoctor(doctorID)
	}
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// Transform from data models to response
	if len(caseIDs) == 0 {
		caseIDs = make([]int64, len(items))
		for i, item := range items {
			caseIDs[i] = item.CaseID
		}
	}

	caseTagsByCaseIDOps := tagging.TONonHiddenOnly
	if rd.PastTrigger {
		caseTagsByCaseIDOps |= tagging.TOPastTrigger
	}
	caseTagsByCaseID, err := h.taggingClient.TagsForCases(caseIDs, caseTagsByCaseIDOps)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	res := &PatientsFeedResponse{
		Items: make([]*PatientsFeedItem, len(items)),
	}
	for i, it := range items {
		res.Items[i] = &PatientsFeedItem{
			// Generate an ID unique to the contents of the item
			ID:               fmt.Sprintf("%d:%s:%d:%d", it.DoctorID, it.PatientID, it.CaseID, it.LastVisitID),
			PatientFirstName: it.PatientFirstName,
			PatientLastName:  it.PatientLastName,
			LastVisitTime:    it.LastVisitTime.Unix(),
			LastVisitDoctor:  it.LastVisitDoctor,
			ActionURL:        app_url.CaseFeedItemAction(it.CaseID, it.PatientID, it.LastVisitID),
			Tags:             []string{it.PathwayName},
			CaseTags:         caseTagsByCaseID[it.CaseID],
		}
	}

	httputil.JSONResponse(w, http.StatusOK, res)
}
