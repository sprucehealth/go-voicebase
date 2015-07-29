package admin

import (
	"net/http"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/diagnosis"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/www"
)

type diagnosisSearchHandler struct {
	dataAPI      api.DataAPI
	diagnosisAPI diagnosis.API
}

type diagnosisSearchResultItem struct {
	Code        string `json:"code"`
	CodeID      string `json:"codeID"`
	Description string `json:"description"`
}

type diagnosisSearchResult struct {
	Results []*diagnosisSearchResultItem `json:"results"`
}

const (
	maxResults = 100
)

func newDiagnosisSearchHandler(dataAPI api.DataAPI, diagnosisAPI diagnosis.API) httputil.ContextHandler {
	return httputil.ContextSupportedMethods(&diagnosisSearchHandler{
		dataAPI:      dataAPI,
		diagnosisAPI: diagnosisAPI,
	}, httputil.Get)
}

func (d *diagnosisSearchHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var response diagnosisSearchResult

	query := r.FormValue("q")

	// account := www.MustCtxAccount(ctx)
	// audit.LogAction(account.ID, "AdminAPI", "DiagnosisSearch", map[string]interface{}{"query": query})

	if query == "" {
		httputil.JSONResponse(w, http.StatusOK, response)
		return
	}

	diagnoses, err := d.diagnosisAPI.SearchDiagnosesByCode(query, maxResults)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	} else if len(diagnoses) == 0 {
		www.APINotFound(w, r)
		return
	}

	response.Results = make([]*diagnosisSearchResultItem, len(diagnoses))
	for i, diagnosisItem := range diagnoses {
		response.Results[i] = &diagnosisSearchResultItem{
			Code:        diagnosisItem.Code,
			CodeID:      diagnosisItem.ID,
			Description: diagnosisItem.Description,
		}
	}

	httputil.JSONResponse(w, http.StatusOK, response)
}
