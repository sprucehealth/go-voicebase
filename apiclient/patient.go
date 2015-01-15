package apiclient

import (
	"net/url"
	"strconv"

	"github.com/sprucehealth/backend/apiservice/apipaths"
	"github.com/sprucehealth/backend/messages"
	"github.com/sprucehealth/backend/patient"
)

type PatientClient struct {
	Config
}

type TreatmentPlanViews struct {
	HeaderViews      []map[string]interface{} `json:"header_views,omitempty"`
	TreatmentViews   []map[string]interface{} `json:"treatment_views,omitempty"`
	InstructionViews []map[string]interface{} `json:"instruction_views,omitempty"`
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

func (pc *PatientClient) TreatmentPlan(tpID int64) (*TreatmentPlanViews, error) {
	var res TreatmentPlanViews
	err := pc.do("GET", apipaths.TreatmentPlanURLPath,
		url.Values{
			"treatment_plan_id": []string{strconv.FormatInt(tpID, 10)},
		}, nil, &res, nil)
	return &res, err
}

func (pc *PatientClient) TreatmentPlanForCase(caseID int64) (*TreatmentPlanViews, error) {
	var res TreatmentPlanViews
	err := pc.do("GET", apipaths.TreatmentPlanURLPath,
		url.Values{
			"case_id": []string{strconv.FormatInt(caseID, 10)},
		}, nil, &res, nil)
	return &res, err
}

func (pc *PatientClient) SignUp(req *patient.SignupPatientRequestData) (*patient.PatientSignedupResponse, error) {
	var res patient.PatientSignedupResponse
	err := pc.do("POST", apipaths.PatientSignupURLPath, nil, req, &res, nil)
	return &res, err
}

func (pc *PatientClient) UpdatePatient(req *patient.UpdateRequest) error {
	return pc.do("PUT", apipaths.PatientUpdateURLPath, nil, req, nil, nil)
}
