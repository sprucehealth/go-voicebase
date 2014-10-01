package www

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/auth"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/third_party/github.com/samuel/go-metrics/metrics"
)

type loginHandler struct {
	authAPI                   api.AuthAPI
	smsAPI                    api.SMSAPI
	template                  *template.Template
	fromNumber                string
	twoFactorExpiration       int
	statFailure               metrics.Counter
	statSuccess2FARequired    metrics.Counter
	statSuccess2FANotRequired metrics.Counter
	statSuccess2FAVerified    metrics.Counter
}

func NewLoginHandler(authAPI api.AuthAPI, smsAPI api.SMSAPI, fromNumber string, twoFactorExpiration int, templateLoader *TemplateLoader, metricsRegistry metrics.Registry) http.Handler {
	h := &loginHandler{
		authAPI:                   authAPI,
		smsAPI:                    smsAPI,
		fromNumber:                fromNumber,
		twoFactorExpiration:       twoFactorExpiration,
		template:                  templateLoader.MustLoadTemplate("login.html", "base.html", nil),
		statSuccess2FARequired:    metrics.NewCounter(),
		statSuccess2FANotRequired: metrics.NewCounter(),
		statSuccess2FAVerified:    metrics.NewCounter(),
		statFailure:               metrics.NewCounter(),
	}
	metricsRegistry.Add("failure", h.statFailure)
	metricsRegistry.Add("success-2fa-required", h.statSuccess2FARequired)
	metricsRegistry.Add("success-2fa-not-required", h.statSuccess2FANotRequired)
	metricsRegistry.Add("success-2fa-verified", h.statSuccess2FAVerified)
	return httputil.SupportedMethods(h, []string{"GET", "POST"})
}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: rate-limit this endpoint

	email := r.FormValue("email")
	next, valid := validateRedirectURL(r.FormValue("next"))
	if !valid {
		next = "/"
	}

	var errorMessage string

	if r.Method == "POST" {
		password := r.PostFormValue("password")
		account, err := h.authAPI.Authenticate(email, password)
		if err != nil {
			switch err {
			case api.LoginDoesNotExist, api.InvalidPassword:
				h.statFailure.Inc(1)
				errorMessage = "Email or password is not valid."
			default:
				InternalServerError(w, r, err)
				return
			}
		} else if account.TwoFactorEnabled {
			var deviceID string
			deviceIDCookie, err := r.Cookie(deviceIDCookieName)
			if err == nil && len(deviceIDCookie.Value) >= common.MinimumTokenLength {
				deviceID = deviceIDCookie.Value

				// See if this device ID is already verified
				//
				// TODO: For now two factor is permanent as long as the device ID cookie remains the same.
				// We should require two factor again after some amount of time.
				device, err := h.authAPI.GetAccountDevice(account.ID, deviceID)
				if err != nil && err != api.NoRowsError {
					InternalServerError(w, r, err)
					return
				} else if device != nil && device.Verified {
					h.statSuccess2FAVerified.Inc(1)
					authenticateResponse(w, r, h.authAPI, account, next)
					return
				}
			} else {
				deviceID, err = common.GenerateToken()
				if err != nil {
					InternalServerError(w, r, err)
					return
				}
				http.SetCookie(w, NewCookie(deviceIDCookieName, deviceID, r))
			}

			h.statSuccess2FARequired.Inc(1)

			token, err := h.authAPI.CreateTempToken(account.ID, h.twoFactorExpiration, api.TwoFactorAuthToken, "")
			if err != nil {
				InternalServerError(w, r, err)
				return
			}

			if _, err := auth.SendTwoFactorCode(h.authAPI, h.smsAPI, h.fromNumber, account.ID, deviceID, h.twoFactorExpiration); err != nil {
				// TODO: return a user friendly error because this could be a bad cell phone number
				InternalServerError(w, r, err)
				return
			}

			params := url.Values{
				"next": []string{next},
				"t":    []string{token},
			}
			ur := "/login/verify?" + params.Encode()
			http.Redirect(w, r, ur, http.StatusSeeOther)
			return
		} else {
			h.statSuccess2FANotRequired.Inc(1)
			authenticateResponse(w, r, h.authAPI, account, next)
			return
		}
	}

	TemplateResponse(w, http.StatusOK, h.template, &BaseTemplateContext{
		Title: "Login | Spruce",
		SubContext: &struct {
			Email string
			Next  string
			Error string
		}{
			Error: errorMessage,
			Email: email,
			Next:  next,
		},
	})
}

// login verification

type loginVerifyHandler struct {
	authAPI                 api.AuthAPI
	template                *template.Template
	statSuccess             metrics.Counter
	statFailureInvalidToken metrics.Counter
	statFailureInvalidCode  metrics.Counter
	statFailureExpired      metrics.Counter
}

func NewLoginVerifyHandler(authAPI api.AuthAPI, templateLoader *TemplateLoader, metricsRegistry metrics.Registry) http.Handler {
	h := &loginVerifyHandler{
		authAPI:                 authAPI,
		template:                templateLoader.MustLoadTemplate("login_verify.html", "base.html", nil),
		statSuccess:             metrics.NewCounter(),
		statFailureInvalidToken: metrics.NewCounter(),
		statFailureInvalidCode:  metrics.NewCounter(),
		statFailureExpired:      metrics.NewCounter(),
	}
	metricsRegistry.Add("success", h.statSuccess)
	metricsRegistry.Add("failure-invalid-token", h.statFailureInvalidToken)
	metricsRegistry.Add("failure-invalid-code", h.statFailureInvalidCode)
	metricsRegistry.Add("failure-expired", h.statFailureExpired)
	return httputil.SupportedMethods(h, []string{"GET", "POST"})
}

func (h *loginVerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: rate-limit this endpoint and only allow a small number of attempts

	next, valid := validateRedirectURL(r.FormValue("next"))
	if !valid {
		next = "/"
	}

	var deviceID string

	deviceIDCookie, err := r.Cookie(deviceIDCookieName)
	if err == nil && len(deviceIDCookie.Value) >= common.MinimumTokenLength {
		deviceID = deviceIDCookie.Value
	}

	tempToken := r.FormValue("t")
	if deviceID == "" || len(tempToken) < common.MinimumTokenLength {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	account, err := h.authAPI.ValidateTempToken(api.TwoFactorAuthToken, tempToken)
	if err == api.TokenDoesNotExist {
		h.statFailureInvalidToken.Inc(1)
	} else if err == api.TokenExpired {
		h.statFailureExpired.Inc(1)
	} else if err != nil {
		InternalServerError(w, r, err)
		return
	}
	if account == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var errorMessage string

	if r.Method == "POST" {
		code := r.PostFormValue("code")
		codeToken := auth.TwoFactorCodeToken(account.ID, deviceID, code)
		account2, err := h.authAPI.ValidateTempToken(api.TwoFactorAuthCode, codeToken)
		if err == api.TokenDoesNotExist {
			errorMessage = "Invalid verification code"
			h.statFailureInvalidCode.Inc(1)
		} else if err == api.TokenExpired {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			h.statFailureExpired.Inc(1)
			return
		} else if err != nil {
			InternalServerError(w, r, err)
			return
		} else if account2.ID != account.ID {
			// This should never ever happen but good to make sure
			golog.Errorf("Accounts don't match: %d != %d", account.ID, account2.ID)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if errorMessage == "" {
			go func() {
				// Mark this "device" as being verified with two factor
				if err := h.authAPI.UpdateAccountDeviceVerification(account.ID, deviceID, true); err != nil {
					golog.Errorf(err.Error())
				}
				if err := h.authAPI.DeleteTempToken(api.TwoFactorAuthCode, codeToken); err != nil {
					golog.Errorf(err.Error())
				}
				if err := h.authAPI.DeleteTempToken(api.TwoFactorAuthToken, tempToken); err != nil {
					golog.Errorf(err.Error())
				}
			}()

			h.statSuccess.Inc(1)
			authenticateResponse(w, r, h.authAPI, account, next)
			return
		}
	}

	numbers, err := h.authAPI.GetPhoneNumbersForAccount(account.ID)
	if err != nil {
		InternalServerError(w, r, err)
		return
	}

	var toNumber string
	for _, n := range numbers {
		if n.Type == api.PHONE_CELL {
			toNumber = n.Phone.String()
			break
		}
	}
	if len(toNumber) < 10 {
		// Shouldn't happen since a account should never have been enabled for two factor
		// if it didn't have a cellphone number attached. However, covering it just to be safe.
		errorMessage = "This account has no cell phone number. Please contact support at support@sprucehealth.com."
		golog.Errorf("Account %d has two factor enabled but no valid cell phone number", account.ID)
	} else {
		toNumber = toNumber[len(toNumber)-2:]
	}

	TemplateResponse(w, http.StatusOK, h.template, &BaseTemplateContext{
		Title: "Login Verification | Spruce",
		SubContext: &struct {
			Next         string
			Error        string
			LastTwoPhone string
		}{
			Next:         next,
			Error:        errorMessage,
			LastTwoPhone: toNumber,
		},
	})
}

// logout

type logoutHandler struct {
	authAPI api.AuthAPI
}

func NewLogoutHandler(authAPI api.AuthAPI) http.Handler {
	return httputil.SupportedMethods(&logoutHandler{
		authAPI: authAPI,
	}, []string{"GET", "POST"})
}

func (h *logoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	next, valid := validateRedirectURL(r.FormValue("next"))
	if !valid {
		next = "/"
	}

	http.SetCookie(w, TomestoneAuthCookie(r))
	http.Redirect(w, r, next, http.StatusSeeOther)
}

//

func authenticateResponse(w http.ResponseWriter, r *http.Request, authAPI api.AuthAPI, account *common.Account, next string) {
	// Must redirect somewhere
	if next == "" {
		next = "/"
	}
	// The root is rarely the place anyone wants to go so redirect appropriately
	// based on the role of the account.
	if next == "/" {
		switch account.Role {
		case api.ADMIN_ROLE:
			next = "/admin"
		}
	}

	token, err := authAPI.CreateToken(account.ID, api.Web, api.RegularAuth)
	if err != nil {
		InternalServerError(w, r, err)
		return
	}
	http.SetCookie(w, NewAuthCookie(token, r))
	http.Redirect(w, r, next, http.StatusSeeOther)
}
