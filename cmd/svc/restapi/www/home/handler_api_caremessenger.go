package home

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sprucehealth/backend/cmd/svc/restapi/common/config"
	"github.com/sprucehealth/backend/cmd/svc/restapi/internal/httputil"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/golog"
)

type messengerBetaRequestAPIHandler struct {
	cfg cfg.Store
}

type messengerWhitePaperRequestAPIHandler struct {
	cfg cfg.Store
}

type messengerContactUsAPIHandler struct {
	cfg cfg.Store
}

type betaWhitePaperPostRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	Practice  string `json:"practice"`
}

func (d *betaWhitePaperPostRequest) Validate() error {
	if d.FirstName == "" {
		return errors.New("Please enter your first name.")
	}
	if d.LastName == "" {
		return errors.New("Please enter your last name.")
	}
	if d.Email == "" {
		return errors.New("Please enter your email address.")
	}
	if d.Phone == "" {
		return errors.New("Please enter your phone number.")
	}
	if d.Role == "" {
		return errors.New("Please enter your role.")
	}
	return nil
}

var messengerWhitepaperRequestSlackWebhookURLDef = &cfg.ValueDef{
	Name:        "SlackURL.Webhook.CareMessenger.WhitepaperRequest",
	Description: "A Slack webhook URL to post the details of a person requesting the Care Messenger whitepaper.",
	Type:        cfg.ValueTypeString,
	Default:     "",
}

var messengerBetaRequestSlackWebhookURLDef = &cfg.ValueDef{
	Name:        "SlackURL.Webhook.CareMessenger.BetaRequest",
	Description: "A Slack webhook URL to post the details of a person requesting beta access to Care Messenger.",
	Type:        cfg.ValueTypeString,
	Default:     "",
}

var messengerContactUsSlackWebhookURLDef = &cfg.ValueDef{
	Name:        "SlackURL.Webhook.CareMessenger.ContactUs",
	Description: "A Slack webhook URL to post the details of a person wanting to reach out to us.",
	Type:        cfg.ValueTypeString,
	Default:     "",
}

func init() {
	config.MustRegisterCfgDef(messengerBetaRequestSlackWebhookURLDef)
	config.MustRegisterCfgDef(messengerWhitepaperRequestSlackWebhookURLDef)
	config.MustRegisterCfgDef(messengerContactUsSlackWebhookURLDef)
}

func newCareMessengerBetaRequestAPIHandler(cfg cfg.Store) http.Handler {
	return httputil.SupportedMethods(&messengerBetaRequestAPIHandler{cfg: cfg}, httputil.Post)
}

func newCareMessengerWhitePaperRequestAPIHandler(cfg cfg.Store) http.Handler {
	return httputil.SupportedMethods(&messengerWhitePaperRequestAPIHandler{cfg: cfg}, httputil.Post)
}

func (h *messengerBetaRequestAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var d betaWhitePaperPostRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&d); err != nil {
		golog.Errorf("Error parsing Care Messenger Beta Request: %s", err.Error())
		www.APIBadRequestError(w, r, "We were unable to process your information. Please double check everything and try again.")
		return
	}
	err = d.Validate()
	if err != nil {
		www.APIBadRequestError(w, r, err.Error())
		return
	}
	textStrings := []string{
		"*New Care Messenger Beta Request*\n\n",
		"_First Name:_\n" + d.FirstName,
		"_Last Name:_\n" + d.LastName,
		"_Email:_\n" + d.Email,
		"_Phone:_\n" + d.Phone,
		"_Role:_\n" + d.Role,
		"_Practice:_\n" + d.Practice,
	}
	text := strings.Join(textStrings, "\n\n")

	url := h.cfg.Snapshot().String(messengerBetaRequestSlackWebhookURLDef.Name)
	if err := postToSlack("DERPebot", text, url); err != nil {
		golog.Errorf("Failed to post beta request form data to Slack; however, we did not return a 500. Error: %s", err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}

func (h *messengerWhitePaperRequestAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var d betaWhitePaperPostRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&d); err != nil {
		golog.Errorf("Error parsing Care Messenger Whitepaper Request: %s", err.Error())
		www.APIBadRequestError(w, r, "We were unable to process your information. Please double check everything and try again.")
		return
	}
	err = d.Validate()
	if err != nil {
		www.APIBadRequestError(w, r, err.Error())
		return
	}

	textStrings := []string{
		"*New Care Messenger Extension Whitepaper Download*\n\n",
		"_First Name:_\n" + d.FirstName,
		"_Last Name:_\n" + d.LastName,
		"_Email:_\n" + d.Email,
		"_Phone:_\n" + d.Phone,
		"_Role:_\n" + d.Role,
		"_Practice:_\n" + d.Practice,
	}
	text := strings.Join(textStrings, "\n\n")

	go func() {
		url := h.cfg.Snapshot().String(messengerWhitepaperRequestSlackWebhookURLDef.Name)
		if err = postToSlack("DERPebot", text, url); err != nil {
			// We silently fail because we don't want to let Slack errors block users from downloading the whitepaper
			golog.Errorf("Failed to post whitepaper request form data to Slack; however, we did not return a 500. Error: %s", err)
		}
	}()

	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}

type contactUsPostRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Practice string `json:"practice"`
	Message  string `json:"message"`
}

func newCareMessengerContactUsAPIHandler(cfg cfg.Store) http.Handler {
	return httputil.SupportedMethods(&messengerContactUsAPIHandler{cfg: cfg}, httputil.Post)
}

func (d *contactUsPostRequest) Validate() error {
	if d.Name == "" {
		return errors.New("Please enter your name.")
	}
	if d.Email == "" {
		return errors.New("Please enter your email address.")
	}
	if d.Phone == "" {
		return errors.New("Please enter your phone number.")
	}
	if d.Practice == "" {
		return errors.New("Please enter your practice name.")
	}
	if d.Message == "" {
		return errors.New("Please enter a message.")
	}
	return nil
}

func (h *messengerContactUsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var d contactUsPostRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&d); err != nil {
		golog.Errorf("Error parsing Care Messenger Contact Us Request: %s", err.Error())
		www.APIBadRequestError(w, r, "We were unable to process your information. Please double check everything and try again.")
		return
	}
	err = d.Validate()
	if err != nil {
		www.APIBadRequestError(w, r, err.Error())
		return
	}
	err = d.Validate()

	textStrings := []string{
		"*New Care Messenger Contact Us*\n\n",
		"_Name:_\n" + d.Name,
		"_Email:_\n" + d.Email,
		"_Phone:_\n" + d.Phone,
		"_Practice:_\n" + d.Practice,
		"_Message:_\n" + d.Message,
	}
	text := strings.Join(textStrings, "\n\n")

	go func() {
		url := h.cfg.Snapshot().String(messengerContactUsSlackWebhookURLDef.Name)
		if err = postToSlack("DERPebot", text, url); err != nil {
			// We silently fail because we don't want to let Slack errors block users from downloading the whitepaper
			golog.Errorf("Failed to post contact us form data to Slack; however, we did not return a 500. Error: %s", err)
		}
	}()

	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}
