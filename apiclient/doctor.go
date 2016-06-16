package apiclient

import (
	"net/url"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice/apipaths"
	"github.com/sprucehealth/backend/cmd/svc/restapi/patient_file"
	"github.com/sprucehealth/backend/common"
	diaghandlers "github.com/sprucehealth/backend/diagnosis/handlers"
	"github.com/sprucehealth/backend/doctor"
	"github.com/sprucehealth/backend/doctor_queue"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/messages"
	"github.com/sprucehealth/backend/responses"
	"github.com/sprucehealth/backend/tagging"
)

type DoctorClient struct {
	Config
}

func (dc *DoctorClient) SetToken(token string) {
	dc.Config.AuthToken = token
}

// Auth signs in as the given doctor account returning the auth response.
// AuthToken is not updated because that could lead to a race condition.
// It is up to the caller to update the struct.
func (dc *DoctorClient) Auth(email, password string) (*doctor.AuthenticationResponse, error) {
	var res doctor.AuthenticationResponse
	err := dc.do("POST", apipaths.DoctorAuthenticateURLPath, nil,
		doctor.AuthenticationRequestData{
			Email:    email,
			Password: password,
		}, &res, nil)
	return &res, err
}

// UpdateTreatmentPlanNote sets the personalized note for a treatment plan.
func (dc *DoctorClient) UpdateTreatmentPlanNote(treatmentPlanID int64, note string) error {
	return dc.do("PUT", apipaths.DoctorSavedNoteURLPath, nil,
		doctor_treatment_plan.DoctorSavedNoteRequestData{
			TreatmentPlanID: treatmentPlanID,
			Message:         note,
		}, nil, nil)
}

func (dc *DoctorClient) Inbox() ([]*doctor_queue.DoctorQueueDisplayItem, error) {
	var items struct {
		Items []*doctor_queue.DoctorQueueDisplayItem `json:"items"`
	}
	err := dc.do("GET", apipaths.DoctorQueueInboxURLPath, nil,
		nil, &items, nil)
	return items.Items, err
}

func (dc *DoctorClient) UnassignedQueue() ([]*doctor_queue.DoctorQueueDisplayItem, error) {
	var items struct {
		Items []*doctor_queue.DoctorQueueDisplayItem `json:"items"`
	}
	err := dc.do("GET", apipaths.DoctorQueueUnassignedURLPath, nil,
		nil, &items, nil)
	return items.Items, err
}

func (dc *DoctorClient) History() ([]*doctor_queue.DoctorQueueDisplayItem, error) {
	var items struct {
		Items []*doctor_queue.DoctorQueueDisplayItem `json:"items"`
	}
	err := dc.do("GET", apipaths.DoctorQueueHistoryURLPath, nil,
		nil, &items, nil)
	return items.Items, err
}

func (dc *DoctorClient) ReviewVisit(patientVisitID int64) (*patient_file.VisitReviewResponse, error) {
	var res patient_file.VisitReviewResponse
	err := dc.do("GET", apipaths.DoctorVisitReviewURLPath, url.Values{
		"patient_visit_id": []string{strconv.FormatInt(patientVisitID, 10)},
	}, nil, &res, nil)
	return &res, err
}

func (dc *DoctorClient) ClaimCase(caseID int64) error {
	return dc.do("POST", apipaths.DoctorCaseClaimURLPath, nil, &doctor_queue.ClaimPatientCaseRequestData{
		PatientCaseID: encoding.DeprecatedNewObjectID(caseID),
	}, nil, nil)
}

// TreatmentPlan fetches the doctor's view of a treatment plan given an ID.
func (dc *DoctorClient) TreatmentPlan(id int64, abridged bool, sections doctor_treatment_plan.Sections) (*responses.TreatmentPlan, error) {
	var res doctor_treatment_plan.DoctorTreatmentPlanResponse
	params := url.Values{"treatment_plan_id": []string{strconv.FormatInt(id, 10)}}
	if abridged {
		params.Set("abridged", "true")
	}
	params.Set("sections", sections.String())
	err := dc.do("GET", apipaths.DoctorTreatmentPlansURLPath, params, nil, &res, nil)
	if err != nil {
		return nil, err
	}
	return res.TreatmentPlan, nil
}

func (dc *DoctorClient) SelectMedication(name, strength string) (*doctor_treatment_plan.NewTreatmentResponse, error) {
	var res doctor_treatment_plan.NewTreatmentResponse
	if err := dc.do("GET", apipaths.DoctorSelectMedicationURLPath, url.Values{
		"drug_internal_name":  []string{name},
		"medication_strength": []string{strength}}, nil, &res, nil); err != nil {
		return nil, err
	}

	return &res, nil
}

func (dc *DoctorClient) AddTreatmentsToTreatmentPlan(treatments []*common.Treatment, tpID int64) (*doctor_treatment_plan.GetTreatmentsResponse, error) {
	var res doctor_treatment_plan.GetTreatmentsResponse
	req := &doctor_treatment_plan.AddTreatmentsRequestBody{
		TreatmentPlanID: encoding.DeprecatedNewObjectID(tpID),
		Treatments:      treatments,
	}

	if err := dc.do("POST", apipaths.DoctorVisitTreatmentsURLPath, nil, req, &res, nil); err != nil {
		return nil, err
	}

	return &res, nil
}

func (dc *DoctorClient) DeleteTreatmentPlan(id int64) error {
	return dc.do("DELETE", apipaths.DoctorTreatmentPlansURLPath,
		url.Values{"treatment_plan_id": []string{strconv.FormatInt(id, 10)}},
		nil, nil, nil)
}

func (dc *DoctorClient) PickTreatmentPlanForVisit(visitID int64, ftp *responses.FavoriteTreatmentPlan) (*responses.TreatmentPlan, error) {
	req := &doctor_treatment_plan.TreatmentPlanRequestData{
		TPParent: &common.TreatmentPlanParent{
			ParentID:   encoding.DeprecatedNewObjectID(visitID),
			ParentType: common.TPParentTypePatientVisit,
		},
	}
	if ftp != nil {
		req.TPContentSource = &common.TreatmentPlanContentSource{
			Type: common.TPContentSourceTypeFTP,
			ID:   ftp.ID,
		}
	}
	var res doctor_treatment_plan.DoctorTreatmentPlanResponse
	if err := dc.do("POST", apipaths.DoctorTreatmentPlansURLPath, nil, req, &res, nil); err != nil {
		return nil, err
	}
	return res.TreatmentPlan, nil
}

func (dc *DoctorClient) SubmitTreatmentPlan(treatmentPlanID int64) error {
	return dc.do("PUT", apipaths.DoctorTreatmentPlansURLPath, nil,
		doctor_treatment_plan.TreatmentPlanRequestData{
			TreatmentPlanID: treatmentPlanID,
		}, nil, nil)
}

func (dc *DoctorClient) ListFavoriteTreatmentPlans() ([]*responses.PathwayFTPGroup, error) {
	params := url.Values{}

	var res doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData
	err := dc.do("GET", apipaths.DoctorFTPURLPath, params, nil, &res, nil)
	if err != nil {
		return nil, err
	}
	return res.FavoriteTreatmentPlansByPathway, nil
}

func (dc *DoctorClient) ListFavoriteTreatmentPlansForTag(pathwayTag string) ([]*responses.FavoriteTreatmentPlan, error) {
	params := url.Values{
		"pathway_id": []string{pathwayTag},
	}

	var res doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData
	err := dc.do("GET", apipaths.DoctorFTPURLPath, params, nil, &res, nil)
	if err != nil {
		return nil, err
	}
	return res.FavoriteTreatmentPlans, nil
}

func (dc *DoctorClient) CreateFavoriteTreatmentPlan(ftp *responses.FavoriteTreatmentPlan) (*responses.FavoriteTreatmentPlan, error) {
	return dc.CreateFavoriteTreatmentPlanFromTreatmentPlan(ftp, 0)
}

func (dc *DoctorClient) CreateFavoriteTreatmentPlanFromTreatmentPlan(ftp *responses.FavoriteTreatmentPlan, tpID int64) (*responses.FavoriteTreatmentPlan, error) {
	var res doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData
	err := dc.do("POST", apipaths.DoctorFTPURLPath, nil,
		&doctor_treatment_plan.DoctorFavoriteTreatmentPlansRequestData{
			FavoriteTreatmentPlan: ftp,
			TreatmentPlanID:       tpID,
		}, &res, nil)
	if err != nil {
		return nil, err
	}
	return res.FavoriteTreatmentPlan, err
}

func (dc *DoctorClient) UpdateFavoriteTreatmentPlan(ftp *responses.FavoriteTreatmentPlan) (*responses.FavoriteTreatmentPlan, error) {
	var res doctor_treatment_plan.DoctorFavoriteTreatmentPlansResponseData
	err := dc.do("PUT", apipaths.DoctorFTPURLPath, nil,
		&doctor_treatment_plan.DoctorFavoriteTreatmentPlansRequestData{
			FavoriteTreatmentPlan: ftp,
		}, &res, nil)
	return res.FavoriteTreatmentPlan, err
}

func (dc *DoctorClient) DeleteFavoriteTreatmentPlan(id int64) error {
	return dc.do("DELETE", apipaths.DoctorFTPURLPath,
		url.Values{"favorite_treatment_plan_id": []string{strconv.FormatInt(id, 10)}},
		nil, nil, nil)
}

func (dc *DoctorClient) CreateRegimenPlan(regimen *common.RegimenPlan) (*common.RegimenPlan, error) {
	var res common.RegimenPlan
	if err := dc.do("POST", apipaths.DoctorRegimenURLPath, nil, regimen, &res, nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (dc *DoctorClient) PostCaseMessage(caseID int64, msg string, attachments []*messages.Attachment) (int64, error) {
	var res messages.PostMessageResponse
	err := dc.do("POST", apipaths.CaseMessagesURLPath, nil,
		&messages.PostMessageRequest{
			CaseID:      caseID,
			Message:     msg,
			Attachments: attachments,
		}, &res, nil)
	return res.MessageID, err
}

func (dc *DoctorClient) ListCaseMessages(caseID int64) ([]*messages.Message, []*messages.Participant, error) {
	var res messages.ListResponse
	err := dc.do("GET", apipaths.CaseMessagesListURLPath,
		url.Values{
			"case_id": []string{strconv.FormatInt(caseID, 10)},
		}, nil, &res, nil)
	return res.Items, res.Participants, err
}

func (dc *DoctorClient) AssignCase(caseID int64, msg string, attachments []*messages.Attachment) (int64, error) {
	var res messages.PostMessageResponse
	err := dc.do("POST", apipaths.DoctorAssignCaseURLPath, nil,
		&messages.PostMessageRequest{
			CaseID:      caseID,
			Message:     msg,
			Attachments: attachments,
		}, &res, nil)
	return res.MessageID, err
}

func (dc *DoctorClient) DoctorCaseHistory() ([]*doctor_queue.PatientsFeedItem, error) {
	var res doctor_queue.PatientsFeedResponse
	err := dc.do("GET", apipaths.DoctorCaseHistoryURLPath, nil, nil, &res, nil)
	return res.Items, err
}

func (dc *DoctorClient) CreateDiagnosisSet(rd *diaghandlers.DiagnosisListRequestData) error {
	return dc.do("PUT", apipaths.DoctorVisitDiagnosisListURLPath, nil, rd, nil, nil)
}

func (dc *DoctorClient) ListDiagnosis(visitID int64) (*diaghandlers.DiagnosisListResponse, error) {
	var res diaghandlers.DiagnosisListResponse
	err := dc.do("GET", apipaths.DoctorVisitDiagnosisListURLPath,
		url.Values{
			"patient_visit_id": []string{strconv.FormatInt(visitID, 10)},
		}, nil, &res, nil)
	return &res, err
}

func (dc *DoctorClient) GetDiagnosis(codeID string) (*diaghandlers.DiagnosisOutputItem, error) {
	var res diaghandlers.DiagnosisOutputItem
	err := dc.do("GET", apipaths.DoctorDiagnosisURLPath,
		url.Values{
			"code_id": []string{codeID},
		}, nil, &res, nil)
	return &res, err
}

func (dc *DoctorClient) SearchDiagnosis(query string) (*diaghandlers.DiagnosisSearchResult, error) {
	var res diaghandlers.DiagnosisSearchResult
	err := dc.do("GET", apipaths.DoctorDiagnosisSearchURLPath,
		url.Values{
			"query": []string{query},
		}, nil, &res, nil)
	return &res, err

}

func (dc *DoctorClient) CasesForPatient(patientID common.PatientID) ([]*responses.Case, error) {
	var res struct {
		Cases []*responses.Case `json:"cases"`
	}
	err := dc.do("GET", apipaths.DoctorPatientCasesListURLPath,
		url.Values{
			"patient_id": []string{patientID.String()},
		}, nil, &res, nil)
	return res.Cases, err
}

func (dc *DoctorClient) ListTreatmentPlanScheduledMessages(treatmentPlanID int64) ([]*responses.ScheduledMessage, error) {
	var res doctor_treatment_plan.ScheduledMessageListResponse
	err := dc.do("GET", apipaths.DoctorTPScheduledMessageURLPath,
		url.Values{
			"treatment_plan_id": []string{strconv.FormatInt(treatmentPlanID, 10)},
		}, nil, &res, nil)
	return res.Messages, err
}

func (dc *DoctorClient) CreateTreatmentPlanScheduledMessage(treatmentPlanID int64, msg *responses.ScheduledMessage) (int64, error) {
	req := &doctor_treatment_plan.ScheduledMessageRequest{
		TreatmentPlanID: treatmentPlanID,
		Message:         msg,
	}
	var res doctor_treatment_plan.ScheduledMessageIDResponse
	err := dc.do("POST", apipaths.DoctorTPScheduledMessageURLPath, nil, req, &res, nil)
	return res.MessageID, err
}

func (dc *DoctorClient) UpdateTreatmentPlanScheduledMessage(treatmentPlanID int64, msg *responses.ScheduledMessage) (int64, error) {
	req := &doctor_treatment_plan.ScheduledMessageRequest{
		TreatmentPlanID: treatmentPlanID,
		Message:         msg,
	}
	var res doctor_treatment_plan.ScheduledMessageIDResponse
	err := dc.do("PUT", apipaths.DoctorTPScheduledMessageURLPath, nil, req, &res, nil)
	return res.MessageID, err
}

func (dc *DoctorClient) DeleteTreatmentPlanScheduledMessages(treatmentPlanID, messageID int64) error {
	return dc.do("DELETE", apipaths.DoctorTPScheduledMessageURLPath,
		url.Values{
			"treatment_plan_id": []string{strconv.FormatInt(treatmentPlanID, 10)},
			"message_id":        []string{strconv.FormatInt(messageID, 10)},
		}, nil, nil, nil)
}

func (dc *DoctorClient) CancelTreatmentPlanScheduledMessage(messageID int64, undo bool) error {
	req := &doctor_treatment_plan.CancelScheduledMessageRequest{
		MessageID: messageID,
		Undo:      undo,
	}
	return dc.do("PUT", apipaths.DoctorCancelTPScheduledMessageURLPath,
		nil, req, nil, nil)
}

func (dc *DoctorClient) AddResourceGuidesToTreatmentPlan(tpID int64, guideIDs []int64) error {
	req := &doctor_treatment_plan.ResourceGuideRequest{
		TreatmentPlanID: tpID,
		GuideIDs:        make([]encoding.ObjectID, len(guideIDs)),
	}
	for i, id := range guideIDs {
		req.GuideIDs[i] = encoding.DeprecatedNewObjectID(id)
	}
	return dc.do("PUT", apipaths.TPResourceGuideURLPath, nil, req, nil, nil)
}

func (dc *DoctorClient) RemoveResourceGuideFromTreatmentPlan(tpID, guideID int64) error {
	return dc.do("DELETE", apipaths.TPResourceGuideURLPath,
		url.Values{
			"treatment_plan_id": []string{strconv.FormatInt(tpID, 10)},
			"resource_guide_id": []string{strconv.FormatInt(guideID, 10)},
		}, nil, nil, nil)
}

func (dc *DoctorClient) ResolveRXErrorByRefillRequestID(refillRequestID int64) error {
	req := &doctor.DoctorPrescriptionErrorIgnoreRequestData{
		RefillRequestID: refillRequestID,
	}
	return dc.do("POST", apipaths.DoctorRXErrorResolveURLPath,
		nil, req, nil, nil)
}

func (dc *DoctorClient) Tags(req *tagging.TagGETRequest) (*tagging.TagGETResponse, error) {
	res := &tagging.TagGETResponse{}
	if err := dc.do("GET", apipaths.TagURLPath, url.Values{
		"text":   req.Text,
		"common": []string{strconv.FormatBool(req.Common)},
	}, nil, &res, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (dc *DoctorClient) DeleteTag(req *tagging.TagDELETERequest) error {
	return dc.do("DELETE", apipaths.TagURLPath, url.Values{
		"id": []string{strconv.FormatInt(req.ID, 10)},
	}, nil, nil, nil)
}

func (dc *DoctorClient) TagCaseMemberships(req *tagging.TagCaseMembershipGETRequest) (*tagging.TagCaseMembershipGETResponse, error) {
	res := &tagging.TagCaseMembershipGETResponse{}
	if err := dc.do("GET", apipaths.TagCaseMembershipURLPath, url.Values{
		"case_id": []string{strconv.FormatInt(req.CaseID, 10)},
	}, nil, &res, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (dc *DoctorClient) PutTagCaseMembership(req *tagging.TagCaseMembershipPUTRequest) error {
	return dc.do("PUT", apipaths.TagCaseMembershipURLPath, nil, req, nil, nil)
}

func (dc *DoctorClient) DeleteTagCaseMembership(req *tagging.TagCaseMembershipDELETERequest) error {
	return dc.do("DELETE", apipaths.TagCaseMembershipURLPath, url.Values{
		"case_id": []string{strconv.FormatInt(req.CaseID, 10)},
		"tag_id":  []string{strconv.FormatInt(req.TagID, 10)},
	}, nil, nil, nil)
}

func (dc *DoctorClient) TagCaseAssociations(req *tagging.TagCaseAssociationGETRequest) (*tagging.TagCaseAssociationGETResponse, error) {
	res := &tagging.TagCaseAssociationGETResponse{}
	if err := dc.do("GET", apipaths.TagCaseAssociationURLPath, url.Values{
		"query":        []string{req.Query},
		"start":        []string{strconv.FormatInt(req.Start, 10)},
		"end":          []string{strconv.FormatInt(req.End, 10)},
		"past_trigger": []string{strconv.FormatBool(req.PastTrigger)},
	}, nil, &res, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (dc *DoctorClient) PostTagCaseAssociation(req *tagging.TagCaseAssociationPOSTRequest) (*tagging.TagCaseAssociationPOSTResponse, error) {
	res := &tagging.TagCaseAssociationPOSTResponse{}
	if err := dc.do("POST", apipaths.TagCaseAssociationURLPath, nil, req, &res, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (dc *DoctorClient) DeleteTagCaseAssociation(req *tagging.TagCaseAssociationDELETERequest) error {
	return dc.do("DELETE", apipaths.TagCaseAssociationURLPath, url.Values{
		"text":    []string{req.Text},
		"case_id": []string{strconv.FormatInt(req.CaseID, 10)},
	}, nil, nil, nil)
}
