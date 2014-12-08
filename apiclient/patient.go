package apiclient

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/sprucehealth/backend/apiservice/apipaths"
	"github.com/sprucehealth/backend/messages"
)

type PatientClient struct {
	BaseURL    string
	AuthToken  string
	HostHeader string
}

func (pc *PatientClient) PostCaseMessage(caseID int64, msg string, attachments []*messages.Attachment) (int64, error) {
	var res messages.PostMessageResponse
	err := pc.do("POST", apipaths.CaseMessagesURLPath, nil,
		&messages.PostMessageRequest{
			CaseID:      caseID,
			Message:     msg,
			Attachments: attachments,
		}, &res, nil)
	return res.MessageID, err
}

func (pc *PatientClient) ListCaseMessages(caseID int64) ([]*messages.Message, []*messages.Participant, error) {
	var res messages.ListResponse
	err := pc.do("GET", apipaths.CaseMessagesListURLPath,
		url.Values{
			"case_id": []string{strconv.FormatInt(caseID, 10)},
		}, nil, &res, nil)
	return res.Items, res.Participants, err
}

func (pc *PatientClient) do(method, path string, params url.Values, req, res interface{}, headers http.Header) error {
	return do(pc.BaseURL, pc.AuthToken, pc.HostHeader, method, path, params, req, res, headers)
}
