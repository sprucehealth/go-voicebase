package doctor_queue

import (
	"fmt"
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/app_url"
	"github.com/sprucehealth/backend/libs/httputil"
)

type itemHandler struct {
	dataAPI api.DataAPI
}

type itemRequest struct {
	Action string `json:"action"`
	ID     string `json:"id"`
}

func NewItemHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(
				&itemHandler{
					dataAPI: dataAPI,
				}), api.RoleCC),
		httputil.Put)
}

func (h *itemHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var rd itemRequest
	if err := apiservice.DecodeRequestData(&rd, r); err != nil {
		apiservice.WriteValidationError(ctx, err.Error(), w, r)
		return
	} else if rd.Action != "remove" {
		apiservice.WriteValidationError(ctx, fmt.Sprintf("%s action not supported", rd.Action), w, r)
		return
	}

	qid, err := queueItemPartsFromID(rd.ID)
	if err != nil {
		apiservice.WriteValidationError(ctx, err.Error(), w, r)
		return
	}

	switch qid.eventType {
	case api.DQEventTypeCaseAssignment, api.DQEventTypeCaseMessage:
		if qid.status != api.DQItemStatusPending {
			apiservice.WriteAccessNotAllowedError(ctx, w, r)
			return
		}
	case api.DQEventTypePatientVisit:
		if qid.status != api.DQItemStatusPending && qid.status != api.DQItemStatusOngoing {
			apiservice.WriteAccessNotAllowedError(ctx, w, r)
			return
		}
	default:
		apiservice.WriteAccessNotAllowedError(ctx, w, r)
		return
	}

	updates := []*api.DoctorQueueUpdate{
		{
			Action: api.DQActionRemove,
			QueueItem: &api.DoctorQueueItem{
				EventType: qid.eventType,
				Status:    qid.status,
				DoctorID:  qid.doctorID,
				ItemID:    qid.itemID,
				QueueType: qid.queueType,
			},
		},
	}
	if qid.eventType == api.DQEventTypePatientVisit {
		account := apiservice.MustCtxAccount(ctx)
		cc, err := h.dataAPI.GetDoctorFromAccountID(account.ID)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
		visit, err := h.dataAPI.GetPatientVisitFromID(qid.itemID)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
		patient, err := h.dataAPI.Patient(visit.PatientID, true)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
		updates = append(updates, &api.DoctorQueueUpdate{
			Action: api.DQActionInsert,
			QueueItem: &api.DoctorQueueItem{
				EventType:        qid.eventType,
				Status:           api.DQItemStatusRemoved,
				DoctorID:         cc.ID.Int64(),
				ItemID:           qid.itemID,
				QueueType:        api.DQTDoctorQueue,
				PatientID:        visit.PatientID,
				Description:      fmt.Sprintf("%s removed visit for %s %s from queue", cc.ShortDisplayName, patient.FirstName, patient.LastName),
				ShortDescription: "Visit removed from queue",
				ActionURL:        app_url.ViewPatientVisitInfoAction(visit.PatientID, qid.itemID, visit.PatientCaseID.Int64()),
			},
		})
	}

	if err := h.dataAPI.UpdateDoctorQueue(updates); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	apiservice.WriteJSONSuccess(w)
}
