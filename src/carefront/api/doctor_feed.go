package api

import (
	"carefront/app_url"
	"carefront/common"
	"carefront/libs/golog"
	"carefront/settings"
	"fmt"
	"time"
)

const (
	DQEventTypePatientVisit                  = "PATIENT_VISIT"
	DQEventTypeTreatmentPlan                 = "TREATMENT_PLAN"
	DQEventTypeRefillRequest                 = "REFILL_REQUEST"
	DQEventTypeTransmissionError             = "TRANSMISSION_ERROR"
	DQEventTypeUnlinkedDNTFTransmissionError = "UNLINKED_DNTF_TRANSMISSION_ERROR"
	DQEventTypeRefillTransmissionError       = "REFILL_TRANSMISSION_ERROR"
	DQEventTypeCaseMessage                   = "CASE_MESSAGE"
	DQItemStatusPending                      = "PENDING"
	DQItemStatusTreated                      = "TREATED"
	DQItemStatusTriaged                      = "TRIAGED"
	DQItemStatusOngoing                      = "ONGOING"
	DQItemStatusRefillApproved               = "APPROVED"
	DQItemStatusRefillDenied                 = "DENIED"
	DQItemStatusReplied                      = "REPLIED"
	DQItemStatusRead                         = "READ"
)

type DoctorQueueItem struct {
	Id                   int64
	DoctorId             int64
	EventType            string
	EnqueueDate          time.Time
	CompletedDate        time.Time
	Expires              time.Time
	ItemId               int64
	Status               string
	PatientCaseId        int64
	PositionInQueue      int
	CareProvidingStateId int64
}

func (d *DoctorQueueItem) GetTitleAndSubtitle(dataApi DataAPI) (string, string, error) {
	var title, subtitle string

	switch d.EventType {
	case DQEventTypePatientVisit, DQEventTypeTreatmentPlan:
		var patient *common.Patient
		var err error

		if d.EventType == DQEventTypeTreatmentPlan {
			patient, err = dataApi.GetPatientFromTreatmentPlanId(d.ItemId)
			if err == NoRowsError {
				golog.Errorf("Did not get patient from treatment plan id (%d)", d.ItemId)
				return "", "", nil
			} else if err != nil {
				return "", "", err
			}
		} else {
			patient, err = dataApi.GetPatientFromPatientVisitId(d.ItemId)
			if err == NoRowsError {
				golog.Errorf("Did not get patient from patient visit id (%d)", d.ItemId)
				return "", "", nil
			} else if err != nil {
				return "", "", err
			}
		}

		switch d.Status {
		case DQItemStatusTreated:
			title = fmt.Sprintf("Treatment Plan completed for %s %s", patient.FirstName, patient.LastName)
		case DQItemStatusPending:
			title = fmt.Sprintf("New visit with %s %s", patient.FirstName, patient.LastName)
			subtitle = getRemainingTimeSubtitleForCaseToBeReviewed(d.EnqueueDate)
		case DQItemStatusOngoing:
			title = fmt.Sprintf("Continue reviewing visit with %s %s", patient.FirstName, patient.LastName)
			subtitle = getRemainingTimeSubtitleForCaseToBeReviewed(d.EnqueueDate)
		case DQItemStatusTriaged:
			title = fmt.Sprintf("Completed and triaged visit for %s %s", patient.FirstName, patient.LastName)
		}

	case DQEventTypeRefillRequest:
		patient, err := dataApi.GetPatientFromRefillRequestId(d.ItemId)
		if err == NoRowsError {
			golog.Errorf("Unable to get patient from refill request id %d", d.ItemId)
			return "", "", nil
		} else if err != nil {
			return "", "", err
		}

		switch d.Status {
		case DQItemStatusPending:
			title = fmt.Sprintf("Refill request for %s %s", patient.FirstName, patient.LastName)
		case DQItemStatusRefillApproved:
			title = fmt.Sprintf("Refill request approved for %s %s", patient.FirstName, patient.LastName)
		case DQItemStatusRefillDenied:
			title = fmt.Sprintf("Refill request denied for %s %s", patient.FirstName, patient.LastName)
		}

	case DQEventTypeRefillTransmissionError:
		patient, err := dataApi.GetPatientFromRefillRequestId(d.ItemId)
		if err == NoRowsError {
			golog.Errorf("Unable to get patient from refill request id %d", d.ItemId)
			return "", "", nil
		} else if err != nil {
			return "", "", err
		}

		switch d.Status {
		case DQItemStatusPending:
			title = fmt.Sprintf("Error completing refill request for %s %s", patient.FirstName, patient.LastName)
		case DQItemStatusTreated:
			title = fmt.Sprintf("Refill request error resolved for %s %s", patient.FirstName, patient.LastName)
		}

	case DQEventTypeTransmissionError:
		patient, err := dataApi.GetPatientFromTreatmentId(d.ItemId)
		if err == NoRowsError {
			golog.Errorf("Unable to get patient from treatment id %d", d.ItemId)
			return "", "", nil
		} else if err != nil {
			return "", "", err
		}

		switch d.Status {
		case DQItemStatusPending, DQItemStatusOngoing:
			title = fmt.Sprintf("Error sending prescription for %s %s", patient.FirstName, patient.LastName)
		case DQItemStatusTreated:
			title = fmt.Sprintf("Error resolved for %s %s", patient.FirstName, patient.LastName)
		}

	case DQEventTypeUnlinkedDNTFTransmissionError:
		unlinkedTreatment, err := dataApi.GetUnlinkedDNTFTreatment(d.ItemId)
		if err == NoRowsError {
			golog.Errorf("Unable to get unlinked dntf treatment from id %d", d.ItemId)
			return "", "", nil
		} else if err != nil {
			return "", "", err
		}

		switch d.Status {
		case DQItemStatusPending, DQItemStatusOngoing:
			title = fmt.Sprintf("Error sending prescription for %s %s", unlinkedTreatment.Patient.FirstName, unlinkedTreatment.Patient.LastName)
		case DQItemStatusTreated:
			title = fmt.Sprintf("Error resolved for %s %s", unlinkedTreatment.Patient.FirstName, unlinkedTreatment.Patient.LastName)
		}
	case DQEventTypeCaseMessage:
		participants, err := dataApi.CaseMessageParticipants(d.ItemId, true)
		if err != nil {
			return "", "", err
		}
		for _, par := range participants {
			person := par.Person
			if person.RoleType == PATIENT_ROLE {
				patient := person.Patient
				switch d.Status {
				case DQItemStatusPending:
					title = fmt.Sprintf("Message from %s %s", patient.FirstName, patient.LastName)
				case DQItemStatusRead:
					title = fmt.Sprintf("Conversation with %s %s", patient.FirstName, patient.LastName)
				case DQItemStatusReplied:
					title = fmt.Sprintf("Replied to %s %s", patient.FirstName, patient.LastName)
				}
				break
			}
		}
	}
	return title, subtitle, nil
}

func getRemainingTimeSubtitleForCaseToBeReviewed(enqueueDate time.Time) string {
	timeLeft := enqueueDate.Add(settings.SLA_TO_SERVICE_CUSTOMER).Sub(time.Now())
	minutesLeft := int64(timeLeft.Minutes()) - (60 * int64(timeLeft.Hours()))
	subtitle := fmt.Sprintf("%dh %dm left", int64(timeLeft.Hours()), int64(minutesLeft))
	return subtitle
}

func (d *DoctorQueueItem) GetImageUrl() *app_url.SpruceAsset {
	switch d.EventType {
	case DQEventTypePatientVisit:
		return app_url.PatientVisitQueueIcon
	}
	return nil
}

func (d *DoctorQueueItem) GetTimestamp() *time.Time {
	if d.EnqueueDate.IsZero() {
		return nil
	}

	return &d.EnqueueDate
}

func (d *DoctorQueueItem) GetDisplayTypes() []string {
	return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}
}

func (d *DoctorQueueItem) ActionUrl(dataApi DataAPI) (*app_url.SpruceAction, error) {
	switch d.EventType {
	case DQEventTypePatientVisit:
		patientVisit, err := dataApi.GetPatientVisitFromId(d.ItemId)
		if err != nil {
			golog.Errorf("Unable to get patient visit based on id: %s", err)
			return nil, err
		}

		switch d.Status {
		case DQItemStatusTreated, DQItemStatusTriaged:
			return app_url.ViewCompletedPatientVisitAction(patientVisit.PatientId.Int64(), d.ItemId, patientVisit.PatientCaseId.Int64()), nil
		case DQItemStatusOngoing, DQItemStatusPending:
			return app_url.BeginPatientVisitReviewAction(patientVisit.PatientId.Int64(), d.ItemId, patientVisit.PatientCaseId.Int64()), nil
		}
	case DQEventTypeTreatmentPlan:

		switch d.Status {
		case DQItemStatusTreated, DQItemStatusTriaged:
			patientVisit, err := dataApi.GetPatientVisitFromTreatmentPlanId(d.ItemId)

			if err == NoRowsError {
				golog.Errorf("Unable to get patient visit  from treatment plan id %d", d.ItemId)
				return nil, nil
			} else if err != nil {
				return nil, err
			}

			return app_url.ViewCompletedPatientVisitAction(patientVisit.PatientId.Int64(), patientVisit.PatientVisitId.Int64(), patientVisit.PatientCaseId.Int64()), nil
		}
	case DQEventTypeRefillTransmissionError:
		patient, err := dataApi.GetPatientFromRefillRequestId(d.ItemId)
		if err != nil {
			golog.Errorf("Unable to get patient from refill request id: %s", err)
			return nil, nil
		}

		return app_url.ViewRefillRequestAction(patient.PatientId.Int64(), d.ItemId), nil
	case DQEventTypeRefillRequest:
		patient, err := dataApi.GetPatientFromRefillRequestId(d.ItemId)
		if err != nil {
			golog.Errorf("Unable to get patient from refill request id %d", d.ItemId)
			return nil, nil
		}

		switch d.Status {
		case DQItemStatusOngoing, DQItemStatusPending:
			return app_url.ViewRefillRequestAction(patient.PatientId.Int64(), d.ItemId), nil
		case DQItemStatusTreated, DQItemStatusRefillApproved, DQItemStatusRefillDenied:
			return app_url.ViewPatientTreatmentsAction(patient.PatientId.Int64()), nil
		}
	case DQEventTypeTransmissionError:
		patient, err := dataApi.GetPatientFromTreatmentId(d.ItemId)
		if err != nil {
			golog.Errorf("Unable to get patient from treatment id : %s", err)
			return nil, nil
		}
		return app_url.ViewTransmissionErrorAction(patient.PatientId.Int64(), d.ItemId), nil
	case DQEventTypeUnlinkedDNTFTransmissionError:
		patient, err := dataApi.GetPatientFromUnlinkedDNTFTreatment(d.ItemId)
		if err != nil {
			golog.Errorf("Unable to get patient from unlinked dntf treatment: %s", err)
			return nil, nil
		}
		return app_url.ViewTransmissionErrorAction(patient.PatientId.Int64(), d.ItemId), nil
	case DQEventTypeCaseMessage:
		participants, err := dataApi.CaseMessageParticipants(d.ItemId, false)
		if err != nil {
			return nil, err
		}
		for _, p := range participants {
			if p.Person.RoleType == PATIENT_ROLE {
				return app_url.ViewPatientConversationsAction(p.Person.RoleId, d.ItemId), nil
			}
		}
	}

	return nil, nil
}
