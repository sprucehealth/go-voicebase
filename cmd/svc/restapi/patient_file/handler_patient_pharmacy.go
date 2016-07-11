package patient_file

import (
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/pharmacy"
	"github.com/sprucehealth/backend/libs/httputil"
	"golang.org/x/net/context"
)

type doctorUpdatePatientPharmacyHandler struct {
	dataAPI api.DataAPI
}

func NewDoctorUpdatePatientPharmacyHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.RequestCacheHandler(
			apiservice.AuthorizationRequired(
				&doctorUpdatePatientPharmacyHandler{
					dataAPI: dataAPI,
				})),
		httputil.Put)
}

type DoctorUpdatePatientPharmacyRequestData struct {
	PatientID common.PatientID       `json:"patient_id"`
	Pharmacy  *pharmacy.PharmacyData `json:"pharmacy"`
}

func (d *doctorUpdatePatientPharmacyHandler) IsAuthorized(ctx context.Context, r *http.Request) (bool, error) {
	account := apiservice.MustCtxAccount(ctx)
	if account.Role != api.RoleDoctor {
		return false, apiservice.NewAccessForbiddenError()
	}

	requestCache := apiservice.MustCtxCache(ctx)

	requestData := &DoctorUpdatePatientPharmacyRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error())
	}
	requestCache[apiservice.CKRequestData] = requestData

	patient, err := d.dataAPI.GetPatientFromID(requestData.PatientID)
	if err != nil {
		return false, err
	}
	requestCache[apiservice.CKPatient] = patient

	doctor, err := d.dataAPI.GetDoctorFromAccountID(account.ID)
	if err != nil {
		return false, err
	}
	requestCache[apiservice.CKDoctor] = doctor

	if err := apiservice.ValidateDoctorAccessToPatientFile(r.Method, account.Role, doctor.ID.Int64(), patient.ID, d.dataAPI); err != nil {
		return false, err
	}

	return true, nil
}

func (d *doctorUpdatePatientPharmacyHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	patient := requestCache[apiservice.CKPatient].(*common.Patient)
	requestData := requestCache[apiservice.CKRequestData].(*DoctorUpdatePatientPharmacyRequestData)

	if err := d.dataAPI.UpdatePatientPharmacy(patient.ID, requestData.Pharmacy); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	apiservice.WriteJSONSuccess(w)
}