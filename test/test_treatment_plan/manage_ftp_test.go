package test_treatment_plan

import (
	"fmt"
	"testing"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/erx"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/responses"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/backend/test/test_integration"
)

func TestManageFTP(t *testing.T) {
	testData := test_integration.SetupTest(t)
	defer testData.Close(t)
	testData.StartAPIServer(t)

	dr, _, _ := test_integration.SignupRandomTestDoctor(t, testData)
	cli := test_integration.DoctorClient(testData, t, dr.DoctorID)

	treatment1 := &common.Treatment{
		DrugDBIDs: map[string]string{
			erx.LexiDrugSynID:     "1234",
			erx.LexiGenProductID:  "12345",
			erx.LexiSynonymTypeID: "123556",
			erx.NDC:               "2415",
		},
		DrugInternalName:        "Drug1 (Route1 - Form1)",
		DosageStrength:          "Strength1",
		DispenseValue:           5,
		DispenseUnitDescription: "Tablet",
		DispenseUnitID:          encoding.DeprecatedNewObjectID(19),
		NumberRefills: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 5,
		},
		SubstitutionsAllowed: false,
		DaysSupply: encoding.NullInt64{
			IsValid:    true,
			Int64Value: 5,
		},
		PatientInstructions: "Take once daily",
		OTC:                 false,
	}

	regimenSections := []*common.RegimenSection{
		{
			Name: "Morning",
			Steps: []*common.DoctorInstructionItem{
				{
					Text: "Step 1",
				},
				{
					Text: "Step 2",
				},
			},
		},
		{
			Name: "Nighttime",
			Steps: []*common.DoctorInstructionItem{
				{
					Text: "Step 1",
				},
				{
					Text: "Step 2",
				},
			},
		},
	}

	_, resourceGuideIDs := test_integration.CreateTestResourceGuides(t, testData)

	favoriteTreatmentPlan := &responses.FavoriteTreatmentPlan{
		DoctorID: dr.DoctorID,
		Name:     "FTP TEST 1",
		RegimenPlan: &common.RegimenPlan{
			Sections: regimenSections,
		},
		TreatmentList: &common.TreatmentList{
			Treatments: []*common.Treatment{treatment1},
		},
		ResourceGuides: []*responses.ResourceGuide{
			{
				ID: resourceGuideIDs[0],
			},
		},
	}

	ftpCreated, err := cli.CreateFavoriteTreatmentPlan(favoriteTreatmentPlan)
	test.OK(t, err)

	/// get ftps for the doctor and ensure they match
	favoriteTreatmentPlans, err := testData.DataAPI.FavoriteTreatmentPlansForDoctor(dr.DoctorID, api.AcnePathwayTag)
	test.OK(t, err)
	test.Equals(t, 1, len(favoriteTreatmentPlans))
	fmt.Println(favoriteTreatmentPlans)
	test.Equals(t, favoriteTreatmentPlans["health_condition_acne"][0].ID.Int64(), ftpCreated.ID.Int64())

	// now lets go ahead and modify the FTP to ensure that it registers
	previousFTPID := ftpCreated.ID.Int64()
	ftpCreated.Name = "FTP TEST 3"
	ftpCreated.RegimenPlan.Sections = ftpCreated.RegimenPlan.Sections[:1]
	ftpCreated.ResourceGuides = append(ftpCreated.ResourceGuides, &responses.ResourceGuide{
		ID: resourceGuideIDs[1],
	})
	// now lets go ahead and make the call to add/modify ftps again
	ftpCreated, err = cli.UpdateFavoriteTreatmentPlan(ftpCreated)
	test.OK(t, err)

	// ensure that the modification was successful
	favoriteTreatmentPlans, err = testData.DataAPI.FavoriteTreatmentPlansForDoctor(dr.DoctorID, api.AcnePathwayTag)
	test.OK(t, err)
	test.Equals(t, 1, len(favoriteTreatmentPlans))
	test.Assert(t, previousFTPID != ftpCreated.ID.Int64(), fmt.Sprintf("Expected the FTP ID returned to have changes but instead remained %d", previousFTPID))
	test.Equals(t, ftpCreated.Name, favoriteTreatmentPlans["health_condition_acne"][0].Name)
	test.Equals(t, 1, len(ftpCreated.RegimenPlan.Sections))

	// now lets go ahead and delete the FTP
	err = cli.DeleteFavoriteTreatmentPlan(ftpCreated.ID.Int64())
	test.OK(t, err)

	favoriteTreatmentPlans, err = testData.DataAPI.FavoriteTreatmentPlansForDoctor(dr.DoctorID, api.AcnePathwayTag)
	test.OK(t, err)
	test.Equals(t, 0, len(favoriteTreatmentPlans))

}
