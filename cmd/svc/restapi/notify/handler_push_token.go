package notify

import (
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common/config"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/device"
)

type notificationHandler struct {
	dataAPI             api.DataAPI
	notificationConfigs *config.NotificationConfigs
	snsClient           snsiface.SNSAPI
}

type requestData struct {
	DeviceToken string `schema:"device_token,required" json:"device_token"`
}

func NewNotificationHandler(dataAPI api.DataAPI, configs *config.NotificationConfigs, snsClient snsiface.SNSAPI) http.Handler {
	return httputil.SupportedMethods(
		apiservice.NoAuthorizationRequired(
			&notificationHandler{
				dataAPI:             dataAPI,
				notificationConfigs: configs,
				snsClient:           snsClient,
			}), httputil.Post)
}

func (n *notificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rData := &requestData{}
	if err := apiservice.DecodeRequestData(rData, r); err != nil {
		apiservice.WriteBadRequestError(err, w, r)
		return
	}

	if rData.DeviceToken == "" {
		apiservice.WriteValidationError("Device token required", w, r)
		return
	}

	sHeaders := device.ExtractSpruceHeaders(w, r)

	// we need the minimum headers set to be able to accept the token
	if sHeaders.Platform == "" || sHeaders.AppEnvironment == "" || sHeaders.AppType == "" {
		apiservice.WriteValidationError("Unable to determine which endpoint to use for push notifications: need platform, app-environment and app-type to be set in request header", w, r)
		return
	}

	// lookup the application config for configuring push notifications
	configName := config.DetermineNotificationConfigName(sHeaders.Platform, sHeaders.AppType, sHeaders.AppEnvironment)
	notificationConfig, err := n.notificationConfigs.Get(configName)
	if err != nil {
		apiservice.WriteError(errors.New("Unable to find right notification config for "+configName), w, r)
		return
	}

	// lookup any existing push config associated with this device token
	existingPushConfigData, err := n.dataAPI.GetPushConfigData(rData.DeviceToken)
	if err != nil && !api.IsErrNotFound(err) {
		apiservice.WriteError(errors.New("Unable to get push config data for device token: "+err.Error()), w, r)
		return
	}

	var pushEndpoint string
	if existingPushConfigData != nil {
		pushEndpoint = existingPushConfigData.PushEndpoint
	}

	// if the device token exists and has changed, register the device token for the user to get the application endpoint
	if existingPushConfigData == nil || rData.DeviceToken != existingPushConfigData.DeviceToken {
		res, err := n.snsClient.CreatePlatformEndpoint(&sns.CreatePlatformEndpointInput{
			PlatformApplicationArn: &notificationConfig.SNSApplicationEndpoint,
			Token: &rData.DeviceToken,
		})
		if err != nil {
			apiservice.WriteError(errors.New("Unable to register token for push notifications: "+err.Error()), w, r)
			return
		}
		pushEndpoint = *res.EndpointArn
	}

	newPushConfigData := &common.PushConfigData{
		AccountID:       apiservice.MustCtxAccount(r.Context()).ID,
		DeviceToken:     rData.DeviceToken,
		PushEndpoint:    pushEndpoint,
		Platform:        sHeaders.Platform,
		PlatformVersion: sHeaders.PlatformVersion,
		AppType:         sHeaders.AppType,
		AppEnvironment:  sHeaders.AppEnvironment,
		AppVersion:      sHeaders.AppVersion.String(),
		Device:          sHeaders.Device,
		DeviceModel:     sHeaders.DeviceModel,
		DeviceID:        sHeaders.DeviceID,
	}

	// update the device token for the user
	if err := n.dataAPI.SetOrReplacePushConfigData(newPushConfigData); err != nil {
		apiservice.WriteError(errors.New("Unable to update push config data: "+err.Error()), w, r)
		return
	}

	// return success
	apiservice.WriteJSONSuccess(w)
}
