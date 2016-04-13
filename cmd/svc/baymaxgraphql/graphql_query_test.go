package main

import (
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/errors"

	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/models"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/raccess"
	ramock "github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/raccess/mock"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/device"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/backend/svc/directory"
	"github.com/sprucehealth/backend/svc/threading"
	"github.com/sprucehealth/backend/test"
	"github.com/sprucehealth/graphql"
	"github.com/sprucehealth/graphql/gqlerrors"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

func TestNodeQuery(t *testing.T) {
	nodeField := queryType.Fields()["node"]

	ra := ramock.New(t)
	acc := &auth.Account{ID: "account_12345"}
	ctx := context.Background()
	ctx = gqlctx.WithAccount(ctx, acc)
	p := graphql.ResolveParams{
		Context: ctx,
		Info: graphql.ResolveInfo{
			RootValue: map[string]interface{}{
				raccess.ParamKey: ra,
				"service":        &service{},
			},
		},
	}

	// Organization

	id := "entity_123"
	p.Args = map[string]interface{}{
		"id": id,
	}
	ra.Expect(mock.NewExpectation(ra.Entity, id, []directory.EntityInformation{
		directory.EntityInformation_CONTACTS,
	}, int64(0)).WithReturns(
		&directory.Entity{
			Type: directory.EntityType_ORGANIZATION,
			ID:   id,
			Info: &directory.EntityInfo{
				DisplayName: "Org",
			},
		}, nil))
	ra.Expect(mock.NewExpectation(ra.EntityForAccountID, id, acc.ID).WithReturns(
		&directory.Entity{
			Type: directory.EntityType_INTERNAL,
			ID:   "entity_222",
			Info: &directory.EntityInfo{
				DisplayName: "Mem",
			},
			Memberships: []*directory.Entity{
				{ID: id},
			},
		}, nil))
	res, err := nodeField.Resolve(p)
	test.OK(t, err)
	test.Equals(t, &models.Organization{
		ID:       id,
		Name:     "Org",
		Contacts: []*models.ContactInfo{},
		Entity: &models.Entity{
			ID:          "entity_222",
			IsEditable:  true,
			DisplayName: "Mem",
			Contacts:    []*models.ContactInfo{},
			IsInternal:  true,
			Gender:      "UNKNOWN",
		},
	}, res)
	mock.FinishAll(ra)

	// Entity

	id = "entity_123"
	p.Args = map[string]interface{}{
		"id": id,
	}
	ra.Expect(mock.NewExpectation(ra.Entity, id, []directory.EntityInformation{
		directory.EntityInformation_CONTACTS}, int64(0)).WithReturns(&directory.Entity{
		Type: directory.EntityType_EXTERNAL,
		ID:   id,
		Info: &directory.EntityInfo{
			DisplayName: "Someone",
			Gender:      directory.EntityInfo_MALE,
		},
	}, nil))
	res, err = nodeField.Resolve(p)
	test.OK(t, err)
	test.Equals(t, &models.Entity{ID: id, IsEditable: true, IsInternal: false, DisplayName: "Someone", Contacts: []*models.ContactInfo{}, Gender: "MALE"}, res)
	mock.FinishAll(ra)

	// Thread

	id = "t_123"
	p.Args = map[string]interface{}{
		"id": id,
	}
	ra.Expect(mock.NewExpectation(ra.Thread, id, "").WithReturns(&threading.Thread{
		ID:              id,
		OrganizationID:  "entity_1",
		PrimaryEntityID: "entity_2",
		Type:            threading.ThreadType_EXTERNAL,
	}, nil))

	ra.Expect(mock.NewExpectation(ra.EntityForAccountID, "entity_1", acc.ID).WithReturns(
		&directory.Entity{
			Type: directory.EntityType_INTERNAL,
			ID:   "entity_222",
			Info: &directory.EntityInfo{
				DisplayName: "Someone",
				Gender:      directory.EntityInfo_FEMALE,
			},
			Memberships: []*directory.Entity{
				{
					Type: directory.EntityType_ORGANIZATION,
					ID:   "entity_1",
				},
			},
		}, nil))
	ra.Expect(mock.NewExpectation(ra.Thread, id, "entity_222").WithReturns(&threading.Thread{
		ID:              id,
		OrganizationID:  "entity_1",
		PrimaryEntityID: "entity_2",
		SystemTitle:     "Someone",
		Type:            threading.ThreadType_EXTERNAL,
		Unread:          true,
		UnreadReference: true,
	}, nil))

	res, err = nodeField.Resolve(p)
	test.OK(t, err)
	res.(*models.Thread).PrimaryEntity = nil
	test.Equals(t, &models.Thread{
		ID: id,
		AllowInternalMessages: true,
		AllowDelete:           true,
		AllowExternalDelivery: true,
		AllowMentions:         true,
		OrganizationID:        "entity_1",
		PrimaryEntityID:       "entity_2",
		Title:                 "Someone",
		LastPrimaryEntityEndpoints: []*models.Endpoint{},
		Type:            models.ThreadTypeExternal,
		Unread:          true,
		UnreadReference: true,
	}, res)
	mock.FinishAll(ra)

	// Thread item

	id = "ti_123"
	p.Args = map[string]interface{}{
		"id": id,
	}
	ra.Expect(mock.NewExpectation(ra.ThreadItem, id).WithReturns(&threading.ThreadItem{
		ID:            id,
		Timestamp:     1234,
		ActorEntityID: "entity_1",
		Internal:      true,
		Type:          threading.ThreadItem_MESSAGE,
		Item: &threading.ThreadItem_Message{
			Message: &threading.Message{
				Title:  "abc",
				Text:   "hello",
				Status: threading.Message_NORMAL,
				Source: &threading.Endpoint{
					ID:      "555-555-5555",
					Channel: threading.Endpoint_VOICE,
				},
				TextRefs: []*threading.Reference{
					{ID: "e2", Type: threading.Reference_ENTITY},
				},
			},
		},
	}, nil))
	res, err = nodeField.Resolve(p)
	test.OK(t, err)
	test.Equals(t, &models.ThreadItem{
		ID:            id,
		Timestamp:     1234,
		ActorEntityID: "entity_1",
		Internal:      true,
		Data: &models.Message{
			ThreadItemID:  id,
			SummaryMarkup: "abc",
			TextMarkup:    "hello",
			Source: &models.Endpoint{
				Channel: models.EndpointChannelVoice,
				ID:      "555-555-5555",
			},
			Refs: []*models.Reference{
				{ID: "e2", Type: "entity"},
			},
		},
	}, res)
	mock.FinishAll(ra)

	// Saved query

	id = "sq_123"
	p.Args = map[string]interface{}{
		"id": id,
	}
	ra.Expect(mock.NewExpectation(ra.SavedQuery, id).WithReturns(&threading.SavedQuery{
		ID:             id,
		OrganizationID: "entity_1",
	}, nil))
	res, err = nodeField.Resolve(p)
	test.OK(t, err)
	test.Equals(t, &models.SavedThreadQuery{ID: id, OrganizationID: "entity_1"}, res)
	mock.FinishAll(ra)
}

func TestTeamThread_OlderVersion(t *testing.T) {
	nodeField := queryType.Fields()["node"]

	ra := ramock.New(t)
	acc := &auth.Account{ID: "account_12345"}
	ctx := context.Background()
	ctx = gqlctx.WithAccount(ctx, acc)
	ctx = gqlctx.WithSpruceHeaders(ctx, &device.SpruceHeaders{
		AppType:         "baymax",
		AppEnvironment:  "dev",
		AppVersion:      &encoding.Version{Major: 1},
		AppBuild:        "001",
		Platform:        common.IOS,
		PlatformVersion: "7.1.1",
		Device:          "Phone",
		DeviceModel:     "iPhone6,1",
		DeviceID:        "12917415",
	})
	p := graphql.ResolveParams{
		Context: ctx,
		Info: graphql.ResolveInfo{
			RootValue: map[string]interface{}{
				raccess.ParamKey: ra,
				"service":        &service{},
			},
		},
	}

	id := "t_123"
	p.Args = map[string]interface{}{
		"id": id,
	}
	ra.Expect(mock.NewExpectation(ra.Thread, id, "").WithReturns(&threading.Thread{
		ID:              id,
		OrganizationID:  "entity_1",
		PrimaryEntityID: "entity_2",
		Type:            threading.ThreadType_TEAM,
	}, nil))

	_, err := nodeField.Resolve(p)
	fe, ok := err.(gqlerrors.FormattedError)
	test.Equals(t, true, ok)
	test.Equals(t, string(errors.ErrTypeNotSupported), fe.Type)
	mock.FinishAll(ra)
}

func TestSubdomainQuery_Unavailable(t *testing.T) {
	subdomainField := queryType.Fields()["subdomain"]

	ra := ramock.New(t)
	acc := &auth.Account{ID: "account:12345"}
	ctx := context.Background()
	ctx = gqlctx.WithAccount(ctx, acc)
	p := graphql.ResolveParams{
		Context: ctx,
		Info: graphql.ResolveInfo{
			RootValue: map[string]interface{}{
				raccess.ParamKey: ra,
			},
		},
	}

	p.Args = map[string]interface{}{
		"value": "mypractice",
	}
	ra.Expect(mock.NewExpectation(ra.EntityDomain, "", "mypractice").WithReturns(
		&directory.LookupEntityDomainResponse{
			EntityID: "dkgj",
			Domain:   "mypractice",
		},
		nil),
	)

	res, err := subdomainField.Resolve(p)
	test.OK(t, err)
	test.Equals(t, &models.Subdomain{
		Available: false,
	}, res)
	mock.FinishAll(ra)
}

func TestSubdomainQuery_Available(t *testing.T) {
	subdomainField := queryType.Fields()["subdomain"]

	ra := ramock.New(t)
	acc := &auth.Account{ID: "account:12345"}
	ctx := context.Background()
	ctx = gqlctx.WithAccount(ctx, acc)
	p := graphql.ResolveParams{
		Context: ctx,
		Info: graphql.ResolveInfo{
			RootValue: map[string]interface{}{
				raccess.ParamKey: ra,
			},
		},
	}

	// Available
	p.Args = map[string]interface{}{
		"value": "anotherpractice",
	}
	ra.Expect(mock.NewExpectation(ra.EntityDomain, "", "anotherpractice").WithReturns(
		&directory.LookupEntityDomainResponse{},
		grpcErrorf(codes.NotFound, "entity_domain not found")),
	)
	res, err := subdomainField.Resolve(p)
	test.OK(t, err)
	test.Equals(t, &models.Subdomain{
		Available: true,
	}, res)
	mock.FinishAll(ra)
}
