package test_promotions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sprucehealth/backend/apiservice/apipaths"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/cost/promotions"
	"github.com/sprucehealth/backend/libs/aws/sqs"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestReferrals_NewPatientReferral(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	stubSQSQueue := &common.SQSQueue{
		QueueURL:     "visit_url",
		QueueService: &sqs.StubSQS{},
	}
	testData.Config.VisitQueue = stubSQSQueue
	testData.StartAPIServer(t)

	admin := test_integration.CreateRandomAdmin(t, testData)

	setupPromotionsTest(testData, t)

	// create referral program template
	title := "dollars off"
	description := "description"
	requestData := map[string]interface{}{
		"type":        "promo_money_off",
		"title":       title,
		"description": description,
		"group":       "new_user",
		"share_text": map[string]interface{}{
			"facebook": "agaHG",
			"sms":      "ADgagh",
			"default":  "aegagh",
		},
		"promotion": map[string]interface{}{
			"display_msg":  "dollars off",
			"success_msg":  "dollars off",
			"short_msg":    "dollars off",
			"for_new_user": true,
			"group":        "new_user",
			"value":        500,
		},
	}

	var responseData map[string]interface{}
	resp, err := testData.AuthPostJSON(testData.APIServer.URL+apipaths.ReferralProgramsTemplateURLPath, admin.AccountID.Int64(), requestData, &responseData)
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// now create patient
	pr1 := test_integration.SignupRandomTestPatientWithPharmacyAndAddress(t, testData)

	// now try to get the referral program for this patient
	resp, err = testData.AuthGet(testData.APIServer.URL+apipaths.ReferralsURLPath, pr1.Patient.AccountID.Int64())
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// now try to get another potential patient to claim the code
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	test.OK(t, err)
	promotionURL := responseData["url"].(string)
	test.OK(t, err)
	slashIndex := strings.LastIndex(promotionURL, "/")
	code := promotionURL[slashIndex+1:]

	_, err = promotions.AssociatePromoCode("kunal@test.com", "California", code, testData.DataAPI, testData.AuthAPI, testData.Config.AnalyticsLogger)
	test.OK(t, err)

	// now signup the patient
	pr := test_integration.SignupTestPatientWithEmail("kunal@test.com", t, testData)
	patientID := pr.Patient.PatientID.Int64()
	patientAccountID := pr.Patient.AccountID.Int64()
	test_integration.AddTestPharmacyForPatient(patientID, testData, t)
	test_integration.AddCreditCardForPatient(patientID, testData, t)
	test_integration.AddTestAddressForPatient(patientID, testData, t)

	// ensure that the interstitial is shown to the patient
	test.Equals(t, true, pr.PromotionConfirmationContent != nil)

	// ensure that the referring patient is informed of the user having associated the code
	referralProgram, err := testData.DataAPI.ActiveReferralProgramForAccount(pr1.Patient.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	rp := referralProgram.Data.(promotions.ReferralProgram)
	test.Equals(t, 1, rp.UsersAssociatedCount())
	test.Equals(t, 0, rp.VisitsSubmittedCount())

	// lets query the price for this user
	cost, lineItems := test_integration.QueryCost(patientAccountID, test_integration.SKUAcneVisit, testData, t)
	test.Equals(t, "$35", cost)
	test.Equals(t, 2, len(lineItems))

	// lets have this user start and submit a visit
	startAndSubmitVisit(patientID, patientAccountID, stubSQSQueue, testData, t)

	// at this point the referral program should account for the submitted visit
	referralProgram, err = testData.DataAPI.ActiveReferralProgramForAccount(pr1.Patient.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	rp = referralProgram.Data.(promotions.ReferralProgram)
	test.Equals(t, 1, rp.UsersAssociatedCount())
	test.Equals(t, 1, rp.VisitsSubmittedCount())

	// lets have one more user use the promo code
	_, err = promotions.AssociatePromoCode("kunal2@test.com", "California", code, testData.DataAPI, testData.AuthAPI, testData.Config.AnalyticsLogger)
	test.OK(t, err)

	// sign the user up
	pr = test_integration.SignupTestPatientWithEmail("kunal2@test.com", t, testData)

	// ensure that the patient is informed of the associated user
	referralProgram, err = testData.DataAPI.ActiveReferralProgramForAccount(pr1.Patient.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	rp = referralProgram.Data.(promotions.ReferralProgram)
	test.Equals(t, 2, rp.UsersAssociatedCount())

	// ensure that if we update the referrals program, the old promotion still works
	requestData["value"] = 1000
	resp, err = testData.AuthPostJSON(testData.APIServer.URL+apipaths.ReferralProgramsTemplateURLPath, admin.AccountID.Int64(), requestData, &responseData)
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// now query the first patient to get the latest referral program
	resp, err = testData.AuthGet(testData.APIServer.URL+apipaths.ReferralsURLPath, pr1.Patient.AccountID.Int64())
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// now get another user to claim the previous code
	_, err = promotions.AssociatePromoCode("kunal3@test.com", "California", code, testData.DataAPI, testData.AuthAPI, testData.Config.AnalyticsLogger)
	test.OK(t, err)

	// sign the user up
	pr3 := test_integration.SignupTestPatientWithEmail("kunal3@test.com", t, testData)

	// there should be a pending promotion for the patient
	pendingPromotions, err := testData.DataAPI.PendingPromotionsForAccount(pr3.Patient.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	test.Equals(t, 1, len(pendingPromotions))

	// count should be 0 for the associated promotions given that the program  was updated and the code was used for the previous promotion
	referralProgram2, err := testData.DataAPI.ActiveReferralProgramForAccount(pr1.Patient.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	test.Equals(t, true, referralProgram.TemplateID != referralProgram2.TemplateID)
	rp = referralProgram2.Data.(promotions.ReferralProgram)
	test.Equals(t, 0, rp.UsersAssociatedCount())
	test.Equals(t, 0, rp.VisitsSubmittedCount())

}

func TestReferrals_ExistingPatientReferral(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	admin := test_integration.CreateRandomAdmin(t, testData)
	setupPromotionsTest(testData, t)

	// create referral program template
	title := "dollars off"
	description := "description"
	requestData := map[string]interface{}{
		"type":        "promo_money_off",
		"title":       title,
		"description": description,
		"group":       "new_user",
		"share_text": map[string]interface{}{
			"facebook": "agaHG",
			"sms":      "ADgagh",
			"default":  "aegagh",
		},
		"promotion": map[string]interface{}{
			"display_msg":  "dollars off",
			"success_msg":  "dollars off",
			"short_msg":    "dollars off",
			"for_new_user": true,
			"group":        "new_user",
			"value":        500,
		},
	}

	var responseData map[string]interface{}
	resp, err := testData.AuthPostJSON(testData.APIServer.URL+apipaths.ReferralProgramsTemplateURLPath, admin.AccountID.Int64(), requestData, &responseData)
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// now create patient
	pr1 := test_integration.SignupRandomTestPatientWithPharmacyAndAddress(t, testData)

	// now try to get the referral program for this patient
	resp, err = testData.AuthGet(testData.APIServer.URL+apipaths.ReferralsURLPath, pr1.Patient.AccountID.Int64())
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&responseData)
	test.OK(t, err)
	promotionURL := responseData["url"].(string)
	test.OK(t, err)
	slashIndex := strings.LastIndex(promotionURL, "/")
	code := promotionURL[slashIndex+1:]

	// now try and get another existing patient to claim the code
	pr2 := test_integration.SignupRandomTestPatientWithPharmacyAndAddress(t, testData)
	_, err = promotions.AssociatePromoCode(pr2.Patient.Email, "California", code, testData.DataAPI, testData.AuthAPI, testData.Config.AnalyticsLogger)
	test.OK(t, err)

	// ensure that the existing user now has a pending promotion
	pendingPromotions, err := testData.DataAPI.PendingPromotionsForAccount(pr2.Patient.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	test.Equals(t, 1, len(pendingPromotions))
}

func TestReferrals_NewDoctorReferral(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	stubSQSQueue := &common.SQSQueue{
		QueueURL:     "visit_url",
		QueueService: &sqs.StubSQS{},
	}
	testData.Config.VisitQueue = stubSQSQueue
	testData.StartAPIServer(t)

	setupPromotionsTest(testData, t)

	// create a doctor to get a referral program created for the doctor
	dr, email, password := test_integration.SignupRandomTestDoctor(t, testData)
	// get the doctor to login so that the doctor picks up the referral program
	params := url.Values{}
	params.Set("email", email)
	params.Set("password", password)
	req, err := http.NewRequest("POST", testData.APIServer.URL+apipaths.DoctorAuthenticateURLPath, strings.NewReader(params.Encode()))
	test.OK(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	// now get an unregistered patient to claim the code
	_, err = promotions.AssociatePromoCode("kunal@test.com", "Florida", fmt.Sprintf("dr%s", doctor.LastName), testData.DataAPI, testData.AuthAPI, testData.Config.AnalyticsLogger)
	test.OK(t, err)

	// now get this patient to signup
	pr := signupPatientWithVisit("kunal@test.com", testData, t)
	test_integration.AddTestPharmacyForPatient(pr.Patient.PatientID.Int64(), testData, t)
	test_integration.AddTestAddressForPatient(pr.Patient.PatientID.Int64(), testData, t)

	patientID := pr.Patient.PatientID.Int64()
	patientAccountID := pr.Patient.AccountID.Int64()
	test_integration.AddCreditCardForPatient(pr.Patient.PatientID.Int64(), testData, t)

	// at this point the doctor's referral program should indicate that the patient signed up
	referralProgram, err := testData.DataAPI.ActiveReferralProgramForAccount(doctor.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	rp := referralProgram.Data.(promotions.ReferralProgram)
	test.Equals(t, 1, rp.UsersAssociatedCount())
	test.Equals(t, 0, rp.VisitsSubmittedCount())

	// now get the patient to submit a visit
	startAndSubmitVisit(patientID, patientAccountID, stubSQSQueue, testData, t)

	// at this point the visit should show up in the doctor's inbox
	pendingItems, err := testData.DataAPI.GetPendingItemsInDoctorQueue(dr.DoctorID)
	test.OK(t, err)
	test.Equals(t, 1, len(pendingItems))

	referralProgram, err = testData.DataAPI.ActiveReferralProgramForAccount(doctor.AccountID.Int64(), promotions.Types)
	test.OK(t, err)
	rp = referralProgram.Data.(promotions.ReferralProgram)
	test.Equals(t, 1, rp.UsersAssociatedCount())
	test.Equals(t, 1, rp.VisitsSubmittedCount())
}

func TestReferrals_ExistingDoctorReferral(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close()
	testData.StartAPIServer(t)

	setupPromotionsTest(testData, t)

	// create a doctor to get a referral program created for the doctor
	dr, email, password := test_integration.SignupRandomTestDoctor(t, testData)
	doctor, err := testData.DataAPI.GetDoctorFromID(dr.DoctorID)
	test.OK(t, err)

	// get the doctor to login so that the doctor picks up the referral program
	params := url.Values{}
	params.Set("email", email)
	params.Set("password", password)
	req, err := http.NewRequest("POST", testData.APIServer.URL+apipaths.DoctorAuthenticateURLPath, strings.NewReader(params.Encode()))
	test.OK(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	test.OK(t, err)
	defer resp.Body.Close()
	test.Equals(t, http.StatusOK, resp.StatusCode)

	// now try and get an existing patient to claim the code
	pr := signupPatientWithVisit("agkn@gmai.com", testData, t)
	test_integration.AddTestAddressForPatient(pr.Patient.PatientID.Int64(), testData, t)
	test_integration.AddTestPharmacyForPatient(pr.Patient.PatientID.Int64(), testData, t)

	_, err = promotions.AssociatePromoCode(pr.Patient.Email, "California", fmt.Sprintf("dr%s", doctor.LastName), testData.DataAPI, testData.AuthAPI, testData.Config.AnalyticsLogger)
	test.OK(t, err)

	// at this point the patient should have a doctor assigned to their care team
	careTeamMembers, err := testData.DataAPI.GetActiveMembersOfCareTeamForPatient(pr.Patient.PatientID.Int64(), false)
	test.OK(t, err)
	test.Equals(t, 1, len(careTeamMembers))
	test.Equals(t, dr.DoctorID, careTeamMembers[0].ProviderID)
}
