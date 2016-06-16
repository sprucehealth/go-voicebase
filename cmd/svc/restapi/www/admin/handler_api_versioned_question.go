package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/responses"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/schema"
	"golang.org/x/net/context" // The base handler struct to handle requests for interacting with versioned questions
)

type versionedQuestionHandler struct {
	dataAPI api.DataAPI
}

// Description of the request to be used for querying this endpoint with GET
type versionedQuestionGETRequest struct {
	ID         int64  `schema:"id"`
	Tag        string `schema:"tag"`
	Version    int64  `schema:"version"`
	LanguageID int64  `schema:"language_id,required"`
}

// Description of the respone object for a GET request
type versionedQuestionGETResponse struct {
	VersionedQuestion *responses.VersionedQuestion `json:"versioned_question"`
}

// Description of the request to be used for inserting new questions via POST
type versionedQuestionPOSTRequest struct {
	Tag                               string                                        `json:"tag"`
	LanguageID                        int64                                         `json:"language_id,string"`
	Type                              string                                        `json:"type"`
	Text                              string                                        `json:"text"`
	ParentQuestionID                  int64                                         `json:"parent_question_id,string"`
	Required                          bool                                          `json:"required"`
	FormattedFieldTags                string                                        `json:"formatted_field_tags"`
	ToAlert                           bool                                          `json:"to_alert"`
	TextHasTokens                     bool                                          `json:"text_has_tokens"`
	Subtext                           string                                        `json:"subtext"`
	SummaryText                       string                                        `json:"summary_text"`
	AlertText                         string                                        `json:"alert_text"`
	VersionedAnswers                  []*versionedAnswerPOSTRequest                 `json:"versioned_answers"`
	VersionedAdditionalQuestionFields *versionedAdditionalQuestionFieldsPOSTRequest `json:"versioned_additional_question_fields"`
	VersionedPhotoSlots               []*versionedPhotoSlotPOSTRequest              `json:"versioned_photo_slots"`
}

type versionedAnswerPOSTRequest struct {
	Tag         string                 `json:"tag"`
	Type        string                 `json:"type"`
	LanguageID  int64                  `json:"language_id,string"`
	Ordering    int64                  `json:"ordering"`
	Text        string                 `json:"text"`
	ToAlert     bool                   `json:"to_alert"`
	SummaryText string                 `json:"summary_text"`
	Status      string                 `json:"status"`
	ClientData  map[string]interface{} `json:"client_data"`
}

type versionedPhotoSlotPOSTRequest struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Required   bool                   `json:"required"`
	ClientData map[string]interface{} `json:"client_data"`
}

type versionedAdditionalQuestionFieldsPOSTRequest map[string]interface{}

// newVersionedQuestionHandler returns a new handler to access the question bank
// Authorization Required: true
// Supported Roles: ADMIN_ROLE
// Supported Method: GET, POST
func newVersionedQuestionHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		&versionedQuestionHandler{
			dataAPI: dataAPI,
		}, httputil.Get, httputil.Post)
}

func (h *versionedQuestionHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		requestData, err := h.parseGETRequest(ctx, r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.serveGET(w, r, requestData)
	case "POST":
		requestData, err := h.parsePOSTRequest(ctx, r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.servePOST(w, r, requestData)
	}
}

// parseGETRequest parses and verifies that a valid combination of GET parameters were supplied to the API
// Valid combinations inclde
//	ID - Returns a question that maps to a specific ID
//	Tag & Version - Returns a question that maps to a specific tag and version
func (h *versionedQuestionHandler) parseGETRequest(ctx context.Context, r *http.Request) (*versionedQuestionGETRequest, error) {
	rd := &versionedQuestionGETRequest{}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}
	if err := schema.NewDecoder().Decode(rd, r.Form); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	// If none of the critical params are provided then we have an invalid request
	// If we have partially completed sets then we have an invalid request
	// If no language ID is present then we have an invalid request
	if rd.ID == 0 && (rd.Tag == "" || rd.LanguageID == 0) {
		return nil, fmt.Errorf("insufficent parameters supplied to form complete query")
	}

	return rd, nil
}

func (h *versionedQuestionHandler) serveGET(w http.ResponseWriter, r *http.Request, rd *versionedQuestionGETRequest) {
	if rd.ID != 0 {
		h.serveQuestionIDGET(w, r, rd.ID, rd.LanguageID)
	} else {
		if rd.Version == 0 {
			version, err := h.dataAPI.MaxQuestionVersion(rd.Tag, rd.LanguageID)
			if err != nil {
				www.APIInternalError(w, r, err)
				return
			}
			rd.Version = version
		}
		h.serveQuestionTagGET(w, r, rd.Tag, rd.Version, rd.LanguageID)
	}
}

func (h *versionedQuestionHandler) serveQuestionIDGET(w http.ResponseWriter, r *http.Request, id, languageID int64) {
	vq, err := h.dataAPI.VersionedQuestionFromID(id)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	h.serveGETPOSTPostFetch(w, r, vq)
}

func (h *versionedQuestionHandler) serveQuestionTagGET(w http.ResponseWriter, r *http.Request, tag string, version, languageID int64) {
	vqs, err := h.dataAPI.VersionedQuestions([]*api.QuestionQueryParams{&api.QuestionQueryParams{LanguageID: languageID, Version: version, QuestionTag: tag}})
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	if len(vqs) != 1 {
		www.APIInternalError(w, r, fmt.Errorf("Expected only 1 result from question tag query but got %d", len(vqs)))
	}
	vq := vqs[0]

	h.serveGETPOSTPostFetch(w, r, vq)
}

func (h *versionedQuestionHandler) serveGETPOSTPostFetch(w http.ResponseWriter, r *http.Request, vq *common.VersionedQuestion) {
	response := versionedQuestionGETResponse{
		VersionedQuestion: responses.NewVersionedQuestionFromDBModel(vq),
	}

	vaqs, err := h.dataAPI.VersionedAdditionalQuestionFields(vq.ID, vq.LanguageID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	vaqsr, err := responses.VersionedAdditionalQuestionFieldsFromDBModels(vaqs)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	response.VersionedQuestion.VersionedAdditionalQuestionFields = vaqsr

	answers, err := answerResponsesForQuestion(h.dataAPI, vq.ID, vq.LanguageID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	response.VersionedQuestion.VersionedAnswers = answers

	photoSlots, err := photoSlotResponsesForQuestion(h.dataAPI, vq.ID, vq.LanguageID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	response.VersionedQuestion.VersionedPhotoSlots = photoSlots

	httputil.JSONResponse(w, http.StatusOK, response)
}

func (h *versionedQuestionHandler) parsePOSTRequest(ctx context.Context, r *http.Request) (*versionedQuestionPOSTRequest, error) {
	rd := &versionedQuestionPOSTRequest{}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		return nil, fmt.Errorf("Unable to parse body: %s", err)
	}

	if rd.Type == "" || rd.Tag == "" || rd.LanguageID == 0 || rd.VersionedAnswers == nil {
		return nil, fmt.Errorf("insufficent parameters supplied to form complete request body")
	}

	ordering := make(map[int64]bool)
	for _, va := range rd.VersionedAnswers {
		if va.Status == "" || va.Tag == "" || va.Type == "" {
			return nil, errors.New("Answer in question answer set is malformed")
		}
		_, ok := ordering[va.Ordering]
		if ok {
			return nil, fmt.Errorf("Found duplicate answer ordering %d", va.Ordering)
		}
		ordering[va.Ordering] = true
	}

	return rd, nil
}

func (h *versionedQuestionHandler) servePOST(w http.ResponseWriter, r *http.Request, rd *versionedQuestionPOSTRequest) {
	vq := &common.VersionedQuestion{
		AlertText:        rd.AlertText,
		LanguageID:       rd.LanguageID,
		ParentQuestionID: &rd.ParentQuestionID,
		SubtextText:      rd.Subtext,
		SummaryText:      rd.SummaryText,
		QuestionTag:      rd.Tag,
		QuestionText:     rd.Text,
		TextHasTokens:    rd.TextHasTokens,
		ToAlert:          rd.ToAlert,
		QuestionType:     rd.Type,
		Required:         rd.Required,
	}

	vas := make([]*common.VersionedAnswer, len(rd.VersionedAnswers))
	for i, va := range rd.VersionedAnswers {
		var clientData []byte
		var err error
		if va.ClientData != nil {
			clientData, err = json.Marshal(va.ClientData)
			if err != nil {
				www.APIInternalError(w, r, err)
				return
			}
		}

		vas[i] = &common.VersionedAnswer{
			AnswerTag:         va.Tag,
			ToAlert:           va.ToAlert,
			Ordering:          va.Ordering,
			LanguageID:        va.LanguageID,
			AnswerText:        va.Text,
			AnswerSummaryText: va.SummaryText,
			Status:            va.Status,
			AnswerType:        va.Type,
			QuestionID:        vq.ID,
			ClientData:        clientData,
		}
	}

	vpss := make([]*common.VersionedPhotoSlot, len(rd.VersionedPhotoSlots))
	for i, vps := range rd.VersionedPhotoSlots {
		clientData, err := json.Marshal(vps.ClientData)
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		vpss[i] = &common.VersionedPhotoSlot{
			Ordering:   int64(i),
			LanguageID: vq.LanguageID,
			Status:     `ACTIVE`,
			Type:       vps.Type,
			Name:       vps.Name,
			ClientData: clientData,
			Required:   vps.Required,
		}
	}

	var vaqf *common.VersionedAdditionalQuestionField
	if rd.VersionedAdditionalQuestionFields != nil {
		jsonBytes, err := json.Marshal(rd.VersionedAdditionalQuestionFields)
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		vaqf = &common.VersionedAdditionalQuestionField{
			LanguageID: vq.LanguageID,
			JSON:       jsonBytes,
		}
	}

	if vq.ParentQuestionID != nil && *vq.ParentQuestionID == 0 {
		vq.ParentQuestionID = nil
	}

	id, err := h.dataAPI.InsertVersionedQuestion(vq, vas, vpss, vaqf)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	// Note: Why waste a read here and look this back up? We want to assert we're giving the user an honest view of the question bank
	// This API is not super latency sensitive
	vq, err = h.dataAPI.VersionedQuestionFromID(id)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	h.serveGETPOSTPostFetch(w, r, vq)
}

func answerResponsesForQuestion(dataAPI api.DataAPI, questionID, languageID int64) ([]*responses.VersionedAnswer, error) {
	vas, err := dataAPI.VersionedAnswers([]*api.AnswerQueryParams{&api.AnswerQueryParams{LanguageID: languageID, QuestionID: questionID}})
	if err != nil {
		return nil, err
	}
	rs := make([]*responses.VersionedAnswer, len(vas))
	for i, va := range vas {
		rs[i], err = responses.NewVersionedAnswerFromDBModel(va)
		if err != nil {
			return nil, err
		}
	}
	return rs, nil
}

func photoSlotResponsesForQuestion(dataAPI api.DataAPI, questionID, languageID int64) ([]*responses.VersionedPhotoSlot, error) {
	vpss, err := dataAPI.VersionedPhotoSlots(questionID, languageID)
	if err != nil {
		return nil, err
	}
	rs := make([]*responses.VersionedPhotoSlot, len(vpss))
	for i, vps := range vpss {
		rs[i], err = responses.NewVersionedPhotoSlotFromDBModel(vps)
		if err != nil {
			return nil, err
		}
	}
	return rs, nil
}
