package layout

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/info_intake"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type doctorLayoutHandler struct {
	dataApi api.DataAPI
	purpose string
}

func NewDoctorLayoutHandler(dataApi api.DataAPI, purpose string) *doctorLayoutHandler {
	return &doctorLayoutHandler{
		dataApi: dataApi,
		purpose: purpose,
	}
}

func (d *doctorLayoutHandler) NonAuthenticated() bool {
	return true
}

func (d *doctorLayoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != apiservice.HTTP_POST {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "unable to parse input parameters: "+err.Error())
		return
	}

	file, _, err := r.FormFile("layout")
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "No layout file or invalid layout file specified: "+err.Error())
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to read in layout file: "+err.Error())
		return
	}

	doctorIntakeLayout := info_intake.GetLayoutModelBasedOnPurpose(d.purpose)
	if err = json.Unmarshal(data, doctorIntakeLayout); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Error parsing layout file: "+err.Error())
		return
	}

	healthConditionTag := doctorIntakeLayout.GetHealthConditionTag()
	if healthConditionTag == "" {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "health condition not specified or invalid in layout")
		return
	}

	healthConditionId, err := d.dataApi.GetHealthConditionInfo(healthConditionTag)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "unable to get health condition id: "+err.Error())
		return
	}

	modelId, err := d.dataApi.CreateLayoutVersion(data, layout_syntax_version, healthConditionId, api.DOCTOR_ROLE, d.purpose, "automatically generated")
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to create record for layout : "+err.Error())
		return
	}

	if err := doctorIntakeLayout.FillInDatabaseInfo(d.dataApi, api.EN_LANGUAGE_ID); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "unable to fill database info into doctor layout: "+err.Error())
		return
	}

	jsonData, err := json.Marshal(doctorIntakeLayout)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to marshal doctor layout: "+err.Error())
		return
	}

	doctorLayoutId, err := d.dataApi.CreateDoctorLayout(jsonData, modelId, healthConditionId)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to record for doctor layout: "+err.Error())
		return
	}

	err = d.dataApi.UpdateDoctorActiveLayouts(modelId, doctorLayoutId, healthConditionId, d.purpose)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to mark record as active: "+err.Error())
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, apiservice.SuccessfulGenericJSONResponse())
}
