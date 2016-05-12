package doctor

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/erx"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/pharmacy"
	"golang.org/x/net/context"
)

type pharmacySearchHandler struct {
	dataAPI api.DataAPI
	erxAPI  erx.ERxAPI
}

func NewPharmacySearchHandler(dataAPI api.DataAPI, erxAPI erx.ERxAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(&pharmacySearchHandler{
				dataAPI: dataAPI,
				erxAPI:  erxAPI,
			}), api.RoleDoctor, api.RoleCC),
		httputil.Get)
}

type PharmacySearchRequestData struct {
	ZipcodeString string   `schema:"zipcode_string"`
	PharmacyTypes []string `schema:"pharmacy_types[]"`
}

type PharmacySearchResponse struct {
	PharmacyResults []*pharmacy.PharmacyData `json:"pharmacy_results"`
}

func (d *pharmacySearchHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestData := &PharmacySearchRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		apiservice.WriteValidationError(ctx, err.Error(), w, r)
		return
	}

	doctor, err := d.dataAPI.GetDoctorFromAccountID(apiservice.MustCtxAccount(ctx).ID)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	pharmacyResults, err := d.erxAPI.SearchForPharmacies(doctor.DoseSpotClinicianID, "", "", requestData.ZipcodeString, "", requestData.PharmacyTypes)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, &PharmacySearchResponse{PharmacyResults: pharmacyResults})
}
