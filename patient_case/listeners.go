package patient_case

import (
	"fmt"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/app_event"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/messages"
	"github.com/sprucehealth/backend/patient_visit"
)

func InitListeners(dataAPI api.DataAPI) {
	dispatch.Default.Subscribe(func(ev *messages.PostEvent) error {

		// delete any pending visit submitted notifications for case
		if err := dataAPI.DeleteCaseNotification(CNVisitSubmitted); err != nil {
			golog.Errorf("Unable to delete case notifications: %s", err)
			return err
		}

		// insert notification into patient case if the doctor
		// sent the patient a message
		if ev.Person.RoleType == api.DOCTOR_ROLE {
			if err := dataAPI.InsertCaseNotification(&common.CaseNotification{
				PatientCaseId:    ev.Case.Id.Int64(),
				NotificationType: CNMessage,
				UID:              fmt.Sprintf("%s:%d", CNMessage, ev.Message.ID),
				Data: &messageNotification{
					MessageId:    ev.Message.ID,
					DoctorId:     ev.Person.Doctor.DoctorId.Int64(),
					DismissOnTap: true,
					CaseId:       ev.Message.CaseID,
				},
			}); err != nil {
				golog.Errorf("Unable to insert notification item for case: %s", err)
				return err
			}
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *doctor_treatment_plan.TreatmentPlanActivatedEvent) error {
		// delete any pending visit submitted notifications for case
		if err := dataAPI.DeleteCaseNotification(CNVisitSubmitted); err != nil {
			golog.Errorf("Unable to delete case notifications: %s", err)
			return err
		}

		// insert a notification into the patient case if the doctor activates a treatment plan
		if err := dataAPI.InsertCaseNotification(&common.CaseNotification{
			PatientCaseId:    ev.Message.CaseID,
			NotificationType: CNTreatmentPlan,
			UID:              fmt.Sprintf("%s:%d", CNTreatmentPlan, ev.TreatmentPlanId),
			Data: &treatmentPlanNotification{
				MessageId:       ev.Message.ID,
				DoctorId:        ev.DoctorId,
				TreatmentPlanId: ev.TreatmentPlanId,
				CaseId:          ev.Message.CaseID,
			},
		}); err != nil {
			golog.Errorf("Unable to insert notification item for case: %s", err)
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *patient_visit.VisitStartedEvent) error {
		if err := dataAPI.InsertCaseNotification(&common.CaseNotification{
			PatientCaseId:    ev.PatientCaseId,
			NotificationType: CNIncompleteVisit,
			UID:              CNIncompleteVisit,
			Data: &incompleteVisitNotification{
				PatientVisitId: ev.VisitId,
			},
		}); err != nil {
			golog.Errorf("Unable to insert notification item for case: %s", err)
			return err
		}

		return nil

	})

	dispatch.Default.Subscribe(func(ev *patient_visit.VisitSubmittedEvent) error {

		// delete the notification that indicates that the user still has to complete
		// the visit
		if err := dataAPI.DeleteCaseNotification(CNIncompleteVisit); err != nil {
			golog.Errorf("Unable to delete case notifications: %s", err)
			return err
		}

		if err := dataAPI.InsertCaseNotification(&common.CaseNotification{
			PatientCaseId:    ev.PatientCaseId,
			NotificationType: CNVisitSubmitted,
			UID:              CNVisitSubmitted,
			Data:             &visitSubmittedNotification{},
		}); err != nil {
			golog.Errorf("Unable to insert notification item for case: %s", err)
			return err
		}

		return nil
	})

	dispatch.Default.Subscribe(func(ev *app_event.AppEvent) error {

		// act on this event if it represents a patient having viewed a treatment plan
		if ev.Resource == "treatment_plan" && ev.Role == api.PATIENT_ROLE && ev.Action == app_event.ViewedAction {
			if err := dataAPI.DeleteCaseNotification(fmt.Sprintf("%s:%d", CNTreatmentPlan, ev.ResourceId)); err != nil {
				golog.Errorf("Unable to delete case notification: %s", err)
				return err
			}
		}

		// act on the event if it represents a patient having viewed a message
		if ev.Resource == "case_message" && ev.Role == api.PATIENT_ROLE && ev.Action == app_event.ViewedAction {
			if err := dataAPI.DeleteCaseNotification(fmt.Sprintf("%s:%d", CNMessage, ev.ResourceId)); err != nil {
				golog.Errorf("Unable to delete case notification: %s", err)
				return err
			}
		}

		return nil
	})

}
