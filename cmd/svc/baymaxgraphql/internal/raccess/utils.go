package raccess

import (
	"context"

	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/svc/directory"
)

// EntityInOrgForAccountID returns the entity in the org specified from the account in the context.
func EntityInOrgForAccountID(ctx context.Context, ram ResourceAccessor, req *directory.LookupEntitiesRequest, orgID string) (*directory.Entity, error) {
	// assert that the lookup entities request is for looking up an entity
	// via externalID
	var accountID string
	switch key := req.LookupKeyOneof.(type) {
	case *directory.LookupEntitiesRequest_AccountID:
		accountID = key.AccountID
	case *directory.LookupEntitiesRequest_ExternalID:
		accountID = key.ExternalID
	default:
		return nil, errors.Errorf("Expected lookup key of type AccountID or ExternalID but got %T", req)
	}

	entities, err := ram.Entities(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		for _, member := range entity.Memberships {
			if member.Type == directory.EntityType_ORGANIZATION && member.ID == orgID {
				return entity, nil
			}
		}
	}

	return nil, errors.Errorf("Did not find entity for account %s and org %s", accountID, orgID)
}

// Entity returns a single expected entity for the directory request.
func Entity(ctx context.Context, ram ResourceAccessor, req *directory.LookupEntitiesRequest) (*directory.Entity, error) {
	return entity(ctx, ram, req)
}

// UnauthorizedEntity returns a single expected entity for the directory request.
func UnauthorizedEntity(ctx context.Context, ram ResourceAccessor, req *directory.LookupEntitiesRequest) (*directory.Entity, error) {
	return entity(ctx, ram, req, EntityQueryOptionUnathorized)
}

// Entity returns a single expected entity for the directory request.
func entity(ctx context.Context, ram ResourceAccessor, req *directory.LookupEntitiesRequest, opts ...EntityQueryOption) (*directory.Entity, error) {
	if req.LookupKeyType != directory.LookupEntitiesRequest_ENTITY_ID && req.LookupKeyType != directory.LookupEntitiesRequest_EXTERNAL_ID {
		return nil, errors.Errorf("Expected lookup of type %s but got %s", directory.LookupEntitiesRequest_ENTITY_ID, req.LookupKeyType)
	}

	entities, err := ram.Entities(ctx, req, opts...)
	if err != nil {
		return nil, err
	} else if len(entities) == 0 {
		return nil, ErrNotFound
	} else if len(entities) != 1 {
		id := req.GetEntityID()
		if id == "" {
			id = req.GetExternalID()
		}
		return nil, errors.Errorf("Expected 1 entity got %d for %s", len(entities), id)
	}

	return entities[0], nil
}
