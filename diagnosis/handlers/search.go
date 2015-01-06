package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/diagnosis"
	"github.com/sprucehealth/backend/libs/httputil"
)

var (
	acneDiagnosisCodeIDs = []string{"diag_l700", "diag_l719", "diag_l710"}
	defaultMaxResults    = 50
)

type searchHandler struct {
	dataAPI      api.DataAPI
	diagnosisAPI diagnosis.API
}

type DiagnosisSearchResult struct {
	Sections []*ResultSection `json:"result_sections"`
}

type ResultSection struct {
	Title string        `json:"title"`
	Items []*ResultItem `json:"items"`
}

type ResultItem struct {
	Title     string             `json:"title"`
	Subtitle  string             `json:"subtitle,omitempty"`
	Diagnosis *AbridgedDiagnosis `json:"abridged_diagnosis"`
}

type AbridgedDiagnosis struct {
	CodeID     string `json:"code_id"`
	Code       string `json:"display_diagnosis_code"`
	Title      string `json:"title"`
	Synonyms   string `json:"synonyms,omitempty"`
	HasDetails bool   `json:"has_details"`
}

func NewSearchHandler(dataAPI api.DataAPI, diagnosisAPI diagnosis.API) http.Handler {
	return apiservice.SupportedRoles(
		httputil.SupportedMethods(
			apiservice.NoAuthorizationRequired(&searchHandler{
				dataAPI:      dataAPI,
				diagnosisAPI: diagnosisAPI,
			}), []string{"GET"}), []string{api.DOCTOR_ROLE})
}

func (s *searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	query := r.FormValue("query")

	// if the query is empty, return the common diagnoses set
	// pertaining to the chief complaint associated with the visit.
	// FIXME: For now always return the acne diagnoses set until we actually
	// have these groups created
	if len(query) == 0 {
		diagnosesMap, err := s.diagnosisAPI.DiagnosisForCodeIDs(acneDiagnosisCodeIDs)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		diagnosesList := make([]*diagnosis.Diagnosis, len(diagnosesMap))
		for i, codeID := range acneDiagnosisCodeIDs {
			diagnosesList[i] = diagnosesMap[codeID]
		}

		response, err := s.createResponseFromDiagnoses(diagnosesList, false, "Common Acne Diagnoses")
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}

		apiservice.WriteJSON(w, response)
		return
	}

	var maxResults int
	var err error

	// parse user specified numResults
	if numResults := r.FormValue("max_results"); numResults == "" {
		maxResults = defaultMaxResults
	} else if maxResults, err = strconv.Atoi(numResults); err != nil {
		apiservice.WriteValidationError(
			fmt.Sprintf("Invalid max_results parameter: %s", err.Error()), w, r)
		return
	}

	var diagnoses []*diagnosis.Diagnosis
	var queriedUsingDiagnosisCode bool

	// search for diagnoses by code if the query resembles a diagnosis code
	if resemblesCode(query) {
		diagnoses, err = s.diagnosisAPI.SearchDiagnosesByCode(query, maxResults)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
		queriedUsingDiagnosisCode = (len(diagnoses) > 0)
	} else if len(query) < 3 {
		apiservice.WriteJSON(w, &DiagnosisSearchResult{})
		return
	}

	// if no diagnoses found, then do a general search for diagnoses
	if len(diagnoses) == 0 {
		diagnoses, err = s.diagnosisAPI.SearchDiagnoses(query, maxResults)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	// if no diagnoses found yet, then fall back to fuzzy string matching
	if len(diagnoses) == 0 {
		diagnoses, err = s.diagnosisAPI.FuzzyTextSearchDiagnoses(query, maxResults)
		if err != nil {
			apiservice.WriteError(err, w, r)
			return
		}
	}

	var title string
	switch l := len(diagnoses); {
	case l == 0:
		title = "0 Results"
	case l > 1:
		title = strconv.Itoa(l) + " Results"
	case l == 1:
		title = "1 Result"
	}

	response, err := s.createResponseFromDiagnoses(diagnoses, queriedUsingDiagnosisCode, title)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSON(w, response)
}

func (s *searchHandler) createResponseFromDiagnoses(
	diagnoses []*diagnosis.Diagnosis,
	queriedUsingDignosisCode bool,
	sectionTitle string) (*DiagnosisSearchResult, error) {
	codeIDs := make([]string, len(diagnoses))
	for i, diagnosis := range diagnoses {
		codeIDs[i] = diagnosis.ID
	}

	// make requests in parallel to get indicators for any codeIDS
	// with additional details and synonyms for diagnoses
	wg := sync.WaitGroup{}
	wg.Add(2)
	errors := make(chan error, 2)

	var codeIDsWithDetails map[string]bool
	go func(codeIDs []string) {
		var err error
		codeIDsWithDetails, err = s.dataAPI.DiagnosesThatHaveDetails(codeIDs)
		if err != nil {
			errors <- err
		}
		wg.Done()
	}(codeIDs)

	var synonymMap map[string][]string
	go func(codeIDs []string) {
		var err error
		synonymMap, err = s.diagnosisAPI.SynonymsForDiagnoses(codeIDs)
		if err != nil {
			errors <- err
		}
		wg.Done()
	}(codeIDs)

	// wait for both calls to finish
	// and return any errors gathered
	wg.Wait()
	select {
	case err := <-errors:
		return nil, err
	default:
	}

	items := make([]*ResultItem, len(diagnoses))
	for i, diagnosis := range diagnoses {

		synonyms := strings.Join(synonymMap[diagnosis.ID], ", ")
		items[i] = &ResultItem{
			Title: diagnosis.Description,
			Diagnosis: &AbridgedDiagnosis{
				CodeID:     diagnosis.ID,
				Code:       diagnosis.Code,
				Title:      diagnosis.Description,
				Synonyms:   synonyms,
				HasDetails: codeIDsWithDetails[diagnosis.ID],
			},
		}

		// appropriately return the subtitle based on whether the
		// user queried using a diagnosis code or synonyms
		if queriedUsingDignosisCode {
			items[i].Subtitle = diagnosis.Code
		} else {
			items[i].Subtitle = synonyms
		}
	}

	return &DiagnosisSearchResult{
		Sections: []*ResultSection{
			{
				Title: sectionTitle,
				Items: items,
			},
		},
	}, nil
}
