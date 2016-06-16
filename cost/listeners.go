package cost

import (
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/cost/promotions"
	"github.com/sprucehealth/backend/libs/dispatch"
)

func InitListeners(dataAPI api.DataAPI, dispatcher *dispatch.Dispatcher) {
	dispatcher.SubscribeAsync(func(ev *VisitChargedEvent) error {
		// looking up any existing referral tracking entry for this patient
		referralTrackingEntry, err := dataAPI.PendingReferralTrackingForAccount(ev.AccountID)
		if api.IsErrNotFound(err) {
			// nothing to do here since there is no feedback to give
			return nil
		} else if err != nil {
			return err
		}

		// lookup the referral program
		referralProgram, err := dataAPI.ReferralProgram(referralTrackingEntry.CodeID, common.PromotionTypes)
		if err != nil {
			return err
		}

		// update the referral program to indicate that the referred patient
		// submitted a visit
		if err := referralProgram.Data.(promotions.ReferralProgram).
			ReferredAccountSubmittedVisit(ev.AccountID, referralTrackingEntry.CodeID, dataAPI); err != nil {
			return err
		}
		return nil
	})
}
