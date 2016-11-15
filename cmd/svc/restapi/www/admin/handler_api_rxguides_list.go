package admin

import (
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
)

type rxGuidesListAPIHandler struct {
	dataAPI api.DataAPI
}

func newRXGuideListAPIHandler(dataAPI api.DataAPI) http.Handler {
	return httputil.SupportedMethods(&rxGuidesListAPIHandler{
		dataAPI: dataAPI,
	}, httputil.Get, httputil.Put)
}

func (h *rxGuidesListAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		h.put(w, r)
		return
	}

	account := www.MustCtxAccount(r.Context())
	audit.LogAction(account.ID, "AdminAPI", "ListRXGuides", nil)

	drugs, err := h.dataAPI.ListDrugDetails()
	if err != nil {
		www.InternalServerError(w, r, err)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, drugs)
}

func (h *rxGuidesListAPIHandler) put(w http.ResponseWriter, r *http.Request) {
	account := www.MustCtxAccount(r.Context())
	audit.LogAction(account.ID, "AdminAPI", "ImportRXGuides", nil)

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	f, _, err := r.FormFile("csv")
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	defer f.Close()

	drugs := make(map[int]*common.DrugDetails)

	section := ""

	rd := csv.NewReader(f)
	for i := 0; ; i++ {
		row, err := rd.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		switch {
		case i == 0:
			for col, l := range row {
				if v := strings.TrimSpace(l); v != "" {
					drugs[col] = &common.DrugDetails{
						Name: v,
					}
				}
			}
		case row[0] == "Generic Name":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.GenericName = strings.TrimSpace(l)
				}
			}
		case row[0] == "Other Names":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.OtherNames = strings.TrimSpace(l)
				}
			}
		case row[0] == "NDC":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.NDC = strings.TrimSpace(l)
				}
			}
		case row[0] == "Image URL":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.ImageURL = strings.TrimSpace(l)
				}
			}
		case row[0] == "Description":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.Description = strings.TrimSpace(l)
				}
			}
		case row[0] == "Route":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.Route = strings.TrimSpace(l)
				}
			}
		case row[0] == "Form":
			for col, l := range row {
				if d := drugs[col]; d != nil {
					d.Form = strings.TrimSpace(l)
				}
			}
		case row[0] == "Comments":
			section = ""
		default:
			if row[0] != "" {
				section = row[0]
			} else if section == "" {
				// TODO: figure out what to do here. shouldn't happen
				continue
			}
			for col, l := range row {
				l = strings.TrimSpace(l)
				if d := drugs[col]; l != "" && d != nil {
					switch section {
					case "Do not take if...", "Warnings":
						d.Warnings = append(d.Warnings, l)
					case "Additional Instructions", "Tips":
						d.Tips = append(d.Tips, l)
					case "Common Side Effects":
						d.CommonSideEffects = append(d.CommonSideEffects, l)
					}
				}
			}
		}
	}

	var details []*common.DrugDetails
	for _, d := range drugs {
		details = append(details, d)
	}
	if err := h.dataAPI.SetDrugDetails(details); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, true)
}
