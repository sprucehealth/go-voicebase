package apiservice

import (
	"carefront/api"
	"carefront/info_intake"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	layout_syntax_version = 1
)

type GenerateClientIntakeModelHandler struct {
	DataApi             api.DataAPI
	CloudStorageApi     api.CloudStorageAPI
	VisualLayoutBucket  string
	PatientLayoutBucket string
	AWSRegion           string
}

type ClientIntakeModelGeneratedResponse struct {
	ClientLayoutUrls []string `json:"clientModelUrls"`
}

func (l *GenerateClientIntakeModelHandler) NonAuthenticated() bool {
	return true
}

func (l *GenerateClientIntakeModelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != HTTP_POST {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	file, _, err := r.FormFile("layout")
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "No layout file or invalid layout file specified")
		return
	}

	healthCondition := &info_intake.InfoIntakeLayout{}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse layout file specified")
		return
	}

	err = json.Unmarshal(data, &healthCondition)
	if err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Error parsing layout file: "+err.Error())
		return
	}

	// determine the healthCondition tag so as to identify what healthCondition this layout belongs to
	healthConditionTag := healthCondition.HealthConditionTag
	if healthConditionTag == "" {
		WriteDeveloperError(w, http.StatusBadRequest, "health condition not specified or invalid in layout: "+err.Error())
		return
	}

	// get the healthConditionId
	healthConditionId, err := l.DataApi.GetHealthConditionInfo(healthConditionTag)

	modelId, err := l.DataApi.CreateLayoutVersion(data, layout_syntax_version, healthConditionId, api.PATIENT_ROLE, api.CONDITION_INTAKE_PURPOSE, "automatically generated")
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Error in creating new layout version: "+err.Error())
		return
	}

	// get all the supported languages
	_, supportedLanguageIds, err := l.DataApi.GetSupportedLanguages()

	// generate a client layout for each language
	clientIntakeModels := make(map[int64]*info_intake.InfoIntakeLayout)
	clientModelVersionIds := make([]int64, len(supportedLanguageIds))

	for i, supportedLanguageId := range supportedLanguageIds {
		clientModel := healthCondition
		if err := clientModel.FillInDatabaseInfo(l.DataApi, supportedLanguageId); err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Unable to populate the layout as expected: "+err.Error())
			return
		}
		clientIntakeModels[supportedLanguageId] = clientModel

		jsonData, err := json.Marshal(&clientModel)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Error generating client digestable layout: "+err.Error())
			return
		}

		// mark the client layout as creating until we have uploaded all client layouts before marking it as ACTIVE
		clientModelId, err := l.DataApi.CreatePatientLayout(jsonData, supportedLanguageId, modelId, clientModel.HealthConditionId)
		if err != nil {
			WriteDeveloperError(w, http.StatusInternalServerError, "Error creating a record for the client layout:"+err.Error())
			return
		}
		clientModelVersionIds[i] = clientModelId
	}

	// update the active layouts to the new current set of layouts
	if err := l.DataApi.UpdatePatientActiveLayouts(modelId, clientModelVersionIds, healthConditionId); err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update patient layouts to be active: "+err.Error())
		return
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, SuccessfulGenericJSONResponse())
}
