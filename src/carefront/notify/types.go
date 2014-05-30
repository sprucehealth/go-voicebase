package notify

import (
	"carefront/apiservice"
	"carefront/app_worker"
	"carefront/common"
	"carefront/common/config"
	"carefront/messages"
	"carefront/visit"
	"fmt"
	"reflect"
)

type internalNotificationView interface {
	renderEmail(event interface{}) (string, string, error)
}

var eventToInternalNotificationMapping map[reflect.Type]internalNotificationView

func getInternalNotificationViewForEvent(ev interface{}) internalNotificationView {
	return eventToInternalNotificationMapping[reflect.TypeOf(ev)]
}

type panicEventView int64

func (panicEventView) renderEmail(event interface{}) (string, string, error) {
	panicEvent, ok := event.(*config.PanicEvent)
	if !ok {
		return "", "", fmt.Errorf("Unexpected type: %T", event)
	}

	subject := fmt.Sprintf("PANIC %s.%s", panicEvent.AppName, panicEvent.Environment)
	body := panicEvent.Body
	return subject, body, nil
}

type patientVisitUnsuitableView int64

func (patientVisitUnsuitableView) renderEmail(event interface{}) (string, string, error) {
	unsuitableVisit, ok := event.(*visit.PatientVisitMarkedUnsuitableEvent)
	if !ok {
		return "", "", fmt.Errorf("Unexpected type: %T", event)
	}

	subject := fmt.Sprintf("Patient Visit %d marked unsuitable for Spruce", unsuitableVisit.PatientVisitId)
	body := "The patient visit id in the subject was marked as unsuitable for Spruce "
	return subject, body, nil
}

// notificationView interface represents the set of possible ways in which
// a notification can be rendered for communicating with a user.
// The idea is to have a notificationView for each of the events we are about.
type notificationView interface {
	renderEmail() string
	renderSMS() string
	renderPush(notificationConfig *config.NotificationConfig, notificationCount int64) interface{}
}

var eventToNotificationViewMapping map[reflect.Type]notificationView

func getNotificationViewForEvent(ev interface{}) notificationView {
	return eventToNotificationViewMapping[reflect.TypeOf(ev)]
}

func init() {
	eventToNotificationViewMapping = map[reflect.Type]notificationView{
		reflect.TypeOf(&apiservice.VisitSubmittedEvent{}):       visitSubmittedNotificationView(0),
		reflect.TypeOf(&apiservice.VisitReviewSubmittedEvent{}): visitReviewedNotificationView(0),
		reflect.TypeOf(&messages.ConversationStartedEvent{}):    newMessageNotificationView(0),
		reflect.TypeOf(&messages.ConversationReplyEvent{}):      newMessageNotificationView(0),
		reflect.TypeOf(&app_worker.RefillRequestCreatedEvent{}): refillRxCreatedNotificationView(0),
		reflect.TypeOf(&app_worker.RxTransmissionErrorEvent{}):  rxTransmissionErrorNotificationView(0),
	}

	eventToInternalNotificationMapping = map[reflect.Type]internalNotificationView{
		reflect.TypeOf(&config.PanicEvent{}):                       panicEventView(0),
		reflect.TypeOf(&visit.PatientVisitMarkedUnsuitableEvent{}): patientVisitUnsuitableView(0),
	}
}

type visitSubmittedNotificationView int64

func (visitSubmittedNotificationView) renderEmail() string {
	// TODO
	return ""
}

func (visitSubmittedNotificationView) renderSMS() string {
	return "You have a new patient visit waiting."
}

func (v visitSubmittedNotificationView) renderPush(notificationConfig *config.NotificationConfig, notificationCount int64) interface{} {
	return renderNotification(notificationConfig, v.renderSMS(), notificationCount)
}

type visitReviewedNotificationView int64

func (visitReviewedNotificationView) renderEmail() string {
	// TODO
	return ""
}

func (visitReviewedNotificationView) renderSMS() string {
	return "Doctor has reviewed your case."
}

func (v visitReviewedNotificationView) renderPush(notificationConfig *config.NotificationConfig, notificationCount int64) interface{} {
	return renderNotification(notificationConfig, v.renderSMS(), notificationCount)
}

type newMessageNotificationView int64

func (newMessageNotificationView) renderEmail() string {
	// TODO
	return ""
}

func (newMessageNotificationView) renderSMS() string {
	return "You have a new message."
}

func (n newMessageNotificationView) renderPush(notificationConfig *config.NotificationConfig, notificationCount int64) interface{} {
	return renderNotification(notificationConfig, n.renderSMS(), notificationCount)
}

type rxTransmissionErrorNotificationView int64

func (rxTransmissionErrorNotificationView) renderEmail() string {
	// TODO
	return ""
}

func (rxTransmissionErrorNotificationView) renderSMS() string {
	return "There was an error routing prescription to pharmacy"
}

func (r rxTransmissionErrorNotificationView) renderPush(notificationConfig *config.NotificationConfig, notificationCount int64) interface{} {
	return renderNotification(notificationConfig, r.renderSMS(), notificationCount)
}

type refillRxCreatedNotificationView int64

func (refillRxCreatedNotificationView) renderEmail() string {
	// TODO
	return ""
}

func (refillRxCreatedNotificationView) renderSMS() string {
	return "You have a new refill request from a patient"
}

func (r refillRxCreatedNotificationView) renderPush(notificationConfig *config.NotificationConfig, notificationCount int64) interface{} {
	return renderNotification(notificationConfig, r.renderSMS(), notificationCount)
}

func renderNotification(notificationConfig *config.NotificationConfig, message string, badgeCount int64) *snsNotification {
	snsNote := &snsNotification{
		DefaultMessage: message,
	}
	switch notificationConfig.Platform {
	case common.Android:
		snsNote.Android = &androidPushNotification{
			Message: snsNote.DefaultMessage,
		}

	case common.IOS:
		iosNotification := &iOSPushNotification{
			Badge: badgeCount,
			Alert: snsNote.DefaultMessage,
		}
		if notificationConfig.IsApnsSandbox {
			snsNote.IOSSandBox = iosNotification
		} else {
			snsNote.IOS = iosNotification
		}
	}

	return snsNote
}
