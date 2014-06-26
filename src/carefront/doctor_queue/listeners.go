package doctor_queue

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/app_worker"
	"carefront/common"
	"carefront/doctor_treatment_plan"
	"carefront/libs/dispatch"
	"carefront/libs/golog"
	"carefront/messages"
	"carefront/notify"
	"carefront/patient_file"
	"carefront/patient_visit"
	"errors"
)

func InitListeners(dataAPI api.DataAPI, notificationManager *notify.NotificationManager) {
	initJumpBallCaseQueueListeners(dataAPI)

	dispatch.Default.Subscribe(func(ev *patient_visit.VisitSubmittedEvent) error {
		// route the incoming visit to a doctor queue
		if err := routeIncomingPatientVisit(ev, dataAPI); err != nil {
			golog.Errorf("Unable to route incoming patient visit: %s", err)
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *patient_file.PatientVisitOpenedEvent) error {
		if err := dataAPI.UpdatePatientVisitStatus(ev.PatientVisit.PatientVisitId.Int64(), "", common.PVStatusReviewing); err != nil {
			return err
		}

		if err := dataAPI.MarkPatientVisitAsOngoingInDoctorQueue(ev.DoctorId, ev.PatientVisit.PatientVisitId.Int64()); err != nil {
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *doctor_treatment_plan.TreatmentPlanActivatedEvent) error {
		// mark the status on the visit in the doctor's queue to move it to the completed tab
		// so that the visit is no longer in the hands of the doctor
		err := dataAPI.MarkGenerationOfTreatmentPlanInVisitQueue(ev.DoctorId,
			ev.VisitId, ev.TreatmentPlanId, api.DQItemStatusOngoing, api.DQItemStatusTreated)
		if err != nil {
			golog.Errorf("Unable to update the status of the patient visit in the doctor queue: " + err.Error())
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *patient_visit.PatientVisitMarkedUnsuitableEvent) error {
		// mark the visit as complete once the doctor submits a diagnosis to indicate that the
		// patient was unsuitable for spruce
		if err := dataAPI.ReplaceItemInDoctorQueue(api.DoctorQueueItem{
			DoctorId:  ev.DoctorId,
			ItemId:    ev.PatientVisitId,
			EventType: api.DQEventTypePatientVisit,
			Status:    api.DQItemStatusTriaged,
		}, api.DQItemStatusOngoing); err != nil {
			golog.Errorf("Unable to insert transmission error resolved into doctor queue: %s", err)
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *app_worker.RxTransmissionErrorEvent) error {
		// Insert item into appropriate doctor queue to make them ever of an erx
		// that had issues being routed to pharmacy
		var eventTypeString string
		switch ev.EventType {
		case common.RefillRxType:
			eventTypeString = api.DQEventTypeRefillTransmissionError
		case common.UnlinkedDNTFTreatmentType:
			eventTypeString = api.DQEventTypeUnlinkedDNTFTransmissionError
		case common.ERxType:
			eventTypeString = api.DQEventTypeTransmissionError
		}
		if err := dataAPI.InsertItemIntoDoctorQueue(api.DoctorQueueItem{
			DoctorId:  ev.DoctorId,
			ItemId:    ev.ItemId,
			Status:    api.STATUS_PENDING,
			EventType: eventTypeString,
		}); err != nil {
			golog.Errorf("Unable to insert transmission error event into doctor queue: %s", err)
			return err
		}

		doctor, err := dataAPI.GetDoctorFromId(ev.DoctorId)
		if err != nil {
			golog.Errorf("Unable to get doctor from id: %s", err)
			return err
		}

		if err := notificationManager.NotifyDoctor(doctor, ev); err != nil {
			golog.Errorf("Unable to notify doctor: %s", err)
			return err
		}

		return nil
	})

	dispatch.Default.Subscribe(func(ev *apiservice.RxTransmissionErrorResolvedEvent) error {
		// Insert item into appropriate doctor queue to indicate resolution of transmission error
		var eventType string
		switch ev.EventType {
		case common.ERxType:
			eventType = api.DQEventTypeTransmissionError
		case common.RefillRxType:
			eventType = api.DQEventTypeRefillTransmissionError
		case common.UnlinkedDNTFTreatmentType:
			eventType = api.DQEventTypeUnlinkedDNTFTransmissionError
		}
		if err := dataAPI.ReplaceItemInDoctorQueue(api.DoctorQueueItem{
			DoctorId:  ev.DoctorId,
			ItemId:    ev.ItemId,
			EventType: eventType,
			Status:    api.DQItemStatusTreated,
		}, api.DQItemStatusPending); err != nil {
			golog.Errorf("Unable to insert transmission error resolved into doctor queue: %s", err)
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *app_worker.RefillRequestCreatedEvent) error {
		// insert refill item into doctor queue as a refill request
		if err := dataAPI.InsertItemIntoDoctorQueue(api.DoctorQueueItem{
			DoctorId:  ev.DoctorId,
			ItemId:    ev.RefillRequestId,
			EventType: api.DQEventTypeRefillRequest,
			Status:    api.STATUS_PENDING,
		}); err != nil {
			golog.Errorf("Unable to insert refill request item into doctor queue: %s", err)
			return err
		}

		doctor, err := dataAPI.GetDoctorFromId(ev.DoctorId)
		if err != nil {
			golog.Errorf("Unable to get doctor from id: %s", err)
			return err
		}

		if err := notificationManager.NotifyDoctor(doctor, ev); err != nil {
			golog.Errorf("Unable to notify doctor: %s", err)
			return err
		}

		return nil
	})

	dispatch.Default.Subscribe(func(ev *apiservice.RefillRequestResolvedEvent) error {
		// Move the queue item for the doctor from the ongoing to the completed state
		if err := dataAPI.ReplaceItemInDoctorQueue(api.DoctorQueueItem{
			DoctorId:  ev.DoctorId,
			ItemId:    ev.RefillRequestId,
			EventType: api.DQEventTypeRefillRequest,
			Status:    ev.Status,
		}, api.DQItemStatusPending); err != nil {
			golog.Errorf("Unable to insert refill request resolved error into doctor queue: %s", err)
			return err
		}
		return nil
	})

	dispatch.Default.Subscribe(func(ev *messages.PostEvent) error {
		// clear the item from the doctor's queue once they respond to a message
		if ev.Person.RoleType == api.DOCTOR_ROLE {
			if err := dataAPI.ReplaceItemInDoctorQueue(api.DoctorQueueItem{
				DoctorId:  ev.Person.RoleId,
				ItemId:    ev.Case.Id,
				EventType: api.DQEventTypeCaseMessage,
				Status:    api.DQItemStatusReplied,
			}, api.DQItemStatusPending); err != nil {
				golog.Errorf("Unable to replace item in doctor queue with a replied item: %s", err)
				return err
			}
			return nil
		}

		// only act on event if the message goes from patient->doctor
		if ev.Person.RoleType != api.PATIENT_ROLE {
			return nil
		}

		// TODO: Once possible, use doctor assigned to the case instead of
		// using the patient's primary.

		careTeam, err := dataAPI.GetCareTeamForPatient(ev.Person.RoleId)
		if err != nil {
			return err
		}
		if careTeam == nil {
			return errors.New("No care team assigned to patient")
		}

		var doctorID int64
		for _, a := range careTeam.Assignments {
			if a.ProviderRole == api.DOCTOR_ROLE {
				doctorID = a.ProviderId
				break
			}
		}
		if doctorID == 0 {
			// No doctor assigned to patient
			return errors.New("No doctor assigned to patient")
		}

		if err := dataAPI.ReplaceItemInDoctorQueue(api.DoctorQueueItem{
			DoctorId:  doctorID,
			ItemId:    ev.Message.CaseID,
			EventType: api.EVENT_TYPE_CASE_MESSAGE,
			Status:    api.QUEUE_ITEM_STATUS_PENDING,
		}, api.QUEUE_ITEM_STATUS_REPLIED); err != nil {
			golog.Errorf("Unable to insert conversation item into doctor queue: %s", err)
			return err
		}

		doctor, err := dataAPI.GetDoctorFromId(doctorID)
		if err != nil {
			return err
		}

		if err := notificationManager.NotifyDoctor(doctor, ev); err != nil {
			golog.Errorf("Unable to notify doctor: %s", err)
			return err
		}

		return nil
	})

	dispatch.Default.Subscribe(func(ev *messages.ReadEvent) error {
		// delete the item from the queue when the doctor marks the conversation
		// as being read
		if ev.Person.RoleType == api.DOCTOR_ROLE {
			if err := dataAPI.ReplaceItemInDoctorQueue(api.DoctorQueueItem{
				DoctorId:  ev.Person.Doctor.DoctorId.Int64(),
				ItemId:    ev.CaseID,
				EventType: api.DQEventTypeCasemessage,
				Status:    api.DQItemStatusRead,
			}, api.DQItemStatusPending); err != nil {
				golog.Errorf("Unable to replace item in doctor queue with a replied item: %s", err)
				return err
			}
		}
		return nil
	})
}
