package passreset

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/third_party/github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/www"
)

type promptHandler struct {
	r            *mux.Router
	dataAPI      api.DataAPI
	authAPI      api.AuthAPI
	emailService email.Service
	supportEmail string
	webDomain    string
	template     *template.Template
}

type verifyHandler struct {
	r                *mux.Router
	dataAPI          api.DataAPI
	authAPI          api.AuthAPI
	smsAPI           api.SMSAPI
	fromNumber       string
	supportEmail     string
	template         *template.Template
	statInvalidToken metrics.Counter
	statExpiredToken metrics.Counter
}

type resetHandler struct {
	r                *mux.Router
	dataAPI          api.DataAPI
	authAPI          api.AuthAPI
	emailService     email.Service
	supportEmail     string
	template         *template.Template
	statInvalidToken metrics.Counter
	statExpiredToken metrics.Counter
}

func SetupRoutes(r *mux.Router, dataAPI api.DataAPI, authAPI api.AuthAPI, smsAPI api.SMSAPI, fromNumber string, emailService email.Service, supportEmail, webDomain string, templateLoader *www.TemplateLoader, metricsRegistry metrics.Registry) {
	templateLoader.MustLoadTemplate("password_reset/base.html", "base.html", nil)

	ph := &promptHandler{
		r:            r,
		dataAPI:      dataAPI,
		authAPI:      authAPI,
		emailService: emailService,
		supportEmail: supportEmail,
		webDomain:    webDomain,
		template:     templateLoader.MustLoadTemplate("password_reset/prompt.html", "password_reset/base.html", nil),
	}

	vh := &verifyHandler{
		r:                r,
		dataAPI:          dataAPI,
		authAPI:          authAPI,
		smsAPI:           smsAPI,
		fromNumber:       fromNumber,
		supportEmail:     supportEmail,
		template:         templateLoader.MustLoadTemplate("password_reset/verify.html", "password_reset/base.html", nil),
		statInvalidToken: metrics.NewCounter(),
		statExpiredToken: metrics.NewCounter(),
	}
	metricsRegistry.Add("verify/fail/invalid_token", vh.statInvalidToken)
	metricsRegistry.Add("verify/fail/expired_token", vh.statExpiredToken)

	rh := &resetHandler{
		r:                r,
		dataAPI:          dataAPI,
		authAPI:          authAPI,
		emailService:     emailService,
		supportEmail:     supportEmail,
		template:         templateLoader.MustLoadTemplate("password_reset/reset.html", "password_reset/base.html", nil),
		statInvalidToken: metrics.NewCounter(),
		statExpiredToken: metrics.NewCounter(),
	}
	metricsRegistry.Add("reset/fail/invalid_token", vh.statInvalidToken)
	metricsRegistry.Add("reset/fail/expired_token", vh.statExpiredToken)

	r.Handle("/reset-password", ph).Name("reset-password-prompt")
	r.Handle("/reset-password/verify", vh).Name("reset-password-verify")
	r.Handle("/reset-password/password", rh).Name("reset-password")
}

func (h *promptHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: rate-limit this endpoint

	var errMsg string

	email := r.FormValue("email")
	if email != "" {
		account, err := h.authAPI.GetAccountForEmail(email)
		if err == api.LoginDoesNotExist {
			errMsg = "There is no account with the provided email. Check that it is entered correctly."
		} else if err != nil {
			www.InternalServerError(w, r, err)
			return
		} else if r.Method == "POST" {
			if err := SendPasswordResetEmail(h.authAPI, h.emailService, h.webDomain, account.ID, email, h.supportEmail); err != nil {
				golog.Errorf("Failed to send password reset email for account %d: %s", account.ID, err.Error())
				errMsg = "Failed to send email. Sorry for the inconvenience, and please try again later."
			} else {
				www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
					Title: "Password Reset | Spruce",
					SubContext: &promptTemplateContext{
						Email:        email,
						Sent:         true,
						SupportEmail: h.supportEmail,
					}})
				return
			}
		}
	} else if r.Method == "POST" {
		errMsg = "Please enter your email"
	}

	www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
		Title: "Password Reset | Spruce",
		SubContext: &promptTemplateContext{
			Email:        email,
			Error:        errMsg,
			SupportEmail: h.supportEmail,
		}})
}

func (h *verifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	account, token, emailAddress, rsent := validateToken(w, r, h.r, h.authAPI, api.LostPassword, h.statInvalidToken, h.statExpiredToken)
	if rsent {
		return
	}

	numbers, err := h.authAPI.GetPhoneNumbersForAccount(account.ID)
	if err != nil {
		www.InternalServerError(w, r, err)
		return
	}

	var toNumber string
	for _, n := range numbers {
		if n.Type == api.PHONE_CELL {
			toNumber = n.Phone.String()
			break
		}
	}

	var lastDigits string
	if len(toNumber) >= 2 {
		lastDigits = toNumber[len(toNumber)-2:]
	}

	if r.Method == "POST" {
		action := r.FormValue("action")
		switch action {
		case "send":
			contact := r.FormValue("method")
			if contact == "sms" {

				code, err := common.GenerateSMSCode()
				if err != nil {
					www.InternalServerError(w, r, err)
					return
				}
				if _, err := h.authAPI.CreateTempToken(account.ID, lostPasswordCodeExpires, api.LostPasswordCode, passResetToken(account.ID, code)); err != nil {
					www.InternalServerError(w, r, err)
					return
				}
				if err := h.smsAPI.Send(h.fromNumber, toNumber, fmt.Sprintf("Your Spruce verification code is %s", code)); err != nil {
					www.InternalServerError(w, r, err)
					return
				}
				www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
					Title: "Password Reset Verification | Spruce",
					SubContext: &verifyTemplateContext{
						Token:         token,
						Email:         emailAddress,
						LastTwoDigits: lastDigits,
						EnterCode:     true,
						SupportEmail:  h.supportEmail,
					}})
				return
			}
		case "validate":
			code := r.FormValue("code")
			codeToken := passResetToken(account.ID, code)
			_, err := h.authAPI.ValidateTempToken(api.LostPasswordCode, codeToken)
			if err != nil {
				switch err {
				case api.TokenExpired:
					h.statExpiredToken.Inc(1)
				case api.TokenDoesNotExist:
					h.statInvalidToken.Inc(1)
				default:
					www.InternalServerError(w, r, err)
					return
				}
				www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
					Title: "Password Reset Verification | Spruce",
					SubContext: &verifyTemplateContext{
						Token:         token,
						Email:         emailAddress,
						LastTwoDigits: lastDigits,
						EnterCode:     true,
						Code:          code,
						Errors:        []string{"Code is incorrect. Check to make sure it's typed correctly."},
						SupportEmail:  h.supportEmail,
					}})
				return
			}

			if err := h.authAPI.DeleteTempToken(api.LostPassword, token); err != nil {
				golog.Errorf("Failed to delete lost password token: %s", err.Error())
			}
			if err := h.authAPI.DeleteTempToken(api.LostPasswordCode, codeToken); err != nil {
				golog.Errorf("Failed to delete lost password code token: %s", err.Error())
			}

			resetToken, err := h.authAPI.CreateTempToken(account.ID, resetPasswordExpires, api.PasswordReset, "")
			if err != nil {
				www.InternalServerError(w, r, err)
				return
			}

			params := url.Values{
				"token": []string{resetToken},
			}
			if emailAddress != "" {
				params.Set("email", emailAddress)
			}
			u, err := h.r.Get("reset-password").URLPath()
			if err != nil {
				www.InternalServerError(w, r, err)
				return
			}
			u.RawQuery = params.Encode()
			http.Redirect(w, r, u.String(), http.StatusSeeOther)
			return
		}
	}

	www.TemplateResponse(w, http.StatusOK, h.template,
		&www.BaseTemplateContext{
			Title: "Password Reset Verification | Spruce",
			SubContext: &verifyTemplateContext{
				Token:         token,
				Email:         emailAddress,
				LastTwoDigits: lastDigits,
				SupportEmail:  h.supportEmail,
			}})
}

func (h *resetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	account, token, emailAddress, rsent := validateToken(w, r, h.r, h.authAPI, api.PasswordReset, h.statInvalidToken, h.statExpiredToken)
	if rsent {
		return
	}

	var errors []string
	var done bool
	if r.Method == "POST" {
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")
		if len(pass1) < api.MinimumPasswordLength {
			// TODO: further validation of length?
			errors = append(errors, fmt.Sprintf("Password must be longer than %d characters.", api.MinimumPasswordLength-1))
		} else if pass1 != pass2 {
			errors = append(errors, "Passwords do not match.")
		} else {
			if err := h.authAPI.SetPassword(account.ID, pass1); err != nil {
				www.InternalServerError(w, r, err)
				return
			}
			if err := h.authAPI.DeleteTempToken(api.PasswordReset, token); err != nil {
				golog.Errorf("Failed to delete password reset token: %s", err.Error())
			}
			done = true
			if err := SendPasswordHasBeenResetEmail(h.emailService, emailAddress, h.supportEmail); err != nil {
				golog.Errorf("Failed to send password reset success email: %s", err.Error())
			}
		}
	}
	www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
		Title: "Password Reset | Spruce",
		SubContext: &resetTemplateContext{
			Token:        token,
			Email:        emailAddress,
			Done:         done,
			Errors:       errors,
			SupportEmail: h.supportEmail,
		}})
}

func validateToken(w http.ResponseWriter, r *http.Request, router *mux.Router, authAPI api.AuthAPI, purpose string, statInvalidToken, statExpiredToken metrics.Counter) (*common.Account, string, string, bool) {
	token := r.FormValue("token")
	emailAddress := r.FormValue("email")
	var account *common.Account
	if token == "" {
		statInvalidToken.Inc(1)
	} else {
		var err error
		account, err = authAPI.ValidateTempToken(purpose, token)
		if err != nil {
			switch err {
			case api.TokenExpired:
				statExpiredToken.Inc(1)
			case api.TokenDoesNotExist:
				statInvalidToken.Inc(1)
			default:
				www.InternalServerError(w, r, err)
				return nil, token, emailAddress, true
			}
		}
	}
	if account == nil {
		// If the token is invalid then redirect to the reset-password page where
		// the person can request a new reset email.
		params := url.Values{}
		if emailAddress != "" {
			params.Set("email", emailAddress)
		}
		u, err := router.Get("reset-password-prompt").URLPath()
		if err != nil {
			www.InternalServerError(w, r, err)
			return nil, token, emailAddress, true
		}
		u.RawQuery = params.Encode()
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return nil, token, emailAddress, true
	}
	return account, token, emailAddress, false
}

func passResetToken(accountID int64, code string) string {
	return fmt.Sprintf("%d:%s", accountID, code)
}
