package demo

import (
	"bytes"
	"carefront/apiservice"
	"carefront/common"
	"carefront/encoding"
	"carefront/libs/golog"
	"carefront/libs/pharmacy"
	"carefront/messages"
	"carefront/patient_visit"
	"encoding/json"
	"fmt"
	"net/http"
)

type questionTag int

const (
	qAcneOnset questionTag = iota
	qAcneWorse
	qAcneChangesWorse
	qAcneSymptoms
	qAcneWorsePeriod
	qSkinDescription
	qAcnePrevTreatmentTypes
	qAcnePrevTreatmentList
	qUsingTreatment
	qEffectiveTreatment
	qTreatmentIrritateSkin
	qLengthTreatment
	qAnythingElseAcne
	qPregnancyPlanning
	qCurrentMedications
	qCurrentMedicationsEntry
	qLengthCurrentMedication
	qAllergicMedications
	qAllergicMedicationEntry
	qPrevSkinConditionDiagnosis
	qListPrevSkinConditionDiagnosis
	qOtherConditionsAcne
	qFacePhotoSection
	qChestPhotoSection
	qOtherLocationPhotoSection
	qPrescriptionPreference
	qAcnePreviousOTCList
	qUsingOTC
	qEffectiveOTC
	qOTCIrritateSkin
	qLengthOTC
)

var (
	questionTags = map[string]questionTag{
		"q_onset_acne":                         qAcneOnset,
		"q_acne_worse":                         qAcneWorse,
		"q_changes_acne_worse":                 qAcneChangesWorse,
		"q_acne_symptoms":                      qAcneSymptoms,
		"q_acne_worse_period":                  qAcneWorsePeriod,
		"q_skin_description":                   qSkinDescription,
		"q_acne_prev_treatment_types":          qAcnePrevTreatmentTypes,
		"q_acne_prev_treatment_list":           qAcnePrevTreatmentList,
		"q_using_treatment":                    qUsingTreatment,
		"q_effective_treatment":                qEffectiveTreatment,
		"q_treatment_irritate_skin":            qTreatmentIrritateSkin,
		"q_length_treatment":                   qLengthTreatment,
		"q_anything_else_acne":                 qAnythingElseAcne,
		"q_pregnancy_planning":                 qPregnancyPlanning,
		"q_current_medications":                qCurrentMedications,
		"q_current_medications_entry":          qCurrentMedicationsEntry,
		"q_length_current_medication":          qLengthCurrentMedication,
		"q_allergic_medications":               qAllergicMedications,
		"q_allergic_medication_entry":          qAllergicMedicationEntry,
		"q_prev_skin_condition_diagnosis":      qPrevSkinConditionDiagnosis,
		"q_list_prev_skin_condition_diagnosis": qListPrevSkinConditionDiagnosis,
		"q_other_conditions_acne":              qOtherConditionsAcne,
		"q_face_photo_section":                 qFacePhotoSection,
		"q_chest_photo_section":                qChestPhotoSection,
		"q_other_location_photo_section":       qOtherLocationPhotoSection,
		"q_prescription_preference":            qPrescriptionPreference,
		"q_acne_prev_otc_list":                 qAcnePreviousOTCList,
		"q_using_otc":                          qUsingOTC,
		"q_effective_otc":                      qEffectiveOTC,
		"q_otc_irritate_skin":                  qOTCIrritateSkin,
		"q_length_otc":                         qLengthOTC,
	}
)

type potentialAnswerTag int

const (
	aSixToTwelveMonths potentialAnswerTag = iota
	aAcneWorseYes
	aDiscoloration
	aScarring
	aPainfulToTouch
	aCysts
	aAcneWorsePeriodNo
	aSkinDescriptionOily
	aPrevTreatmentsTypeOTC
	aUsingTreatmentYes
	aSomewhatEffectiveTreatment
	aIrritateSkinYes
	aLengthTreatmentLessThanMonth
	aCurrentlyPregnant
	aCurrentMedicationsYes
	aTwoToFiveMonthsLength
	aAllergicMedicationsYes
	aPrevSkinConditionDiagnosisYes
	aListPrevSkinConditionDiagnosisAcne
	aListPrevSkinConditionDiagnosisPsoriasis
	aNoneOfTheAboveOtherConditions
	aGenericRxOnly
	aPickedOrSqueezed
	aCreatedScars
	aLengthOTCSixEleventMonths
	aOTCIrritateSkinYes
	aEffectiveOTCSomewhat
	aUsingOTCNo
)

var (
	answerTags = map[string]potentialAnswerTag{
		"a_six_twelve_months_ago":                     aSixToTwelveMonths,
		"a_yes_acne_worse":                            aAcneWorseYes,
		"a_discoloration":                             aDiscoloration,
		"a_scarring":                                  aScarring,
		"a_painful_touch":                             aPainfulToTouch,
		"a_cysts":                                     aCysts,
		"a_acne_worse_no":                             aAcneWorsePeriodNo,
		"a_oil_skin":                                  aSkinDescriptionOily,
		"a_otc_prev_treatment_type":                   aPrevTreatmentsTypeOTC,
		"a_using_treatment_yes":                       aUsingTreatmentYes,
		"a_effective_treatment_somewhat":              aSomewhatEffectiveTreatment,
		"a_irritate_skin_yes":                         aIrritateSkinYes,
		"a_length_treatment_less_one":                 aLengthTreatmentLessThanMonth,
		"a_pregnant":                                  aCurrentlyPregnant,
		"a_current_medications_yes":                   aCurrentMedicationsYes,
		"a_length_current_medication_two_five_months": aTwoToFiveMonthsLength,
		"a_yes_allergic_medications":                  aAllergicMedicationsYes,
		"a_yes_prev_skin_diagnosis":                   aPrevSkinConditionDiagnosisYes,
		"a_acne_skin_diagnosis":                       aListPrevSkinConditionDiagnosisAcne,
		"a_psoriasis_skin_diagnosis":                  aListPrevSkinConditionDiagnosisPsoriasis,
		"a_other_condition_acne_none":                 aNoneOfTheAboveOtherConditions,
		"a_generic_only":                              aGenericRxOnly,
		"a_picked_or_squeezed":                        aPickedOrSqueezed,
		"a_created_scars":                             aCreatedScars,
		"a_effective_otc_somewhat":                    aEffectiveOTCSomewhat,
		"a_otc_irritate_skin_yes":                     aOTCIrritateSkinYes,
		"a_length_otc_two_six_eleven_months":          aLengthOTCSixEleventMonths,
		"a_using_otc_no":                              aUsingOTCNo,
	}

	sampleMessages = []string{
		"I forgot to mention I'm allergic to sulfa drugs.",
		"Could you recommend a sunscreen that won't make me break out?",
		"Could you recommend a facial wash for oily skin?",
	}
)

type photoSlotType int

const (
	photoSlotFaceFront photoSlotType = iota
	photoSlotFaceRight
	photoSlotFaceLeft
	photoSlotOther
	photoSlotBack
	photoSlotChest
)

var (
	photoSlotTypes = map[string]photoSlotType{
		"photo_slot_face_right": photoSlotFaceRight,
		"photo_slot_face_left":  photoSlotFaceLeft,
		"photo_slot_face_front": photoSlotFaceFront,
		"photo_slot_other":      photoSlotOther,
		"photo_slot_chest":      photoSlotChest,
		"photo_slot_back":       photoSlotBack,
	}
)

const (
	signupPatientUrl         = "http://127.0.0.1:8080/v1/patient"
	updatePatientPharmacyUrl = "http://127.0.0.1:8080/v1/patient/pharmacy"
	patientVisitUrl          = "http://127.0.0.1:8080/v1/patient/visit"
	answerQuestionsUrl       = "http://127.0.0.1:8080/v1/patient/visit/answer"
	photoIntakeUrl           = "http://127.0.0.1:8080/v1/patient/visit/photo_answer"
	conversationUrl          = "http://127.0.0.1:8080/v1/patient/conversation"
	demoPhotosBucketFormat   = "%s-carefront-demo"
	frontPhoto               = "profile_front.jpg"
	profileRightPhoto        = "profile_right.jpg"
	profileLeftPhoto         = "profile_left.jpg"
	neckPhoto                = "neck.jpg"
	chestPhoto               = "chest.jpg"
	failure                  = 0
	success                  = 1
)

func populatePatientIntake(questionIds map[questionTag]int64, answerIds map[potentialAnswerTag]int64) []*apiservice.AnswerToQuestionItem {

	return []*apiservice.AnswerToQuestionItem{
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcneOnset],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aSixToTwelveMonths],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcneWorse],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aAcneWorseYes],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcneChangesWorse],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					AnswerText: "I've starting working out again so wonder if sweat could be a contributing factor?",
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcneSymptoms],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aDiscoloration],
				},
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aCreatedScars],
				},
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aCysts],
				},
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aPainfulToTouch],
				},
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aPickedOrSqueezed],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcneWorsePeriod],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aAcneWorsePeriodNo],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qSkinDescription],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aSkinDescriptionOily],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcnePrevTreatmentTypes],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aPrevTreatmentsTypeOTC],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcnePreviousOTCList],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					AnswerText: "Proactiv",
					SubQuestionAnswerIntakes: []*apiservice.SubQuestionAnswerIntake{
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qUsingOTC],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aUsingOTCNo],
								},
							},
						},
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qEffectiveOTC],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aEffectiveOTCSomewhat],
								},
							},
						},
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qOTCIrritateSkin],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aOTCIrritateSkinYes],
								},
							},
						},
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qLengthOTC],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aLengthOTCSixEleventMonths],
								},
							},
						},
					},
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAcnePrevTreatmentList],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					AnswerText: "Benzoyl Peroxide 10% Wash",
					SubQuestionAnswerIntakes: []*apiservice.SubQuestionAnswerIntake{
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qUsingTreatment],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aUsingTreatmentYes],
								},
							},
						},
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qEffectiveTreatment],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aSomewhatEffectiveTreatment],
								},
							},
						},
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qTreatmentIrritateSkin],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aIrritateSkinYes],
								},
							},
						},
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qLengthTreatment],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aLengthTreatmentLessThanMonth],
								},
							},
						},
					},
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAnythingElseAcne],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					AnswerText: "I've noticed that my acne flares up when I wait longer between changing razor blades. Also, my acne typically concentrates around my lips.",
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qCurrentMedications],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aCurrentMedicationsYes],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qCurrentMedicationsEntry],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					AnswerText: "Clyndamycin",
					SubQuestionAnswerIntakes: []*apiservice.SubQuestionAnswerIntake{
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qLengthCurrentMedication],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aTwoToFiveMonthsLength],
								},
							},
						},
					},
				},
				&apiservice.AnswerItem{
					AnswerText: "Tretinoin Topical",
					SubQuestionAnswerIntakes: []*apiservice.SubQuestionAnswerIntake{
						&apiservice.SubQuestionAnswerIntake{
							QuestionId: questionIds[qLengthCurrentMedication],
							AnswerIntakes: []*apiservice.AnswerItem{
								&apiservice.AnswerItem{
									PotentialAnswerId: answerIds[aTwoToFiveMonthsLength],
								},
							},
						},
					},
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAllergicMedications],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aAllergicMedicationsYes],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qAllergicMedicationEntry],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					AnswerText: "Penicillin",
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qPrevSkinConditionDiagnosis],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aPrevSkinConditionDiagnosisYes],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qListPrevSkinConditionDiagnosis],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aListPrevSkinConditionDiagnosisAcne],
				},
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aListPrevSkinConditionDiagnosisPsoriasis],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qOtherConditionsAcne],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aNoneOfTheAboveOtherConditions],
				},
			},
		},
		&apiservice.AnswerToQuestionItem{
			QuestionId: questionIds[qPrescriptionPreference],
			AnswerIntakes: []*apiservice.AnswerItem{
				&apiservice.AnswerItem{
					PotentialAnswerId: answerIds[aGenericRxOnly],
				},
			},
		},
	}
}

func startPatientIntakeSubmission(answersToQuestions []*apiservice.AnswerToQuestionItem, patientVisitId int64, patientAuthToken string, signal chan int) {

	go func() {

		answerIntakeRequestBody := &apiservice.AnswerIntakeRequestBody{
			PatientVisitId: patientVisitId,
			Questions:      answersToQuestions,
		}

		jsonData, _ := json.Marshal(answerIntakeRequestBody)
		answerQuestionsRequest, err := http.NewRequest("POST", answerQuestionsUrl, bytes.NewReader(jsonData))
		answerQuestionsRequest.Header.Set("Content-Type", "application/json")
		answerQuestionsRequest.Header.Set("Authorization", "token "+patientAuthToken)

		resp, err := http.DefaultClient.Do(answerQuestionsRequest)
		if err != nil || resp.StatusCode != http.StatusOK {
			golog.Errorf("Error while submitting patient intake: %+v", err)
			signal <- failure
			return
		}
		signal <- success
	}()
}

func (c *Handler) startSendingMessageToDoctor(token, message string, signal chan int) {
	go func() {
		requestData := &messages.NewConversationRequest{
			Message: message,
			TopicId: 1,
		}
		jsonData, _ := json.Marshal(requestData)
		newConversationRequest, err := http.NewRequest("POST", conversationUrl, bytes.NewReader(jsonData))
		newConversationRequest.Header.Set("Content-Type", "application/json")
		newConversationRequest.Header.Set("Authorization", "token "+token)

		resp, err := http.DefaultClient.Do(newConversationRequest)
		if err != nil || resp.StatusCode != http.StatusOK {
			golog.Errorf("Error while starting new conversation for patient: %+v", err)
			signal <- failure
			return
		}
		signal <- success
	}()
}

func (c *Handler) startPhotoSubmissionForPatient(questionId, patientVisitId int64, photoSections []*common.PhotoIntakeSection, patientAuthToken string, signal chan int) {

	go func() {

		patient, err := c.dataApi.GetPatientFromPatientVisitId(patientVisitId)
		if err != nil {
			golog.Errorf("Unable to get patient id from patient visit id: %s", err)
			signal <- failure
			return
		}

		for _, photoSection := range photoSections {
			for _, photo := range photoSection.Photos {
				// get the key of the photo under the assumption that the caller of this method populated
				// the photo key into the photo url
				photoKey := photo.PhotoUrl

				// get the url of the image so as to add the photo to the photos table
				imageUrl, err := c.cloudStorageApi.GetUnsignedUrlForObjectAtLocation(fmt.Sprintf(demoPhotosBucketFormat, c.environment), photoKey, c.awsRegion)
				if err != nil {
					golog.Errorf("Error while getting unsigned url for object at location: %s", err)
					signal <- failure
					return
				}

				if photoId, err := c.dataApi.AddPhoto(patient.PersonId, imageUrl, "image/jpeg"); err != nil {
					golog.Errorf("Unable to add photo to photo table: %s ", err)
					signal <- failure
					return
				} else {
					photo.PhotoId = photoId
				}
			}
		}

		// prepare the request to submit the photo sections
		requestData := patient_visit.PhotoAnswerIntakeRequestData{
			PatientVisitId: patientVisitId,
			PhotoQuestions: []*patient_visit.PhotoAnswerIntakeQuestionItem{
				&patient_visit.PhotoAnswerIntakeQuestionItem{
					QuestionId:    questionId,
					PhotoSections: photoSections,
				},
			},
		}

		jsonData, err := json.Marshal(&requestData)
		if err != nil {
			golog.Errorf("Unable to marshal json for photo intake: %s", err)
			signal <- failure
			return
		}

		photoIntakeRequest, err := http.NewRequest("POST", photoIntakeUrl, bytes.NewReader(jsonData))
		photoIntakeRequest.Header.Set("Content-Type", "application/json")
		photoIntakeRequest.Header.Set("Authorization", "token "+patientAuthToken)
		resp, err := http.DefaultClient.Do(photoIntakeRequest)
		if err != nil || resp.StatusCode != http.StatusOK {
			golog.Errorf("Error while trying submit photo for intake: %+v", err)
			signal <- failure
			return
		}
		signal <- success
	}()
}

func prepareSurescriptsPatients() []*common.Patient {

	patients := make([]*common.Patient, 8)

	patients[0] = &common.Patient{
		FirstName: "Ci",
		LastName:  "Li",
		Gender:    "Male",
		Dob: encoding.Dob{
			Year:  1923,
			Month: 10,
			Day:   18,
		},
		ZipCode: "94115",
		PhoneNumbers: []*common.PhoneInformation{&common.PhoneInformation{
			Phone:     "2068773590",
			PhoneType: "Home",
		},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "12345 Main Street",
			AddressLine2: "Apt 1112",
			City:         "San Francisco",
			State:        "California",
			ZipCode:      "94115",
		},
	}

	patients[1] = &common.Patient{
		Prefix:    "Mr",
		FirstName: "Howard",
		LastName:  "Plower",
		Gender:    "Male",
		Dob: encoding.Dob{
			Year:  1923,
			Month: 10,
			Day:   18,
		},
		ZipCode: "19102",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "215-988-6723",
				PhoneType: "Home",
			},
			&common.PhoneInformation{
				Phone:     "4137762738",
				PhoneType: "Cell",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "76 Deerlake Road",
			City:         "Philadelphia",
			State:        "Pennsylvania",
			ZipCode:      "19102",
		},
	}

	patients[2] = &common.Patient{
		FirstName: "Kara",
		LastName:  "Whiteside",
		Gender:    "Female",
		Dob: encoding.Dob{
			Year:  1952,
			Month: 10,
			Day:   11,
		},
		ZipCode: "44306",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "3305547754",
				PhoneType: "Home",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "23230 Seaport",
			City:         "Akron",
			State:        "Ohio",
			ZipCode:      "44306",
		},
	}

	patients[3] = &common.Patient{
		Prefix:    "Ms",
		FirstName: "Debra",
		LastName:  "Tucker",
		Gender:    "Female",
		Dob: encoding.Dob{
			Year:  1970,
			Month: 11,
			Day:   01,
		},
		ZipCode: "44103",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "4408450398",
				PhoneType: "Home",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "8331 Everwood Dr.",
			AddressLine2: "Apt 342",
			City:         "Cleveland",
			State:        "Ohio",
			ZipCode:      "44103",
		},
	}

	patients[4] = &common.Patient{
		Prefix:     "Ms",
		FirstName:  "Felicia",
		LastName:   "Flounders",
		MiddleName: "Ann",
		Gender:     "Female",
		Dob: encoding.Dob{
			Year:  1980,
			Month: 11,
			Day:   01,
		},
		ZipCode: "20187",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "3108620035x2345",
				PhoneType: "Home",
			},
			&common.PhoneInformation{
				Phone:     "3019289283",
				PhoneType: "Cell",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "6715 Swanson Ave",
			AddressLine2: "Apt 102",
			City:         "Bethesda",
			State:        "Maryland",
			ZipCode:      "20187",
		},
	}

	patients[5] = &common.Patient{
		FirstName:  "Douglas",
		LastName:   "Richardson",
		MiddleName: "R",
		Gender:     "Male",
		Dob: encoding.Dob{
			Year:  1968,
			Month: 9,
			Day:   1,
		},
		ZipCode: "01040",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "4137760938",
				PhoneType: "Home",
			},
			&common.PhoneInformation{
				Phone:     "4137762738",
				PhoneType: "Cell",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "2556 Lane Rd",
			AddressLine2: "Apt 101",
			City:         "Smittyville",
			State:        "Virginia",
			ZipCode:      "01040-2239",
		},
	}

	patients[6] = &common.Patient{
		FirstName: "David",
		LastName:  "Thrower",
		Gender:    "Male",
		Dob: encoding.Dob{
			Year:  1933,
			Month: 2,
			Day:   22,
		},
		ZipCode: "34737",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "3526685547",
				PhoneType: "Home",
			},
			&common.PhoneInformation{
				Phone:     "4137762738",
				PhoneType: "Cell",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "64 Violet Lane",
			AddressLine2: "Apt 101",
			City:         "Howey In The Hills",
			State:        "Florida",
			ZipCode:      "34737",
		},
	}

	patients[7] = &common.Patient{
		Prefix:     "Patient II",
		FirstName:  "!\"#$%'+,-/:;=?@[\\]^_`{|}~0000&",
		LastName:   "!\"#$%'+,-/:;=?@[\\]^_`{|}~0000&",
		MiddleName: "!\"#$%'+,-/:;=?@[\\]^_`{|}~0000&",
		Suffix:     "Junior iii",
		Gender:     "Male",
		Dob: encoding.Dob{
			Year:  1948,
			Month: 1,
			Day:   1,
		},
		ZipCode: "34737",
		PhoneNumbers: []*common.PhoneInformation{
			&common.PhoneInformation{
				Phone:     "5719212122x1234567890444",
				PhoneType: "Home",
			},
			&common.PhoneInformation{
				Phone:     "7034445523x4473",
				PhoneType: "Cell",
			},
			&common.PhoneInformation{
				Phone:     "7034445524x4474",
				PhoneType: "Work",
			},
			&common.PhoneInformation{
				Phone:     "7034445522x4472",
				PhoneType: "Work",
			},
			&common.PhoneInformation{
				Phone:     "7034445526x4476",
				PhoneType: "Home",
			},
		},
		Pharmacy: &pharmacy.PharmacyData{
			SourceId:     "47731",
			AddressLine1: "116 New Montgomery St",
			Name:         "CA pharmacy store 10.6",
			City:         "San Francisco",
			State:        "CA",
			Postal:       "92804",
			Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
		},
		PatientAddress: &common.Address{
			AddressLine1: "!\"#$%'+,-/:;=?@[\\]^_`{|}~0000&",
			AddressLine2: "!\"#$%'+,-/:;=?@[\\]^_`{|}~0000&",
			City:         "!\"#$%'+,-/:;=?@[\\]^_`{|}~0000&",
			State:        "Colorado",
			ZipCode:      "94115",
		},
	}
	return patients
}

func prepareDemoPatients(n int64) []*common.Patient {
	patients := make([]*common.Patient, n)
	for i := int64(0); i < n; i++ {
		patients[i] = &common.Patient{
			FirstName: "Kunal",
			LastName:  "Jham",
			Gender:    "male",
			Dob: encoding.Dob{
				Year:  1987,
				Month: 11,
				Day:   8,
			},
			ZipCode: "94115",
			PhoneNumbers: []*common.PhoneInformation{&common.PhoneInformation{
				Phone:     "2068773590",
				PhoneType: "Home",
			},
			},
			Pharmacy: &pharmacy.PharmacyData{
				SourceId:     "47731",
				AddressLine1: "116 New Montgomery St",
				Name:         "CA pharmacy store 10.6",
				City:         "San Francisco",
				State:        "CA",
				Postal:       "92804",
				Source:       pharmacy.PHARMACY_SOURCE_SURESCRIPTS,
			},
			PatientAddress: &common.Address{
				AddressLine1: "12345 Main Street",
				AddressLine2: "Apt 1112",
				City:         "San Francisco",
				State:        "California",
				ZipCode:      "94115",
			},
		}
	}
	return patients
}
