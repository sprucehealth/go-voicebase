package notify

import "github.com/sprucehealth/backend/libs/golog"

func (n *NotificationManager) sendSMSToUser(toNumber, message string) error {
	if n.twilioClient == nil {
		return nil
	}

	go func() {
		_, _, err := n.twilioClient.Messages.SendSMS(n.fromNumber, toNumber, message)
		if err != nil {
			n.statSMSFailed.Inc(1)
			golog.Errorf("Error sending sms: %s", err.Error())
		} else {
			n.statSMSSent.Inc(1)
		}
	}()
	return nil
}
