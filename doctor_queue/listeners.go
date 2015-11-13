package doctor_queue

import (
	"fmt"

	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/app_worker"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/cost"
	"github.com/sprucehealth/backend/doctor"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/messages"
	"github.com/sprucehealth/backend/notify"
	"github.com/sprucehealth/backend/patient_visit"
	"github.com/sprucehealth/backend/tagging"
)

const (
	UnsuitableTag         = "Unsuitable"
	caseAssignmentMessage = "A Spruce patient case has been assigned to you."
)

var PublicUnsuitableMessageEnabledDef = &cfg.ValueDef{
	Name:        "Unsuitable.Message.Public.Enabled",
	Description: "Enable or disable making the unsuitable for spruce message from the doctor public.",
	Type:        cfg.ValueTypeBool,
	Default:     true,
}

func InitListeners(dataAPI api.DataAPI, analyticsLogger analytics.Logger, dispatcher *dispatch.Dispatcher,
	notificationManager *notify.NotificationManager, statsRegistry metrics.Registry, jbcqMinutesThreshold int,
	customerSupportEmail string, cfgStore cfg.Store, taggingClient tagging.Client) {
	initJumpBallCaseQueueListeners(dataAPI, analyticsLogger, dispatcher, statsRegistry, jbcqMinutesThreshold)

	routeSuccess := metrics.NewCounter()
	routeFailure := metrics.NewCounter()
	statsRegistry.Add("route/success", routeSuccess)
	statsRegistry.Add("route/failure", routeFailure)

	// Register out server config for enabling public unsuitable messages
	cfgStore.Register(PublicUnsuitableMessageEnabledDef)

	dispatcher.Subscribe(func(ev *cost.VisitChargedEvent) error {
		// route the incoming visit to a doctor queue
		if err := routeIncomingPatientVisit(ev, dataAPI, notificationManager); err != nil {
			routeFailure.Inc(1)
			golog.Errorf("Unable to route incoming patient visit: %s", err)
			return err
		}
		routeSuccess.Inc(1)
		return nil
	})

	dispatcher.Subscribe(func(ev *doctor_treatment_plan.TreatmentPlanSubmittedEvent) error {
		assignments, err := dataAPI.GetActiveMembersOfCareTeamForCase(ev.TreatmentPlan.PatientCaseID.Int64(), false)
		if err != nil {
			golog.Errorf("Unable to get care team of patient case: %s", err)
			return err
		}

		var maID int64
		var doctorID int64
		for _, assignment := range assignments {
			switch assignment.ProviderRole {
			case api.RoleCC:
				maID = assignment.ProviderID
			case api.RoleDoctor:
				doctorID = assignment.ProviderID
			}
		}

		doctor, err := dataAPI.Doctor(doctorID, true)
		if err != nil {
			golog.Errorf("Doctor lookup failed: %s", err.Error())
		}

		patient, err := dataAPI.Patient(ev.TreatmentPlan.PatientID, true)
		if err != nil {
			golog.Errorf("Patient lookup failed: %s", err.Error())
		}

		if doctor != nil {
			patientCase, err := dataAPI.GetPatientCaseFromID(ev.TreatmentPlan.PatientCaseID.Int64())
			if err != nil {
				golog.Errorf("Unable to get case from id: %s", err)
				return err
			}

			if patient != nil {
				err := dataAPI.CompleteVisitOnTreatmentPlanGeneration(
					ev.TreatmentPlan.DoctorID.Int64(),
					ev.VisitID,
					ev.TreatmentPlan.ID.Int64(),
					[]*api.DoctorQueueUpdate{
						{
							Action: api.DQActionRemove,
							QueueItem: &api.DoctorQueueItem{
								DoctorID:  ev.TreatmentPlan.DoctorID.Int64(),
								EventType: api.DQEventTypeCaseAssignment,
								Status:    api.DQItemStatusPending,
								ItemID:    ev.TreatmentPlan.PatientCaseID.Int64(),
							},
						},
						{
							Action: api.DQActionInsert,
							QueueItem: &api.DoctorQueueItem{
								DoctorID:         ev.TreatmentPlan.DoctorID.Int64(),
								PatientID:        ev.TreatmentPlan.PatientID,
								ItemID:           ev.TreatmentPlan.ID.Int64(),
								EventType:        api.DQEventTypeTreatmentPlan,
								Status:           api.DQItemStatusTreated,
								Description:      fmt.Sprintf("%s completed treatment plan for %s %s", doctor.ShortDisplayName, patient.FirstName, patient.LastName),
								ShortDescription: fmt.Sprintf("Treatment plan by %s", doctor.ShortDisplayName),
								ActionURL:        app_url.ViewCompletedTreatmentPlanAction(ev.TreatmentPlan.PatientID, ev.TreatmentPlan.ID.Int64(), ev.TreatmentPlan.PatientCaseID.Int64()),
								Tags:             []string{patientCase.Name},
							},
						},
					},
				)
				if err != nil {
					golog.Errorf("Unable to update the status of the patient visit in the doctor queue: " + err.Error())
				}
			}
		}

		// also notify the MA if part of the care team that a treatment plan was generated by the doctor

		if maID > 0 {
			ma, err := dataAPI.Doctor(maID, true)
			if err != nil {
				golog.Errorf("Unable to get ma for patient case: %s", err)
				return err
			}

			if err := notificationManager.NotifyDoctor(
				api.RoleCC,
				ma.ID.Int64(),
				ma.AccountID.Int64(), &notify.Message{
					ShortMessage: "A treatment plan was created for a patient.",
				}); err != nil {
				golog.Errorf("Unable to notify the ma of the treatment plan generation: %s", err)
			}

		}

		return nil
	})

	dispatcher.Subscribe(func(ev *patient_visit.PatientVisitMarkedUnsuitableEvent) error {

		patient, err := dataAPI.Patient(ev.PatientID, true)
		if err != nil {
			golog.Errorf("Unable to get patient info: %s", err.Error())
			return err
		}

		patientCase, err := dataAPI.GetPatientCaseFromID(ev.CaseID)
		if err != nil {
			golog.Errorf("Unable to get case from id: %s", err.Error())
			return err
		}

		doctor, err := dataAPI.GetDoctorFromID(ev.DoctorID)
		if err != nil {
			golog.Errorf("Unable to get doctor from id: %s", err)
			routeFailure.Inc(1)
			return err
		}

		conc.Go(func() {
			if err := tagging.ApplyCaseTag(taggingClient, UnsuitableTag, ev.CaseID, nil, tagging.TONone); err != nil {
				golog.Errorf("Unable to tag case as unsuitable: %s", err)
			}
		})

		// mark the visit as complete once the doctor submits a diagnosis to indicate that the
		// patient was unsuitable for spruce
		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action:       api.DQActionReplace,
				CurrentState: api.DQItemStatusOngoing,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:  ev.DoctorID,
					PatientID: ev.PatientID,
					ItemID:    ev.PatientVisitID,
					EventType: api.DQEventTypePatientVisit,
					Status:    api.DQItemStatusTriaged,
					Description: fmt.Sprintf("%s completed and triaged visit for %s %s",
						doctor.ShortDisplayName, patient.FirstName, patient.LastName),
					ShortDescription: fmt.Sprintf("Visit triaged by %s", doctor.ShortDisplayName),
					ActionURL:        app_url.ViewPatientVisitInfoAction(ev.PatientID, ev.PatientVisitID, ev.CaseID),
					Tags:             []string{patientCase.Name},
				},
			},
		}); err != nil {
			golog.Errorf("Unable to insert transmission error resolved into doctor queue: %s", err)
			routeFailure.Inc(1)
			return err
		}

		// assign the case to the MA
		assignments, err := dataAPI.GetActiveMembersOfCareTeamForCase(ev.CaseID, false)
		if err != nil {
			routeFailure.Inc(1)
			golog.Errorf("Unable to get active members of care team for case: %s", err)
			return err
		}

		var maID int64
		for _, assignment := range assignments {
			if assignment.ProviderRole == api.RoleCC {
				maID = assignment.ProviderID
				break
			}
		}

		ma, err := dataAPI.GetDoctorFromID(maID)
		if err != nil {
			golog.Errorf("Unable to get MA from id: %s", err)
			routeFailure.Inc(1)
			return err
		}

		public := cfgStore.Snapshot().Bool(PublicUnsuitableMessageEnabledDef.Name)
		message := &common.CaseMessage{
			CaseID:    ev.CaseID,
			PersonID:  doctor.PersonID,
			Body:      ev.Reason,
			IsPrivate: !public, // Utilize the server config to make the unsuitable message public or private
			EventText: fmt.Sprintf("assigned to %s", ma.LongDisplayName),
		}

		if _, err := dataAPI.CreateCaseMessage(message); err != nil {
			golog.Errorf("Unable to create message to assign case to MA: %s", err)
			routeFailure.Inc(1)
			return err
		}

		if public {
			people, err := dataAPI.GetPeople([]int64{doctor.PersonID})
			if err != nil {
				golog.Errorf("Unable to get doctor person object to publish PostEvent: %s", err)
				routeFailure.Inc(1)
				return err
			}
			postEvent := &messages.PostEvent{
				Message: message,
				Person:  people[doctor.PersonID],
				Case:    patientCase,
			}

			dispatcher.PublishAsync(postEvent)
		}

		// insert a pending item into the MA's queue
		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action: api.DQActionInsert,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         ma.ID.Int64(),
					PatientID:        ev.PatientID,
					ItemID:           ev.CaseID,
					EventType:        api.DQEventTypeCaseAssignment,
					Status:           api.DQItemStatusPending,
					Description:      fmt.Sprintf("%s %s's case assigned to %s", patient.FirstName, patient.LastName, ma.ShortDisplayName),
					ShortDescription: fmt.Sprintf("Reassigned by %s", doctor.ShortDisplayName),
					ActionURL:        app_url.ViewPatientMessagesAction(ev.PatientID, ev.CaseID),
					Tags:             []string{patientCase.Name},
				},
			},
		}); err != nil {
			golog.Errorf("Unable to insert case assignment item into doctor queue: %s", err)
			routeFailure.Inc(1)
			return err
		}

		// notify the ma of the case assignment
		if err := notificationManager.NotifyDoctor(
			api.RoleCC,
			ma.ID.Int64(),
			ma.AccountID.Int64(), &notify.Message{
				ShortMessage: caseAssignmentMessage,
			}); err != nil {
			golog.Errorf("Unable to notify assigned provider of event %T: %s", ev, err)
			routeFailure.Inc(1)
			return err
		}

		routeSuccess.Inc(1)
		return nil
	})

	dispatcher.Subscribe(func(ev *app_worker.RxTransmissionErrorEvent) error {

		// Insert item into appropriate doctor queue to make them ever of an erx
		// that had issues being routed to pharmacy
		var eventTypeString string
		var actionURL *app_url.SpruceAction
		switch ev.EventType {
		case common.RefillRxType:
			eventTypeString = api.DQEventTypeRefillTransmissionError
			actionURL = app_url.ViewRefillRequestAction(ev.Patient.ID, ev.ItemID)
		case common.UnlinkedDNTFTreatmentType:
			eventTypeString = api.DQEventTypeUnlinkedDNTFTransmissionError
			actionURL = app_url.ViewDNTFTransmissionErrorAction(ev.Patient.ID, ev.ItemID)
		case common.ERxType:
			eventTypeString = api.DQEventTypeTransmissionError
			actionURL = app_url.ViewTransmissionErrorAction(ev.Patient.ID, ev.ItemID)
		}

		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action: api.DQActionInsert,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         ev.ProviderID,
					PatientID:        ev.Patient.ID,
					ItemID:           ev.ItemID,
					Status:           api.StatusPending,
					EventType:        eventTypeString,
					Description:      fmt.Sprintf("Error sending prescription for %s %s", ev.Patient.FirstName, ev.Patient.LastName),
					ShortDescription: "Prescription error",
					ActionURL:        actionURL,
				},
			},
		}); err != nil {
			routeFailure.Inc(1)
			golog.Errorf("Unable to insert transmission error event into doctor queue: %s", err)
			return err
		}
		routeSuccess.Inc(1)

		doctor, err := dataAPI.GetDoctorFromID(ev.ProviderID)
		if err != nil {
			golog.Errorf("Unable to get doctor from id: %s", err)
			return err
		}

		if err := notificationManager.NotifyDoctor(
			ev.ProviderRole,
			doctor.ID.Int64(),
			doctor.AccountID.Int64(),
			&notify.Message{
				ShortMessage: "There was an error routing prescriptions to a pharmacy on Spruce.",
			}); err != nil {
			golog.Errorf("Unable to notify doctor: %s", err)
			return err
		}

		return nil
	})

	dispatcher.Subscribe(func(ev *doctor.RxTransmissionErrorResolvedEvent) error {
		// Insert item into appropriate doctor queue to indicate resolution of transmission error
		var eventType string
		var actionURL *app_url.SpruceAction
		var description, shortDescription string
		switch ev.EventType {
		case common.ERxType:
			eventType = api.DQEventTypeTransmissionError
			description = fmt.Sprintf("%s resolved error for %s %s", ev.Doctor.ShortDisplayName, ev.Patient.FirstName, ev.Patient.LastName)
			shortDescription = fmt.Sprintf("Prescription error resolved by %s", ev.Doctor.ShortDisplayName)
			actionURL = app_url.ViewTransmissionErrorAction(ev.Patient.ID, ev.ItemID)
		case common.RefillRxType:
			eventType = api.DQEventTypeRefillTransmissionError
			description = fmt.Sprintf("%s resolved refill request error for %s %s", ev.Doctor.ShortDisplayName, ev.Patient.FirstName, ev.Patient.LastName)
			shortDescription = fmt.Sprintf("Refill request error resolved by %s", ev.Doctor.ShortDisplayName)
			actionURL = app_url.ViewRefillRequestAction(ev.Patient.ID, ev.ItemID)
		case common.UnlinkedDNTFTreatmentType:
			eventType = api.DQEventTypeUnlinkedDNTFTransmissionError
			description = fmt.Sprintf("%s resolved error for %s %s", ev.Doctor.ShortDisplayName, ev.Patient.FirstName, ev.Patient.LastName)
			shortDescription = fmt.Sprintf("Prescription error resolved by %s", ev.Doctor.ShortDisplayName)
			actionURL = app_url.ViewDNTFTransmissionErrorAction(ev.Patient.ID, ev.ItemID)
		}

		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action:       api.DQActionReplace,
				CurrentState: api.DQItemStatusPending,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         ev.Doctor.ID.Int64(),
					PatientID:        ev.Patient.ID,
					ItemID:           ev.ItemID,
					EventType:        eventType,
					Status:           api.DQItemStatusTreated,
					Description:      description,
					ShortDescription: shortDescription,
					ActionURL:        actionURL,
				},
			},
		}); err != nil {
			golog.Errorf("Unable to insert transmission error resolved into doctor queue: %s", err)
			return err
		}
		return nil
	})

	dispatcher.Subscribe(func(ev *app_worker.RefillRequestCreatedEvent) error {
		// insert refill item into doctor queue as a refill request
		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action: api.DQActionInsert,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         ev.DoctorID,
					PatientID:        ev.Patient.ID,
					ItemID:           ev.RefillRequestID,
					EventType:        api.DQEventTypeRefillRequest,
					Status:           api.StatusPending,
					Description:      fmt.Sprintf("Refill request for %s %s", ev.Patient.FirstName, ev.Patient.LastName),
					ShortDescription: "Refill request",
					ActionURL:        app_url.ViewRefillRequestAction(ev.Patient.ID, ev.RefillRequestID),
				},
			},
		}); err != nil {
			routeFailure.Inc(1)
			golog.Errorf("Unable to insert refill request item into doctor queue: %s", err)
			return err
		}
		routeSuccess.Inc(1)

		doctor, err := dataAPI.GetDoctorFromID(ev.DoctorID)
		if err != nil {
			golog.Errorf("Unable to get doctor from id: %s", err)
			return err
		}

		if err := notificationManager.NotifyDoctor(
			api.RoleDoctor,
			doctor.ID.Int64(),
			doctor.AccountID.Int64(),
			&notify.Message{
				ShortMessage: "You have a new refill request from a Spruce patient.",
			}); err != nil {
			golog.Errorf("Unable to notify doctor: %s", err)
			return err
		}

		return nil
	})

	dispatcher.Subscribe(func(ev *doctor.RefillRequestResolvedEvent) error {

		var description, shortDescription string

		switch ev.Status {
		case api.DQItemStatusRefillApproved:
			description = fmt.Sprintf("%s approved refill request for %s %s",
				ev.Doctor.ShortDisplayName, ev.Patient.FirstName, ev.Patient.LastName)
			shortDescription = fmt.Sprintf("Refill request approved by %s", ev.Doctor.ShortDisplayName)
		case api.DQItemStatusRefillDenied:
			description = fmt.Sprintf("%s denied refill request for %s %s",
				ev.Doctor.ShortDisplayName, ev.Patient.FirstName, ev.Patient.LastName)
			shortDescription = fmt.Sprintf("Refill request denied by %s", ev.Doctor.ShortDisplayName)
		}

		// Move the queue item for the doctor from the ongoing to the completed state
		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action:       api.DQActionReplace,
				CurrentState: api.DQItemStatusPending,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         ev.Doctor.ID.Int64(),
					PatientID:        ev.Patient.ID,
					ItemID:           ev.RefillRequestID,
					EventType:        api.DQEventTypeRefillRequest,
					Status:           ev.Status,
					Description:      description,
					ShortDescription: shortDescription,
					ActionURL:        app_url.ViewRefillRequestAction(ev.Patient.ID, ev.RefillRequestID),
				},
			},
		}); err != nil {
			golog.Errorf("Unable to insert refill request resolved error into doctor queue: %s", err)
			return err
		}
		return nil
	})

	dispatcher.Subscribe(func(ev *messages.PostEvent) error {
		assignments, err := dataAPI.GetActiveMembersOfCareTeamForCase(ev.Case.ID.Int64(), false)
		if err != nil {
			golog.Errorf("Unable to get doctors assignend to case: %s", err)
			return nil
		}

		var maID int64
		for _, assignment := range assignments {
			switch assignment.Status {
			case api.StatusActive:
				if assignment.ProviderRole == api.RoleCC {
					maID = assignment.ProviderID
				}
			}
		}

		patient, err := dataAPI.Patient(ev.Case.PatientID, true)
		if err != nil {
			golog.Errorf("Patient lookup failed: %s", err.Error())
			return nil
		}

		switch ev.Person.RoleType {
		case api.RoleDoctor, api.RoleCC:
			// don't touch doctor or cc queue
			// if the message sent was autmomated
			if ev.IsAutomated {
				return nil
			}

			doctor, err := dataAPI.Doctor(ev.Person.RoleID, true)
			if err != nil {
				golog.Errorf("Doctor lookup failed for doctorID %d : %s", ev.Person.RoleID, err.Error())
				return nil
			}

			// clear the item from the doctor's queue once they respond to a message
			// delete any pending case assignments since the doctor/MA has passed the case onto the other party.
			if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
				{
					Action: api.DQActionRemove,
					QueueItem: &api.DoctorQueueItem{
						DoctorID:  ev.Person.RoleID,
						ItemID:    ev.Case.ID.Int64(),
						EventType: api.DQEventTypeCaseAssignment,
						Status:    api.DQItemStatusPending,
					},
				},
				{
					Action: api.DQActionRemove,
					QueueItem: &api.DoctorQueueItem{
						DoctorID:  ev.Person.RoleID,
						ItemID:    ev.Case.ID.Int64(),
						EventType: api.DQEventTypeCaseMessage,
						Status:    api.DQItemStatusPending,
					},
				},
				{
					Action: api.DQActionInsert,
					QueueItem: &api.DoctorQueueItem{
						DoctorID:         ev.Person.RoleID,
						PatientID:        ev.Case.PatientID,
						ItemID:           ev.Case.ID.Int64(),
						EventType:        api.DQEventTypeCaseMessage,
						Status:           api.DQItemStatusReplied,
						Description:      fmt.Sprintf("%s replied to %s %s", doctor.ShortDisplayName, patient.FirstName, patient.LastName),
						ShortDescription: fmt.Sprintf("Messaged by %s", doctor.ShortDisplayName),
						ActionURL:        app_url.ViewPatientMessagesAction(patient.ID, ev.Case.ID.Int64()),
						Tags:             []string{ev.Case.Name},
					},
				},
			}); err != nil {
				golog.Errorf("Unable to replace item in doctor queue with a replied item: %s", err)
				return err
			}

			return nil
		}

		// only act on event if the message goes from patient->doctor
		if ev.Person.RoleType != api.RolePatient {
			return nil
		}

		// send the patient message to the MA, or to the doctor if the MA doesn't exist
		if maID == 0 {
			// No doctor or ma assigned to patient
			return errors.New("No ma assigned to patient case")
		}

		ma, err := dataAPI.Doctor(maID, true)
		if err != nil {
			golog.Errorf("Unable to get doctor from id: %d", maID)
			return err
		}

		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action: api.DQActionInsert,
				Dedupe: true,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         maID,
					PatientID:        ev.Case.PatientID,
					ItemID:           ev.Message.CaseID,
					EventType:        api.DQEventTypeCaseMessage,
					Status:           api.DQItemStatusPending,
					Description:      fmt.Sprintf("Message from %s %s", patient.FirstName, patient.LastName),
					ShortDescription: "New message",
					ActionURL:        app_url.ViewPatientMessagesAction(patient.ID, ev.Case.ID.Int64()),
					Tags:             []string{ev.Case.Name},
				},
			},
		}); err != nil {
			routeFailure.Inc(1)
			golog.Errorf("Unable to insert conversation item into doctor queue: %s", err)
			return err
		}
		routeSuccess.Inc(1)

		if err := notificationManager.NotifyDoctor(
			api.RoleCC,
			ma.ID.Int64(),
			ma.AccountID.Int64(),
			&notify.Message{
				ShortMessage: "You have a new message on Spruce.",
			}); err != nil {
			golog.Errorf("Unable to notify doctor: %s", err)
			return err
		}

		return nil
	})

	dispatcher.Subscribe(func(ev *messages.CaseAssignEvent) error {

		// identify the provider the case is being assigned to
		var assignedProvider *common.Doctor
		var assigneeProvider *common.Doctor
		var assignedProviderRole string
		if ev.Person.RoleType == api.RoleDoctor {
			assignedProvider = ev.MA
			assigneeProvider = ev.Doctor
			assignedProviderRole = api.RoleCC
		} else {
			assignedProvider = ev.Doctor
			assigneeProvider = ev.MA
			assignedProviderRole = api.RoleDoctor
		}

		patient, err := dataAPI.Patient(ev.Case.PatientID, true)
		if err != nil {
			golog.Errorf("Unable to get patient info: %s", err.Error())
			return err
		}

		// create an item in the history tab for the provider assigning the case
		// also delete an pending case assignments given that the doctor just handled the assignment.
		if !ev.IsAutomated {
			if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
				{
					Action: api.DQActionRemove,
					QueueItem: &api.DoctorQueueItem{
						DoctorID:  ev.Person.RoleID,
						ItemID:    ev.Case.ID.Int64(),
						EventType: api.DQEventTypeCaseAssignment,
						Status:    api.DQItemStatusPending,
					},
				},
				{
					Action: api.DQActionRemove,
					QueueItem: &api.DoctorQueueItem{
						DoctorID:  ev.Person.RoleID,
						ItemID:    ev.Case.ID.Int64(),
						EventType: api.DQEventTypeCaseMessage,
						Status:    api.DQItemStatusPending,
					},
				},
				{
					Action: api.DQActionInsert,
					QueueItem: &api.DoctorQueueItem{
						DoctorID:  ev.Person.RoleID,
						PatientID: ev.Case.PatientID,
						ItemID:    ev.Case.ID.Int64(),
						EventType: api.DQEventTypeCaseAssignment,
						Status:    api.DQItemStatusReplied,
						Description: fmt.Sprintf("%s assigned %s %s's case to %s", assigneeProvider.ShortDisplayName,
							patient.FirstName, patient.LastName, assignedProvider.ShortDisplayName),
						ShortDescription: fmt.Sprintf("Assigned to %s", assignedProvider.ShortDisplayName),
						ActionURL:        app_url.ViewPatientMessagesAction(patient.ID, ev.Case.ID.Int64()),
						Tags:             []string{ev.Case.Name},
					},
				},
			}); err != nil {
				golog.Errorf("Unable to insert case assignment item into doctor queue: %s", err)
				routeFailure.Inc(1)
				return err
			}
		}

		// insert a pending item into the queue of the assigned provider
		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action: api.DQActionInsert,
				Dedupe: true,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:  assignedProvider.ID.Int64(),
					PatientID: ev.Case.PatientID,
					ItemID:    ev.Case.ID.Int64(),
					EventType: api.DQEventTypeCaseAssignment,
					Status:    api.DQItemStatusPending,
					Description: fmt.Sprintf("%s %s's case assigned to %s", patient.FirstName, patient.LastName,
						assignedProvider.ShortDisplayName),
					ShortDescription: fmt.Sprintf("Reassigned by %s", assigneeProvider.ShortDisplayName),
					ActionURL:        app_url.ViewPatientMessagesAction(patient.ID, ev.Case.ID.Int64()),
					Tags:             []string{ev.Case.Name},
				},
			},
		}); err != nil {
			golog.Errorf("Unable to insert case assignment item into doctor queue: %s", err)
			routeFailure.Inc(1)
			return err
		}

		// notify the assigned provider
		if err := notificationManager.NotifyDoctor(
			assignedProviderRole,
			assignedProvider.ID.Int64(),
			assignedProvider.AccountID.Int64(),
			&notify.Message{
				ShortMessage: caseAssignmentMessage,
			}); err != nil {
			golog.Errorf("Unable to notify assigned provider of event %T: %s", ev, err)
			routeFailure.Inc(1)
			return err
		}

		routeSuccess.Inc(1)
		return nil
	})

	dispatcher.Subscribe(func(ev *doctor_treatment_plan.TreatmentPlanScheduledMessageCancelledEvent) error {

		patient, err := dataAPI.Patient(ev.PatientID, true)
		if err != nil {
			return errors.Trace(err)
		}

		doctor, err := dataAPI.Doctor(ev.DoctorID, true)
		if err != nil {
			return errors.Trace(err)
		}

		pc, err := dataAPI.GetPatientCaseFromID(ev.CaseID)
		if err != nil {
			return errors.Trace(err)
		}

		var description, shortDescription string
		if ev.Undone {
			description = fmt.Sprintf("%s undid scheduled message cancellation for %s %s's.", doctor.ShortDisplayName, patient.FirstName, patient.LastName)
			shortDescription = fmt.Sprintf("Scheduled message cancellation undone")
		} else {
			description = fmt.Sprintf("%s cancelled scheduled message for %s %s's.", doctor.ShortDisplayName, patient.FirstName, patient.LastName)
			shortDescription = fmt.Sprintf("Scheduled message cancelled")
		}

		if err := dataAPI.UpdateDoctorQueue([]*api.DoctorQueueUpdate{
			{
				Action: api.DQActionInsert,
				Dedupe: false,
				QueueItem: &api.DoctorQueueItem{
					DoctorID:         ev.DoctorID,
					PatientID:        ev.PatientID,
					ItemID:           ev.CaseID,
					EventType:        api.DQEventTypeCaseMessage,
					Status:           api.DQItemStatusCancelled,
					Description:      description,
					ShortDescription: shortDescription,
					ActionURL:        app_url.ViewPatientMessagesAction(ev.PatientID, ev.CaseID),
					Tags:             []string{pc.Name},
				},
			},
		}); err != nil {
			golog.Errorf("Unable to item into doctor queue: %s", err)
			routeFailure.Inc(1)
			return errors.Trace(err)
		}

		return nil
	})
}
