package promotions

import (
	"errors"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/common"
)

type giveReferralProgram struct {
	referralProgramParams
	Group           string                  `json:"group"`
	Promotion       *moneyDiscountPromotion `json:"promotion"`
	AssociatedCount int                     `json:"associated_count"`
	SubmittedCount  int                     `json:"visit_submitted_count"`
}

func (g *giveReferralProgram) TypeName() string {
	return giveReferralType
}

func (g *giveReferralProgram) HomeCardText() string {
	if g.referralProgramParams.HomeCard == nil {
		return ""
	}
	return g.referralProgramParams.HomeCard.Text
}

func (g *giveReferralProgram) HomeCardImageURL() *app_url.SpruceAsset {
	if g.referralProgramParams.HomeCard == nil {
		return app_url.IconPromoLogo
	}
	return g.referralProgramParams.HomeCard.ImageURL
}

func (g *giveReferralProgram) Title() string {
	return g.referralProgramParams.Title
}

func (g *giveReferralProgram) Description() string {
	return g.referralProgramParams.Description
}

func (g *giveReferralProgram) ShareTextInfo() *ShareTextParams {
	return g.referralProgramParams.ShareText
}

func (g *giveReferralProgram) SetOwnerAccountID(accountID int64) {
	g.OwnerAccountID = accountID
}

func (g *giveReferralProgram) Validate() error {
	if err := g.referralProgramParams.Validate(); err != nil {
		return err
	}

	if g.Group == "" {
		return errors.New("missing group")
	}

	if g.Promotion == nil {
		return errors.New("missing promotion on referral")
	}

	if err := g.Promotion.Validate(); err != nil {
		return err
	}

	return nil
}

func NewGiveReferralProgram(title, description, group string, homeCard *HomeCardConfig, promotion *moneyDiscountPromotion, shareTextParams *ShareTextParams) ReferralProgram {
	return &giveReferralProgram{
		referralProgramParams: referralProgramParams{
			Title:       title,
			Description: description,
			ShareText:   shareTextParams,
			HomeCard:    homeCard,
		},
		Group:     group,
		Promotion: promotion,
	}
}

func (g *giveReferralProgram) PromotionForReferredAccount(code string) *common.Promotion {
	return &common.Promotion{
		Code:  code,
		Group: g.Group,
		Data:  g.Promotion,
	}
}

func (g *giveReferralProgram) ReferredAccountAssociatedCode(accountID, codeID int64, dataAPI api.DataAPI) error {
	//  update the associated count for the original promotion and update the database
	g.AssociatedCount += 1
	if err := dataAPI.UpdateReferralProgram(g.referralProgramParams.OwnerAccountID, codeID, g); err != nil {
		return err
	}

	if err := dataAPI.TrackAccountReferral(&common.ReferralTrackingEntry{
		CodeID:             codeID,
		ClaimingAccountID:  accountID,
		ReferringAccountID: g.referralProgramParams.OwnerAccountID,
		Status:             common.RTSPending,
	}); err != nil {
		return err
	}

	return nil
}

func (g *giveReferralProgram) ReferredAccountSubmittedVisit(accountID, codeID int64, dataAPI api.DataAPI) error {

	g.SubmittedCount += 1
	if err := dataAPI.UpdateReferralProgram(g.referralProgramParams.OwnerAccountID, codeID, g); err != nil {
		return err
	}

	if err := dataAPI.UpdateAccountReferral(accountID, common.RTSCompleted); err != nil {
		return err
	}

	return nil
}

func (g *giveReferralProgram) UsersAssociatedCount() int {
	return g.AssociatedCount
}

func (g *giveReferralProgram) VisitsSubmittedCount() int {
	return g.SubmittedCount
}
