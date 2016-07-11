package patient_file

import (
	"time"

	"github.com/sprucehealth/backend/cmd/svc/restapi/analytics"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
)

type PatientVisitOpenedEvent struct {
	PatientVisit *common.PatientVisit
	PatientID    common.PatientID
	DoctorID     int64
	Role         string
}

func (e *PatientVisitOpenedEvent) Events() []analytics.Event {
	return []analytics.Event{
		&analytics.ServerEvent{
			Event:     "visit_opened",
			Timestamp: analytics.Time(time.Now()),
			PatientID: e.PatientID.Int64(),
			DoctorID:  e.DoctorID,
			VisitID:   e.PatientVisit.ID.Int64(),
			CaseID:    e.PatientVisit.PatientCaseID.Int64(),
			Role:      e.Role,
		},
	}
}