package main

import (
	"context"
	"fmt"

	segment "github.com/segmentio/analytics-go"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/apiaccess"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/errors"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/models"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/raccess"
	"github.com/sprucehealth/backend/device/devicectx"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/analytics"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/graphql"
	"github.com/sprucehealth/graphql/gqlerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// TOOD: this is a stubbed entity currently used in the primaryEntity for a thread. see comment there for more information
var stubEntity = &models.Entity{
	ID:                    "entity_stub",
	Gender:                genderUnknown,
	LastModifiedTimestamp: 1458949057,
}

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"me": &graphql.Field{
				Type: graphql.NewNonNull(meType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := p.Context
					acc := gqlctx.Account(p.Context)

					headers := devicectx.SpruceHeaders(ctx)
					if acc == nil {
						return nil, errors.ErrNotAuthenticated(p.Context)
					}
					cek := gqlctx.ClientEncryptionKey(p.Context)

					var platform string
					if headers != nil {
						platform = headers.Platform.String()
					}
					analytics.SegmentIdentify(&segment.Identify{
						UserId: acc.ID,
						Traits: map[string]interface{}{
							"platform": platform,
						},
						Context: map[string]interface{}{
							"ip":        remoteAddrFromParams(p),
							"userAgent": userAgentFromParams(p),
						},
					})
					return &models.Me{Account: transformAccountToResponse(acc), ClientEncryptionKey: cek}, nil
				},
			},
			"node": &graphql.Field{
				Type: graphql.NewNonNull(nodeInterfaceType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ram := raccess.ResourceAccess(p)
					svc := serviceFromParams(p)
					ctx := p.Context
					acc := gqlctx.Account(ctx)
					if acc == nil {
						return nil, errors.ErrNotAuthenticated(ctx)
					}
					id := p.Args["id"].(string)
					switch nodeIDPrefix(id) {
					case "account":
						return lookupAccount(ctx, ram, id)
					case "cp":
						return lookupCarePlan(ctx, ram, id)
					case "entity":
						// TOOD: this is a stubbed entity currently used in the primaryEntity for a thread. see comment there for more information
						if id == "entity_stub" {
							return stubEntity, nil
						}
						return lookupEntity(ctx, svc, ram, id)
					case "entityProfile":
						return lookupProfile(ctx, ram, id)
					case "ipc":
						return lookupCall(ctx, ram, id)
					case "payment":
						return lookupPaymentRequest(ctx, svc, ram, id)
					case "sq":
						return lookupSavedQuery(ctx, ram, id)
					case "t":
						return lookupThreadWithReadStatus(ctx, ram, acc, id)
					case "ti":
						return lookupThreadItem(ctx, ram, id, svc.webDomain, svc.mediaAPIDomain)
					case "visit":
						return lookupVisit(ctx, svc, ram, id)
					case "visitCategory":
						return lookupVisitCategory(ctx, svc, id)
					case "visitLayout":
						return lookupVisitLayout(ctx, svc, id)
					case "visitLayoutVersion":
						return lookupVisitLayoutVersion(ctx, svc, id)
					}
					return nil, fmt.Errorf("unknown ID '%s' in node query", id)
				},
			},
			"organization": &graphql.Field{
				Type: graphql.NewNonNull(organizationType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: apiaccess.Authenticated(func(p graphql.ResolveParams) (interface{}, error) {
					ram := raccess.ResourceAccess(p)
					ctx := p.Context
					svc := serviceFromParams(p)
					return lookupEntity(ctx, svc, ram, p.Args["id"].(string))
				}),
			},
			"savedThreadQuery": &graphql.Field{
				Type: graphql.NewNonNull(savedThreadQueryType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: apiaccess.Authenticated(func(p graphql.ResolveParams) (interface{}, error) {
					ram := raccess.ResourceAccess(p)
					ctx := p.Context
					return lookupSavedQuery(ctx, ram, p.Args["id"].(string))
				}),
			},
			"thread": &graphql.Field{
				Type: graphql.NewNonNull(threadType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: apiaccess.Authenticated(func(p graphql.ResolveParams) (interface{}, error) {
					ram := raccess.ResourceAccess(p)
					ctx := p.Context
					acc := gqlctx.Account(ctx)
					return lookupThreadWithReadStatus(ctx, ram, acc, p.Args["id"].(string))
				}),
			},
			"visit": &graphql.Field{
				Type: graphql.NewNonNull(visitType),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: apiaccess.Authenticated(func(p graphql.ResolveParams) (interface{}, error) {
					svc := serviceFromParams(p)
					ram := raccess.ResourceAccess(p)
					ctx := p.Context
					return lookupVisit(ctx, svc, ram, p.Args["id"].(string))
				}),
			},
			"subdomain": &graphql.Field{
				Type: graphql.NewNonNull(subdomainType),
				Args: graphql.FieldConfigArgument{
					"value": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: apiaccess.Authenticated(func(p graphql.ResolveParams) (interface{}, error) {
					ram := raccess.ResourceAccess(p)
					ctx := p.Context
					domain := p.Args["value"].(string)

					var available bool
					_, err := ram.EntityDomain(ctx, "", domain)
					if grpc.Code(err) == codes.NotFound {
						available = true
					} else if err != nil {
						return nil, err
					}

					return &models.Subdomain{
						Available: available,
					}, nil
				}),
			},
			"entity":                  entityQuery,
			"call":                    callQuery,
			"carePlan":                carePlanQuery,
			"forceUpgradeStatus":      forceUpgradeQuery,
			"medicationSearch":        medicationSearchQuery,
			"paymentRequest":          paymentRequestQuery,
			"pendingCalls":            pendingCallsQuery,
			"savedMessages":           savedMessagesQuery,
			"setting":                 settingsQuery,
			"threadsSearch":           threadsSearchQuery,
			"visitAutocompleteSearch": visitAutocompleteSearchQuery,
		},
	},
)

// TODO: This double read is inefficent/incorrect in the sense that we need the org ID to get the correct entity. We will use this for now until we can encode the organization ID into the thread ID
func lookupThreadWithReadStatus(ctx context.Context, ram raccess.ResourceAccessor, acc *auth.Account, id string) (*models.Thread, error) {
	th, err := lookupThread(ctx, ram, id, "")
	if grpc.Code(err) == codes.NotFound {
		return nil, errors.ErrNotFound(ctx, id)
	} else if e, ok := err.(gqlerrors.FormattedError); ok {
		return nil, e
	} else if err != nil {
		return nil, errors.InternalError(ctx, fmt.Errorf("account=%+v threadID=%s: %s", gqlctx.Account(ctx), id, err))
	}

	headers := devicectx.SpruceHeaders(ctx)
	if th.Type == models.ThreadTypeTeam {
		if !headers.AppVersion.GreaterThanOrEqualTo(&encoding.Version{Major: 1, Minor: 1, Patch: 0}) {
			return nil, errors.UserError(ctx, errors.ErrTypeNotSupported, "Team Conversations does not work on this version. Please refresh your browser or update your app to open this thread.")
		}
	}

	ent, err := entityInOrgForAccountID(ctx, ram, th.OrganizationID, acc)
	if errors.Type(err) == errors.ErrTypeNotFound {
		return nil, errors.UserError(ctx, errors.ErrTypeNotAuthorized, "You are not a member of the organzation")
	} else if err != nil {
		return nil, err
	}

	return lookupThread(ctx, ram, id, ent.ID)
}