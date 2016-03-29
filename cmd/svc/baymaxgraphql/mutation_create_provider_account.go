package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/analytics-go"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/errors"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/models"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/raccess"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/phone"
	"github.com/sprucehealth/backend/libs/validate"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/backend/svc/directory"
	"github.com/sprucehealth/backend/svc/invite"
	"github.com/sprucehealth/backend/svc/threading"
	"github.com/sprucehealth/graphql"
	"google.golang.org/grpc"
)

const (
	supportThreadTitle    = "Spruce Support"
	onboardingThreadTitle = "Setup"
	teamSpruceInitialText = `This is your personal support conversation with Spruce Health.

If you're unsure about anything or need some help, send us a message here and a member of the Spruce Health team will respond.`
)

type createProviderAccountOutput struct {
	ClientMutationID    string                  `json:"clientMutationId,omitempty"`
	Success             bool                    `json:"success"`
	ErrorCode           string                  `json:"errorCode,omitempty"`
	ErrorMessage        string                  `json:"errorMessage,omitempty"`
	Token               string                  `json:"token,omitempty"`
	Account             *models.ProviderAccount `json:"account,omitempty"`
	ClientEncryptionKey string                  `json:"clientEncryptionKey,omitempty"`
}

var createProviderAccountInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateProviderAccountInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"clientMutationId":       newClientMutationIDInputField(),
		"uuid":                   newUUIDInputField(),
		"email":                  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"password":               &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"phoneNumber":            &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"firstName":              &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"lastName":               &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"shortTitle":             &graphql.InputObjectFieldConfig{Type: graphql.String},
		"longTitle":              &graphql.InputObjectFieldConfig{Type: graphql.String},
		"organizationName":       &graphql.InputObjectFieldConfig{Type: graphql.String},
		"phoneVerificationToken": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
	},
})

var createProviderAccountOutputType = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateProviderAccountPayload",
	Fields: graphql.Fields{
		"clientMutationId":    newClientmutationIDOutputField(),
		"success":             &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"errorCode":           &graphql.Field{Type: createAccountErrorCodeEnum},
		"errorMessage":        &graphql.Field{Type: graphql.String},
		"token":               &graphql.Field{Type: graphql.String},
		"account":             &graphql.Field{Type: accountInterfaceType},
		"clientEncryptionKey": &graphql.Field{Type: graphql.String},
	},
	IsTypeOf: func(value interface{}, info graphql.ResolveInfo) bool {
		_, ok := value.(*createProviderAccountOutput)
		return ok
	},
})

var createProviderAccountMutation = &graphql.Field{
	Type: graphql.NewNonNull(createProviderAccountOutputType),
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createAccountInputType)},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return createProviderAccount(p)
	},
}

func createProviderAccount(p graphql.ResolveParams) (*createProviderAccountOutput, error) {
	svc := serviceFromParams(p)
	ram := raccess.ResourceAccess(p)
	ctx := p.Context
	input := p.Args["input"].(map[string]interface{})

	mutationID, _ := input["clientMutationId"].(string)

	inv, err := svc.inviteInfo(ctx)
	if err != nil {
		return nil, errors.InternalError(ctx, err)
	}
	// Sanity check to make sure we fail early in case we forgot to handle all new invite types
	if inv != nil && inv.Type != invite.LookupInviteResponse_COLLEAGUE {
		return nil, errors.InternalError(ctx, fmt.Errorf("unknown invite type %s", inv.Type.String()))
	}

	req := &auth.CreateAccountRequest{
		Email:    input["email"].(string),
		Password: input["password"].(string),
	}
	req.Email = strings.TrimSpace(req.Email)
	if !validate.Email(req.Email) {
		return &createProviderAccountOutput{
			ClientMutationID: mutationID,
			Success:          false,
			ErrorCode:        createAccountErrorCodeInvalidEmail,
			ErrorMessage:     "Please enter a valid email address.",
		}, nil
	}
	entityInfo, err := entityInfoFromInput(input)
	if err != nil {
		return nil, errors.InternalError(ctx, err)
	}

	req.FirstName = strings.TrimSpace(entityInfo.FirstName)
	req.LastName = strings.TrimSpace(entityInfo.LastName)
	if req.FirstName == "" || !isValidPlane0Unicode(req.FirstName) {
		return &createProviderAccountOutput{
			ClientMutationID: mutationID,
			Success:          false,
			ErrorCode:        createAccountErrorCodeInvalidFirstName,
			ErrorMessage:     "Please enter a valid first name.",
		}, nil
	}
	if req.LastName == "" || !isValidPlane0Unicode(req.LastName) {
		return &createProviderAccountOutput{
			ClientMutationID: mutationID,
			Success:          false,
			ErrorCode:        createAccountErrorCodeInvalidLastName,
			ErrorMessage:     "Please enter a valid last name.",
		}, nil
	}

	var organizationName string
	if inv == nil {
		organizationName, _ = input["organizationName"].(string)
		organizationName = strings.TrimSpace(organizationName)
		if organizationName == "" || !isValidPlane0Unicode(organizationName) {
			return &createProviderAccountOutput{
				ClientMutationID: mutationID,
				Success:          false,
				ErrorCode:        createAccountErrorCodeInvalidOrganizationName,
				ErrorMessage:     "Please enter a valid organization name.",
			}, nil
		}
	}
	verifiedValue, err := ram.VerifiedValue(ctx, input["phoneVerificationToken"].(string))
	if grpc.Code(err) == auth.ValueNotYetVerified {
		return nil, errors.New("The phone number for the provided token has not yet been verified.")
	} else if err != nil {
		return nil, err
	}
	vpn, err := phone.ParseNumber(verifiedValue)
	if err != nil {
		return nil, errors.InternalError(ctx, err)
	}
	pn, err := phone.ParseNumber(input["phoneNumber"].(string))
	if err != nil {
		return &createProviderAccountOutput{
			ClientMutationID: mutationID,
			Success:          false,
			ErrorCode:        createAccountErrorCodeInvalidPhoneNumber,
			ErrorMessage:     "Please enter a valid phone number.",
		}, nil
	}
	req.PhoneNumber = pn.String()
	if vpn.String() != pn.String() {
		golog.Debugf("The provided phone number %q does not match the number validated by the provided token %s", pn.String(), vpn.String())
		return nil, fmt.Errorf("The provided phone number %q does not match the number validated by the provided token", req.PhoneNumber)
	}
	res, err := ram.CreateAccount(ctx, req)
	if err != nil {
		switch grpc.Code(err) {
		case auth.DuplicateEmail:
			return &createProviderAccountOutput{
				ClientMutationID: mutationID,
				Success:          false,
				ErrorCode:        createAccountErrorCodeAccountExists,
				ErrorMessage:     "An account already exists with the entered email address.",
			}, nil
		case auth.InvalidEmail:
			return &createProviderAccountOutput{
				ClientMutationID: mutationID,
				Success:          false,
				ErrorCode:        createAccountErrorCodeInvalidEmail,
				ErrorMessage:     "Please enter a valid email address.",
			}, nil
		case auth.InvalidPhoneNumber:
			return &createProviderAccountOutput{
				ClientMutationID: mutationID,
				Success:          false,
				ErrorCode:        createAccountErrorCodeInvalidPhoneNumber,
				ErrorMessage:     "Please enter a valid phone number.",
			}, nil
		}
		return nil, errors.InternalError(ctx, err)
	}
	// TODO: updating the gqlctx this is safe for now because the GraphQL pkg serializes mutations.
	// that likely won't change, but this still isn't a great way to update the gqlctx.
	gqlctx.InPlaceWithAccount(ctx, res.Account)

	var orgEntityID string
	var accEntityID string
	{
		if inv == nil {
			// Create organization
			ent, err := ram.CreateEntity(ctx, &directory.CreateEntityRequest{
				EntityInfo: &directory.EntityInfo{
					GroupName:   organizationName,
					DisplayName: organizationName,
				},
				Type: directory.EntityType_ORGANIZATION,
			})
			if err != nil {
				return nil, err
			}
			orgEntityID = ent.ID
		} else {
			orgEntityID = inv.GetColleague().OrganizationEntityID
		}

		contacts := []*directory.Contact{
			{
				ContactType: directory.ContactType_PHONE,
				Value:       req.PhoneNumber,
				Provisioned: false,
			},
		}
		entityInfo.DisplayName, err = buildDisplayName(entityInfo, contacts)
		if err != nil {
			return nil, errors.InternalError(ctx, err)
		}
		// Create entity
		ent, err := ram.CreateEntity(ctx, &directory.CreateEntityRequest{
			EntityInfo:                entityInfo,
			Type:                      directory.EntityType_INTERNAL,
			ExternalID:                res.Account.ID,
			InitialMembershipEntityID: orgEntityID,
			Contacts:                  contacts,
		})
		if err != nil {
			return nil, err
		}
		accEntityID = ent.ID
	}

	// Create a default saved query
	if err = ram.CreateSavedQuery(ctx, &threading.CreateSavedQueryRequest{
		OrganizationID: orgEntityID,
		EntityID:       accEntityID,
		// TODO: query
	}); err != nil {
		return nil, errors.InternalError(ctx, err)
	}

	if inv == nil {
		// Create initial threads, but don't fail entirely on errors as this isn't critical to the account existing,
		// and because a hard fail leaves the account around but makes it look like it failed it's best just to
		// log and continue. Once the account creation is idempotent then can have this be a hard fail.

		// These are created synchronously to enforce strict ordering

		// Create a support thread (linked to Spruce support org) and the primary entities for them
		var tsEnt1, tsEnt2 *directory.Entity
		par := conc.NewParallel()
		par.Go(func() error {
			var err error
			tsEnt1, err = ram.CreateEntity(ctx, &directory.CreateEntityRequest{
				EntityInfo: &directory.EntityInfo{
					GroupName:   supportThreadTitle,
					DisplayName: supportThreadTitle,
				},
				Type: directory.EntityType_SYSTEM,
				InitialMembershipEntityID: orgEntityID,
			})
			return err
		})
		remoteSupportThreadTitle := fmt.Sprintf(supportThreadTitle+" (%s)", organizationName)
		par.Go(func() error {
			var err error
			tsEnt2, err = ram.CreateEntity(ctx, &directory.CreateEntityRequest{
				EntityInfo: &directory.EntityInfo{
					GroupName:   remoteSupportThreadTitle,
					DisplayName: remoteSupportThreadTitle,
				},
				Type: directory.EntityType_SYSTEM,
				InitialMembershipEntityID: svc.spruceOrgID,
			})
			return err
		})
		if err := par.Wait(); err != nil {
			golog.Errorf("Failed to create entity for support thread for org %s: %s", orgEntityID, err)
		} else {
			_, err = ram.CreateLinkedThreads(ctx, &threading.CreateLinkedThreadsRequest{
				Organization1ID:      orgEntityID,
				Organization2ID:      svc.spruceOrgID,
				PrimaryEntity1ID:     tsEnt1.ID,
				PrimaryEntity2ID:     tsEnt2.ID,
				PrependSenderThread1: false,
				PrependSenderThread2: true,
				Summary:              supportThreadTitle + ": " + teamSpruceInitialText[:128],
				Text:                 teamSpruceInitialText,
				Type:                 threading.ThreadType_SUPPORT,
				SystemTitle1:         supportThreadTitle,
				SystemTitle2:         remoteSupportThreadTitle,
			})
			if err != nil {
				golog.Errorf("Failed to create linked support threads for org %s: %s", orgEntityID, err)
			}
		}

		// Create an onboarding thread and related system entity
		onbEnt, err := ram.CreateEntity(ctx, &directory.CreateEntityRequest{
			EntityInfo: &directory.EntityInfo{
				GroupName:   onboardingThreadTitle,
				DisplayName: onboardingThreadTitle,
			},
			Type: directory.EntityType_SYSTEM,
			InitialMembershipEntityID: orgEntityID,
		})
		if err != nil {
			golog.Errorf("Failed to create entity for onboarding thread for org %s: %s", orgEntityID, err)
		} else {
			_, err = ram.CreateOnboardingThread(ctx, &threading.CreateOnboardingThreadRequest{
				OrganizationID:  orgEntityID,
				PrimaryEntityID: onbEnt.ID,
			})
			if err != nil {
				golog.Errorf("Failed to create onboarding thread for org %s: %s", orgEntityID, err)
			}
		}
	}

	// Record analytics
	headers := gqlctx.SpruceHeaders(ctx)
	var platform string
	if headers != nil {
		platform = headers.Platform.String()
		golog.Infof("Account created. ID = %s Device = %s", res.Account.ID, headers.DeviceID)
	}
	orgName := organizationName
	if inv != nil {
		oe, err := ram.Entity(ctx, orgEntityID, nil, 0)
		if err != nil {
			golog.Errorf("Failed to lookup organization %s: %s", orgEntityID, err)
		} else {
			orgName = oe.Info.DisplayName
		}
	}
	conc.Go(func() {
		svc.segmentio.Identify(&analytics.Identify{
			UserId: res.Account.ID,
			Traits: map[string]interface{}{
				"name":              res.Account.FirstName + " " + res.Account.LastName,
				"first_name":        res.Account.FirstName,
				"last_name":         res.Account.LastName,
				"email":             req.Email,
				"title":             entityInfo.ShortTitle,
				"organization_name": orgName,
				"platform":          platform,
				"createdAt":         time.Now().Unix(),
			},
			Context: map[string]interface{}{
				"ip":        remoteAddrFromParams(p),
				"userAgent": userAgentFromParams(p),
			},
		})
		svc.segmentio.Group(&analytics.Group{
			UserId:  res.Account.ID,
			GroupId: orgEntityID,
			Traits: map[string]interface{}{
				"name": orgName,
			},
		})
		props := map[string]interface{}{
			"entity_id":       accEntityID,
			"organization_id": orgEntityID,
		}
		if inv != nil {
			props["invite"] = inv.Type.String()
		}
		svc.segmentio.Track(&analytics.Track{
			Event:      "signedup",
			UserId:     res.Account.ID,
			Properties: props,
		})
	})

	result := p.Info.RootValue.(map[string]interface{})["result"].(conc.Map)
	result.Set("auth_token", res.Token.Value)
	result.Set("auth_expiration", time.Unix(int64(res.Token.ExpirationEpoch), 0))

	return &createProviderAccountOutput{
		ClientMutationID:    mutationID,
		Success:             true,
		Token:               res.Token.Value,
		Account:             transformAccountToResponse(res.Account),
		ClientEncryptionKey: res.Token.ClientEncryptionKey,
	}, nil
}