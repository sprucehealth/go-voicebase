package treatment_plan

import (
	"fmt"
	"net/http"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
)

type treatmentPlanHandler struct {
	dataApi api.DataAPI
}

func NewTreatmentPlanHandler(dataApi api.DataAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.AuthorizationRequired(
				&treatmentPlanHandler{
					dataApi: dataApi,
				}), []string{api.PATIENT_ROLE, api.DOCTOR_ROLE}),
		[]string{"GET"})
}

type TreatmentPlanRequest struct {
	TreatmentPlanID int64 `schema:"treatment_plan_id"`
	PatientCaseId   int64 `schema:"case_id"`
}

type treatmentPlanViewsResponse struct {
	HeaderViews      []tpView `json:"header_views,omitempty"`
	TreatmentViews   []tpView `json:"treatment_views,omitempty"`
	InstructionViews []tpView `json:"instruction_views,omitempty"`
}

func (p *treatmentPlanHandler) IsAuthorized(r *http.Request) (bool, error) {
	ctxt := apiservice.GetContext(r)

	requestData := &TreatmentPlanRequest{}
	if err := apiservice.DecodeRequestData(requestData, r); err != nil {
		return false, apiservice.NewValidationError(err.Error(), r)
	}
	ctxt.RequestCache[apiservice.RequestData] = requestData

	switch ctxt.Role {
	case api.PATIENT_ROLE:
		if requestData.TreatmentPlanID == 0 && requestData.PatientCaseId == 0 {
			return false, apiservice.NewValidationError("either treatment_plan_id or patient_case_id must be specified", r)
		}

		patient, err := p.dataApi.GetPatientFromAccountId(ctxt.AccountId)
		if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.Patient] = patient

		var treatmentPlan *common.TreatmentPlan
		if requestData.TreatmentPlanID != 0 {
			treatmentPlan, err = p.dataApi.GetTreatmentPlanForPatient(patient.PatientId.Int64(), requestData.TreatmentPlanID)
		} else {
			treatmentPlan, err = p.dataApi.GetActiveTreatmentPlanForCase(requestData.PatientCaseId)
		}
		if err == api.NoRowsError {
			return false, apiservice.NewResourceNotFoundError("treatment plan not found", r)
		} else if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.TreatmentPlan] = treatmentPlan

		if treatmentPlan.PatientId != patient.PatientId.Int64() {
			return false, apiservice.NewAccessForbiddenError()
		}

		if !treatmentPlan.IsReadyForPatient() {
			return false, apiservice.NewResourceNotFoundError("Inactive/active treatment_plan not found", r)
		}

		doctor, err := p.dataApi.GetDoctorFromId(treatmentPlan.DoctorId.Int64())
		if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.Doctor] = doctor

	case api.DOCTOR_ROLE:
		if requestData.TreatmentPlanID == 0 {
			return false, apiservice.NewValidationError("treatment_plan_id must be specified", r)
		}

		doctor, err := p.dataApi.GetDoctorFromAccountId(ctxt.AccountId)
		if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.Doctor] = doctor

		patient, err := p.dataApi.GetPatientFromTreatmentPlanId(requestData.TreatmentPlanID)
		if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.Patient] = patient

		treatmentPlan, err := p.dataApi.GetTreatmentPlanForPatient(patient.PatientId.Int64(), requestData.TreatmentPlanID)
		if err == api.NoRowsError {
			return false, apiservice.NewResourceNotFoundError("treatment plan not found", r)
		} else if err != nil {
			return false, err
		}
		ctxt.RequestCache[apiservice.TreatmentPlan] = treatmentPlan

		if err = apiservice.ValidateAccessToPatientCase(r.Method, ctxt.Role, doctor.DoctorId.Int64(), patient.PatientId.Int64(),
			treatmentPlan.PatientCaseId.Int64(), p.dataApi); err != nil {
			return false, err
		}
	default:
		return false, apiservice.NewAccessForbiddenError()
	}
	return true, nil
}

func (p *treatmentPlanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctxt := apiservice.GetContext(r)
	doctor := ctxt.RequestCache[apiservice.Doctor].(*common.Doctor)
	patient := ctxt.RequestCache[apiservice.Patient].(*common.Patient)
	treatmentPlan := ctxt.RequestCache[apiservice.TreatmentPlan].(*common.TreatmentPlan)

	err := populateTreatmentPlan(p.dataApi, treatmentPlan)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	res, err := treatmentPlanResponse(p.dataApi, treatmentPlan, doctor, patient)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}
	apiservice.WriteJSON(w, res)
}

func treatmentPlanResponse(dataApi api.DataAPI, treatmentPlan *common.TreatmentPlan, doctor *common.Doctor, patient *common.Patient) (*treatmentPlanViewsResponse, error) {
	var headerViews, treatmentViews, instructionViews []tpView

	// HEADER VIEWS
	headerViews = append(headerViews,
		&tpHeroHeaderView{
			Title:           fmt.Sprintf("%s's\nTreatment Plan", patient.FirstName),
			Subtitle:        fmt.Sprintf("Created by %s", doctor.ShortDisplayName),
			CreatedDateText: fmt.Sprintf("on %s", treatmentPlan.CreationDate.Format("January 2, 2006")),
		})

	// TREATMENT VIEWS
	if len(treatmentPlan.TreatmentList.Treatments) > 0 {
		treatmentViews = append(treatmentViews, generateViewsForTreatments(treatmentPlan, doctor, dataApi, false)...)
		treatmentViews = append(treatmentViews,
			&tpCardView{
				Views: []tpView{
					&tpCardTitleView{
						Title: "Prescription Pickup",
					},
					&tpPharmacyView{
						Text:     "Your prescriptions should be ready soon. Call your pharmacy to confirm a pickup time.",
						Pharmacy: patient.Pharmacy,
					},
				},
			},
			&tpButtonFooterView{
				FooterText: fmt.Sprintf("If you have any questions about your treatment plan, message your care team."),
				ButtonText: "Send a Message",
				IconURL:    app_url.IconMessage,
				TapURL:     app_url.SendCaseMessageAction(treatmentPlan.PatientCaseId.Int64()),
			},
		)
	}

	// INSTRUCTION VIEWS
	if treatmentPlan.RegimenPlan != nil && len(treatmentPlan.RegimenPlan.Sections) > 0 {
		for _, regimenSection := range treatmentPlan.RegimenPlan.Sections {
			cView := &tpCardView{
				Views: []tpView{},
			}
			instructionViews = append(instructionViews, cView)

			cView.Views = append(cView.Views, &tpCardTitleView{
				Title:   regimenSection.Name,
				IconURL: app_url.IconRegimen.String(),
			})

			for i, regimenStep := range regimenSection.Steps {
				cView.Views = append(cView.Views, &tpListElementView{
					ElementStyle: numberedStyle,
					Number:       i + 1,
					Text:         regimenStep.Text,
				})
			}
		}
	}

	instructionViews = append(instructionViews, &tpButtonFooterView{
		FooterText:       "If you have any questions about your treatment plan, message your care team.",
		ButtonText:       "Send a Message",
		IconURL:          app_url.IconMessage,
		TapURL:           app_url.SendCaseMessageAction(treatmentPlan.PatientCaseId.Int64()),
		CenterFooterText: true,
	})

	for _, vContainer := range [][]tpView{headerViews, treatmentViews, instructionViews} {
		for _, v := range vContainer {
			if err := v.Validate(); err != nil {
				return nil, err
			}
		}
	}

	return &treatmentPlanViewsResponse{
		HeaderViews:      headerViews,
		TreatmentViews:   treatmentViews,
		InstructionViews: instructionViews,
	}, nil
}
