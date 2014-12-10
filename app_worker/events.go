package app_worker

import "github.com/sprucehealth/backend/common"

type RxTransmissionErrorEvent struct {
	Patient   *common.Patient
	DoctorID  int64
	ItemID    int64
	EventType common.ERxSourceType
}

type RefillRequestCreatedEvent struct {
	Patient         *common.Patient
	DoctorID        int64
	RefillRequestID int64
	Status          string
}
