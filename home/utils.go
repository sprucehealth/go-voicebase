package home

import (
	"errors"
	"fmt"

	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/common"
)

type homeState int64

const (
	noAccountState homeState = iota
	noCaseState
	casesExistState
)

func getHomeCards(hState homeState, ctxt map[string]interface{}) ([]PHView, error) {
	var views []PHView

	switch hState {
	case noAccountState:
		views = append(views, getStartVisitCard())
		views = append(views, getDefaultCards()...)
	case noCaseState:
		views = append(views, getStartVisitCard())
		views = append(views, getDefaultCards()...)
	case casesExistState:
	default:
		return nil, errors.New("Unidentified home state")
	}

	for _, v := range views {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	return views, nil
}

func getStartVisitCard() PHView {
	return &PHPrimaryActionView{
		Title:       "Start an Acne Visit",
		ActionURL:   app_url.StartVisitAction(),
		ButtonTitle: "Get Started",
	}
}

func getCompleteVisitCard(patientVisitId int64) PHView {
	return &PHPrimaryActionView{
		Title:     "Continue acne visit",
		ActionURL: app_url.ContinueVisitAction(patientVisitId),
	}
}

func getViewCaseCard(patientCase *common.PatientCase, notificationView PHView) PHView {
	return &PHCaseView{
		Title:            "View Acne Case",
		Subtitle:         fmt.Sprintf("Started on %s", patientCase.CreationDate.Format(apiservice.TimeFormatLayout)),
		ActionURL:        app_url.ViewCaseAction(patientCase.Id.Int64()),
		NotificationView: notificationView,
	}
}

func getSampleTreatmentPlanCard() PHView {
	return &PHSmallIconText{
		Title:       "See a sample treatment plan",
		IconURL:     app_url.IconBlueTreatmentPlan,
		ActionURL:   app_url.ViewSampleTreatmentPlanAction(),
		RoundedIcon: true,
	}
}

func getSeeSpruceDermsCard() PHView {
	return &PHSmallIconText{
		Title:       "Meet the Spruce Dermatologists",
		IconURL:     app_url.IconSpruceDoctors,
		ActionURL:   app_url.ViewSampleDoctorProfilesAction(),
		RoundedIcon: true,
	}
}

func getLearnSpruceCard() PHView {
	return &PHSmallIconText{
		Title:       "Learn how a Spruce Visit Works",
		IconURL:     app_url.IconLearnSpruce,
		ActionURL:   app_url.ViewTutorialAction(),
		RoundedIcon: true,
	}
}

func getDefaultCards() []PHView {
	return []PHView{
		getSeeSpruceDermsCard(),
		getLearnSpruceCard(),
		getSampleTreatmentPlanCard(),
	}
}
