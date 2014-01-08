package apiservice

import (
	"carefront/api"
	"carefront/common"
	"github.com/gorilla/schema"
	"net/http"
)

type PatientVisitFollowUpHandler struct {
	DataApi   api.DataAPI
	accountId int64
}

type PatientVisitFollowUpRequestResponse struct {
	PatientVisitId      int64  `schema:"patient_visit_id"`
	CurrentTimeOnClient int64  `schema:"client_time"`
	FollowUpValue       int64  `schema:"follow_up_value"`
	FollowUpUnit        string `schema:"follow_up_unit"`
}

type PatientVisitFollowupResponse struct {
	Result string `json:"result,omitempty"`
	*common.FollowUp
}

func NewPatientVisitFollowUpHandler(dataApi api.DataAPI) *PatientVisitFollowUpHandler {
	return &PatientVisitFollowUpHandler{DataApi: dataApi, accountId: 0}
}

func (p *PatientVisitFollowUpHandler) AccountIdFromAuthToken(accountId int64) {
	p.accountId = accountId
}

func (p *PatientVisitFollowUpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p.getFollowupForPatientVisit(w, r)
	case "POST":
		p.updatePatientVisitFollowup(w, r)
	}
}

func (p *PatientVisitFollowUpHandler) getFollowupForPatientVisit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestData := new(PatientVisitFollowUpRequestResponse)
	decoder := schema.NewDecoder()
	err := decoder.Decode(requestData, r.Form)

	_, _, _, statusCode, err := ValidateDoctorAccessToPatientVisitAndGetRelevantData(requestData.PatientVisitId, p.accountId, p.DataApi)
	if err != nil {
		WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	followupTime, followupValue, followupUnit, err := p.DataApi.GetFollowUpTimeForPatientVisit(requestData.PatientVisitId)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to get follow up for patient visit: "+err.Error())
		return
	}

	response := &PatientVisitFollowupResponse{}
	if followupValue != 0 && followupUnit != "" {
		response.FollowUpTime = followupTime
		response.FollowUpUnit = followupUnit
		response.FollowUpValue = followupValue
		response.PatientVisitId = requestData.PatientVisitId

	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, response)
}

func (p *PatientVisitFollowUpHandler) updatePatientVisitFollowup(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestData := new(PatientVisitFollowUpRequestResponse)
	decoder := schema.NewDecoder()
	err := decoder.Decode(requestData, r.Form)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	switch requestData.FollowUpUnit {
	case api.FOLLOW_UP_WEEK, api.FOLLOW_UP_DAY, api.FOLLOW_UP_MONTH:
	default:
		WriteDeveloperError(w, http.StatusBadRequest, "Follow up unit should be week, month or day")
		return
	}

	doctorId, _, _, statusCode, err := ValidateDoctorAccessToPatientVisitAndGetRelevantData(requestData.PatientVisitId, p.accountId, p.DataApi)
	if err != nil {
		WriteDeveloperError(w, statusCode, err.Error())
		return
	}

	err = p.DataApi.UpdateFollowUpTimeForPatientVisit(requestData.PatientVisitId, requestData.CurrentTimeOnClient, doctorId, requestData.FollowUpValue, requestData.FollowUpUnit)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update followup for patient visit")
		return
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, &PatientVisitFollowupResponse{Result: "success"})
}
