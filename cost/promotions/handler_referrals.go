package promotions

import (
	"net/http"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/libs/httputil"
)

type referralProgramHandler struct {
	domain  string
	dataAPI api.DataAPI
}

// ShareTextInfo represents the data that the client should use to populate the various "Share" aspects of the app such as Refer A Friend via Facebook
type ShareTextInfo struct {
	EmailSubject string `json:"email_subject"`
	EmailBody    string `json:"email_body"`
	SMS          string `json:"sms"`
	Twitter      string `json:"twitter"`
	Facebook     string `json:"facebook"`
	Pasteboard   string `json:"pasteboard"`
	Default      string `json:"default"`
}

// ReferralDisplayInfo represents the display info to be consumed by the client when rendering the Refer A Friend views
type ReferralDisplayInfo struct {
	CTATitle           string          `json:"account_screen_cta_title"`
	NavBarTitle        string          `json:"nav_bar_title"`
	Title              string          `json:"title"`
	Body               string          `json:"body_text"`
	URLDisplayText     string          `json:"url_display_text"`
	URL                string          `json:"url"`
	ButtonTitle        string          `json:"button_title"`
	DismissButtonTitle string          `json:"dismiss_button_title"`
	ImageURL           string          `json:"image_url"`
	ImageWidth         int             `json:"image_width"`
	ImageHeight        int             `json:"image_height"`
	ShareText          *ShareTextInfo  `json:"share_text"`
	ReferralProgram    ReferralProgram `json:"-"`
}

// NewReferralProgramHandler returns a new instance of the referralProgramHandler
func NewReferralProgramHandler(dataAPI api.DataAPI, domain string) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(&referralProgramHandler{
				dataAPI: dataAPI,
				domain:  domain,
			}),
			api.RolePatient),
		httputil.Get)
}

func (p *referralProgramHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	account := apiservice.MustCtxAccount(ctx)
	referralDisplayInfo, err := CreateReferralDisplayInfo(p.dataAPI, p.domain, account.ID)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, referralDisplayInfo)
}
