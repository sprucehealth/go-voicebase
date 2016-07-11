package notify

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common/config"
	"github.com/sprucehealth/backend/device"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/golog"
)

var jsonStr = "json"

type errorWithCode interface {
	Code() string
}

func (n *NotificationManager) pushNotificationToUser(
	accountID int64,
	role string,
	msg *Message,
	notificationCount int64,
) error {
	if n.snsClient == nil {
		golog.Errorf("No sns client configured when one was expected")
		return nil
	}

	// identify all devices associated with this user
	pushConfigDataList, err := n.dataAPI.GetPushConfigDataForAccount(accountID)
	if err != nil {
		return err
	}

	// render the notification and push for each device and send to each device
	for _, pushConfigData := range pushConfigDataList {

		// lookup config to use to determine endpoint to push to
		configName := config.DetermineNotificationConfigName(pushConfigData.Platform, pushConfigData.AppType, pushConfigData.AppEnvironment)
		notificationConfig, err := n.notificationConfigs.Get(configName)
		if err != nil {
			return err
		}

		pushEndpoint := pushConfigData.PushEndpoint
		// send push notifications in parallel
		conc.Go(func() {
			note := renderNotification(notificationConfig, msg, notificationCount)
			js, err := json.Marshal(note)
			if err != nil {
				n.statPushFailed.Inc(1)
				golog.Errorf("Failed to marshal SNS notification: %s", err)
				return
			}
			jsStr := string(js)
			_, err = n.snsClient.Publish(&sns.PublishInput{
				Message:          &jsStr,
				MessageStructure: &jsonStr,
				TargetArn:        &pushEndpoint,
			})
			if err != nil {
				if aError, ok := err.(errorWithCode); ok {
					// delete the preference to push notifications for user if the endpoint
					// is disabled such that we revert to other mechanisms for communicating with patient.
					if aError.Code() == "EndpointDisabled" {
						if err := n.dataAPI.DeletePushCommunicationPreferenceForAccount(accountID); err != nil {
							golog.Errorf("Unable to delete push preference for account %d. Error: %s", accountID, err)
							return
						}
					}
				} else {
					// don't return err so that we attempt to send push to as many devices as possible
					n.statPushFailed.Inc(1)
					golog.Errorf("Error sending push notification: %s", err)
				}
			} else {
				n.statPushSent.Inc(1)
			}
		})
	}

	return nil
}

func renderNotification(notificationConfig *config.NotificationConfig, message *Message, badgeCount int64) *snsNotification {
	snsNote := &snsNotification{
		DefaultMessage: message.ShortMessage,
	}
	switch notificationConfig.Platform {
	case device.Android:
		jsonData, err := json.Marshal(&androidPushNotification{
			Data: androidPushData{
				Message: snsNote.DefaultMessage,
				PushID:  message.PushID,
			},
		})
		if err != nil {
			golog.Infof("Unable to marshal json: %s", err)
		} else {
			snsNote.Android = string(jsonData)
		}

	case device.IOS:
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