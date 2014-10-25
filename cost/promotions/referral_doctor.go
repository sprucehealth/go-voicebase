package promotions

import (
	"errors"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
)

type routeDoctorReferralProgram struct {
	referralProgramParams
	Group                string                `json:"group"`
	RouteDoctorPromotion *routeDoctorPromotion `json:"route_doctor_promotion"`
	AssociatedCount      int                   `json:"associated_count"`
	SubmittedCount       int                   `json:"visit_submitted_count"`
}

func NewDoctorReferralProgram(accountID int64, title, description, group string,
	promotion *routeDoctorPromotion) ReferralProgram {
	return &routeDoctorReferralProgram{
		referralProgramParams: referralProgramParams{
			Title:          title,
			Description:    description,
			OwnerAccountID: accountID,
		},
		Group:                group,
		RouteDoctorPromotion: promotion,
	}
}

func (r *routeDoctorReferralProgram) TypeName() string {
	return routeWithDiscountReferralType
}

func (r *routeDoctorReferralProgram) Title() string {
	return r.referralProgramParams.Title
}

func (r *routeDoctorReferralProgram) Description() string {
	return r.referralProgramParams.Description
}

func (r *routeDoctorReferralProgram) SetOwnerAccountID(accountID int64) {
	r.OwnerAccountID = accountID
}

func (r *routeDoctorReferralProgram) Validate() error {
	if err := r.referralProgramParams.Validate(); err != nil {
		return err
	}

	if r.Group == "" {
		return errors.New("missing group")
	}

	if r.RouteDoctorPromotion == nil {
		return errors.New("missing route doctor promotion")
	}

	if err := r.RouteDoctorPromotion.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *routeDoctorReferralProgram) PromotionForReferredPatient(code string) *common.Promotion {
	return &common.Promotion{
		Code:  code,
		Group: r.Group,
		Data:  r.RouteDoctorPromotion,
	}
}

func (r *routeDoctorReferralProgram) ReferredPatientAssociatedCode(patientID, codeID int64, dataAPI api.DataAPI) error {
	r.AssociatedCount += 1
	if err := dataAPI.UpdateReferralProgram(r.referralProgramParams.OwnerAccountID, codeID, r); err != nil {
		return err
	}

	if err := dataAPI.TrackPatientReferral(&common.ReferralTrackingEntry{
		CodeID:             codeID,
		ClaimingPatientID:  patientID,
		ReferringAccountID: r.referralProgramParams.OwnerAccountID,
		Status:             common.RTSPending,
	}); err != nil {
		return err
	}

	return nil
}

func (r *routeDoctorReferralProgram) ReferredPatientSubmittedVisit(patientID, codeID int64, dataAPI api.DataAPI) error {

	r.SubmittedCount += 1
	if err := dataAPI.UpdateReferralProgram(r.referralProgramParams.OwnerAccountID, codeID, r); err != nil {
		return err
	}

	if err := dataAPI.UpdatePatientReferral(patientID, common.RTSCompleted); err != nil {
		return err
	}

	return nil
}

func (r *routeDoctorReferralProgram) UsersAssociatedCount() int {
	return r.AssociatedCount
}

func (r *routeDoctorReferralProgram) VisitsSubmittedCount() int {
	return r.SubmittedCount
}
