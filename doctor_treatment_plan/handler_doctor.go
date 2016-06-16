package doctor_treatment_plan

import (
	"fmt"
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/erx"
	"github.com/sprucehealth/backend/cmd/svc/restapi/mediastore"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/responses"
	"golang.org/x/net/context"
)

type doctorTreatmentPlanHandler struct {
	dataAPI         api.DataAPI
	mediaStore      *mediastore.Store
	erxAPI          erx.ERxAPI
	dispatcher      *dispatch.Dispatcher
	erxRoutingQueue *common.SQSQueue
	erxStatusQueue  *common.SQSQueue
	routeErx        bool
}

func NewDoctorTreatmentPlanHandler(
	dataAPI api.DataAPI,
	erxAPI erx.ERxAPI,
	mediaStore *mediastore.Store,
	dispatcher *dispatch.Dispatcher,
	erxRoutingQueue *common.SQSQueue,
	erxStatusQueue *common.SQSQueue,
	routeErx bool,
) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.RequestCacheHandler(
			apiservice.AuthorizationRequired(
				&doctorTreatmentPlanHandler{
					dataAPI:         dataAPI,
					erxAPI:          erxAPI,
					mediaStore:      mediaStore,
					dispatcher:      dispatcher,
					erxRoutingQueue: erxRoutingQueue,
					erxStatusQueue:  erxStatusQueue,
					routeErx:        routeErx,
				})),
		httputil.Get, httputil.Put, httputil.Post, httputil.Delete)
}

type TreatmentPlanRequestData struct {
	DoctorFavoriteTreatmentPlanID int64                              `json:"dr_favorite_treatment_plan_id,string" schema:"dr_favorite_treatment_plan_id"`
	TreatmentPlanID               int64                              `json:"treatment_plan_id,string" schema:"treatment_plan_id" `
	PatientVisitID                int64                              `json:"patient_visit_id,string" schema:"patient_visit_id" `
	Abridged                      bool                               `json:"abridged" schema:"abridged"`
	TPContentSource               *common.TreatmentPlanContentSource `json:"content_source"`
	TPParent                      *common.TreatmentPlanParent        `json:"parent"`
	Message                       string                             `json:"message"`
	Sections                      string                             `json:"sections,omitempty"`
}

type DoctorTreatmentPlanResponse struct {
	TreatmentPlan *responses.TreatmentPlan `json:"treatment_plan"`
}

func (d *doctorTreatmentPlanHandler) IsAuthorized(ctx context.Context, r *http.Request) (bool, error) {
	requestCache := apiservice.MustCtxCache(ctx)
	account := apiservice.MustCtxAccount(ctx)

	requestData := &TreatmentPlanRequestData{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error())
	}
	requestCache[apiservice.CKRequestData] = requestData

	doctorID, err := d.dataAPI.GetDoctorIDFromAccountID(account.ID)
	if err != nil {
		return false, err
	}
	requestCache[apiservice.CKDoctorID] = doctorID

	switch r.Method {
	case httputil.Get:
		if requestData.TreatmentPlanID == 0 {
			return false, apiservice.NewValidationError("treatment_plan_id must be specified")
		}

		treatmentPlan, err := d.dataAPI.GetAbridgedTreatmentPlan(requestData.TreatmentPlanID, doctorID)
		if err != nil {
			return false, err
		}
		requestCache[apiservice.CKTreatmentPlan] = treatmentPlan

		if err := apiservice.ValidateAccessToPatientCase(
			r.Method,
			account.Role,
			doctorID,
			treatmentPlan.PatientID,
			treatmentPlan.PatientCaseID.Int64(),
			d.dataAPI); err != nil {
			return false, err
		}

		// if we are dealing with a draft, and the owner of the treatment plan does not match the doctor requesting it,
		// return an error because this should never be the case
		if treatmentPlan.InDraftMode() && treatmentPlan.DoctorID.Int64() != doctorID {
			return false, apiservice.NewAccessForbiddenError()
		}

	case httputil.Put, httputil.Delete:
		if requestData.TreatmentPlanID == 0 {
			return false, apiservice.NewValidationError("treatment_plan_id must be specified")
		}

		treatmentPlan, err := d.dataAPI.GetAbridgedTreatmentPlan(requestData.TreatmentPlanID, doctorID)
		if err != nil {
			return false, err
		}
		requestCache[apiservice.CKTreatmentPlan] = treatmentPlan

		if err := apiservice.ValidateAccessToPatientCase(
			r.Method,
			account.Role,
			doctorID,
			treatmentPlan.PatientID,
			treatmentPlan.PatientCaseID.Int64(),
			d.dataAPI); err != nil {
			return false, err
		}

		// ensure that doctor is owner of the treatment plan
		// and that the treatment plan is in draft mode
		if doctorID != treatmentPlan.DoctorID.Int64() {
			return false, apiservice.NewAccessForbiddenError()
		}

	case httputil.Post:
		if requestData.TPParent == nil || requestData.TPParent.ParentID.Int64() == 0 {
			return false, apiservice.NewValidationError("parent_id must be specified")
		}

		patientVisitID := requestData.TPParent.ParentID.Int64()
		switch requestData.TPParent.ParentType {
		case common.TPParentTypeTreatmentPlan:
			// ensure that parent treatment plan is ACTIVE
			parentTreatmentPlan, err := d.dataAPI.GetAbridgedTreatmentPlan(requestData.TPParent.ParentID.Int64(), doctorID)
			if err != nil {
				return false, err
			} else if parentTreatmentPlan.Status != api.StatusActive {
				return false, apiservice.NewValidationError("parent treatment plan has to be ACTIVE")
			}

			patientVisitID, err = d.dataAPI.GetPatientVisitIDFromTreatmentPlanID(requestData.TPParent.ParentID.Int64())
			if err != nil {
				return false, err
			}
		case common.TPParentTypePatientVisit:
		default:
			return false, apiservice.NewValidationError("Expected the parent type to either by PATIENT_VISIT or TREATMENT_PLAN")
		}
		requestCache[apiservice.CKPatientVisitID] = patientVisitID

		patientCase, err := d.dataAPI.GetPatientCaseFromPatientVisitID(patientVisitID)
		if err != nil {
			return false, err
		}
		requestCache[apiservice.CKPatientCase] = patientCase

		if err := apiservice.ValidateAccessToPatientCase(
			r.Method,
			account.Role,
			doctorID,
			patientCase.PatientID,
			patientCase.ID.Int64(),
			d.dataAPI); err != nil {
			return false, err
		}

	default:
		return false, nil
	}

	return true, nil
}

func (d *doctorTreatmentPlanHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httputil.Get:
		d.getTreatmentPlan(ctx, w, r)
	case httputil.Post:
		d.pickATreatmentPlan(ctx, w, r)
	case httputil.Put:
		d.submitTreatmentPlan(ctx, w, r)
	case httputil.Delete:
		d.deleteTreatmentPlan(ctx, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (d *doctorTreatmentPlanHandler) deleteTreatmentPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	treatmentPlan := requestCache[apiservice.CKTreatmentPlan].(*common.TreatmentPlan)

	// Ensure treatment plan is a draft
	if !treatmentPlan.InDraftMode() {
		apiservice.WriteValidationError(ctx, "only draft treatment plan can be deleted", w, r)
		return
	}

	// Delete treatment plan
	if err := d.dataAPI.DeleteTreatmentPlan(treatmentPlan.ID.Int64()); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	apiservice.WriteJSONSuccess(w)
}

func (d *doctorTreatmentPlanHandler) submitTreatmentPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	requestData := requestCache[apiservice.CKRequestData].(*TreatmentPlanRequestData)
	treatmentPlan := requestCache[apiservice.CKTreatmentPlan].(*common.TreatmentPlan)

	note, err := d.dataAPI.GetTreatmentPlanNote(requestData.TreatmentPlanID)
	if err != nil && !api.IsErrNotFound(err) {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	if note == "" {
		apiservice.WriteValidationError(ctx, "Please include a personal note to the patient before submitting the treatment plan.", w, r)
		return
	}

	// replace any tokens in the note
	p := conc.NewParallel()

	var patient *common.Patient
	p.Go(func() error {
		var err error
		patient, err = d.dataAPI.Patient(treatmentPlan.PatientID, true)
		return err
	})

	var doctor *common.Doctor
	p.Go(func() error {
		var err error
		doctor, err = d.dataAPI.Doctor(treatmentPlan.DoctorID.Int64(), true)
		return err
	})

	if err := p.Wait(); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	t := newPatientDoctorTokenizer(patient, doctor)
	updatedNote, err := t.replace(note)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	if updatedNote != note {
		note = updatedNote

		// update the note in the database
		if err := d.dataAPI.SetTreatmentPlanNote(
			treatmentPlan.DoctorID.Int64(),
			requestData.TreatmentPlanID,
			updatedNote,
		); err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
	}

	var patientVisitID int64
	switch treatmentPlan.Parent.ParentType {
	case common.TPParentTypePatientVisit:
		patientVisitID = treatmentPlan.Parent.ParentID.Int64()
	case common.TPParentTypeTreatmentPlan:
		var err error
		patientVisitID, err = d.dataAPI.GetPatientVisitIDFromTreatmentPlanID(requestData.TreatmentPlanID)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}

		// if the parent of the treatment plan is a previous version of a treatment plan, ensure that it is an ACTIVE
		// treatment plan
		treatmentPlan, err := d.dataAPI.GetAbridgedTreatmentPlan(treatmentPlan.Parent.ParentID.Int64(), treatmentPlan.DoctorID.Int64())
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		} else if treatmentPlan.Status != api.StatusActive {
			apiservice.WriteValidationError(ctx, fmt.Sprintf("Expected the parent treatment plan to be in the active state but its in %s state", treatmentPlan.Status), w, r)
			return
		}

	default:
		apiservice.WriteValidationError(ctx, fmt.Sprintf("Parent of treatment plan is unexpected parent of type %s", treatmentPlan.Parent.ParentType), w, r)
		return
	}

	// mark the treatment plan as submitted
	status := common.TPStatusSubmitted
	if err := d.dataAPI.UpdateTreatmentPlan(treatmentPlan.ID.Int64(), &api.TreatmentPlanUpdate{
		Status: &status,
	}); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	d.dispatcher.Publish(&TreatmentPlanSubmittedEvent{
		VisitID:       patientVisitID,
		TreatmentPlan: treatmentPlan,
	})

	if d.routeErx {
		if err := apiservice.QueueUpJob(d.erxRoutingQueue, &erxRouteMessage{
			TreatmentPlanID: requestData.TreatmentPlanID,
			PatientID:       treatmentPlan.PatientID,
			DoctorID:        treatmentPlan.DoctorID.Int64(),
			Message:         note,
		}); err != nil {
			golog.Errorf("Failed to queue erx routing job: %s", err)
		}
	} else {
		if err := d.dataAPI.ActivateTreatmentPlan(treatmentPlan.ID.Int64(), treatmentPlan.DoctorID.Int64()); err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}

		doctor, err := d.dataAPI.GetDoctorFromID(treatmentPlan.DoctorID.Int64())
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}

		if err := sendCaseMessageAndPublishTPActivatedEvent(d.dataAPI, d.dispatcher, treatmentPlan, doctor, note); err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
	}

	apiservice.WriteJSONSuccess(w)
}

func (d *doctorTreatmentPlanHandler) getTreatmentPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	requestData := requestCache[apiservice.CKRequestData].(*TreatmentPlanRequestData)
	doctorID := requestCache[apiservice.CKDoctorID].(int64)
	treatmentPlan := requestCache[apiservice.CKTreatmentPlan].(*common.TreatmentPlan)
	account := apiservice.MustCtxAccount(ctx)

	// only return the small amount of information retreived about the treatment plan
	if requestData.Abridged {
		tpRes, err := responses.TransformTPToResponse(d.dataAPI, d.mediaStore, scheduledMessageMediaExpirationDuration, treatmentPlan, account.Role)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
		httputil.JSONResponse(w, http.StatusOK, &DoctorTreatmentPlanResponse{TreatmentPlan: tpRes})
		return
	}

	if err := populateTreatmentPlan(treatmentPlan, doctorID, d.dataAPI, parseSections(requestData.Sections)); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	tpRes, err := responses.TransformTPToResponse(d.dataAPI, d.mediaStore, scheduledMessageMediaExpirationDuration, treatmentPlan, account.Role)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	httputil.JSONResponse(w, http.StatusOK,
		&DoctorTreatmentPlanResponse{TreatmentPlan: tpRes})
}

func (d *doctorTreatmentPlanHandler) pickATreatmentPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	requestData := requestCache[apiservice.CKRequestData].(*TreatmentPlanRequestData)
	doctorID := requestCache[apiservice.CKDoctorID].(int64)
	patientVisitID := requestCache[apiservice.CKPatientVisitID].(int64)
	patientCase := requestCache[apiservice.CKPatientCase].(*common.PatientCase)
	account := apiservice.MustCtxAccount(ctx)
	if requestData.TPContentSource != nil {
		switch requestData.TPContentSource.Type {
		case common.TPContentSourceTypeFTP, common.TPContentSourceTypeTreatmentPlan:
		default:
			apiservice.WriteValidationError(ctx, "Invalid content source for treatment plan", w, r)
			return
		}
	}

	// NOTE: for now lets always determine the parent of the treatment plan to
	// be the patient visit context. The parent concept is not effectively used
	// in the doctor app and should be considered for removal in a future update.
	// One thing to note is that the doctor app uses it to show a button on the treatment plan
	// to indicate the patient visit or treatment plan to open. So we should ensure that removing
	// this functionality in the future would not break the client.
	tp := &common.TreatmentPlan{
		PatientID:     patientCase.PatientID,
		PatientCaseID: patientCase.ID,
		DoctorID:      encoding.DeprecatedNewObjectID(doctorID),
		Parent: &common.TreatmentPlanParent{
			ParentID:   encoding.DeprecatedNewObjectID(patientVisitID),
			ParentType: common.TPParentTypePatientVisit,
		},
		ContentSource: requestData.TPContentSource,
	}

	if err := copyContentSourceIntoTreatmentPlan(tp, d.dataAPI, doctorID); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	treatmentPlanID, err := d.dataAPI.StartNewTreatmentPlan(patientVisitID, tp)
	if err != nil {
		apiservice.WriteError(ctx, fmt.Errorf("Unable to start new treatment plan for patient visit: %s", err.Error()), w, r)
		return
	}

	// get the treatment plan just created so that it populates it with all the necessary metadata
	tp, err = d.dataAPI.GetAbridgedTreatmentPlan(treatmentPlanID, doctorID)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	if err := populateTreatmentPlan(tp, doctorID, d.dataAPI, AllSections); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	d.dispatcher.Publish(&NewTreatmentPlanStartedEvent{
		PatientID:       tp.PatientID,
		DoctorID:        doctorID,
		Case:            patientCase,
		CaseID:          tp.PatientCaseID.Int64(),
		VisitID:         patientVisitID,
		TreatmentPlanID: treatmentPlanID,
	})

	tpRes, err := responses.TransformTPToResponse(d.dataAPI, d.mediaStore, scheduledMessageMediaExpirationDuration, tp, account.Role)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	httputil.JSONResponse(w, http.StatusOK,
		&DoctorTreatmentPlanResponse{TreatmentPlan: tpRes})
}
