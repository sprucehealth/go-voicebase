package schedmsg

import (
	"fmt"
	"time"

	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/messages"
	"github.com/sprucehealth/backend/cmd/svc/restapi/patient"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/ptr"
)

var (
	defaultTimePeriod = 20
)

type Worker struct {
	dataAPI       api.DataAPI
	authAPI       api.AuthAPI
	publisher     dispatch.Publisher
	timePeriod    int
	stopCh        chan bool
	statSucceeded *metrics.Counter
	statFailed    *metrics.Counter
	statAge       metrics.Histogram
}

func StartWorker(
	dataAPI api.DataAPI, authAPI api.AuthAPI, publisher dispatch.Publisher,
	metricsRegistry metrics.Registry, timePeriod int,
) *Worker {
	w := NewWorker(dataAPI, authAPI, publisher, metricsRegistry, timePeriod)
	w.Start()
	return w
}

func NewWorker(
	dataAPI api.DataAPI, authAPI api.AuthAPI, publisher dispatch.Publisher,
	metricsRegistry metrics.Registry, timePeriod int,
) *Worker {
	tPeriod := timePeriod
	if tPeriod == 0 {
		tPeriod = defaultTimePeriod
	}
	w := &Worker{
		dataAPI:       dataAPI,
		authAPI:       authAPI,
		publisher:     publisher,
		timePeriod:    tPeriod,
		stopCh:        make(chan bool),
		statSucceeded: metrics.NewCounter(),
		statFailed:    metrics.NewCounter(),
		statAge:       metrics.NewUnbiasedHistogram(),
	}
	metricsRegistry.Add("age", w.statAge)
	metricsRegistry.Add("succeeded", w.statSucceeded)
	metricsRegistry.Add("failed", w.statFailed)
	return w
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case <-w.stopCh:
				return
			default:
			}

			msgConsumed, err := w.ConsumeMessage()
			if err != nil {
				golog.Errorf(err.Error())
			}

			if !msgConsumed {
				select {
				case <-w.stopCh:
					return
				case <-time.After(time.Duration(w.timePeriod) * time.Second):
				}
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

func (w *Worker) ConsumeMessage() (bool, error) {
	scheduledMessage, err := w.dataAPI.RandomlyPickAndStartProcessingScheduledMessage(ScheduledMsgTypes)
	if api.IsErrNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, errors.Trace(err)
	}

	w.statAge.Update(time.Since(scheduledMessage.Scheduled).Nanoseconds() / 1e9)

	if err := w.processMessage(scheduledMessage); err != nil {
		switch errors.Cause(err) {
		case api.ErrNotFound("sku"), patient.ErrInitialVisitNotTreated, patient.ErrFollowupNotSupportedOnApp, patient.ErrNoInitialVisitFound, patient.ErrOpenFollowupExists:
			// Record this as a success since it's a handled error
			w.statSucceeded.Inc(1)
			golog.Errorf("Can't send scheduled message %d: %s", scheduledMessage.ID, err.Error())
			if err := w.dataAPI.UpdateScheduledMessage(scheduledMessage.ID, &api.ScheduledMessageUpdate{
				Status: ptr.String(common.SMError.String()),
			}); err != nil {
				return false, errors.Trace(err)
			}
			return false, errors.Trace(err)
		}
		w.statFailed.Inc(1)
		// revert the status back to being in the scheduled state
		if err := w.dataAPI.UpdateScheduledMessage(scheduledMessage.ID, &api.ScheduledMessageUpdate{
			Status: ptr.String(common.SMScheduled.String()),
		}); err != nil {
			return false, errors.Trace(err)
		}
		return false, errors.Trace(err)
	}

	w.statSucceeded.Inc(1)

	// update the status to indicate that the message was succesfully sent
	if err := w.dataAPI.UpdateScheduledMessage(scheduledMessage.ID, &api.ScheduledMessageUpdate{
		Status:        ptr.String(common.SMSent.String()),
		CompletedTime: ptr.Time(time.Now()),
	}); err != nil {
		return false, errors.Trace(err)
	}

	return true, nil
}

func (w *Worker) processMessage(schedMsg *common.ScheduledMessage) error {
	switch schedMsg.Message.TypeName() {
	case common.SMCaseMessageType:
		appMessage := schedMsg.Message.(*CaseMessage)

		patientCase, err := w.dataAPI.GetPatientCaseFromID(appMessage.PatientCaseID)
		if err != nil {
			return errors.Trace(err)
		}

		people, err := w.dataAPI.GetPeople([]int64{appMessage.SenderPersonID})
		if err != nil {
			return errors.Trace(err)
		}

		msg := &common.CaseMessage{
			PersonID: appMessage.SenderPersonID,
			Body:     appMessage.Message,
			CaseID:   appMessage.PatientCaseID,
		}

		if err := messages.CreateMessageAndAttachments(msg, appMessage.Attachments,
			appMessage.SenderPersonID, appMessage.ProviderID, appMessage.SenderRole, w.dataAPI); err != nil {
			return errors.Trace(err)
		}

		w.publisher.Publish(&messages.PostEvent{
			Message:     msg,
			Case:        patientCase,
			Person:      people[appMessage.SenderPersonID],
			IsAutomated: appMessage.IsAutomated,
		})

	case common.SMTreatmanPlanMessageType:
		sm := schedMsg.Message.(*TreatmentPlanMessage)

		// Make sure treatment plan is still active. This will happen if a treatment plan was revised after messages
		// were scheduled. It's fine. Just want to make sure not to send the messages.
		tp, err := w.dataAPI.GetAbridgedTreatmentPlan(sm.TreatmentPlanID, 0)
		if err != nil {
			return errors.Trace(err)
		}
		if tp.Status != api.StatusActive {
			golog.Infof("Treatment plan %d not active when trying to send scheduled message %d", sm.TreatmentPlanID, sm.MessageID)
			return nil
		}

		msg, err := w.dataAPI.TreatmentPlanScheduledMessage(sm.MessageID)
		if err != nil {
			return errors.Trace(err)
		}

		pcase, err := w.dataAPI.GetPatientCaseFromID(tp.PatientCaseID.Int64())
		if err != nil {
			return errors.Trace(err)
		}

		careTeams, err := w.dataAPI.CaseCareTeams([]int64{pcase.ID.Int64()})
		if err != nil {
			return errors.Trace(err)
		}
		if len(careTeams) != 1 {
			return fmt.Errorf("Expected to find 1 care team for patient case %d but found %d", pcase.ID.Int64(), len(careTeams))
		}

		_, ok := careTeams[pcase.ID.Int64()]
		if !ok {
			return fmt.Errorf("No care team found for patient %s for case %d", tp.PatientID, tp.PatientCaseID.Int64())
		}

		var careCoordinator *common.Doctor
		var activeDoctorID int64
		for _, x := range careTeams[pcase.ID.Int64()].Assignments {
			switch x.ProviderRole {
			case api.RoleCC:
				careCoordinator, err = w.dataAPI.Doctor(x.ProviderID, true)
				if err != nil {
					return errors.Trace(err)
				}
			case api.RoleDoctor:
				activeDoctorID = x.ProviderID
			}
		}

		if careCoordinator == nil {
			golog.Errorf("Unable to find care coordinator in care team for patient case %d - continuing but this is suspicious. This case will not be reassigned.", tp.PatientCaseID.Int64())
		} else if activeDoctorID == 0 {
			return errors.Trace(fmt.Errorf("Unable to find active doctor on the case care team on behalf of whom to send the scheduled treatment plan (id: %d) message.", sm.TreatmentPlanID))
		}

		personID, err := w.dataAPI.GetPersonIDByRole(api.RoleDoctor, activeDoctorID)
		if err != nil {
			return errors.Trace(err)
		}

		people, err := w.dataAPI.GetPeople([]int64{personID})
		if err != nil {
			return errors.Trace(err)
		}

		// Create follow-up visits when necessary
		for _, a := range msg.Attachments {
			if a.ItemType == common.AttachmentTypeFollowupVisit {
				pat, err := w.dataAPI.GetPatientFromID(tp.PatientID)
				if err != nil {
					return errors.Trace(err)
				}

				fvisit, err := patient.CreatePendingFollowup(pat, pcase, w.dataAPI, w.authAPI, w.publisher)
				if err != nil {
					return errors.Trace(err)
				}

				a.ItemID = fvisit.ID.Int64()
				break
			}
		}

		cmsg := &common.CaseMessage{
			PersonID:    personID,
			Body:        msg.Message,
			CaseID:      pcase.ID.Int64(),
			Attachments: msg.Attachments,
		}

		msg.ID, err = w.dataAPI.CreateCaseMessage(cmsg)
		if err != nil {
			return errors.Trace(err)
		}

		w.publisher.Publish(&messages.PostEvent{
			Message: cmsg,
			Case:    pcase,
			Person:  people[personID],
		})

		if careCoordinator != nil {
			// Whenever a TP sched message goes out we should reassign to the CC if one exists
			w.publisher.Publish(&messages.CaseAssignEvent{
				Message:     cmsg,
				Person:      people[personID],
				Case:        pcase,
				Doctor:      people[personID].Doctor,
				MA:          careCoordinator,
				IsAutomated: true,
			})
		}

		// update to indicate the the treatment plan scheduled message was sent
		if err := w.dataAPI.UpdateTreatmentPlanScheduledMessage(sm.MessageID, &api.TreatmentPlanScheduledMessageUpdate{
			SentTime: ptr.Time(time.Now()),
		}); err != nil {
			golog.Errorf("Unable to update treatment plan scheduled message (id %d) to indicate that it was sent: %s", sm.MessageID, err.Error())
		}
	default:
		return errors.Trace(fmt.Errorf("Unknown message type: %s", schedMsg.Message.TypeName()))
	}

	return nil
}
