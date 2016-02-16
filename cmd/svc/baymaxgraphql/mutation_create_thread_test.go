package main

import (
	"encoding/json"
	"testing"

	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/directory"
	"github.com/sprucehealth/backend/svc/threading"
	"github.com/sprucehealth/backend/test"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func init() {
	conc.Testing = true
}

func TestCreateThreadMutation_NoExistingThreads(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	acc := &account{
		ID: "a_1",
	}
	ctx = ctxWithAccount(ctx, acc)

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_EXTERNAL_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_ExternalID{
			ExternalID: "a_1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
				directory.EntityInformation_CONTACTS,
			},
		},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_creator",
				Type: directory.EntityType_INTERNAL,
				Memberships: []*directory.Entity{
					{ID: "e_org", Type: directory.EntityType_ORGANIZATION},
				},
			},
		},
	}, nil))

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: "someone@example.com",
		RequestedInformation: &directory.RequestedInformation{
			Depth:             1,
			EntityInformation: []directory.EntityInformation{directory.EntityInformation_MEMBERSHIPS},
		},
	}).WithReturns((*directory.LookupEntitiesByContactResponse)(nil), grpc.Errorf(codes.NotFound, "No entities found")))

	entityInfo := &directory.EntityInfo{
		FirstName:     "firstName",
		MiddleInitial: "middleInitial",
		LastName:      "lastName",
		GroupName:     "groupName",
		ShortTitle:    "shortTitle",
		DisplayName:   "firstName middleInitial. lastName, shortTitle",
		LongTitle:     "longTitle",
		Note:          "note",
	}
	contacts := []*directory.Contact{
		{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
		{ContactType: directory.ContactType_PHONE, Value: "+14155555555", Label: "Phone"},
	}
	g.dirC.Expect(mock.NewExpectation(g.dirC.CreateEntity, &directory.CreateEntityRequest{
		Type: directory.EntityType_EXTERNAL,
		InitialMembershipEntityID: "e_org",
		Contacts:                  contacts,
		EntityInfo:                entityInfo,
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
		},
	}).WithReturns(&directory.CreateEntityResponse{
		Entity: &directory.Entity{
			ID:       "e_patient",
			Type:     directory.EntityType_EXTERNAL,
			Contacts: contacts,
			Info:     entityInfo,
		},
	}, nil))

	g.thC.Expect(mock.NewExpectation(g.thC.CreateEmptyThread, &threading.CreateEmptyThreadRequest{
		UUID:           "zztop",
		OrganizationID: "e_org",
		FromEntityID:   "e_creator",
		Source: &threading.Endpoint{
			Channel: threading.Endpoint_APP,
			ID:      "e_creator",
		},
		PrimaryEntityID: "e_patient",
		Summary:         "New conversation", // TODO: not sure what we want here. it's a fallback if there's no posts made in the thread.

	}).WithReturns(&threading.CreateEmptyThreadResponse{
		Thread: &threading.Thread{
			ID: "t_1",
		},
	}, nil))

	res := g.query(ctx, `
		mutation _ {
			createThread(input: {
				clientMutationId: "a1b2c3",
				uuid: "zztop",
				organizationID: "e_org",
				entityInfo: {
					firstName:     "firstName",
					middleInitial: "middleInitial",
					lastName:      "lastName",
					groupName:     "groupName",
					shortTitle:    "shortTitle",
					longTitle:     "longTitle",
					note:          "note",
					contactInfos:  [
						{type: PHONE, value: "+14155555555", label: "Phone"},
						{type: EMAIL, value: "someone@example.com", label: "Email"},
						{type: PHONE, value: "1", label: "Invalid"}
					],
				},
				createForContactInfo: {type: EMAIL, value: "someone@example.com"},
			}) {
				clientMutationId
				success
				thread {
					id
					allowInternalMessages
					isDeletable
				}
			}
		}`, nil)
	b, err := json.MarshalIndent(res, "", "\t")
	test.OK(t, err)
	test.Equals(t, `{
	"data": {
		"createThread": {
			"clientMutationId": "a1b2c3",
			"success": true,
			"thread": {
				"allowInternalMessages": true,
				"id": "t_1",
				"isDeletable": true
			}
		}
	}
}`, string(b))
}

func TestCreateThreadMutation_DifferentOrg(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	acc := &account{
		ID: "a_1",
	}
	ctx = ctxWithAccount(ctx, acc)

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_EXTERNAL_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_ExternalID{
			ExternalID: "a_1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
				directory.EntityInformation_CONTACTS,
			},
		},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_creator",
				Type: directory.EntityType_INTERNAL,
				Memberships: []*directory.Entity{
					{ID: "e_org", Type: directory.EntityType_ORGANIZATION},
				},
			},
		},
	}, nil))

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: "someone@example.com",
		RequestedInformation: &directory.RequestedInformation{
			Depth:             1,
			EntityInformation: []directory.EntityInformation{directory.EntityInformation_MEMBERSHIPS},
		},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_existing_1",
				Type: directory.EntityType_EXTERNAL,
				Contacts: []*directory.Contact{
					{ContactType: directory.ContactType_PHONE, Value: "+14155555555", Label: "Phone"},
					{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
				},
				Info: &directory.EntityInfo{
					FirstName:     "differentName",
					MiddleInitial: "middleInitial",
					LastName:      "lastName",
					DisplayName:   "firstName middleInitial. lastName, shortTitle",
					GroupName:     "groupName",
					ShortTitle:    "shortTitle",
					LongTitle:     "longTitle",
					Note:          "note",
				},
				Memberships: []*directory.Entity{
					{ID: "differentOrg"},
				},
			},
		},
	}, nil))

	// The rest should behave like a create because the found entity doesn't match the org

	entityInfo := &directory.EntityInfo{
		FirstName:     "firstName",
		MiddleInitial: "middleInitial",
		LastName:      "lastName",
		DisplayName:   "firstName middleInitial. lastName, shortTitle",
		GroupName:     "groupName",
		ShortTitle:    "shortTitle",
		LongTitle:     "longTitle",
		Note:          "note",
	}
	contacts := []*directory.Contact{
		{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
		{ContactType: directory.ContactType_PHONE, Value: "+14155555555", Label: "Phone"},
	}
	g.dirC.Expect(mock.NewExpectation(g.dirC.CreateEntity, &directory.CreateEntityRequest{
		Type: directory.EntityType_EXTERNAL,
		InitialMembershipEntityID: "e_org",
		Contacts:                  contacts,
		EntityInfo:                entityInfo,
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
		},
	}).WithReturns(&directory.CreateEntityResponse{
		Entity: &directory.Entity{
			ID:       "e_patient",
			Type:     directory.EntityType_EXTERNAL,
			Contacts: contacts,
			Info:     entityInfo,
		},
	}, nil))

	g.thC.Expect(mock.NewExpectation(g.thC.CreateEmptyThread, &threading.CreateEmptyThreadRequest{
		UUID:           "zztop",
		OrganizationID: "e_org",
		FromEntityID:   "e_creator",
		Source: &threading.Endpoint{
			Channel: threading.Endpoint_APP,
			ID:      "e_creator",
		},
		PrimaryEntityID: "e_patient",
		Summary:         "New conversation", // TODO: not sure what we want here. it's a fallback if there's no posts made in the thread.

	}).WithReturns(&threading.CreateEmptyThreadResponse{
		Thread: &threading.Thread{
			ID: "t_1",
		},
	}, nil))

	res := g.query(ctx, `
		mutation _ {
			createThread(input: {
				clientMutationId: "a1b2c3",
				uuid: "zztop",
				organizationID: "e_org",
				entityInfo: {
					firstName:     "firstName",
					middleInitial: "middleInitial",
					lastName:      "lastName",
					groupName:     "groupName",
					shortTitle:    "shortTitle",
					longTitle:     "longTitle",
					note:          "note",
					contactInfos:  [
						{type: PHONE, value: "+14155555555", label: "Phone"},
						{type: EMAIL, value: "someone@example.com", label: "Email"},
					],
				},
				createForContactInfo: {type: EMAIL, value: "someone@example.com"},
			}) {
				clientMutationId
				success
				thread {
					id
					title
					allowInternalMessages
					isDeletable
				}
			}
		}`, nil)
	b, err := json.MarshalIndent(res, "", "\t")
	test.OK(t, err)
	test.Equals(t, `{
	"data": {
		"createThread": {
			"clientMutationId": "a1b2c3",
			"success": true,
			"thread": {
				"allowInternalMessages": true,
				"id": "t_1",
				"isDeletable": true,
				"title": "firstName middleInitial. lastName, shortTitle"
			}
		}
	}
}`, string(b))
}

func TestCreateThreadMutation_ExistingThreads_DifferentName(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	acc := &account{
		ID: "a_1",
	}
	ctx = ctxWithAccount(ctx, acc)

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_EXTERNAL_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_ExternalID{
			ExternalID: "a_1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
				directory.EntityInformation_CONTACTS,
			},
		},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_creator",
				Type: directory.EntityType_INTERNAL,
				Memberships: []*directory.Entity{
					{ID: "e_org", Type: directory.EntityType_ORGANIZATION},
				},
			},
		},
	}, nil))

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: "someone@example.com",
		RequestedInformation: &directory.RequestedInformation{
			Depth:             1,
			EntityInformation: []directory.EntityInformation{directory.EntityInformation_MEMBERSHIPS},
		},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_existing_1",
				Type: directory.EntityType_EXTERNAL,
				Contacts: []*directory.Contact{
					{ContactType: directory.ContactType_PHONE, Value: "+14155555555", Label: "Phone"},
					{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
				},
				Info: &directory.EntityInfo{
					FirstName:     "differentName",
					MiddleInitial: "middleInitial",
					LastName:      "lastName",
					GroupName:     "groupName",
					ShortTitle:    "shortTitle",
					LongTitle:     "longTitle",
					Note:          "note",
				},
				Memberships: []*directory.Entity{
					{ID: "e_org"},
				},
			},
			{
				ID:   "e_existing_2",
				Type: directory.EntityType_EXTERNAL,
				Contacts: []*directory.Contact{
					{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
					{ContactType: directory.ContactType_PHONE, Value: "+16305555555", Label: "Phone"},
				},
				Info: &directory.EntityInfo{
					FirstName:     "otherName",
					MiddleInitial: "middleInitial",
					LastName:      "lastName",
					GroupName:     "groupName",
					ShortTitle:    "shortTitle",
					LongTitle:     "longTitle",
					Note:          "note",
				},
				Memberships: []*directory.Entity{
					{ID: "e_org"},
				},
			},
		},
	}, nil))

	g.thC.Expect(mock.NewExpectation(g.thC.ThreadsForMember, &threading.ThreadsForMemberRequest{
		EntityID:    "e_existing_1",
		PrimaryOnly: true,
	}).WithReturns(&threading.ThreadsForMemberResponse{
		Threads: []*threading.Thread{
			{ID: "t_1", OrganizationID: "e_org", PrimaryEntityID: "e_existing_1"},
		},
	}, nil))

	g.thC.Expect(mock.NewExpectation(g.thC.ThreadsForMember, &threading.ThreadsForMemberRequest{
		EntityID:    "e_existing_2",
		PrimaryOnly: true,
	}).WithReturns(&threading.ThreadsForMemberResponse{
		Threads: []*threading.Thread{
			{ID: "t_2", OrganizationID: "e_org", PrimaryEntityID: "e_existing_2"},
		},
	}, nil))

	res := g.query(ctx, `
		mutation _ {
			createThread(input: {
				clientMutationId: "a1b2c3",
				uuid: "zztop",
				organizationID: "e_org",
				entityInfo: {
					firstName:     "firstName",
					middleInitial: "middleInitial",
					lastName:      "lastName",
					groupName:     "groupName",
					shortTitle:    "shortTitle",
					longTitle:     "longTitle",
					note:          "note",
					contactInfos:  [
						{type: PHONE, value: "+14155555555", label: "Phone"},
						{type: EMAIL, value: "someone@example.com", label: "Email"},
					],
				},
				createForContactInfo: {type: EMAIL, value: "someone@example.com"},
			}) {
				clientMutationId
				success
				errorCode
				thread {
					allowInternalMessages
					id
					title
				}
				existingThreads {
					allowInternalMessages
					id
					title
				}
				nameDiffers
			}
		}`, nil)
	b, err := json.MarshalIndent(res, "", "\t")
	test.OK(t, err)
	test.Equals(t, `{
	"data": {
		"createThread": {
			"clientMutationId": "a1b2c3",
			"errorCode": "EXISTING_THREAD",
			"existingThreads": [
				{
					"allowInternalMessages": true,
					"id": "t_1",
					"title": "(415) 555-5555"
				},
				{
					"allowInternalMessages": true,
					"id": "t_2",
					"title": "someone@example.com"
				}
			],
			"nameDiffers": true,
			"success": false,
			"thread": {
				"allowInternalMessages": true,
				"id": "t_1",
				"title": "(415) 555-5555"
			}
		}
	}
}`, string(b))
}

func TestCreateThreadMutation_ExistingThreads_SameName(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	acc := &account{
		ID: "a_1",
	}
	ctx = ctxWithAccount(ctx, acc)

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_EXTERNAL_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_ExternalID{
			ExternalID: "a_1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
				directory.EntityInformation_CONTACTS,
			},
		},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_creator",
				Type: directory.EntityType_INTERNAL,
				Memberships: []*directory.Entity{
					{ID: "e_org", Type: directory.EntityType_ORGANIZATION},
				},
			},
		},
	}, nil))

	g.dirC.Expect(mock.NewExpectation(g.dirC.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: "someone@example.com",
		RequestedInformation: &directory.RequestedInformation{
			Depth:             1,
			EntityInformation: []directory.EntityInformation{directory.EntityInformation_MEMBERSHIPS},
		},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				ID:   "e_existing_1",
				Type: directory.EntityType_EXTERNAL,
				Contacts: []*directory.Contact{
					{ContactType: directory.ContactType_PHONE, Value: "+14155555555", Label: "Phone"},
					{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
				},
				Info: &directory.EntityInfo{
					FirstName:     "differentName",
					MiddleInitial: "middleInitial",
					LastName:      "lastName",
					GroupName:     "groupName",
					ShortTitle:    "shortTitle",
					LongTitle:     "longTitle",
					Note:          "note",
				},
				Memberships: []*directory.Entity{
					{ID: "e_org"},
				},
			},
			{
				ID:   "e_existing_2",
				Type: directory.EntityType_EXTERNAL,
				Contacts: []*directory.Contact{
					{ContactType: directory.ContactType_EMAIL, Value: "someone@example.com", Label: "Email"},
					{ContactType: directory.ContactType_PHONE, Value: "+16305555555", Label: "Phone"},
				},
				Info: &directory.EntityInfo{
					FirstName:     "firstName",
					MiddleInitial: "middleInitial",
					LastName:      "lastName",
					GroupName:     "groupName",
					ShortTitle:    "shortTitle",
					LongTitle:     "longTitle",
					Note:          "note",
				},
				Memberships: []*directory.Entity{
					{ID: "e_org"},
				},
			},
		},
	}, nil))

	g.thC.Expect(mock.NewExpectation(g.thC.ThreadsForMember, &threading.ThreadsForMemberRequest{
		EntityID:    "e_existing_1",
		PrimaryOnly: true,
	}).WithReturns(&threading.ThreadsForMemberResponse{
		Threads: []*threading.Thread{
			{ID: "t_1", OrganizationID: "e_org", PrimaryEntityID: "e_existing_1"},
		},
	}, nil))

	g.thC.Expect(mock.NewExpectation(g.thC.ThreadsForMember, &threading.ThreadsForMemberRequest{
		EntityID:    "e_existing_2",
		PrimaryOnly: true,
	}).WithReturns(&threading.ThreadsForMemberResponse{
		Threads: []*threading.Thread{
			{ID: "t_2", OrganizationID: "e_org", PrimaryEntityID: "e_existing_2"},
		},
	}, nil))

	res := g.query(ctx, `
		mutation _ {
			createThread(input: {
				clientMutationId: "a1b2c3",
				uuid: "zztop",
				organizationID: "e_org",
				entityInfo: {
					firstName:     "firstName",
					middleInitial: "middleInitial",
					lastName:      "lastName",
					groupName:     "groupName",
					shortTitle:    "shortTitle",
					longTitle:     "longTitle",
					note:          "note",
					contactInfos:  [
						{type: PHONE, value: "+14155555555", label: "Phone"},
						{type: EMAIL, value: "someone@example.com", label: "Email"},
					],
				},
				createForContactInfo: {type: EMAIL, value: "someone@example.com"},
			}) {
				clientMutationId
				success
				errorCode
				thread {
					id
					allowInternalMessages
				}
				existingThreads {
					id
					allowInternalMessages
				}
				nameDiffers
			}
		}`, nil)
	b, err := json.MarshalIndent(res, "", "\t")
	test.OK(t, err)
	test.Equals(t, `{
	"data": {
		"createThread": {
			"clientMutationId": "a1b2c3",
			"errorCode": "EXISTING_THREAD",
			"existingThreads": [
				{
					"allowInternalMessages": true,
					"id": "t_1"
				},
				{
					"allowInternalMessages": true,
					"id": "t_2"
				}
			],
			"nameDiffers": false,
			"success": false,
			"thread": {
				"allowInternalMessages": true,
				"id": "t_2"
			}
		}
	}
}`, string(b))
}
