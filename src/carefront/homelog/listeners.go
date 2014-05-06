package homelog

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/common"
	"carefront/libs/dispatch"
	"carefront/libs/golog"
	"carefront/messages"
	patientApiService "carefront/patient"
	"errors"
	"fmt"
)

func InitListeners(dataAPI api.DataAPI) {
	dispatch.Default.Subscribe(func(ev *apiservice.VisitStartedEvent) error {
		// Insert an incomplete notification when a patient starts a visit
		_, err := dataAPI.InsertPatientNotification(ev.PatientId, &common.Notification{
			UID:             incompleteVisit,
			Dismissible:     false,
			DismissOnAction: false,
			Priority:        1000,
			Data: &incompleteVisitNotification{
				VisitId: ev.VisitId,
			},
		})
		return err
	})

	dispatch.Default.Subscribe(func(ev *apiservice.VisitSubmittedEvent) error {
		// Remove the incomplete visit notification when the patient submits a visit
		if err := dataAPI.DeletePatientNotificationByUID(ev.PatientId, incompleteVisit); err != nil {
			golog.Errorf("Failed to remove incomplete visit notification for patient %d: %s", ev.PatientId, err.Error())
		}

		doctor, err := dataAPI.GetDoctorFromId(ev.DoctorId)
		if err != nil {
			return err
		}

		// Add "visit submitted" to health log
		if _, err := dataAPI.InsertOrUpdatePatientHealthLogItem(ev.PatientId, &common.HealthLogItem{
			UID: fmt.Sprintf("visit_submitted:%d", ev.VisitId),
			Data: &titledLogItem{
				Title:    "Visit Submitted",
				Subtitle: fmt.Sprintf("With Dr. %s", doctor.LastName),
				IconURL:  "spruce:///image/icon_log_visit",
				TapURL:   fmt.Sprintf("spruce:///action/view_visit?visit_id=%d", ev.VisitId),
			},
		}); err != nil {
			golog.Errorf("Failed to insert visit submitted into health log for patient %d: %s", ev.PatientId, err.Error())
		}

		return nil
	})

	dispatch.Default.Subscribe(func(ev *apiservice.VisitReviewSubmittedEvent) error {
		doctor, err := dataAPI.GetDoctorFromId(ev.DoctorId)
		if err != nil {
			return err
		}

		// Add "treatment plan created" notification
		if _, err := dataAPI.InsertPatientNotification(ev.PatientId, &common.Notification{
			UID:             visitReviewed,
			Dismissible:     true,
			DismissOnAction: true,
			Priority:        1000,
			Data: &visitReviewedNotification{
				VisitId:  ev.VisitId,
				DoctorId: doctor.DoctorId.Int64(),
			},
		}); err != nil {
			golog.Errorf("Failed to insert treatment plan created into noficiation queue for patient %d: %s", ev.PatientId, err.Error())
		}

		// Add "treatment plan created" to health log
		if _, err := dataAPI.InsertOrUpdatePatientHealthLogItem(ev.PatientId, &common.HealthLogItem{
			UID: fmt.Sprintf("treatment_plan_created:%d", ev.TreatmentPlanId),
			Data: &titledLogItem{
				Title:    "Treatment Plan",
				Subtitle: fmt.Sprintf("Created By. %s", doctor.LastName),
				IconURL:  "spruce:///image/icon_log_treatment_plan",
				TapURL:   fmt.Sprintf("spruce:///action/view_treatment_plan?treatment_plan_id=%d", ev.TreatmentPlanId),
			},
		}); err != nil {
			golog.Errorf("Failed to insert visit treatment plan created into health log for patient %d: %s", ev.PatientId, err.Error())
		}

		return nil
	})

	dispatch.Default.Subscribe(func(ev *patientApiService.CareTeamAssingmentEvent) error {
		for _, a := range ev.Assignments {
			if a.ProviderRole == api.DOCTOR_ROLE {
				doctor, err := dataAPI.GetDoctorFromId(a.ProviderId)
				if err != nil {
					golog.Errorf("Failed to lookup doctor %d: %s", a.ProviderId, err.Error())
				} else {
					if _, err := dataAPI.InsertOrUpdatePatientHealthLogItem(ev.PatientId, &common.HealthLogItem{
						UID: fmt.Sprintf("doctor_added:%d", a.ProviderId),
						Data: &textLogItem{
							Text:    fmt.Sprintf("%s %s, M.D., added to your care team.", doctor.FirstName, doctor.LastName),
							IconURL: fmt.Sprintf("spruce:///image/thumbnail_care_team_%d", doctor.DoctorId.Int64()), // TODO
							TapURL:  "spruce:///action/view_care_team",
						},
					}); err != nil {
						golog.Errorf("Failed to insert visit treatment plan created into health log for patient %d: %s", ev.PatientId, err.Error())
					}
				}
			}
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *messages.ConversationStartedEvent) error {
		people, err := dataAPI.GetPeople([]int64{ev.FromId, ev.ToId})
		if err != nil {
			return err
		}
		from := people[ev.FromId]
		if from == nil {
			return errors.New("failed to find person conversation is from")
		}
		to := people[ev.ToId]
		if to == nil {
			return errors.New("failed to find person conversation is addressed to")
		}

		// TODO: currently only supporting patient notifications
		if to.RoleType != api.PATIENT_ROLE || from.RoleType != api.DOCTOR_ROLE {
			return nil
		}

		_, err = dataAPI.InsertPatientNotification(to.RoleId, &common.Notification{
			UID:             fmt.Sprintf("conversation:%d", ev.ConversationId),
			Dismissible:     true,
			DismissOnAction: true,
			Priority:        1000,
			Data: &newConversationNotification{
				ConversationId: ev.ConversationId,
				DoctorId:       from.RoleId,
			},
		})
		return err
	})

	dispatch.Default.Subscribe(func(ev *messages.ConversationReplyEvent) error {
		people, err := dataAPI.GetPeople([]int64{ev.FromId})
		if err != nil {
			return err
		}
		from := people[ev.FromId]
		if from == nil {
			return errors.New("failed to find person conversation is from")
		}

		// TODO: currently only supporting patient notifications
		if from.RoleType != api.DOCTOR_ROLE {
			return nil
		}

		con, err := dataAPI.GetConversation(ev.ConversationId)
		if err != nil {
			return err
		}
		// Find the patient participant
		for _, p := range con.Participants {
			if p.RoleType == api.PATIENT_ROLE {
				_, err = dataAPI.InsertPatientNotification(p.RoleId, &common.Notification{
					UID:             fmt.Sprintf("conversation:%d", ev.ConversationId),
					Dismissible:     true,
					DismissOnAction: true,
					Priority:        1000,
					Data: &conversationReplyNotification{
						ConversationId: ev.ConversationId,
						DoctorId:       from.RoleId,
					},
				})
				return err
			}
		}
		return nil
	})
}
