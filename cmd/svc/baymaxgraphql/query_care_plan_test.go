package main

import (
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/backend/svc/care"
	"golang.org/x/net/context"
)

func TestCarePlanQuery(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	acc := &auth.Account{
		ID:   "account_12345",
		Type: auth.AccountType_PROVIDER,
	}
	ctx = gqlctx.WithAccount(ctx, acc)

	g.ra.Expect(mock.NewExpectation(g.ra.CarePlan, "cpid").WithReturns(&care.CarePlan{
		ID:           "cpID",
		Name:         "cpname",
		Instructions: []*care.CarePlanInstruction{{Title: "title", Steps: []string{"one", "two"}}},
		Treatments: []*care.CarePlanTreatment{
			{
				EPrescribe:           true,
				Name:                 "tname",
				MedicationID:         "mid",
				Dosage:               "dosage",
				DispenseType:         "dispenseType",
				DispenseNumber:       1,
				Refills:              2,
				SubstitutionsAllowed: true,
				DaysSupply:           3,
				Sig:                  "sig",
				PharmacyID:           "pharm",
				PharmacyInstructions: "pharmInst",
			},
		},
	}, nil))

	res := g.query(ctx, `
		query _ {
			carePlan(id: "cpid") {
				id
				name
				instructions {
					title
					steps
				}
				treatments {
					ePrescribe
					name
					dosage
					dispenseType
					dispenseNumber
					refills
					substitutionsAllowed
					daysSupply
					sig
					pharmacyInstructions
				}
			}
		}`, nil)
	responseEquals(t, `{
	"data": {
		"carePlan": {
			"id": "cpID",
			"name": "cpname",
			"instructions": [{
				"title": "title",
				"steps": ["one", "two"]
			}],
			"treatments": [{
				"ePrescribe": true,
				"name": "tname",
				"dosage": "dosage",
				"dispenseType": "dispenseType",
				"dispenseNumber": 1,
				"refills": 2,
				"substitutionsAllowed": true,
				"daysSupply": 3,
				"sig": "sig",
				"pharmacyInstructions": "pharmInst"
			}]
		}
	}
}`, res)
}
