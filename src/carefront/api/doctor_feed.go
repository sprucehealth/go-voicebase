package api

import (
	"carefront/settings"
	"fmt"
	"time"
)

const (
	EVENT_TYPE_PATIENT_VISIT            = "PATIENT_VISIT"
	EVENT_TYPE_TREATMENT_PLAN           = "TREATMENT_PLAN"
	EVENT_TYPE_REFILL_REQUEST           = "REFILL_REQUEST"
	EVENT_TYPE_TRANSMISSION_ERROR       = "TRANSMISSION_ERROR"
	patientVisitImageTag                = "patient_visit_queue_icon"
	beginPatientVisitReviewAction       = "begin_patient_visit"
	viewTreatedPatientVisitReviewAction = "view_treated_patient_visit"
	viewRefillRequestAction             = "view_refill_request"
	viewTransmissionErrorAction         = "view_transmission_error"
)

type DoctorQueueItem struct {
	Id              int64
	DoctorId        int64
	EventType       string
	EnqueueDate     time.Time
	CompletedDate   time.Time
	ItemId          int64
	Status          string
	PositionInQueue int
}

func (d *DoctorQueueItem) GetTitleAndSubtitle(dataApi DataAPI) (string, string, error) {
	var title, subtitle string

	switch d.EventType {
	case EVENT_TYPE_PATIENT_VISIT, EVENT_TYPE_TREATMENT_PLAN:
		var patientVisitId int64
		var err error

		if d.EventType == EVENT_TYPE_TREATMENT_PLAN {
			patientVisitId, err = dataApi.GetPatientVisitIdFromTreatmentPlanId(d.ItemId)
			if err != nil {
				return "", "", err
			}
		} else {
			patientVisitId = d.ItemId
		}

		patientId, err := dataApi.GetPatientIdFromPatientVisitId(patientVisitId)
		if err != nil {
			return "", "", err
		}
		patient, err := dataApi.GetPatientFromId(patientId)
		if err != nil {
			return "", "", err
		}

		switch d.Status {
		case QUEUE_ITEM_STATUS_COMPLETED:
			title = fmt.Sprintf("Treatment Plan completed for %s %s", patient.FirstName, patient.LastName)
			formattedTime := d.EnqueueDate.Format("3:04pm")
			subtitle = fmt.Sprintf("%s %d at %s", d.EnqueueDate.Month().String(), d.EnqueueDate.Day(), formattedTime)
		case QUEUE_ITEM_STATUS_PENDING:
			title = fmt.Sprintf("New visit with %s %s", patient.FirstName, patient.LastName)
			subtitle = getRemainingTimeSubtitleForCaseToBeReviewed(d.EnqueueDate)
		case QUEUE_ITEM_STATUS_ONGOING:
			title = fmt.Sprintf("Continue reviewing visit with %s %s", patient.FirstName, patient.LastName)
			subtitle = getRemainingTimeSubtitleForCaseToBeReviewed(d.EnqueueDate)
		case QUEUE_ITEM_STATUS_PHOTOS_REJECTED:
			title = fmt.Sprintf("Photos rejected for visit with %s %s", patient.FirstName, patient.LastName)
			formattedTime := d.EnqueueDate.Format("3:04pm")
			subtitle = fmt.Sprintf("%s %d at %s", d.EnqueueDate.Month().String(), d.EnqueueDate.Day(), formattedTime)
		case QUEUE_ITEM_STATUS_TRIAGED:
			title = fmt.Sprintf("Completed and triaged visit for %s %s", patient.FirstName, patient.LastName)
			formattedTime := d.EnqueueDate.Format("3:04pm")
			subtitle = fmt.Sprintf("%s %d at %s", d.EnqueueDate.Month().String(), d.EnqueueDate.Day(), formattedTime)
		}

	case EVENT_TYPE_REFILL_REQUEST:
		patient, err := dataApi.GetPatientFromRefillRequestId(d.ItemId)
		if err != nil {
			return "", "", err
		}

		if patient == nil {
			return "", "", nil
		}

		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING:
			title = fmt.Sprintf("Refill request for %s %s", patient.FirstName, patient.LastName)
		case QUEUE_ITEM_STATUS_ONGOING:
			title = fmt.Sprintf("Continue refill request for %s %s", patient.FirstName, patient.LastName)
		case QUEUE_ITEM_STATUS_REFILL_APPROVED:
			title = fmt.Sprintf("Refill request approved for %s %s", patient.FirstName, patient.LastName)
			formattedTime := d.EnqueueDate.Format("3:04pm")
			subtitle = fmt.Sprintf("%s %d at %s", d.EnqueueDate.Month().String(), d.EnqueueDate.Day(), formattedTime)
		case QUEUE_ITEM_STATUS_REFILL_DENIED:
			title = fmt.Sprintf("Refill request denied for %s %s", patient.FirstName, patient.LastName)
			formattedTime := d.EnqueueDate.Format("3:04pm")
			subtitle = fmt.Sprintf("%s %d at %s", d.EnqueueDate.Month().String(), d.EnqueueDate.Day(), formattedTime)
		}
	case EVENT_TYPE_TRANSMISSION_ERROR:
		patient, err := dataApi.GetPatientFromTreatmentId(d.ItemId)
		if err != nil {
			return "", "", err
		}

		if patient == nil {
			return "", "", nil
		}

		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING, QUEUE_ITEM_STATUS_ONGOING:
			title = fmt.Sprintf("Error sending prescription for %s %s", patient.FirstName, patient.LastName)
		case QUEUE_ITEM_STATUS_COMPLETED:
			title = fmt.Sprintf("Error resolved for %s %s", patient.FirstName, patient.LastName)
			formattedTime := d.EnqueueDate.Format("3:04pm")
			subtitle = fmt.Sprintf("%s %d at %s", d.EnqueueDate.Month().String(), d.EnqueueDate.Day(), formattedTime)
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

func (d *DoctorQueueItem) GetImageUrl() string {
	switch d.EventType {
	case EVENT_TYPE_PATIENT_VISIT:
		return fmt.Sprintf("%s%s", SpruceImageBaseUrl, patientVisitImageTag)
	}
	return ""
}

func (d *DoctorQueueItem) GetDisplayTypes() []string {
	switch d.EventType {
	case EVENT_TYPE_PATIENT_VISIT, EVENT_TYPE_TREATMENT_PLAN:
		switch d.Status {

		case QUEUE_ITEM_STATUS_PHOTOS_REJECTED:
			return []string{DISPLAY_TYPE_TITLE_SUBTITLE_NONACTIONABLE}

		case QUEUE_ITEM_STATUS_COMPLETED, QUEUE_ITEM_STATUS_TRIAGED:
			return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}

		case QUEUE_ITEM_STATUS_PENDING, QUEUE_ITEM_STATUS_ONGOING:
			if d.PositionInQueue == 0 {
				return []string{DISPLAY_TYPE_TITLE_SUBTITLE_BUTTON}
			} else {
				return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}
			}
		}
	case EVENT_TYPE_REFILL_REQUEST:
		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING, QUEUE_ITEM_STATUS_ONGOING:
			if d.PositionInQueue == 0 {
				return []string{DISPLAY_TYPE_TITLE_SUBTITLE_BUTTON}
			} else {
				return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}
			}
		case QUEUE_ITEM_STATUS_REFILL_APPROVED, QUEUE_ITEM_STATUS_REFILL_DENIED:
			return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}
		}
	case EVENT_TYPE_TRANSMISSION_ERROR:
		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING, QUEUE_ITEM_STATUS_ONGOING:
			if d.PositionInQueue == 0 {
				return []string{DISPLAY_TYPE_TITLE_SUBTITLE_BUTTON}
			} else {
				return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}
			}
		case QUEUE_ITEM_STATUS_COMPLETED:
			return []string{DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE}
		}
	}
	return nil
}

func (d *DoctorQueueItem) GetActionUrl() string {
	switch d.EventType {
	case EVENT_TYPE_PATIENT_VISIT:
		switch d.Status {
		case QUEUE_ITEM_STATUS_COMPLETED, QUEUE_ITEM_STATUS_TRIAGED:
			return fmt.Sprintf("%s%s?patient_visit_id=%d", SpruceButtonBaseActionUrl, viewTreatedPatientVisitReviewAction, d.ItemId)
		case QUEUE_ITEM_STATUS_ONGOING, QUEUE_ITEM_STATUS_PENDING:
			return fmt.Sprintf("%s%s?patient_visit_id=%d", SpruceButtonBaseActionUrl, beginPatientVisitReviewAction, d.ItemId)
		}
	case EVENT_TYPE_TREATMENT_PLAN:
		return fmt.Sprintf("%s%s?treatment_plan_id=%d", SpruceButtonBaseActionUrl, viewTreatedPatientVisitReviewAction, d.ItemId)
	case EVENT_TYPE_REFILL_REQUEST:
		return fmt.Sprintf("%s%s?refill_request_id=%d", SpruceButtonBaseActionUrl, viewRefillRequestAction, d.ItemId)
	case EVENT_TYPE_TRANSMISSION_ERROR:
		return fmt.Sprintf("%s%s?treatment_id=%d", SpruceButtonBaseActionUrl, viewTransmissionErrorAction, d.ItemId)
	}
	return ""
}

func (d *DoctorQueueItem) GetButton() *Button {
	switch d.EventType {
	case EVENT_TYPE_PATIENT_VISIT:
		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING:
			if d.PositionInQueue != 0 {
				return nil
			}
			button := &Button{}
			button.ButtonText = "Begin"
			button.ButtonActionUrl = fmt.Sprintf("%s%s?patient_visit_id=%d", SpruceButtonBaseActionUrl, beginPatientVisitReviewAction, d.ItemId)
			return button
		case QUEUE_ITEM_STATUS_ONGOING:
			button := &Button{}
			button.ButtonText = "Continue"
			button.ButtonActionUrl = fmt.Sprintf("%s%s?patient_visit_id=%d", SpruceButtonBaseActionUrl, beginPatientVisitReviewAction, d.ItemId)
			return button
		}
	case EVENT_TYPE_REFILL_REQUEST:
		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING:
			if d.PositionInQueue != 0 {
				return nil
			}
			button := &Button{}
			button.ButtonText = "Begin"
			button.ButtonActionUrl = fmt.Sprintf("%s%s?refill_request_id=%d", SpruceButtonBaseActionUrl, viewRefillRequestAction, d.ItemId)
			return button
		case QUEUE_ITEM_STATUS_ONGOING:
			if d.PositionInQueue != 0 {
				return nil
			}
			button := &Button{}
			button.ButtonText = "Continue"
			button.ButtonActionUrl = fmt.Sprintf("%s%s?refill_request_id=%d", SpruceButtonBaseActionUrl, viewRefillRequestAction, d.ItemId)
			return button
		}
	case EVENT_TYPE_TRANSMISSION_ERROR:
		switch d.Status {
		case QUEUE_ITEM_STATUS_PENDING:
			if d.PositionInQueue != 0 {
				return nil
			}
			button := &Button{}
			button.ButtonText = "Resolve Error"
			button.ButtonActionUrl = fmt.Sprintf("%s%s?treatment_id=%d", SpruceButtonBaseActionUrl, viewTransmissionErrorAction, d.ItemId)
			return button
		case QUEUE_ITEM_STATUS_ONGOING:
			if d.PositionInQueue != 0 {
				return nil
			}
			button := &Button{}
			button.ButtonText = "Resolve Error"
			button.ButtonActionUrl = fmt.Sprintf("%s%s?treatment_id=%d", SpruceButtonBaseActionUrl, viewTransmissionErrorAction, d.ItemId)
			return button
		}
	}
	return nil
}
