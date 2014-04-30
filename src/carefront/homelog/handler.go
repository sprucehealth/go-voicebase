package homelog

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/libs/golog"
	"fmt"
	"net/http"
	"strconv"
)

type ListHandler struct {
	dataAPI api.DataAPI
}

type DismissHandler struct {
	dataAPI api.DataAPI
}

type response struct {
	Notifications []view `json:"notifications"`
	LogItems      []view `json:"log_items"`
}

func NewListHandler(dataAPI api.DataAPI) *ListHandler {
	return &ListHandler{
		dataAPI: dataAPI,
	}
}

func NewDismissHandler(dataAPI api.DataAPI) *DismissHandler {
	return &DismissHandler{
		dataAPI: dataAPI,
	}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_GET {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	patientId, err := h.dataAPI.GetPatientIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "home/list: failed to get patient: "+err.Error())
		return
	}

	notes, _, err := h.dataAPI.GetNotificationsForPatient(patientId, notifyTypes)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "home/list: failed to get patient notifications: "+err.Error())
		return
	}
	noteViews := make([]view, 0, len(notes))
	for _, n := range notes {
		view, err := n.Data.(notification).makeView(h.dataAPI, patientId)
		if err != nil {
			golog.Errorf("Failed to create view for notification %d of type %s", n.Id, n.Data.TypeName())
			continue
		}
		noteViews = append(noteViews, view)
	}

	log, _, err := h.dataAPI.GetHealthLogForPatient(patientId, logItemTypes)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "home/list: failed to get health log: "+err.Error())
		return
	}
	logViews := make([]view, 0, len(log))
	for _, lit := range log {
		view, err := lit.Data.(logItem).makeView(h.dataAPI, patientId, lit)
		if err != nil {
			golog.Errorf("home/list: failed to create view for notification %d of type %s", lit.Id, lit.Data.TypeName())
			continue
		}
		logViews = append(logViews, view)
	}

	res := &response{
		Notifications: noteViews,
		LogItems:      logViews,
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, res)
}

func (h *DismissHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_POST {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		apiservice.WriteUserError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	var noteIDs []int64
	for _, s := range r.PostForm["notification_id"] {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			apiservice.WriteUserError(w, http.StatusBadRequest, fmt.Sprintf("home/dismiss: notification ID '%s' not an integer", s))
			return
		}
		noteIDs = append(noteIDs, id)
	}
	if len(noteIDs) == 0 {
		apiservice.WriteUserError(w, http.StatusBadRequest, "notification_id required")
		return
	}

	if err := h.dataAPI.DeletePatientNotifications(noteIDs); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "home/dismiss: failed to delete notifications: "+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, apiservice.SuccessfulGenericJSONResponse())
}
