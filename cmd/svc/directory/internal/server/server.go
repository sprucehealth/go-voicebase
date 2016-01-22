package server

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/cmd/svc/directory/internal/dal"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/validate"
	"github.com/sprucehealth/backend/svc/directory"
	"golang.org/x/net/context"
)

var grpcErrorf = grpc.Errorf

// DAL represents the methods required to provide data access layer functionality
type DAL interface {
	InsertEntity(model *dal.Entity) (dal.EntityID, error)
	Entity(id dal.EntityID) (*dal.Entity, error)
	Entities(ids []dal.EntityID) ([]*dal.Entity, error)
	UpdateEntity(id dal.EntityID, update *dal.EntityUpdate) (int64, error)
	DeleteEntity(id dal.EntityID) (int64, error)
	InsertExternalEntityID(model *dal.ExternalEntityID) error
	ExternalEntityIDs(externalID string) ([]*dal.ExternalEntityID, error)
	ExternalEntityIDsForEntities(entityID []dal.EntityID) ([]*dal.ExternalEntityID, error)
	InsertEntityMembership(model *dal.EntityMembership) error
	EntityMemberships(id dal.EntityID) ([]*dal.EntityMembership, error)
	EntityMembers(id dal.EntityID) ([]*dal.Entity, error)
	InsertEntityContact(model *dal.EntityContact) (dal.EntityContactID, error)
	EntityContacts(id dal.EntityID) ([]*dal.EntityContact, error)
	EntityContact(id dal.EntityContactID) (*dal.EntityContact, error)
	EntityContactsForValue(value string) ([]*dal.EntityContact, error)
	UpdateEntityContact(id dal.EntityContactID, update *dal.EntityContactUpdate) (int64, error)
	DeleteEntityContact(id dal.EntityContactID) (int64, error)
	InsertEvent(model *dal.Event) (dal.EventID, error)
	Event(id dal.EventID) (*dal.Event, error)
	UpdateEvent(id dal.EventID, update *dal.EventUpdate) (int64, error)
	DeleteEvent(id dal.EventID) (int64, error)
	Transact(trans func(dal dal.DAL) error) (err error)
}

// Server describes the methods exposed by the server
type Server interface {
	CreateContact(context.Context, *directory.CreateContactRequest) (*directory.CreateContactResponse, error)
	CreateEntity(context.Context, *directory.CreateEntityRequest) (*directory.CreateEntityResponse, error)
	CreateMembership(context.Context, *directory.CreateMembershipRequest) (*directory.CreateMembershipResponse, error)
	ExternalIDs(context.Context, *directory.ExternalIDsRequest) (*directory.ExternalIDsResponse, error)
	LookupEntities(context.Context, *directory.LookupEntitiesRequest) (*directory.LookupEntitiesResponse, error)
	LookupEntitiesByContact(context.Context, *directory.LookupEntitiesByContactRequest) (*directory.LookupEntitiesByContactResponse, error)
}

var (
	// ErrNotImplemented is returned from RPC calls that have yet to be implemented
	ErrNotImplemented = errors.New("Not Implemented")
)

type server struct {
	dl DAL
}

// New returns an initialized instance of server
func New(dl DAL) Server {
	return &server{dl: dl}
}

// This is a helper to avoid duplicating this nil check everywhere
func riEntityInformation(ri *directory.RequestedInformation) []directory.EntityInformation {
	if ri == nil {
		return nil
	}
	return ri.EntityInformation
}

// This is a helper to avoid duplicating this nil check everywhere
func riDepth(ri *directory.RequestedInformation) int64 {
	if ri == nil {
		return 0
	}
	return ri.Depth
}

func (s *server) LookupEntities(ctx context.Context, rd *directory.LookupEntitiesRequest) (*directory.LookupEntitiesResponse, error) {
	golog.Debugf("Entering server.server.LookupEntities: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.LookupEntities...") }()
	}
	var entityIDs []dal.EntityID
	switch rd.LookupKeyType {
	case directory.LookupEntitiesRequest_EXTERNAL_ID:
		externalEntityIDs, err := s.dl.ExternalEntityIDs(rd.GetExternalID())
		if err != nil {
			return nil, grpcErrorf(codes.Internal, err.Error())
		}
		entityIDs = make([]dal.EntityID, len(externalEntityIDs))
		for i, v := range externalEntityIDs {
			entityIDs[i] = v.EntityID
		}
	case directory.LookupEntitiesRequest_ENTITY_ID:
		entityID, err := dal.ParseEntityID(rd.GetEntityID())
		if err != nil {
			return nil, grpcErrorf(codes.InvalidArgument, "Unable to parse entity id")
		}
		entityIDs = append(entityIDs, entityID)
	default:
		return nil, grpcErrorf(codes.Internal, "Unknown lookup key type %d", rd.LookupKeyType)
	}
	entities, err := s.dl.Entities(entityIDs)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	if len(entities) == 0 {
		return nil, grpcErrorf(codes.NotFound, "No entities located matching query")
	}
	pbEntities, err := getPBEntities(s.dl, entities, riEntityInformation(rd.RequestedInformation), riDepth(rd.RequestedInformation))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &directory.LookupEntitiesResponse{
		Entities: pbEntities,
	}, nil
}

func (s *server) CreateEntity(ctx context.Context, rd *directory.CreateEntityRequest) (*directory.CreateEntityResponse, error) {
	golog.Debugf("Entering server.server.CreateEntity: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.CreateEntity...") }()
	}
	if err := s.validateCreateEntityRequest(rd); err != nil {
		return nil, err
	}

	entityType, err := dal.ParseEntityType(directory.EntityType_name[int32(rd.Type)])
	if err != nil {
		return nil, errors.Trace(err)
	}
	var pbEntity *directory.Entity
	if err = s.dl.Transact(func(dl dal.DAL) error {
		entityID, err := dl.InsertEntity(&dal.Entity{
			Name:   rd.Name,
			Type:   entityType,
			Status: dal.EntityStatusActive,
		})
		if err != nil {
			return errors.Trace(err)
		}
		if rd.ExternalID != "" {
			if err := dl.InsertExternalEntityID(&dal.ExternalEntityID{
				EntityID:   entityID,
				ExternalID: rd.ExternalID,
			}); err != nil {
				return errors.Trace(err)
			}
		}
		if rd.InitialMembershipEntityID != "" {
			targetEntityID, err := dal.ParseEntityID(rd.InitialMembershipEntityID)
			if err != nil {
				return errors.Trace(err)
			}
			if err := dl.InsertEntityMembership(&dal.EntityMembership{
				EntityID:       entityID,
				TargetEntityID: targetEntityID,
				Status:         dal.EntityMembershipStatusActive,
			}); err != nil {
				return errors.Trace(err)
			}
		}
		for _, contact := range rd.Contacts {
			contactType, err := dal.ParseEntityContactType(directory.ContactType_name[int32(contact.ContactType)])
			if err != nil {
				return errors.Trace(err)
			}
			if _, err := dl.InsertEntityContact(&dal.EntityContact{
				EntityID:    entityID,
				Type:        contactType,
				Value:       contact.Value,
				Provisioned: contact.Provisioned,
			}); err != nil {
				return errors.Trace(err)
			}
		}
		entity, err := dl.Entity(entityID)
		if err != nil {
			return errors.Trace(err)
		}

		pbEntity, err = getPBEntity(dl, entity, riEntityInformation(rd.RequestedInformation), riDepth(rd.RequestedInformation))
		return errors.Trace(err)
	}); err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	return &directory.CreateEntityResponse{
		Entity: pbEntity,
	}, nil
}

func (s *server) ExternalIDs(ctx context.Context, rd *directory.ExternalIDsRequest) (*directory.ExternalIDsResponse, error) {
	golog.Debugf("Entering server.server.ExternalIDs: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.ExternalIDs...") }()
	}
	ids := make([]dal.EntityID, len(rd.EntityIDs))
	for i, id := range rd.EntityIDs {
		eID, err := dal.ParseEntityID(id)
		if err != nil {
			return nil, grpcErrorf(codes.InvalidArgument, "Unable to parse entity id")
		}
		ids[i] = eID
	}
	externalIDs, err := s.dl.ExternalEntityIDsForEntities(ids)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	return &directory.ExternalIDsResponse{
		ExternalIDs: transformExternalIDs(externalIDs),
	}, nil
}

func (s *server) validateCreateEntityRequest(rd *directory.CreateEntityRequest) error {
	golog.Debugf("Entering server.server.validateCreateEntityRequest: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.validateCreateEntityRequest...") }()
	}
	if rd.Type != directory.EntityType_EXTERNAL && rd.Name == "" {
		return grpcErrorf(codes.InvalidArgument, "Name cannot be empty for non external entities")
	}
	if rd.InitialMembershipEntityID != "" {
		eID, err := dal.ParseEntityID(rd.InitialMembershipEntityID)
		if err != nil {
			return grpcErrorf(codes.InvalidArgument, "Unable to parse entity id")
		}
		exists, err := doesEntityExist(s.dl, eID)
		if err != nil {
			return grpcErrorf(codes.Internal, err.Error())
		}
		if !exists {
			return grpcErrorf(codes.NotFound, "Entity not found %s", rd.InitialMembershipEntityID)
		}
	}
	for _, contact := range rd.Contacts {
		if contact.Value == "" {
			return grpcErrorf(codes.InvalidArgument, "Contact value cannot be empty")
		}
		if err := validateContact(contact); err != nil {
			return grpcErrorf(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}

func doesEntityExist(dl dal.DAL, entityID dal.EntityID) (bool, error) {
	golog.Debugf("Entering server.doesEntityExist: %s", entityID)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.doesEntityExist...") }()
	}
	_, err := dl.Entity(entityID)
	if api.IsErrNotFound(err) {
		return false, nil
	}
	return true, errors.Trace(err)
}

func (s *server) CreateMembership(ctx context.Context, rd *directory.CreateMembershipRequest) (*directory.CreateMembershipResponse, error) {
	golog.Debugf("Entering server.server.CreateMembership: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.CreateMembership...") }()
	}
	entityID, err := dal.ParseEntityID(rd.EntityID)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, "Unable to parse entity id")
	}
	targetEntityID, err := dal.ParseEntityID(rd.TargetEntityID)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, "Unable to parse entity id")
	}
	exists, err := doesEntityExist(s.dl, entityID)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	if !exists {
		return nil, grpcErrorf(codes.NotFound, "Entity not found %s", rd.EntityID)
	}
	exists, err = doesEntityExist(s.dl, targetEntityID)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	if !exists {
		return nil, grpcErrorf(codes.NotFound, "Entity not found %s", rd.TargetEntityID)
	}

	if err := s.dl.InsertEntityMembership(&dal.EntityMembership{
		EntityID:       entityID,
		TargetEntityID: targetEntityID,
		Status:         dal.EntityMembershipStatusActive,
	}); err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	entity, err := s.dl.Entity(entityID)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	pbEntity, err := getPBEntity(s.dl, entity, riEntityInformation(rd.RequestedInformation), riDepth(rd.RequestedInformation))
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	return &directory.CreateMembershipResponse{
		Entity: pbEntity,
	}, nil
}

func (s *server) LookupEntitiesByContact(ctx context.Context, rd *directory.LookupEntitiesByContactRequest) (*directory.LookupEntitiesByContactResponse, error) {
	golog.Debugf("Entering server.server.LookupEntitiesByContact: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.LookupEntitiesByContact...") }()
	}
	entityContacts, err := s.dl.EntityContactsForValue(strings.TrimSpace(rd.ContactValue))
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	if len(entityContacts) == 0 {
		return nil, grpcErrorf(codes.NotFound, "Contact with value %s not found", rd.ContactValue)
	}
	uniqueEntityIDs := make(map[uint64]struct{})
	var entityIDs []dal.EntityID
	for _, ec := range entityContacts {
		if _, ok := uniqueEntityIDs[ec.EntityID.ObjectID.Val]; !ok {
			entityIDs = append(entityIDs, ec.EntityID)
		}
	}
	entities, err := s.dl.Entities(entityIDs)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}

	pbEntities, err := getPBEntities(s.dl, entities, riEntityInformation(rd.RequestedInformation), riDepth(rd.RequestedInformation))
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	return &directory.LookupEntitiesByContactResponse{
		Entities: pbEntities,
	}, nil
}

func (s *server) CreateContact(ctx context.Context, rd *directory.CreateContactRequest) (*directory.CreateContactResponse, error) {
	golog.Debugf("Entering server.server.CreateContact: %+v", rd)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.server.CreateContact...") }()
	}
	entityID, err := dal.ParseEntityID(rd.EntityID)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, "Unable to parse entity id")
	}
	if exists, err := doesEntityExist(s.dl, entityID); err != nil {
		return nil, errors.Trace(err)
	} else if !exists {
		return nil, grpcErrorf(codes.NotFound, "Entity %s not found", rd.EntityID)
	}
	if err := validateContact(rd.GetContact()); err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, err.Error())
	}

	contactType, err := dal.ParseEntityContactType(directory.ContactType_name[int32(rd.GetContact().ContactType)])
	if err != nil {
		return nil, errors.Trace(err)
	}
	var pbEntity *directory.Entity
	if err := s.dl.Transact(func(dl dal.DAL) error {
		if _, err := dl.InsertEntityContact(&dal.EntityContact{
			EntityID:    entityID,
			Type:        contactType,
			Value:       rd.GetContact().Value,
			Provisioned: rd.GetContact().Provisioned,
		}); err != nil {
			return errors.Trace(err)
		}
		entity, err := dl.Entity(entityID)
		if err != nil {
			return errors.Trace(err)
		}
		pbEntity, err = getPBEntity(dl, entity, riEntityInformation(rd.RequestedInformation), riDepth(rd.RequestedInformation))
		return errors.Trace(err)
	}); err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	return &directory.CreateContactResponse{
		Entity: pbEntity,
	}, nil
}

func validateContact(contact *directory.Contact) error {
	switch contact.ContactType {
	case directory.ContactType_EMAIL:
		if !validate.Email(contact.Value) {
			return fmt.Errorf("Invalid email: %s", contact.Value)
		}
	case directory.ContactType_PHONE:
		/*
			if _, err := common.ParsePhone(contact.Value); err != nil {
				return fmt.Errorf("Invalid phone number: %s", contact.Value)
			}
		*/
	}
	return nil
}

func transformExternalIDs(dExternalEntityIDs []*dal.ExternalEntityID) []*directory.ExternalID {
	pExternalID := make([]*directory.ExternalID, len(dExternalEntityIDs))
	for i, eID := range dExternalEntityIDs {
		pExternalID[i] = &directory.ExternalID{
			ID:       eID.ExternalID,
			EntityID: eID.EntityID.String(),
		}
	}
	return pExternalID
}

func getPBEntities(dl dal.DAL, dEntities []*dal.Entity, entityInformation []directory.EntityInformation, depth int64) ([]*directory.Entity, error) {
	golog.Debugf("Entering server.getPBEntities: dEntities: %+v, entityInformation, %+v, depth: %d", dEntities, entityInformation, depth)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.getPBEntities...") }()
	}
	pbEntities := make([]*directory.Entity, len(dEntities))
	for i, e := range dEntities {
		pbEntity, err := getPBEntity(dl, e, entityInformation, depth)
		if err != nil {
			return nil, errors.Trace(err)
		}
		pbEntities[i] = pbEntity
	}
	return pbEntities, nil
}

// Note: How we optimize this deep crawl is very likely to change
func getPBEntity(dl dal.DAL, dEntity *dal.Entity, entityInformation []directory.EntityInformation, depth int64) (*directory.Entity, error) {
	golog.Debugf("Entering server.getPBEntity: dEntity: %+v, entityInformation, %+v, depth: %d", dEntity, entityInformation, depth)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.getPBEntity...") }()
	}
	entity := dalEntityAsPBEntity(dEntity)
	if depth >= 0 {
		if hasRequestedInfo(entityInformation, directory.EntityInformation_MEMBERSHIPS) {
			memberships, err := getPBMemberships(dl, dEntity.ID, entityInformation, depth-1)
			if err != nil {
				return nil, errors.Trace(err)
			}
			entity.Memberships = memberships
		}
		if hasRequestedInfo(entityInformation, directory.EntityInformation_MEMBERS) {
			members, err := getPBMembers(dl, dEntity.ID, entityInformation, depth-1)
			if err != nil {
				return nil, errors.Trace(err)
			}
			entity.Members = members
		}
		if hasRequestedInfo(entityInformation, directory.EntityInformation_EXTERNAL_IDS) {
			externalIDs, err := getPBExternalIDs(dl, dEntity.ID)
			if err != nil {
				return nil, errors.Trace(err)
			}
			entity.ExternalIDs = externalIDs
		}
		if hasRequestedInfo(entityInformation, directory.EntityInformation_CONTACTS) {
			contacts, err := getPBEntityContacts(dl, dEntity.ID)
			if err != nil {
				return nil, errors.Trace(err)
			}
			entity.Contacts = contacts
		}
	}
	entity.IncludedInformation = entityInformation
	return entity, nil
}

func getPBMemberships(dl dal.DAL, entityID dal.EntityID, entityInformation []directory.EntityInformation, depth int64) ([]*directory.Entity, error) {
	golog.Debugf("Entering server.getPBMemberships - EntityID: %s, entityInformation: %+v, depth: %d", entityID, entityInformation, depth)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.getPBMemberships...") }()
	}
	memberships, err := dl.EntityMemberships(entityID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	entityIDs := make([]dal.EntityID, len(memberships))
	for i, membership := range memberships {
		entityIDs[i] = membership.TargetEntityID
	}
	entities, err := dl.Entities(entityIDs)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pbEntities, err := getPBEntities(dl, entities, entityInformation, depth)
	return pbEntities, errors.Trace(err)
}

func getPBMembers(dl dal.DAL, entityID dal.EntityID, entityInformation []directory.EntityInformation, depth int64) ([]*directory.Entity, error) {
	golog.Debugf("Entering server.getPBMembers - EntityID: %s, RequestedInformation: %v, depth: %d", entityID, entityInformation, depth)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.getPBMembers...") }()
	}
	members, err := dl.EntityMembers(entityID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pbEntities, err := getPBEntities(dl, members, entityInformation, depth)
	return pbEntities, errors.Trace(err)
}

func getPBExternalIDs(dl dal.DAL, entityID dal.EntityID) ([]string, error) {
	golog.Debugf("Entering server.getPBExternalIDs - EntityID: %s", entityID)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.getPBExternalIDs...") }()
	}
	externalIDs, err := dl.ExternalEntityIDsForEntities([]dal.EntityID{entityID})
	if err != nil {
		return nil, errors.Trace(err)
	}
	externalIDStrings := make([]string, len(externalIDs))
	for i, eID := range externalIDs {
		externalIDStrings[i] = eID.ExternalID
	}
	return externalIDStrings, nil
}

func getPBEntityContacts(dl dal.DAL, entityID dal.EntityID) ([]*directory.Contact, error) {
	golog.Debugf("Entering server.getPBEntityContacts - EntityID: %s", entityID)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.getPBEntityContacts...") }()
	}
	entityContacts, err := dl.EntityContacts(entityID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pbEntityContacts := make([]*directory.Contact, len(entityContacts))
	for i, ec := range entityContacts {
		pbEntityContacts[i] = dalEntityContactAsPBContact(ec)
	}
	return pbEntityContacts, nil
}

func dalEntityContactAsPBContact(dEntityContact *dal.EntityContact) *directory.Contact {
	golog.Debugf("Entering server.dalEntityContactAsPBContact: %+v...", dEntityContact)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.dalEntityContactAsPBContact...") }()
	}
	contact := &directory.Contact{}
	contact.Provisioned = dEntityContact.Provisioned
	contactType, ok := directory.ContactType_value[dEntityContact.Type.String()]
	if !ok {
		golog.Errorf("Unknown contact type %s when converting to PB format", dEntityContact.Type)
	}
	contact.ContactType = directory.ContactType(contactType)
	contact.Value = dEntityContact.Value
	return contact
}

func dalEntityAsPBEntity(dEntity *dal.Entity) *directory.Entity {
	golog.Debugf("Entering server.dalEntityAsPBEntity: %+v...", dEntity)
	if golog.Default().L(golog.DEBUG) {
		defer func() { golog.Debugf("Leaving server.dalEntityAsPBEntity...") }()
	}
	entity := &directory.Entity{}
	entity.ID = dEntity.ID.String()
	entity.Name = dEntity.Name
	entityType, ok := directory.EntityType_value[dEntity.Type.String()]
	if !ok {
		golog.Errorf("Unknown entity type %s when converting to PB format", dEntity.Type)
	}
	entity.Type = directory.EntityType(entityType)
	return entity
}

// Note: Assuming this is small it's not a big deal but might want to be more efficient if it grows
func hasRequestedInfo(requestedInformation []directory.EntityInformation, target directory.EntityInformation) bool {
	for _, v := range requestedInformation {
		if v == target {
			return true
		}
	}
	return false
}