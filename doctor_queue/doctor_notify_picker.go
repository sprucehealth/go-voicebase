package doctor_queue

import (
	"math/rand"
	"time"

	"github.com/sprucehealth/backend/api"
)

// DoctorNotifyPicker is an interface to provide different
// ways in which to pick a doctor to notify for a paritcular care providing state
type DoctorNotifyPickerConfig struct {
	CareProvidingStateID                   int64
	TimePeriodBetweenNotifyingAtStateLevel time.Duration
	TimePeriodBeforeNotifyingSameDoctor    time.Duration
	StatesToAvoid                          []int64
}

type DoctorNotifyPicker interface {
	PickDoctorToNotify(config *DoctorNotifyPickerConfig) (int64, error)
}

// defaultDoctorPicker picks a doctor to notify of a case in a state when:
// a) No doctor has been notified of a case in that state for the specified time period
// b) There is a doctor that either:
// 		b.1) Has never been notified of a case OR
// 		b.2) Has been notified, but not within the minimum time required before notifying the same doctor
// 		WHILE also biasing towards doctors that are not registered in previous states for which a doctor
// 		was just notified
type defaultDoctorPicker struct {
	dataAPI api.DataAPI
}

func (d *defaultDoctorPicker) PickDoctorToNotify(config *DoctorNotifyPickerConfig) (int64, error) {

	// only notify at a state level once per 15 minute period
	lastNotifiedTime, err := d.dataAPI.LastNotifiedTimeForCareProvidingState(config.CareProvidingStateID)
	if err != api.NoRowsError && err != nil {
		return 0, err
	} else if err != api.NoRowsError &&
		!lastNotifiedTime.Add(config.TimePeriodBetweenNotifyingAtStateLevel).Before(time.Now()) {
		return 0, nil
	}

	// don't notify the same doctor within the specified period
	timeThreshold := time.Now().Add(-config.TimePeriodBeforeNotifyingSameDoctor)
	for i := len(config.StatesToAvoid); i >= 0; i-- {

		elligibleDoctors, err := d.dataAPI.DoctorsToNotifyInCareProvidingState(config.CareProvidingStateID,
			config.StatesToAvoid[:i], timeThreshold)
		if err != nil {
			return 0, err
		} else if len(elligibleDoctors) == 0 {
			continue
		}

		// populate all doctors that have never been notified so as to give preference to picking these
		// doctors before we start to pick from doctors that have already been notified
		doctorsNeverNotified := make([]*api.DoctorNotify, 0, len(elligibleDoctors))
		for _, dNotify := range elligibleDoctors {
			if dNotify.LastNotified == nil {
				doctorsNeverNotified = append(doctorsNeverNotified, dNotify)
			}
		}
		if len(doctorsNeverNotified) > 0 {
			return doctorsNeverNotified[rand.Intn(len(doctorsNeverNotified))].DoctorID, nil
		}

		// randomly pick one of the doctors
		return elligibleDoctors[rand.Intn(len(elligibleDoctors))].DoctorID, nil
	}

	return 0, noDoctorFound
}
