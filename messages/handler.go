package messages

import (
	"errors"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/httputil"
)

type handler struct {
	dataAPI    api.DataAPI
	dispatcher *dispatch.Dispatcher
}

type PostMessageRequest struct {
	CaseID      int64         `json:"case_id,string"`
	Message     string        `json:"message"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

func (r *PostMessageRequest) Validate() error {
	if r.CaseID <= 0 {
		return errors.New("case_id missing or invalid")
	}
	if r.Message == "" {
		return errors.New("message must not be blank")
	}
	return nil
}

type PostMessageResponse struct {
	MessageID int64 `json:"message_id,string"`
}

type Attachment struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	MimeType string `json:"mimetype,omitempty"`
	ID       int64  `json:"id,string"`
	URL      string `json:"url,omitempty"`
}

func NewHandler(dataAPI api.DataAPI, dispatcher *dispatch.Dispatcher) http.Handler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(
			&handler{
				dataAPI:    dataAPI,
				dispatcher: dispatcher}), []string{"POST"})
}

func (h *handler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)

	var req PostMessageRequest
	if err := apiservice.DecodeRequestData(&req, r); err != nil {
		return false, apiservice.NewValidationError(err.Error(), r)
	}
	ctxt.RequestCache[apiservice.RequestData] = &req

	if err := req.Validate(); err != nil {
		return false, apiservice.NewValidationError(err.Error(), r)
	}

	cas, err := h.dataAPI.GetPatientCaseFromID(req.CaseID)
	if err == api.NoRowsError {
		return false, err
	}
	ctxt.RequestCache[apiservice.PatientCase] = cas

	personID, doctorID, err := validateAccess(h.dataAPI, r, cas)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.PersonID] = personID
	ctxt.RequestCache[apiservice.DoctorID] = doctorID

	return true, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	req := ctxt.RequestCache[apiservice.RequestData].(*PostMessageRequest)
	personID := ctxt.RequestCache[apiservice.PersonID].(int64)
	doctorID := ctxt.RequestCache[apiservice.DoctorID].(int64)
	cas := ctxt.RequestCache[apiservice.PatientCase].(*common.PatientCase)

	people, err := h.dataAPI.GetPeople([]int64{personID})
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}
	person := people[personID]

	msg := &common.CaseMessage{
		CaseID:   req.CaseID,
		PersonID: personID,
		Body:     req.Message,
	}

	if err := CreateMessageAndAttachments(msg, req.Attachments, personID, doctorID, ctxt.Role, h.dataAPI); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	h.dispatcher.Publish(&PostEvent{
		Message: msg,
		Case:    cas,
		Person:  person,
	})

	res := &PostMessageResponse{
		MessageID: msg.ID,
	}
	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, res)
}
