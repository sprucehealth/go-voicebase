package main

import (
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/errors"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/models"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/raccess"
	"github.com/sprucehealth/backend/device/devicectx"
	lerrors "github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/svc/directory"
	"github.com/sprucehealth/backend/svc/media"
	"github.com/sprucehealth/graphql"
	"golang.org/x/net/context"
)

var profileSectionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ProfileSection",
	Fields: graphql.Fields{
		"title": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"body":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var profileType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Profile",
	Interfaces: []*graphql.Interface{
		nodeInterfaceType,
	},
	Fields: graphql.Fields{
		"id":       &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"entityID": &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"title":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"sections": &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(profileSectionType))},
		"image": &graphql.Field{
			Type: imageType,
			Args: NewImageArguments(nil),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				svc := serviceFromParams(p)
				ram := raccess.ResourceAccess(p)
				imgArgs := ParseImageArguments(p.Args)
				ctx := p.Context
				profile := p.Source.(*models.Profile)
				ent, err := raccess.Entity(ctx, ram, &directory.LookupEntitiesRequest{
					LookupKeyType: directory.LookupEntitiesRequest_ENTITY_ID,
					LookupKeyOneof: &directory.LookupEntitiesRequest_EntityID{
						EntityID: profile.EntityID,
					},
				})
				if lerrors.Cause(err) == raccess.ErrNotFound {
					return nil, errors.ErrNotFound(ctx, profile.EntityID)
				} else if err != nil {
					return nil, errors.InternalError(ctx, err)
				}
				// Fallback to the avatar
				if ent.ImageMediaID == "" {
					rEnt, err := transformEntityToResponse(svc.staticURLPrefix, ent, devicectx.SpruceHeaders(ctx), gqlctx.Account(ctx))
					if err != nil {
						golog.Errorf("Error while transforming entity to response for profile image creation: %s", err)
						return nil, nil
					}
					return rEnt.Avatar, nil
				}
				return &models.Image{
					URL:    media.ThumbnailURL(svc.mediaAPIDomain, ent.ImageMediaID, imgArgs.Height, imgArgs.Width, imgArgs.Crop),
					Width:  imgArgs.Width,
					Height: imgArgs.Height,
				}, nil
			},
		},
		"allowEdit":             &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"lastModifiedTimestamp": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
	},
})

func lookupEntityProfile(ctx context.Context, ram raccess.ResourceAccessor, entityID string) (interface{}, error) {
	profile, err := ram.Profile(ctx, &directory.ProfileRequest{
		LookupKeyType: directory.ProfileRequest_ENTITY_ID,
		LookupKeyOneof: &directory.ProfileRequest_EntityID{
			EntityID: entityID,
		},
	})
	if lerrors.Cause(err) == raccess.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return transformProfileToResponse(ctx, ram, profile), nil
}

func lookupProfile(ctx context.Context, ram raccess.ResourceAccessor, entityProfileID string) (interface{}, error) {
	profile, err := ram.Profile(ctx, &directory.ProfileRequest{
		LookupKeyType: directory.ProfileRequest_PROFILE_ID,
		LookupKeyOneof: &directory.ProfileRequest_ProfileID{
			ProfileID: entityProfileID,
		},
	})
	if lerrors.Cause(err) == raccess.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return transformProfileToResponse(ctx, ram, profile), nil
}
