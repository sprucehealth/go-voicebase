package patient_visit

import (
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/info_intake"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/golog"
)

type diagnosePatientHandler struct {
	dataApi    api.DataAPI
	authApi    api.AuthAPI
	dispatcher *dispatch.Dispatcher
}

func NewDiagnosePatientHandler(dataApi api.DataAPI, authApi api.AuthAPI, dispatcher *dispatch.Dispatcher) *diagnosePatientHandler {
	cacheInfoForUnsuitableVisit(dataApi)
	return &diagnosePatientHandler{
		dataApi:    dataApi,
		authApi:    authApi,
		dispatcher: dispatcher,
	}
}

type GetDiagnosisResponse struct {
	DiagnosisLayout *info_intake.DiagnosisIntake `json:"diagnosis"`
}

type DiagnosePatientRequestData struct {
	PatientVisitId int64 `schema:"patient_visit_id,required"`
}

func (d *diagnosePatientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case apiservice.HTTP_GET:
		d.getDiagnosis(w, r)
	case apiservice.HTTP_POST:
		d.diagnosePatient(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (d *diagnosePatientHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)

	doctorId, err := d.dataApi.GetDoctorIdFromAccountId(ctxt.AccountId)
	if err != nil {
		return false, err
	}
	ctxt.RequestCache[apiservice.DoctorID] = doctorId

	switch r.Method {
	case apiservice.HTTP_GET:
		requestData := new(DiagnosePatientRequestData)
		if err := apiservice.DecodeRequestData(requestData, r); err != nil {
			return false, apiservice.NewValidationError(err.Error(), r)
		} else if requestData.PatientVisitId == 0 {
			return false, apiservice.NewValidationError("patient_id must be specified", r)
		}
		ctxt.RequestCache[apiservice.RequestData] = requestData

		patientVisit, err := d.dataApi.GetPatientVisitFromId(requestData.PatientVisitId)
		if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.PatientVisit] = patientVisit

		if err := apiservice.ValidateAccessToPatientCase(r.Method, ctxt.Role, doctorId, patientVisit.PatientId.Int64(), patientVisit.PatientCaseId.Int64(), d.dataApi); err != nil {
			return false, err
		}

		if ctxt.Role == api.MA_ROLE {
			// identify the doctor on the case to surface the diagnosis to the MA
			assignments, err := d.dataApi.GetActiveMembersOfCareTeamForCase(patientVisit.PatientCaseId.Int64(), false)
			if err != nil {
				return false, err
			}
			var doctorOnCase *common.Doctor
			for _, assignment := range assignments {
				if assignment.ProviderRole == api.DOCTOR_ROLE {
					doctorOnCase, err = d.dataApi.GetDoctorFromId(assignment.ProviderID)
					if err != nil {
						return false, err
					}
					ctxt.RequestCache[apiservice.DoctorID] = doctorOnCase.DoctorId.Int64()
					break
				}
			}

		}
	case apiservice.HTTP_POST:
		answerIntakeRequestBody := &apiservice.AnswerIntakeRequestBody{}
		if err := apiservice.DecodeRequestData(answerIntakeRequestBody, r); err != nil {
			return false, apiservice.NewValidationError(err.Error(), r)
		} else if answerIntakeRequestBody.PatientVisitId == 0 {
			return false, apiservice.NewValidationError("patient_visit_id must be specified", r)
		}
		ctxt.RequestCache[apiservice.RequestData] = answerIntakeRequestBody

		patientVisit, err := d.dataApi.GetPatientVisitFromId(answerIntakeRequestBody.PatientVisitId)
		if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.PatientVisit] = patientVisit

		if err := apiservice.ValidateAccessToPatientCase(r.Method, ctxt.Role, doctorId, patientVisit.PatientId.Int64(), patientVisit.PatientCaseId.Int64(), d.dataApi); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (d *diagnosePatientHandler) getDiagnosis(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	doctorId := ctxt.RequestCache[apiservice.DoctorID].(int64)
	patientVisit := ctxt.RequestCache[apiservice.PatientVisit].(*common.PatientVisit)

	diagnosisLayout, err := GetDiagnosisLayout(d.dataApi, patientVisit.PatientVisitId.Int64(), doctorId)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSONToHTTPResponseWriter(w, http.StatusOK, &GetDiagnosisResponse{DiagnosisLayout: diagnosisLayout})
}

func (d *diagnosePatientHandler) diagnosePatient(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	answerIntakeRequestBody := ctxt.RequestCache[apiservice.RequestData].(*apiservice.AnswerIntakeRequestBody)
	doctorId := ctxt.RequestCache[apiservice.DoctorID].(int64)
	patientVisit := ctxt.RequestCache[apiservice.PatientVisit].(*common.PatientVisit)
	if err := apiservice.EnsurePatientVisitInExpectedStatus(d.dataApi, answerIntakeRequestBody.PatientVisitId, common.PVStatusReviewing); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, err.Error())
		return
	}

	layoutVersionId, err := d.dataApi.GetLayoutVersionIdOfActiveDiagnosisLayout(api.HEALTH_CONDITION_ACNE_ID)
	if err != nil {
		apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to get the layout version id of the diagnosis layout "+err.Error())
		return
	}

	answersToStorePerQuestion := make(map[int64][]*common.AnswerIntake)
	for _, questionItem := range answerIntakeRequestBody.Questions {
		// enumerate the answers to store from the top level questions as well as the sub questions
		answersToStorePerQuestion[questionItem.QuestionId] = apiservice.PopulateAnswersToStoreForQuestion(api.DOCTOR_ROLE, questionItem, answerIntakeRequestBody.PatientVisitId, doctorId, layoutVersionId)
	}

	if err := d.dataApi.DeactivatePreviousDiagnosisForPatientVisit(answerIntakeRequestBody.PatientVisitId, doctorId); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	diagnosisIntake := &api.DiagnosisIntake{
		DoctorID:       doctorId,
		PatientVisitID: answerIntakeRequestBody.PatientVisitId,
		LVersionID:     layoutVersionId,
		Intake:         answersToStorePerQuestion,
	}

	if err := d.dataApi.StoreAnswersForQuestion(diagnosisIntake); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// check if the doctor diagnosed the patient's visit as being unsuitable for spruce
	unsuitableReason, wasMarkedUnsuitable := wasVisitMarkedUnsuitableForSpruce(answerIntakeRequestBody)
	if wasMarkedUnsuitable {
		err = d.dataApi.ClosePatientVisit(answerIntakeRequestBody.PatientVisitId, common.PVStatusTriaged)
		if err != nil {
			apiservice.WriteDeveloperError(w, http.StatusInternalServerError, "Unable to update the status of the visit to closed: "+err.Error())
			return
		}

		d.dispatcher.Publish(&PatientVisitMarkedUnsuitableEvent{
			DoctorID:       doctorId,
			PatientID:      patientVisit.PatientId.Int64(),
			CaseID:         patientVisit.PatientCaseId.Int64(),
			PatientVisitID: answerIntakeRequestBody.PatientVisitId,
			InternalReason: unsuitableReason,
		})

	} else {
		diagnosis := determineDiagnosisFromAnswers(answerIntakeRequestBody)

		if err := d.dataApi.UpdateDiagnosisForVisit(patientVisit.PatientVisitId.Int64(), doctorId, diagnosis); err != nil {
			golog.Errorf("Unable to update diagnosis for patient visit: %s", err)
		}

		d.dispatcher.Publish(&DiagnosisModifiedEvent{
			DoctorID:       doctorId,
			PatientID:      patientVisit.PatientId.Int64(),
			PatientVisitID: answerIntakeRequestBody.PatientVisitId,
			PatientCaseID:  patientVisit.PatientCaseId.Int64(),
			Diagnosis:      diagnosis,
		})
	}

	apiservice.WriteJSONSuccess(w)
}
