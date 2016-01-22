package main

import (
	"errors"
	"strings"

	"golang.org/x/net/context"

	"github.com/graphql-go/graphql"
	"github.com/sprucehealth/backend/libs/golog"
)

var errNotAuthenticated = errors.New("not authenticated")

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"me": &graphql.Field{
				Type: graphql.NewNonNull(accountType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acc := accountFromContext(p.Context)
					if acc == nil {
						return nil, errNotAuthenticated
					}
					return acc, nil
				},
			},
			"node": &graphql.Field{
				Type: graphql.NewNonNull(nodeInterfaceType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					svc := serviceFromParams(p)
					ctx := p.Context
					acc := accountFromContext(ctx)
					if acc == nil {
						return nil, errNotAuthenticated
					}
					id := p.Args["id"].(string)
					i := strings.IndexByte(id, '_')
					prefix := id[:i]
					switch prefix {
					case "entity":
						return lookupEntity(ctx, svc, id)
					case "account":
						if id == acc.ID {
							return acc, nil
						}
						return lookupAccount(ctx, svc, id)
					case "sq":
						return lookupSavedQuery(ctx, svc, id)
					case "t":
						return lookupThreadWithReadStatus(ctx, svc, acc, id)
					case "ti":
						return lookupThreadItem(ctx, svc, id)
					}
					return nil, errors.New("unknown node type")
				},
			},
			// "listSavedThreadQueries": &graphql.Field{
			// 	Type: graphql.NewList(graphql.NewNonNull(savedThreadQueryType)),
			// 	Args: graphql.FieldConfigArgument{
			// 		"orgID": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			// 	},
			// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 		return nil, nil
			// 	},
			// },
			"organization": &graphql.Field{
				Type: graphql.NewNonNull(organizationType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					svc := serviceFromParams(p)
					ctx := p.Context
					acc := accountFromContext(ctx)
					if acc == nil {
						return nil, errNotAuthenticated
					}
					id := p.Args["id"].(string)
					return lookupEntity(ctx, svc, id)
				},
			},
			"savedThreadQuery": &graphql.Field{
				Type: graphql.NewNonNull(savedThreadQueryType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					svc := serviceFromParams(p)
					ctx := p.Context
					acc := accountFromContext(ctx)
					if acc == nil {
						return nil, errNotAuthenticated
					}
					id := p.Args["id"].(string)
					return lookupSavedQuery(ctx, svc, id)
				},
			},
			"thread": &graphql.Field{
				Type: graphql.NewNonNull(threadType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					svc := serviceFromParams(p)
					ctx := p.Context
					acc := accountFromContext(ctx)
					if acc == nil {
						return nil, errNotAuthenticated
					}
					it, err := lookupThreadWithReadStatus(ctx, svc, acc, p.Args["id"].(string))
					return it, err
				},
			},
		},
	},
)

// TODO: This double read is inefficent/incorrect in the sense that we need the org ID to get the correct entity. We will use this for now until we can encode the organization ID into the thread ID
func lookupThreadWithReadStatus(ctx context.Context, svc *service, acc *account, id string) (interface{}, error) {
	ith, err := lookupThread(ctx, svc, id, "")
	if err != nil {
		golog.Errorf(err.Error())
		return nil, errors.New("Unable to retrieve thread")
	}
	th, ok := ith.(*thread)
	if !ok {
		golog.Errorf("Expected *thread to be returned but got %T:%+v", ith, ith)
		return nil, errors.New("Unable to retrieve thread")
	}

	ent, err := svc.entityForAccountID(ctx, th.OrganizationID, acc.ID)
	if err != nil || ent == nil {
		golog.Errorf("Unable to find entity for account/org: %s/%s - %s", acc.ID, th.OrganizationID, err)
		return nil, errors.New("Unable to retrieve thread")
	}
	return lookupThread(ctx, svc, id, ent.ID)
}
