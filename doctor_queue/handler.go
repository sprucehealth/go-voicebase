package doctor_queue

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
)

const (
	stateCompleted = "completed"
	stateLocal     = "local"
	stateGlobal    = "global"
)

type queueHandler struct {
	dataAPI api.DataAPI
}

type DoctorQueueItemsResponseData struct {
	Items []*DisplayFeedItem `json:"items"`
}

func NewQueueHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(
				&queueHandler{
					dataAPI: dataAPI,
				}), []string{api.DOCTOR_ROLE, api.MA_ROLE}),
		[]string{"GET"})
}

type DoctorQueueRequestData struct {
	State string `schema:"state"`
}

func (d *queueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	requestData := &DoctorQueueRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteValidationError("Unable to parse input parameters", w, r)
		return
	} else if requestData.State == "" {
		apiservice.WriteValidationError("State (local,global,completed) required", w, r)
		return
	}

	doctorID, err := d.dataAPI.GetDoctorIDFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// only add auth url for items in global queue so that
	// the doctor can first be granted acess to the case before opening the case
	var addAuthUrl bool
	var queueItems []*api.DoctorQueueItem
	switch requestData.State {
	case stateLocal:
		queueItems, err = d.dataAPI.GetPendingItemsInDoctorQueue(doctorID)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	case stateGlobal:
		if apiservice.GetContext(r).Role == api.MA_ROLE {
			queueItems, err = d.dataAPI.GetPendingItemsForClinic()
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}
		} else {
			addAuthUrl = true
			queueItems, err = d.dataAPI.GetElligibleItemsInUnclaimedQueue(doctorID)
			if err != nil && err != api.NoRowsError {
				apiservice.WriteError(err, w, r)
				return
			}
		}
	case stateCompleted:
		if apiservice.GetContext(r).Role == api.MA_ROLE {
			queueItems, err = d.dataAPI.GetCompletedItemsForClinic()
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}
		} else {
			queueItems, err = d.dataAPI.GetCompletedItemsInDoctorQueue(doctorID)
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}
		}
	default:
		apiservice.WriteValidationError("Unexpected state value. Can only be local, global or completed", w, r)
		return
	}

	feedItems := make([]*DisplayFeedItem, 0, len(queueItems))
	for i, doctorQueueItem := range queueItems {
		doctorQueueItem.PositionInQueue = i
		doctorQueueItem.DoctorContextID = doctorID
		feedItem, err := converQueueItemToDisplayFeedItem(d.dataAPI, doctorQueueItem)
		if err != nil {
			golog.Errorf("Unable to convert item (Id: %d, EventType: %s, Status: %s, ItemId: %d) into display item", doctorQueueItem.ID,
				doctorQueueItem.EventType, doctorQueueItem.Status, doctorQueueItem.ItemID)
			continue
		}
		if addAuthUrl {
			feedItem.AuthUrl = app_url.ClaimPatientCaseAction(doctorQueueItem.PatientCaseID)
		}

		feedItems = append(feedItems, feedItem)
	}
	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &DoctorQueueItemsResponseData{Items: feedItems})
}
