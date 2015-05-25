package patient_case

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/cost/promotions"
	"github.com/sprucehealth/backend/errors"
	"github.com/sprucehealth/backend/responses"
)

type auxillaryHomeCard int

const (
	learnMoreAboutSpruceCard auxillaryHomeCard = 1 << iota
	contactUsCard
	referralCard
	careTeamCard
	noCards = 0
)

func getHomeCards(cases []*common.PatientCase,
	cityStateInfo *address.CityState,
	isSpruceAvailable bool,
	dataAPI api.DataAPI,
	apiCDNDomain string,
	webDomain string,
	r *http.Request,
) ([]common.ClientView, error) {
	var views []common.ClientView
	var err error

	if len(cases) == 0 {
		views, err = homeCardsForUnAuthenticatedUser(dataAPI, cityStateInfo, isSpruceAvailable, r)
	} else {
		views, err = homeCardsForAuthenticatedUser(dataAPI, cases, cityStateInfo, apiCDNDomain, webDomain, r)
	}

	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, v := range views {
		if v == nil {
			continue
		}
		if err := v.Validate(); err != nil {
			return nil, errors.Trace(err)
		}
	}

	return views, nil
}

func homeCardsForUnAuthenticatedUser(
	dataAPI api.DataAPI,
	cityStateInfo *address.CityState,
	isSpruceAvailable bool,
	r *http.Request,
) ([]common.ClientView, error) {

	// TODO: assume Acne
	pathway, err := dataAPI.PathwayForTag(api.AcnePathwayTag, api.PONone)
	if err != nil {
		return nil, err
	}

	views := make([]common.ClientView, 2)
	views[1] = getLearnAboutSpruceSection(pathway.Tag)

	if isSpruceAvailable {
		views[0] = getStartVisitCard()
	} else {
		spruceHeaders := apiservice.ExtractSpruceHeaders(r)
		entryExists, err := dataAPI.FormEntryExists("form_notify_me", spruceHeaders.DeviceID)
		if err != nil {
			return nil, err
		}

		if entryExists {
			views[0] = getNotifyMeConfirmationCard(cityStateInfo.State)
		} else {
			views[0] = getNotifyMeCard(cityStateInfo.State)
		}
	}

	return views, nil
}

func homeCardsForAuthenticatedUser(
	dataAPI api.DataAPI,
	cases []*common.PatientCase,
	cityStateInfo *address.CityState,
	apiCDNDomain string,
	webDomain string,
	r *http.Request,
) ([]common.ClientView, error) {

	// If we're authenticated then we need account information to look up appropriate promotions/refer a friend
	patient, err := dataAPI.Patient(cases[0].PatientID.Int64(), true)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// get notifications for all cases for a patient
	notificationMap, err := dataAPI.NotificationsForCases(patient.PatientID.Int64(), NotifyTypes)
	if err != nil {
		return nil, err
	}

	// get the care teams for all cases for a patient
	caseIDs := make([]int64, len(cases))
	for i, pc := range cases {
		caseIDs[i] = pc.ID.Int64()
	}
	careTeams, err := dataAPI.CaseCareTeams(caseIDs)
	if err != nil {
		return nil, err
	}

	var views []common.ClientView
	var auxillaryCardOptions auxillaryHomeCard
	var caseWithCompletedVisit bool

	// only show the care team if the patient has a case for which:
	// a) they have submitted a visit
	// b) they have not received a TP OR they have recieved but not viewed their TP
	if len(cases) == 1 {
		visits, err := dataAPI.GetVisitsForCase(cases[0].ID.Int64(), common.TreatedPatientVisitStates())
		if err != nil {
			return nil, err
		}

		if len(visits) == 1 {
			tps, err := dataAPI.GetTreatmentPlansForCase(cases[0].ID.Int64())
			if err != nil {
				return nil, err
			}

			if len(tps) == 0 || (len(tps) == 1 && !tps[0].PatientViewed) {
				auxillaryCardOptions |= careTeamCard
			}
		}
	}

	// iterate through the cases to populate the view for each case card
	for _, patientCase := range cases {

		caseNotifications := notificationMap[patientCase.ID.Int64()]
		assignments := careTeams[patientCase.ID.Int64()].Assignments

		// get current doctor assigned to case
		var doctorAssignment, maAssignment *common.CareProviderAssignment
		for _, assignment := range assignments {
			if assignment.Status != api.StatusActive {
				continue
			}
			switch assignment.ProviderRole {
			case api.RoleDoctor:
				doctorAssignment = assignment
			case api.RoleMA:
				maAssignment = assignment
			}
		}

		// identify the number of renderable case notifications to display the count
		// as the call to action is to view the case details page and the notification
		// count on the home card should map to the number of renderable case notifications
		var renderableCaseNotifications int64
		for _, notificationItem := range caseNotifications {
			if notificationItem.Data.(notification).canRenderCaseNotificationView() {
				renderableCaseNotifications++
			}
		}

		// populate home cards based on the notification types and the number of notifications in the case
		switch l := renderableCaseNotifications; {

		case len(caseNotifications) == 1, l == 1:
			hView, err := caseNotifications[0].Data.(notification).makeHomeCardView(&caseData{
				APIDomain:       apiCDNDomain,
				Notification:    caseNotifications[0],
				CareTeamMembers: assignments,
				Case:            patientCase,
			})
			if err != nil {
				return nil, err
			}

			switch caseNotifications[0].NotificationType {

			case CNPreSubmissionTriage:
				views = append(views, hView)

			case CNIncompleteVisit:
				views = append(views, hView)
				auxillaryCardOptions |= contactUsCard

			case CNVisitSubmitted, CNTreatmentPlan, CNStartFollowup, CNIncompleteFollowup, CNMessage:
				views = append(views, getViewCaseCard(patientCase, doctorAssignment, hView))
				auxillaryCardOptions |= referralCard
				caseWithCompletedVisit = true
			}

		case l > 1:

			// treating the fact that multiple notifications exist to indicate that the patient
			// has completed a visit within a case.
			// NOTE: This might be fragile in that
			// we may change the functionality in the future where there could be situations with no CTA
			// when the user has not started or completed a visit, but its a lot more expensive to figure out
			// if a visit has been completed or not
			caseWithCompletedVisit = true

			auxillaryCardOptions |= referralCard

			a := maAssignment
			if doctorAssignment != nil {
				a = doctorAssignment
			}

			views = append(views, getViewCaseCard(patientCase, doctorAssignment, &phCaseNotificationStandardView{
				Title:       "You have" + spellNumber(int(l)) + "new updates.",
				ButtonTitle: "View Case",
				ActionURL:   app_url.ViewCaseAction(patientCase.ID.Int64()),
				IconURL:     app_url.ThumbnailURL(apiCDNDomain, a.ProviderRole, a.ProviderID),
			}))

		case l == 0:

			// treating the lack of a notification to indicate that the patient has completed a visit
			// within a case.
			// NOTE: This might be fragile in that
			// we may change the functionality in the future where there could be situations with no CTA
			// when the user has not started or completed a visit, but its a lot more expensive to figure out
			// if a visit has been completed or not
			caseWithCompletedVisit = true

			auxillaryCardOptions |= referralCard

			imageURL := app_url.IconCaseLarge.String()
			if doctorAssignment != nil {
				imageURL = app_url.ThumbnailURL(apiCDNDomain, doctorAssignment.ProviderRole, doctorAssignment.ProviderID)
			}

			views = append(views,
				getViewCaseCard(patientCase, doctorAssignment, &phCaseNotificationStandardView{
					ButtonTitle: "View Case",
					ActionURL:   app_url.ViewCaseAction(patientCase.ID.Int64()),
					IconURL:     imageURL,
				}),
			)
		}
	}

	// only show the learn more about spruce section if there is no case with a completed visit
	if !caseWithCompletedVisit {
		auxillaryCardOptions |= learnMoreAboutSpruceCard
	}

	if auxillaryCardOptions&careTeamCard != 0 {
		views = append(views, getMeetCareTeamSection(careTeams[cases[0].ID.Int64()].Assignments, cases[0], apiCDNDomain))
	}
	if auxillaryCardOptions&referralCard != 0 {
		spruceHeaders := apiservice.ExtractSpruceHeaders(r)
		view, err := getShareSpruceSection(spruceHeaders.AppVersion, dataAPI, webDomain, patient.AccountID.Int64())
		if err != nil {
			return nil, err
		} else if view != nil {
			views = append(views, view)
		}
	}
	if auxillaryCardOptions&contactUsCard != 0 {
		views = append(views, getSendUsMessageSection())
	}
	if auxillaryCardOptions&learnMoreAboutSpruceCard != 0 {
		// TODO: For now default to showing the sample treatment plan for Acne
		views = append(views, getLearnAboutSpruceSection(api.AcnePathwayTag))
	}
	return views, nil
}

func getViewCaseCard(patientCase *common.PatientCase, careProvider *common.CareProviderAssignment, notificationView common.ClientView) common.ClientView {
	if patientCase.Claimed {
		return &phCaseView{
			Title:            fmt.Sprintf("%s Case", patientCase.Name),
			Subtitle:         fmt.Sprintf("With %s", careProvider.ShortDisplayName),
			ActionURL:        app_url.ViewCaseAction(patientCase.ID.Int64()),
			CaseID:           patientCase.ID.Int64(),
			NotificationView: notificationView,
		}
	}
	return &phCaseView{
		Title:            fmt.Sprintf("%s Case", patientCase.Name),
		Subtitle:         "Pending Review",
		ActionURL:        app_url.ViewCaseAction(patientCase.ID.Int64()),
		CaseID:           patientCase.ID.Int64(),
		NotificationView: notificationView,
	}
}

func getStartVisitCard() common.ClientView {
	return &phStartVisit{
		Title:     "Start Your First Visit",
		IconURL:   app_url.IconVisitLarge,
		ActionURL: app_url.StartVisitAction(),
		ImageURLs: []string{
			"https://d2bln09x7zhlg8.cloudfront.net/start_visit_doc_1.jpg",
			"https://d2bln09x7zhlg8.cloudfront.net/start_visit_doc_2.jpg",
			"https://d2bln09x7zhlg8.cloudfront.net/start_visit_doc_3.jpg",
			"https://d2bln09x7zhlg8.cloudfront.net/start_visit_doc_4.jpg",
		},
		ButtonTitle: "Get Started",
		Description: "Receive an effective, personalized treatment plan from a dermatologist within 24 hours.",
	}
}

func getMeetCareTeamSection(careTeamAssignments []*common.CareProviderAssignment, patientCase *common.PatientCase, apiCDNDomain string) common.ClientView {
	sectionView := &phSectionView{
		Title: fmt.Sprintf("Meet your %s care team", patientCase.Name),
		Views: make([]common.ClientView, 0, len(careTeamAssignments)),
	}

	sort.Sort(api.ByCareProviderRole(careTeamAssignments))

	for _, assignment := range careTeamAssignments {
		sectionView.Views = append(sectionView.Views, &phCareProviderView{
			CareProvider: responses.TransformCareTeamMember(assignment, apiCDNDomain),
		})
	}

	return sectionView
}

func getShareSpruceSection(currentAppVersion *common.Version, dataAPI api.DataAPI, webDomain string, accountID int64) (common.ClientView, error) {
	// FIXME: Improve the way we do app version/view mapping.
	// The current version checking mechanism will be difficult to maintain
	// Version 1.1.0 - Initial refer a friend spruce action homecard version
	// Version 2.0.2 - Improved direct refer a friend link homecard

	// The initial refer a friend launch version
	referFriendLaunchVersion110 := &common.Version{
		Major: 1,
		Minor: 1,
		Patch: 0,
	}

	// Improved refer a friend home card
	referFriendVersion202 := &common.Version{
		Major: 2,
		Minor: 0,
		Patch: 2,
	}

	switch {
	case currentAppVersion.GreaterThanOrEqualTo(referFriendLaunchVersion110) && currentAppVersion.LessThan(referFriendVersion202):

		activeTemplate, err := dataAPI.ActiveReferralProgramTemplate(api.RolePatient, common.PromotionTypes)
		if api.IsErrNotFound(err) {
			return nil, nil
		} else if err != nil {
			return nil, errors.Trace(err)
		}

		referralProgram, ok := activeTemplate.Data.(promotions.ReferralProgram)
		if !ok {
			return nil, nil
		}

		return &phSmallIconText{
			Title:       referralProgram.HomeCardText(),
			IconURL:     referralProgram.HomeCardImageURL(),
			ActionURL:   app_url.ViewReferFriendAction().String(),
			RoundedIcon: true,
		}, nil
	case currentAppVersion.GreaterThanOrEqualTo(referFriendVersion202):
		referralDisplayInfo, err := promotions.CreateReferralDisplayInfo(dataAPI, webDomain, accountID)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return &phReferFriend{
			ReferFriendContent: referralDisplayInfo,
		}, nil
	}

	return nil, nil
}

func getSendUsMessageSection() common.ClientView {
	return &phSmallIconText{
		Title:       "Have a question? Send us a message.",
		IconURL:     app_url.IconSupport,
		ActionURL:   app_url.ViewSupportAction().String(),
		RoundedIcon: true,
	}
}

func getNotifyMeCard(state string) common.ClientView {
	return &phNotifyMeView{
		Title:       fmt.Sprintf("Sign up to be notified when Spruce is available in %s.", state),
		Placeholder: "Your Email Address",
		ButtonTitle: "Sign Up",
		ActionURL:   app_url.NotifyMeAction(),
	}
}

func getNotifyMeConfirmationCard(state string) common.ClientView {
	return &phHeroIconView{
		Title:       "Thanks!",
		Description: fmt.Sprintf("We'll notify you when Spruce is available in %s.", state),
		IconURL:     app_url.IconBlueSuccess,
	}
}

func getLearnAboutSpruceSection(pathwayTag string) common.ClientView {
	return &phSectionView{
		Views: []common.ClientView{
			&phSmallIconText{
				Title:       "Meet the doctors",
				IconURL:     app_url.IconSpruceDoctors,
				ActionURL:   app_url.ViewSampleDoctorProfilesAction().String(),
				RoundedIcon: true,
			},
			&phSmallIconText{
				Title:       "Frequently asked questions",
				IconURL:     app_url.IconFAQ,
				ActionURL:   app_url.ViewSpruceFAQAction().String(),
				RoundedIcon: true,
			},
		},
	}
}

func spellNumber(count int) string {
	switch count {
	case 2:
		return " two "
	case 3:
		return " three "
	case 4:
		return " four "
	case 5:
		return " five "
	case 6:
		return " six "
	case 7:
		return " seven "
	case 8:
		return " eight "
	case 9:
		return " nine "
	case 10:
		return " ten "
	}
	return ""
}
