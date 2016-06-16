package medrecord

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"golang.org/x/net/context"
)

type apiHandler struct {
	dataAPI api.DataAPI
	queue   *common.SQSQueue
}

type RequestResponse struct {
	MedicalRecordID int64 `json:"medical_record_id"`
}

func NewRequestAPIHandler(dataAPI api.DataAPI, queue *common.SQSQueue) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(&apiHandler{
				dataAPI: dataAPI,
				queue:   queue,
			}), api.RolePatient), httputil.Post)
}

func (h *apiHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := apiservice.MustCtxAccount(ctx)
	patientID, err := h.dataAPI.GetPatientIDFromAccountID(account.ID)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	mrID, err := h.dataAPI.CreateMedicalRecord(patientID)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	js, err := json.Marshal(&queueMessage{
		MedicalRecordID: mrID,
		PatientID:       patientID,
	})
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	jsStr := string(js)
	if _, err := h.queue.QueueService.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &h.queue.QueueURL,
		MessageBody: &jsStr,
	}); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, &RequestResponse{
		MedicalRecordID: mrID,
	})
}
