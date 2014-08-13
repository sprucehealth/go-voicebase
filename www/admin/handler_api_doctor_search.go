package admin

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/www"
)

type doctorSearchAPIHandler struct {
	dataAPI api.DataAPI
}

func NewDoctorSearchAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&doctorSearchAPIHandler{
		dataAPI: dataAPI,
	}, []string{"GET"})
}

func (h *doctorSearchAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var results []*common.DoctorSearchResult

	query := r.FormValue("q")
	if query != "" {
		var err error
		results, err = h.dataAPI.SearchDoctors(query)
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}
	}

	www.JSONResponse(w, r, http.StatusOK, &struct {
		Results []*common.DoctorSearchResult `json:"results"`
	}{
		Results: results,
	})
}
