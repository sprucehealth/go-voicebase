package main

import (
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/backend/svc/directory"
	"github.com/sprucehealth/backend/svc/threading"
	"golang.org/x/net/context"
)

func TestCreateTeamThreadMutation(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	acc := &auth.Account{
		ID: "a_1",
	}
	organizationID := "e_org"
	ctx = gqlctx.WithAccount(ctx, acc)

	g.ra.Expect(mock.NewExpectation(g.ra.EntityForAccountID, organizationID, acc.ID).WithReturns(
		&directory.Entity{
			ID:   "e_creator",
			Type: directory.EntityType_INTERNAL,
			Memberships: []*directory.Entity{
				{ID: "e_org", Type: directory.EntityType_ORGANIZATION},
			},
		}, nil))

	g.ra.Expect(mock.NewExpectation(g.ra.CreateEmptyThread, &threading.CreateEmptyThreadRequest{
		UUID:            "zztop",
		OrganizationID:  organizationID,
		FromEntityID:    "e_creator",
		Summary:         "New conversation",
		UserTitle:       "thetitle",
		MemberEntityIDs: []string{"e1", "e2", "e3", "e_creator"},
		Type:            threading.ThreadType_TEAM,
	}).WithReturns(&threading.Thread{
		ID:          "t_1",
		Type:        threading.ThreadType_TEAM,
		UserTitle:   "thetitle",
		SystemTitle: "Person1, Poster",
	}, nil))

	res := g.query(ctx, `
		mutation _ {
			createTeamThread(input: {
				clientMutationId: "a1b2c3",
				uuid: "zztop",
				organizationID: "e_org",
				title: "thetitle",
				memberEntityIDs: ["e1", "e2", "e3"],
			}) {
				clientMutationId
				success
				thread {
					id
					allowInternalMessages
					allowDelete
					allowAddMembers
					allowRemoveMembers
					allowLeave
					allowUpdateTitle
					title
				}
			}
		}`, nil)
	responseEquals(t, `{
		"data": {
			"createTeamThread": {
				"clientMutationId": "a1b2c3",
				"success": true,
				"thread": {
					"allowAddMembers": true,
					"allowDelete": true,
					"allowInternalMessages": false,
					"allowLeave": true,
					"allowRemoveMembers": true,
					"allowUpdateTitle": true,
					"id": "t_1",
					"title": "thetitle"
				}
			}
		}}`, res)
}
