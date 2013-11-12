package apiservice

import (
	"bytes"
	"carefront/api"
	"carefront/info_intake"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	CAREFRONT_LAYOUT_BUCKET        = "carefront-layout"
	CAREFRONT_CLIENT_LAYOUT_BUCKET = "carefront-client-layout"
	LAYOUT_SYNTAX_VERSION          = 1
)

type GenerateClientIntakeModelHandler struct {
	DataApi         api.DataAPI
	CloudStorageApi api.CloudStorageAPI
}

type ClientIntakeModelGeneratedResponse struct {
	ClientLayoutUrls []string `json:"clientModelUrls"`
}

type ClientIntakeModelErrorResponse struct {
	PhotoUploadErrorString string `json:"error"`
}

func (l *GenerateClientIntakeModelHandler) NonAuthenticated() bool {
	return true
}

func (l *GenerateClientIntakeModelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	file, handler, err := r.FormFile("layout")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
		return
	}

	// ensure that the file is a valid healthCondition layout, by trying to parse it
	// into the structure
	healthCondition := &info_intake.HealthCondition{}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
		return
	}

	err = json.Unmarshal(data, &healthCondition)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
		return
	}

	// determine the healthCondition tag so as to identify what healthCondition this layout belongs to
	healthConditionTag := healthCondition.HealthConditionTag
	if healthConditionTag == "" {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
		return
	}

	// check if the current active layout is the same as the layout trying to be uploaded
	currentActiveBucket, currentActiveKey, currentActiveRegion, _ := l.DataApi.GetActiveLayoutInfoForHealthCondition(healthConditionTag)
	if currentActiveBucket != "" {
		rawData, err := l.CloudStorageApi.GetObjectAtLocation(currentActiveBucket, currentActiveKey, currentActiveRegion)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
			return
		}
		res := bytes.Compare(data, rawData)
		// nothing to do if the layouts are exactly the same
		if res == 0 {
			WriteJSONToHTTPResponseWriter(w, ClientIntakeModelGeneratedResponse{nil})
			return
		}
	}

	// upload the layout version to S3 and get back an object storage id
	objectId, _, err := l.CloudStorageApi.PutObjectToLocation(CAREFRONT_LAYOUT_BUCKET,
		strconv.Itoa(int(time.Now().Unix())), api.US_EAST_1, handler.Header.Get("Content-Type"), data, time.Now().Add(10*time.Minute), l.DataApi)

	// get the healthConditionId
	healthConditionId, err := l.DataApi.GetHealthConditionInfo(healthConditionTag)

	// once that is successful, create a record for the layout version and mark is as CREATING
	modelId, err := l.DataApi.MarkNewLayoutVersionAsCreating(objectId, LAYOUT_SYNTAX_VERSION, healthConditionId, "automatically generated")

	// get all the supported languages
	_, supportedLanguageIds, err := l.DataApi.GetSupportedLanguages()

	// generate a client layout for each language
	clientIntakeModels := make(map[int64]*info_intake.HealthCondition)
	clientModelProcessor := &info_intake.HealthConditionIntakeModelProcessor{l.DataApi}
	clientModelVersionIds := make([]int64, len(supportedLanguageIds))
	clientModelUrls := make([]string, len(supportedLanguageIds))

	for i, supportedLanguageId := range supportedLanguageIds {
		clientModel := *healthCondition
		clientModelProcessor.FillInDetailsFromDatabase(&clientModel, supportedLanguageId)
		clientIntakeModels[supportedLanguageId] = &clientModel

		jsonData, err := json.Marshal(&clientModel)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
			return
		}
		// put each client layout that is generated into S3
		objectId, clientModelUrl, err := l.CloudStorageApi.PutObjectToLocation(CAREFRONT_CLIENT_LAYOUT_BUCKET,
			strconv.Itoa(int(time.Now().Unix())), api.US_EAST_1, handler.Header.Get("Content-Type"), jsonData, time.Now().Add(10*time.Minute), l.DataApi)
		clientModelUrls[i] = clientModelUrl
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
			return
		}

		// mark the client layout as creating until we have uploaded all client layouts before marking it as ACTIVE
		clientModelId, err := l.DataApi.MarkNewPatientLayoutVersionAsCreating(objectId, supportedLanguageId, modelId, clientModel.HealthConditionId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			WriteJSONToHTTPResponseWriter(w, ClientIntakeModelErrorResponse{err.Error()})
			return
		}
		clientModelVersionIds[i] = clientModelId
	}

	// update the active layouts to the new current set of layouts
	l.DataApi.UpdateActiveLayouts(modelId, clientModelVersionIds, 1)

	WriteJSONToHTTPResponseWriter(w, ClientIntakeModelGeneratedResponse{clientModelUrls})
}
