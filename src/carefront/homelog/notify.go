/*
Package homelog provides the implementation of the home feed notifications and log.
*/
package homelog

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/libs/dispatch"
	"fmt"

	"reflect"
)

const (
	bodyButton                   = "body_button"
	incompleteVisit              = "incomplete_visit"
	visitReviewed                = "visit_reviewed"
	patientNotificationNamespace = "patient_notification"
)

type notification interface {
	TypeName() string
	makeView(dataAPI api.DataAPI, patientId int64) (view, error)
}

type IncompleteVisitNotification struct {
	VisitId int64
}

type VisitReviewedNotification struct {
	DoctorId int64
	VisitId  int64
}

func (*IncompleteVisitNotification) TypeName() string {
	return incompleteVisit
}

func (*VisitReviewedNotification) TypeName() string {
	return visitReviewed
}

func (n *IncompleteVisitNotification) makeView(dataAPI api.DataAPI, patientId int64) (view, error) {
	patient, err := dataAPI.GetPatientFromId(patientId)
	if err != nil {
		return nil, err
	}
	doctor, err := apiservice.GetPrimaryDoctorInfoBasedOnPatient(dataAPI, patient, "")
	if err != nil {
		return nil, err
	}

	return &incompleteVisitView{
		Type:           patientNotificationNamespace + ":" + incompleteVisit,
		Title:          fmt.Sprintf("Complete your visit with Dr. %s.", doctor.LastName),
		IconURL:        fmt.Sprintf("spruce:///image/thumbnail_care_team_%d", doctor.DoctorId.Int64()), // TODO
		ButtonText:     "Continue Your Visit",
		TapURL:         fmt.Sprintf("spruce:///action/view_visit?visit_id=%d", n.VisitId),
		PatientVisitId: n.VisitId,
	}, nil
}

func (n *VisitReviewedNotification) makeView(dataAPI api.DataAPI, patientId int64) (view, error) {
	doctor, err := dataAPI.GetDoctorFromId(n.DoctorId)
	if err != nil {
		return nil, err
	}
	planID, err := dataAPI.GetActiveTreatmentPlanForPatientVisit(n.DoctorId, n.VisitId)
	if err != nil {
		return nil, err
	}
	tapURL := fmt.Sprintf("spruce:///action/view_treatment_plan?treatment_plan_id=%d", planID)
	return &bodyButtonView{
		Dismissible:       true,
		DismissOnAction:   true,
		Type:              patientNotificationNamespace + ":" + bodyButton,
		Title:             fmt.Sprintf("Dr. %s created your treatment plan.", doctor.LastName),
		IconURL:           fmt.Sprintf("spruce:///image/thumbnail_care_team_%d", doctor.DoctorId.Int64()), // TODO
		TapURL:            tapURL,
		BodyButtonIconURL: "spruce:///image/icon_blue_treatment_plan",
		BodyButtonText:    "Treatment Plan",
		BodyButtonTapURL:  tapURL,
	}, nil
}

var notifyTypes = map[string]reflect.Type{}

func init() {
	registerNotificationType(&IncompleteVisitNotification{})
	registerNotificationType(&VisitReviewedNotification{})
}

func registerNotificationType(n notification) {
	notifyTypes[n.TypeName()] = reflect.TypeOf(reflect.Indirect(reflect.ValueOf(n)))
}

func InitListeners(dataAPI api.DataAPI) {
	// Insert an incomplete notification when a patient starts a visit
	dispatch.Default.Subscribe(func(ev *apiservice.VisitStartedEvent) error {
		_, err := dataAPI.InsertHomeNotification(&common.HomeNotification{
			PatientId:       ev.PatientId,
			UID:             incompleteVisit,
			Dismissible:     false,
			DismissOnAction: false,
			Priority:        1000,
			Data: &IncompleteVisitNotification{
				VisitId: ev.VisitId,
			},
		})
		return err
	})

	// Remove the incomplete visit notification when the patient submits a visit
	dispatch.Default.Subscribe(func(ev *apiservice.VisitSubmittedEvent) error {
		return dataAPI.DeleteHomeNotificationByUID(ev.PatientId, incompleteVisit)
	})
}
