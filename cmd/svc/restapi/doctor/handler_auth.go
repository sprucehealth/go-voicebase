package doctor

import (
	"net/http"
	"strings"

	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/auth"
	"github.com/sprucehealth/backend/cmd/svc/restapi/responses"
	"github.com/sprucehealth/backend/device"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/ratelimit"
)

type authenticationHandler struct {
	authAPI              api.AuthAPI
	dataAPI              api.DataAPI
	apiDomain            string
	smsAPI               api.SMSAPI
	fromNumber           string
	dispatch             *dispatch.Dispatcher
	twoFactorExpiration  int
	dispatcher           *dispatch.Dispatcher
	rateLimiter          ratelimit.KeyedRateLimiter
	statLoginAttempted   *metrics.Counter
	statLoginSucceeded   *metrics.Counter
	statLogin2FARequired *metrics.Counter
	statLoginRateLimited *metrics.Counter
}

type AuthenticationRequestData struct {
	Email    string `schema:"email,required" json:"email"`
	Password string `schema:"password,required" json:"password"`
}

type AuthenticationResponse struct {
	Token             string            `json:"token,omitempty"`
	Doctor            *responses.Doctor `json:"doctor,omitempty"`
	LastFourPhone     string            `json:"last_four_phone,omitempty"`
	TwoFactorToken    string            `json:"two_factor_token,omitempty"`
	TwoFactorRequired bool              `json:"two_factor_required"`
}

func NewAuthenticationHandler(
	dataAPI api.DataAPI,
	authAPI api.AuthAPI,
	smsAPI api.SMSAPI,
	apiDomain string,
	dispatcher *dispatch.Dispatcher,
	fromNumber string,
	twoFactorExpiration int,
	rateLimiter ratelimit.KeyedRateLimiter,
	metricsRegistry metrics.Registry,
) http.Handler {
	h := &authenticationHandler{
		dataAPI:              dataAPI,
		authAPI:              authAPI,
		smsAPI:               smsAPI,
		apiDomain:            apiDomain,
		fromNumber:           fromNumber,
		twoFactorExpiration:  twoFactorExpiration,
		dispatcher:           dispatcher,
		rateLimiter:          rateLimiter,
		statLoginAttempted:   metrics.NewCounter(),
		statLoginSucceeded:   metrics.NewCounter(),
		statLogin2FARequired: metrics.NewCounter(),
		statLoginRateLimited: metrics.NewCounter(),
	}
	metricsRegistry.Add("login.attempted", h.statLoginAttempted)
	metricsRegistry.Add("login.succeeded", h.statLoginSucceeded)
	metricsRegistry.Add("login.2fa-required", h.statLogin2FARequired)
	metricsRegistry.Add("login.rate-limited", h.statLoginRateLimited)
	return httputil.SupportedMethods(apiservice.NoAuthorizationRequired(h), httputil.Post)
}

func (h *authenticationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.statLoginAttempted.Inc(1)

	// rate limit on IP address (prevent scanning accounts)
	if ok, err := h.rateLimiter.Check("login:"+r.RemoteAddr, 1); err != nil {
		golog.Errorf("Rate limit check failed: %s", err.Error())
	} else if !ok {
		h.statLoginRateLimited.Inc(1)
		apiservice.WriteAccessNotAllowedError(w, r)
		return
	}

	var requestData AuthenticationRequestData
	if err := apiservice.DecodeRequestData(&requestData, r); err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	requestData.Email = strings.TrimSpace(strings.ToLower(requestData.Email))

	// rate limit on account (prevent trying one account from multiple IPs)
	if ok, err := h.rateLimiter.Check("login:"+requestData.Email, 1); err != nil {
		golog.Errorf("Rate limit check failed: %s", err.Error())
	} else if !ok {
		h.statLoginRateLimited.Inc(1)
		apiservice.WriteAccessNotAllowedError(w, r)
		return
	}

	account, err := h.authAPI.Authenticate(requestData.Email, requestData.Password)
	if err != nil {
		switch err {
		case api.ErrLoginDoesNotExist, api.ErrInvalidPassword:
			apiservice.WriteUserError(w, http.StatusForbidden, "Invalid email/password combination")
			return
		}
		apiservice.WriteError(err, w, r)
		return
	}

	// Patient trying to sign in on doctor app
	if account.Role != api.RoleDoctor && account.Role != api.RoleCC {
		apiservice.WriteUserError(w, http.StatusForbidden, "Invalid email/password combination")
		return
	}

	if account.TwoFactorEnabled {
		appHeaders := device.ExtractSpruceHeaders(w, r)
		device, err := h.authAPI.GetAccountDevice(account.ID, appHeaders.DeviceID)
		if err != nil && !api.IsErrNotFound(err) {
			apiservice.WriteError(err, w, r)
			return
		}
		if device == nil || !device.Verified {
			// Create a temporary token to the client can use to authenticate the code submission request
			token, err := h.authAPI.CreateTempToken(account.ID, h.twoFactorExpiration, api.TwoFactorAuthToken, "")
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}

			phone, err := auth.SendTwoFactorCode(h.authAPI, h.smsAPI, h.fromNumber, account.ID, appHeaders.DeviceID, h.twoFactorExpiration)
			if err != nil {
				apiservice.WriteError(err, w, r)
				return
			}

			h.statLogin2FARequired.Inc(1)
			h.statLoginSucceeded.Inc(1)

			httputil.JSONResponse(w, http.StatusOK, &AuthenticationResponse{
				LastFourPhone:     phone[len(phone)-4:],
				TwoFactorToken:    token,
				TwoFactorRequired: true,
			})
			return
		}
	}

	var ctOpt api.CreateTokenOption
	if account.Role == api.RoleCC {
		ctOpt |= api.CreateTokenAllowMany
	}
	token, err := h.authAPI.CreateToken(account.ID, api.Mobile, ctOpt)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	doctor, err := h.dataAPI.GetDoctorFromAccountID(account.ID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	h.dispatcher.Publish(&DoctorLoggedInEvent{
		Doctor: doctor,
	})

	headers := device.ExtractSpruceHeaders(w, r)
	h.dispatcher.PublishAsync(&auth.AuthenticatedEvent{
		AccountID:     doctor.AccountID.Int64(),
		SpruceHeaders: headers,
	})

	h.statLoginSucceeded.Inc(1)

	httputil.JSONResponse(w, http.StatusOK, &AuthenticationResponse{
		Token:  token,
		Doctor: responses.TransformDoctor(doctor, h.apiDomain)})
}
