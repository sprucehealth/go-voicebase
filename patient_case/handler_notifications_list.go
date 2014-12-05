package patient_case

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
)

type notificationsListHandler struct {
	dataAPI   api.DataAPI
	apiDomain string
}

type notificationsListRequestData struct {
	PatientCaseId int64 `schema:"case_id"`
}

type notificationsListResponseData struct {
	Items []common.ClientView `json:"items"`
}

func NewNotificationsListHandler(dataAPI api.DataAPI, apiDomain string) http.Handler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(&notificationsListHandler{
			dataAPI:   dataAPI,
			apiDomain: apiDomain,
		}), []string{"GET"})
}

func (n *notificationsListHandler) IsAuthorized(r *http.Request) (bool, error) {
	if apiservice.GetContext(r).Role != api.PATIENT_ROLE {
		return false, apiservice.NewAccessForbiddenError()
	}

	return true, nil
}

func (n *notificationsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestData := &notificationsListRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteError(err, w, r)
		return
	} else if requestData.PatientCaseId == 0 {
		apiservice.WriteValidationError("case_id must be specified", w, r)
		return
	}

	notificationItems, err := n.dataAPI.GetNotificationsForCase(requestData.PatientCaseId, NotifyTypes)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	nViewItems := make([]common.ClientView, 0, len(notificationItems))
	for _, notificationItem := range notificationItems {
		nDataItem := notificationItem.Data.(notification)
		if nDataItem.canRenderCaseNotificationView() {
			viewItem, err := nDataItem.makeCaseNotificationView(n.dataAPI, n.apiDomain, notificationItem)
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}

			nViewItems = append(nViewItems, viewItem)
		}
	}

	apiservice.WriteJSON(w, &notificationsListResponseData{Items: nViewItems})
}
