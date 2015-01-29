package patient

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/info_intake"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/sku"
)

func IntakeLayoutForVisit(
	dataAPI api.DataAPI,
	apiDomain string,
	store storage.Store,
	expirationDuration time.Duration,
	visit *common.PatientVisit) (*info_intake.InfoIntakeLayout, error) {

	// if there is an active patient visit record, then ensure to lookup the layout to send to the patient
	// based on what layout was shown to the patient at the time of opening of the patient visit, NOT the current
	// based on what is the current active layout because that may have potentially changed and we want to ensure
	// to not confuse the patient by changing the question structure under their feet for this particular patient visit
	// in other words, want to show them what they have already seen in terms of a flow.
	visitLayout, err := apiservice.GetPatientLayoutForPatientVisit(visit, api.EN_LANGUAGE_ID, dataAPI, apiDomain)
	if err != nil {
		return nil, err
	}

	err = populateLayoutWithAnswers(
		visitLayout,
		dataAPI,
		store,
		expirationDuration,
		visit)

	return visitLayout, err
}

func populateLayoutWithAnswers(
	visitLayout *info_intake.InfoIntakeLayout,
	dataAPI api.DataAPI,
	store storage.Store,
	expirationDuration time.Duration,
	patientVisit *common.PatientVisit,
) error {

	patientID := patientVisit.PatientID.Int64()
	visitID := patientVisit.PatientVisitID.Int64()

	photoQuestionIDs := visitLayout.PhotoQuestionIDs()
	photosForVisit, err := dataAPI.PatientPhotoSectionsForQuestionIDs(photoQuestionIDs, patientID, visitID)
	if err != nil {
		return err
	}

	// create photoURLs for each answer
	expirationTime := time.Now().Add(expirationDuration)
	for _, photoSections := range photosForVisit {
		for _, photoSection := range photoSections {
			ps := photoSection.(*common.PhotoIntakeSection)
			for _, intakeSlot := range ps.Photos {
				media, err := dataAPI.GetMedia(intakeSlot.PhotoID)
				if err != nil {
					return err
				}

				if ok, err := dataAPI.MediaHasClaim(intakeSlot.PhotoID, common.ClaimerTypePhotoIntakeSection, ps.ID); err != nil {
					return err
				} else if !ok {
					return errors.New("ClaimerID does not match PhotoIntakeSectionID")
				}

				intakeSlot.PhotoURL, err = store.GetSignedURL(media.URL, expirationTime)
				if err != nil {
					return err
				}
			}
		}

	}

	nonPhotoQuestionIDs := visitLayout.NonPhotoQuestionIDs()
	answersForVisit, err := dataAPI.AnswersForQuestions(nonPhotoQuestionIDs, &api.PatientIntake{
		PatientID:      patientID,
		PatientVisitID: visitID,
	})
	if err != nil {
		return err
	}

	// merge answers into one map
	for questionID, answers := range photosForVisit {
		answersForVisit[questionID] = answers
	}

	// keep track of any question that is to be prefilled
	// and doesn't have an answer for this visit yet
	prefillQuestionsWithNoAnswers := make(map[int64]*info_intake.Question)
	var prefillQuestionIDs []int64
	// populate layout with the answers for each question
	for _, section := range visitLayout.Sections {
		for _, screen := range section.Screens {
			for _, question := range screen.Questions {
				question.Answers = answersForVisit[question.QuestionID]
				if question.ToPrefill && len(question.Answers) == 0 {
					prefillQuestionsWithNoAnswers[question.QuestionID] = question
					prefillQuestionIDs = append(prefillQuestionIDs, question.QuestionID)
				}
			}
		}
	}

	// if visit is still open, prefill any questions currently unanswered
	// with answers by the patient from a previous visit
	if patientVisit.Status == common.PVStatusOpen {
		previousAnswers, err := dataAPI.PreviousPatientAnswersForQuestions(
			prefillQuestionIDs, patientID, patientVisit.CreationDate)
		if err != nil {
			return err
		}

		// populate the questions with previous answers by the patient
		for questionID, answers := range previousAnswers {
			prefillQuestionsWithNoAnswers[questionID].PrefilledWithPreviousAnswers = true
			prefillQuestionsWithNoAnswers[questionID].Answers = answers
		}
	}

	return nil
}

func createPatientVisit(
	patient *common.Patient,
	doctorID int64,
	pathwayTag string,
	dataAPI api.DataAPI,
	apiDomain string,
	dispatcher *dispatch.Dispatcher,
	store storage.Store,
	expirationDuration time.Duration,
	r *http.Request,
	context *apiservice.VisitLayoutContext,
) (*PatientVisitResponse, error) {

	var clientLayout *info_intake.InfoIntakeLayout
	var patientVisit *common.PatientVisit

	patientCases, err := dataAPI.CasesForPathway(patient.PatientID.Int64(), pathwayTag, common.ActivePatientCaseStates())
	if err != nil {
		return nil, err
	} else if err == nil {
		switch l := len(patientCases); {
		case l == 0:
		case l == 1:
			// if there exists open visits against an active case for this pathwayTag, return
			// the last created patient visit. Technically, the patient should not have more than a single open
			// patient visit against a case.
			patientVisits, err := dataAPI.GetVisitsForCase(patientCases[0].ID.Int64(), common.OpenPatientVisitStates())
			if err != nil {
				return nil, err
			} else if len(patientVisits) > 0 {
				sort.Reverse(common.ByPatientVisitCreationDate(patientVisits))
				patientVisit = patientVisits[0]
			}
		default:
			return nil, fmt.Errorf("Only a single active case per pathway can exist for now. Pathway %s has %d active cases.", pathwayTag, len(patientCases))
		}
	}

	if patientVisit == nil {
		pathway, err := dataAPI.PathwayForTag(pathwayTag, api.PONone)
		if err != nil {
			return nil, err
		}

		// start a new visit
		var layoutVersionID int64
		sHeaders := apiservice.ExtractSpruceHeaders(r)
		clientLayout, layoutVersionID, err = apiservice.GetCurrentActiveClientLayoutForPathway(dataAPI,
			pathway.ID, api.EN_LANGUAGE_ID, sku.AcneVisit,
			sHeaders.AppVersion, sHeaders.Platform, nil)
		if err != nil {
			return nil, err
		}

		// TODO: Fix SKU
		patientVisit = &common.PatientVisit{
			PatientID:       patient.PatientID,
			PathwayTag:      pathway.Tag,
			Status:          common.PVStatusOpen,
			LayoutVersionID: encoding.NewObjectID(layoutVersionID),
			SKU:             sku.AcneVisit,
		}

		_, err = dataAPI.CreatePatientVisit(patientVisit)
		if err != nil {
			return nil, err
		}

		// assign the doctor to the case if the doctor is specified
		if doctorID > 0 {
			if err := dataAPI.AddDoctorToPatientCase(doctorID, patientVisit.PatientCaseID.Int64()); err != nil {
				return nil, err
			}
		}

		dispatcher.Publish(&VisitStartedEvent{
			PatientID:     patient.PatientID.Int64(),
			VisitID:       patientVisit.PatientVisitID.Int64(),
			PatientCaseID: patientVisit.PatientCaseID.Int64(),
		})
	} else {
		// return current visit
		clientLayout, err = IntakeLayoutForVisit(dataAPI, apiDomain, store, expirationDuration, patientVisit)
		if err != nil {
			return nil, err
		}
	}

	return &PatientVisitResponse{
		PatientVisitID: patientVisit.PatientVisitID.Int64(),
		Status:         patientVisit.Status,
		ClientLayout:   clientLayout,
	}, nil
}
