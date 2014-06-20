package doctor_queue

import (
	"carefront/api"
	"carefront/apiservice"
	"net/http"

	"github.com/gorilla/schema"
)

const (
	state_completed = "completed"
	state_pending   = "pending"
)

type QueueHandler struct {
	dataApi api.DataAPI
}

func NewQueueHandler(dataApi api.DataAPI) *QueueHandler {
	return &QueueHandler{
		dataApi: dataApi,
	}
}

type DoctorQueueRequestData struct {
	State string `schema:"state"`
}

func (d *QueueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_GET {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	var requestData DoctorQueueRequestData
	if err := schema.NewDecoder().Decode(&requestData, r.Form); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	doctorId, err := d.dataApi.GetDoctorIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get doctor id from account id:"+err.Error())
		return
	}

	var pendingItemsDoctorQueue, unclaimedItemsDoctorQueue, completedItemsDoctorQueue []*api.DoctorQueueItem

	if requestData.State == "" || requestData.State == state_pending {
		pendingItemsDoctorQueue, err = d.dataApi.GetPendingItemsInDoctorQueue(doctorId)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		unclaimedItemsDoctorQueue, err = d.dataApi.GetElligibleItemsInUnclaimedQueue(doctorId)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	if requestData.State == "" || requestData.State == state_completed {
		completedItemsDoctorQueue, err = d.dataApi.GetCompletedItemsInDoctorQueue(doctorId)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	doctorDisplayFeed, err := d.convertDoctorQueueIntoDisplayQueue(pendingItemsDoctorQueue, unclaimedItemsDoctorQueue, completedItemsDoctorQueue)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &doctorDisplayFeed)
}

func (d *QueueHandler) convertDoctorQueueIntoDisplayQueue(pendingItems, unclaimedItems, completedItems []*api.DoctorQueueItem) (*DisplayFeedTabs, error) {
	var doctorDisplayFeedTabs DisplayFeedTabs

	var pendingOrOngoingDisplayFeed, completedDisplayFeed, unclaimedDisplayFeed *DisplayFeed
	doctorDisplayFeedTabs.Tabs = make([]*DisplayFeed, 0, 3)

	if pendingItems != nil {
		pendingOrOngoingDisplayFeed = &DisplayFeed{
			Title: "Pending",
		}
		doctorDisplayFeedTabs.Tabs = append(doctorDisplayFeedTabs.Tabs, pendingOrOngoingDisplayFeed)
	}

	if unclaimedItems != nil {
		unclaimedDisplayFeed = &DisplayFeed{
			Title: "Unclaimed",
		}
		doctorDisplayFeedTabs.Tabs = append(doctorDisplayFeedTabs.Tabs, unclaimedDisplayFeed)
	}

	if completedItems != nil {
		completedDisplayFeed = &DisplayFeed{
			Title: "Completed",
		}
		doctorDisplayFeedTabs.Tabs = append(doctorDisplayFeedTabs.Tabs, completedDisplayFeed)
	}

	if len(pendingItems) > 0 {

		// put the first item in the queue into the first section of the display feed
		visitsSection := &DisplayFeedSection{}
		for i, doctorQueueItem := range pendingItems {
			doctorQueueItem.PositionInQueue = i
			item, err := converQueueItemToDisplayFeedItem(d.dataApi, doctorQueueItem)
			if err != nil {
				return nil, err
			}
			visitsSection.Items = append(visitsSection.Items, item)
		}

		pendingOrOngoingDisplayFeed.Sections = []*DisplayFeedSection{visitsSection}
	}

	if len(unclaimedItems) > 0 {
		currentDisplaySection := &DisplayFeedSection{}
		for i, unclaimedItem := range unclaimedItems {
			unclaimedItem.PositionInQueue = i
			displayItem, err := converQueueItemToDisplayFeedItem(d.dataApi, unclaimedItem)
			if err != nil {
				return nil, err
			}
			currentDisplaySection.Items = append(currentDisplaySection.Items, displayItem)
		}
		unclaimedDisplayFeed.Sections = []*DisplayFeedSection{currentDisplaySection}
	}

	if len(completedItems) > 0 {
		currentDisplaySection := &DisplayFeedSection{}
		for i, completedItem := range completedItems {
			completedItem.PositionInQueue = i
			displayItem, err := converQueueItemToDisplayFeedItem(d.dataApi, completedItem)
			if err != nil {
				return nil, err
			}
			currentDisplaySection.Items = append(currentDisplaySection.Items, displayItem)
		}
		completedDisplayFeed.Sections = []*DisplayFeedSection{currentDisplaySection}
	}

	return &doctorDisplayFeedTabs, nil
}
