package patient_visit

import (
	"bytes"
	"carefront/api"
	"carefront/apiservice"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
)

type PhotoAnswerIntakeHandler struct {
	DataApi             api.DataAPI
	CloudStorageApi     api.CloudStorageAPI
	PatientVisitBucket  string
	MaxInMemoryForPhoto int64
	AWSRegion           string
}

type PhotoAnswerIntakeResponse struct {
	Result string `json:"result"`
}

type PhotoAnswerIntakeRequestData struct {
	QuestionId        int64 `schema:"question_id,required"`
	PotentialAnswerId int64 `schema:"potential_answer_id,required"`
	PatientVisitId    int64 `schema:"patient_visit_id,required"`
}

func NewPhotoAnswerIntakeHandler(dataApi api.DataAPI, cloudStorageApi api.CloudStorageAPI, bucketLocation, region string, maxMemoryForPhotoMB int64) *PhotoAnswerIntakeHandler {
	return &PhotoAnswerIntakeHandler{
		DataApi:             dataApi,
		CloudStorageApi:     cloudStorageApi,
		PatientVisitBucket:  bucketLocation,
		MaxInMemoryForPhoto: maxMemoryForPhotoMB,
		AWSRegion:           region,
	}
}

func (p *PhotoAnswerIntakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_POST {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := r.ParseMultipartForm(p.MaxInMemoryForPhoto)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse out the form values for the request: "+err.Error())
		return
	}

	var requestData PhotoAnswerIntakeRequestData
	if err := schema.NewDecoder().Decode(&requestData, r.Form); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	patientId, err := p.DataApi.GetPatientIdFromAccountId(apiservice.GetContext(r).AccountId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError,
			"Unable to get patientId from the accountId retrieved from auth token: "+err.Error())
		return
	}

	patientIdFromPatientVisitId, err := p.DataApi.GetPatientIdFromPatientVisitId(requestData.PatientVisitId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to get the patient id from the patient visit id: "+err.Error())
		return
	}

	if patientIdFromPatientVisitId != patientId {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "patient id retrieved from the patient_visit_id does not match patient id retrieved from auth token")
		return
	}

	layoutVersionId, err := p.DataApi.GetLayoutVersionIdForPatientVisit(requestData.PatientVisitId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Error getting latest layout version id: "+err.Error())
		return
	}

	questionType, err := p.DataApi.GetQuestionType(requestData.QuestionId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Error getting question type: "+err.Error())
		return
	}

	if questionType != "q_type_single_photo" && questionType != "q_type_multiple_photo" && questionType != "q_type_photo" {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "This api is only for uploading pictures")
		return
	}

	file, handler, err := r.FormFile("photo")
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Missing or invalid photo in parameters: "+err.Error())
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Error reading data from photo: "+err.Error())
		return
	}

	// in the event that we are dealing with a question that can only have one photo set for the potential answer,
	// mark the previously set answer to the quesiton as inactive
	if questionType == "q_type_single_photo" {
		err = p.DataApi.MakeCurrentPhotoAnswerInactive(api.PATIENT_ROLE, patientId, requestData.QuestionId, requestData.PatientVisitId, requestData.PotentialAnswerId, layoutVersionId)
		if err != nil {
			apiservice.WriteDeveloperError(w, http.StatusInternalServerError,
				"Error marking the current active photo answer as inactive: "+err.Error())
			return
		}
	}

	// create the record for answer input and mark it as pending upload
	patientAnswerInfoIntakeId, err := p.DataApi.CreatePhotoAnswerForQuestionRecord(api.PATIENT_ROLE, patientId,
		requestData.QuestionId, requestData.PatientVisitId, requestData.PotentialAnswerId, layoutVersionId)
	var buffer bytes.Buffer
	buffer.WriteString(strconv.Itoa(int(requestData.PatientVisitId)))
	buffer.WriteString("/")
	buffer.WriteString(strconv.FormatInt(patientAnswerInfoIntakeId, 10))

	parts := strings.Split(handler.Filename, ".")
	if len(parts) > 1 {
		buffer.WriteString(".")
		buffer.WriteString(parts[1])
	}

	objectStorageId, _, err := p.CloudStorageApi.PutObjectToLocation(p.PatientVisitBucket, buffer.String(), p.AWSRegion,
		handler.Header.Get("Content-Type"), data, time.Now().Add(10*time.Minute), p.DataApi)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Error uploading image to patient-visit bucket in s3: "+err.Error())
		return
	}

	// once the upload is complete, go ahead and mark the record as active with the object storage id linked
	err = p.DataApi.UpdatePhotoAnswerRecordWithObjectStorageId(patientAnswerInfoIntakeId, objectStorageId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, `Unable to update photo answer record with 
			object storage id after uploading picture: `+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, PhotoAnswerIntakeResponse{Result: "success"})
}
