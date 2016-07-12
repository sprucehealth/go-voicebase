package patient_file

import (
	"errors"
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/libs/httputil"
)

type doctorPatientTreatmentsHandler struct {
	DataAPI api.DataAPI
}

func NewDoctorPatientTreatmentsHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.RequestCacheHandler(
			apiservice.AuthorizationRequired(
				&doctorPatientTreatmentsHandler{
					DataAPI: dataAPI,
				})),
		httputil.Get)
}

type requestData struct {
	PatientID common.PatientID `schema:"patient_id,required"`
}

type doctorPatientTreatmentsResponse struct {
	Treatments             []*common.Treatment         `json:"treatments,omitempty"`
	UnlinkedDNTFTreatments []*common.Treatment         `json:"unlinked_dntf_treatments,omitempty"`
	RefillRequests         []*common.RefillRequestItem `json:"refill_requests,omitempty"`
}

func (d *doctorPatientTreatmentsHandler) IsAuthorized(ctx context.Context, r *http.Request) (bool, error) {
	requestCache := apiservice.MustCtxCache(ctx)
	account := apiservice.MustCtxAccount(ctx)
	if account.Role != api.RoleDoctor {
		return false, apiservice.NewAccessForbiddenError()
	}

	requestData := &requestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error())
	}
	requestCache[apiservice.CKRequestData] = requestData

	currentDoctor, err := d.DataAPI.GetDoctorFromAccountID(account.ID)
	if err != nil {
		return false, err
	}
	requestCache[apiservice.CKDoctor] = currentDoctor

	patient, err := d.DataAPI.GetPatientFromID(requestData.PatientID)
	if err != nil {
		return false, err
	}
	requestCache[apiservice.CKPatient] = patient

	if err := apiservice.ValidateDoctorAccessToPatientFile(r.Method, account.Role, currentDoctor.ID.Int64(), patient.ID, d.DataAPI); err != nil {
		return false, err
	}

	return true, nil
}

func (d *doctorPatientTreatmentsHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	requestData := requestCache[apiservice.CKRequestData].(*requestData)

	treatments, err := d.DataAPI.GetTreatmentsForPatient(requestData.PatientID)
	if err != nil {
		apiservice.WriteError(ctx, errors.New("Unable to get treatments for patient: "+err.Error()), w, r)
		return
	}

	refillRequests, err := d.DataAPI.GetRefillRequestsForPatient(requestData.PatientID)
	if err != nil {
		apiservice.WriteError(ctx, errors.New("Unable to get refill requests for patient: "+err.Error()), w, r)
		return
	}

	unlinkedDNTFTreatments, err := d.DataAPI.GetUnlinkedDNTFTreatmentsForPatient(requestData.PatientID)
	if err != nil {
		apiservice.WriteError(ctx, errors.New("Unable to get unlinked dntf treatments for patient: "+err.Error()), w, r)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, &doctorPatientTreatmentsResponse{
		Treatments:             treatments,
		RefillRequests:         refillRequests,
		UnlinkedDNTFTreatments: unlinkedDNTFTreatments,
	})
}
