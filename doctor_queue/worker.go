package doctor_queue

import (
	"errors"
	"time"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/doctor"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/notify"
	"github.com/sprucehealth/backend/third_party/github.com/samuel/go-metrics/metrics"
)

var (
	noDoctorFound                          = errors.New("No doctor found to notify")
	timePeriodBetweenNotificationChecks    = time.Minute
	minimumTimeBeforeNotifyingSameDoctor   = time.Hour
	minimumTimeBeforeNotifyingForSameState = 15 * time.Minute
)

type Worker struct {
	dataAPI                                api.DataAPI
	notificationManager                    *notify.NotificationManager
	lockAPI                                api.LockAPI
	stopChan                               chan bool
	doctorPicker                           DoctorNotifyPicker
	statNotificationCycle                  *metrics.Counter
	statNoDoctorsToNotify                  *metrics.Counter
	timePeriodBetweenChecks                time.Duration
	minimumTimeBeforeNotifyingForSameState time.Duration
	minimumTimeBeforeNotifyingSameDoctor   time.Duration
}

func StartWorker(dataAPI api.DataAPI, authAPI api.AuthAPI, lockAPI api.LockAPI,
	notificationManager *notify.NotificationManager,
	metricsRegistry metrics.Registry) *Worker {

	statNotificationCycle := metrics.NewCounter()
	statNoDoctorsToNotify := metrics.NewCounter()

	metricsRegistry.Add("cycle", statNotificationCycle)
	metricsRegistry.Add("nodoctors", statNoDoctorsToNotify)

	w := &Worker{
		dataAPI:                                dataAPI,
		notificationManager:                    notificationManager,
		lockAPI:                                lockAPI,
		statNotificationCycle:                  statNotificationCycle,
		statNoDoctorsToNotify:                  statNoDoctorsToNotify,
		doctorPicker:                           &defaultDoctorPicker{dataAPI: dataAPI, authAPI: authAPI},
		stopChan:                               make(chan bool),
		timePeriodBetweenChecks:                timePeriodBetweenNotificationChecks,
		minimumTimeBeforeNotifyingForSameState: minimumTimeBeforeNotifyingForSameState,
		minimumTimeBeforeNotifyingSameDoctor:   minimumTimeBeforeNotifyingSameDoctor,
	}
	w.start()
	return w
}

func (w *Worker) start() {
	go func() {
		defer w.lockAPI.Release()
		for {
			if !w.lockAPI.Wait() {
				return
			}

			select {
			case <-w.stopChan:
				return
			default:
			}

			if err := w.notifyDoctorsOfUnclaimedCases(); err != nil {
				golog.Errorf(err.Error())
			}
			w.statNotificationCycle.Inc(1)
			time.Sleep(w.timePeriodBetweenChecks)
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) notifyDoctorsOfUnclaimedCases() error {

	// identify the distinct states in which we currently have unclaimed cases
	careProvidingStateIDs, err := w.dataAPI.CareProvidingStatesWithUnclaimedCases()
	if err != nil {
		return err
	}

	// iterate through the states to notify a doctor per state
	for i, careProvidingStateID := range careProvidingStateIDs {

		doctorToNotify, err := w.doctorPicker.PickDoctorToNotify(&DoctorNotifyPickerConfig{
			CareProvidingStateID:                   careProvidingStateID,
			StatesToAvoid:                          careProvidingStateIDs[:i],
			MinimumTimeBeforeNotifyingForSameState: w.minimumTimeBeforeNotifyingForSameState,
			MinimumTimeBeforeNotifyingSameDoctor:   w.minimumTimeBeforeNotifyingSameDoctor,
		})
		if err == noDoctorFound {
			w.statNoDoctorsToNotify.Inc(1)
			continue
		} else if err != nil {
			return err
		} else if doctorToNotify == 0 {
			continue
		}

		accountID, err := w.dataAPI.GetAccountIDFromDoctorID(doctorToNotify)
		if err != nil {
			return err
		}

		if err := w.notificationManager.NotifyDoctor(api.DOCTOR_ROLE, doctorToNotify, accountID, &doctor.NotifyDoctorOfUnclaimedCaseEvent{
			DoctorID: doctorToNotify,
		}); err != nil {
			return err
		}
	}

	return nil
}
