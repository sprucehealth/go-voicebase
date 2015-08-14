package patient_visit

import (
	"time"

	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/common"
)

type DiagnosisModifiedEvent struct {
	PatientID       common.PatientID
	DoctorID        int64
	PatientVisitID  int64
	TreatmentPlanID int64
	PatientCaseID   int64
}

func (e *DiagnosisModifiedEvent) Events() []analytics.Event {
	return []analytics.Event{
		&analytics.ServerEvent{
			Event:           "diagnosis_modified",
			Timestamp:       analytics.Time(time.Now()),
			PatientID:       e.PatientID.Int64(),
			DoctorID:        e.DoctorID,
			VisitID:         e.PatientVisitID,
			CaseID:          e.PatientCaseID,
			TreatmentPlanID: e.TreatmentPlanID,
		},
	}
}

type PatientVisitMarkedUnsuitableEvent struct {
	PatientVisitID int64
	PatientID      common.PatientID
	CaseID         int64
	DoctorID       int64
	Reason         string
}

func (e *PatientVisitMarkedUnsuitableEvent) Events() []analytics.Event {
	return []analytics.Event{
		&analytics.ServerEvent{
			Event:     "visit_marked_unsuitable",
			Timestamp: analytics.Time(time.Now()),
			PatientID: e.PatientID.Int64(),
			DoctorID:  e.DoctorID,
			VisitID:   e.PatientVisitID,
			CaseID:    e.CaseID,
		},
	}
}

type PreSubmissionVisitTriageEvent struct {
	VisitID       int64
	CaseID        int64
	Title         string
	ActionMessage string
	ActionURL     string
}

func (e *PreSubmissionVisitTriageEvent) Events() []analytics.Event {
	return []analytics.Event{
		&analytics.ServerEvent{
			Event:     "visit_pre_submission_triage",
			Timestamp: analytics.Time(time.Now()),
			VisitID:   e.VisitID,
			CaseID:    e.CaseID,
		},
	}
}
