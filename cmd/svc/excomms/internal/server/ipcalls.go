package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/sprucehealth/backend/cmd/svc/excomms/internal/dal"
	"github.com/sprucehealth/backend/cmd/svc/excomms/internal/models"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/twilio"
	"github.com/sprucehealth/backend/svc/directory"
	"github.com/sprucehealth/backend/svc/excomms"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

const ipCallTokenTTL = 6 * 60 * 60

type ipCallStateTransition struct {
	from, to models.IPCallState
}

var validIPCallParicipantStateTransitions = map[ipCallStateTransition]struct{}{
	{from: models.IPCallStatePending, to: models.IPCallStateAccepted}:    {},
	{from: models.IPCallStatePending, to: models.IPCallStateDeclined}:    {},
	{from: models.IPCallStateAccepted, to: models.IPCallStateConnected}:  {},
	{from: models.IPCallStateAccepted, to: models.IPCallStateFailed}:     {},
	{from: models.IPCallStateConnected, to: models.IPCallStateCompleted}: {},
}

func (e *excommsService) InitiateIPCall(ctx context.Context, req *excomms.InitiateIPCallRequest) (*excomms.InitiateIPCallResponse, error) {
	// For now only allow 2 party calls
	if len(req.RecipientEntityIDs) != 1 {
		return nil, grpcErrorf(codes.InvalidArgument, "Must provide exactly one participant")
	}
	if req.RecipientEntityIDs[0] == req.CallerEntityID {
		return nil, grpcErrorf(codes.InvalidArgument, "Recipient may not be the same entity as the caller")
	}

	entityIDs := append(req.RecipientEntityIDs, req.CallerEntityID)
	leres, err := e.directory.LookupEntities(ctx, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_BATCH_ENTITY_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_BatchEntityID{
			BatchEntityID: &directory.IDList{IDs: entityIDs},
		},
	})
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	if len(leres.Entities) != len(entityIDs) {
		return nil, grpcErrorf(codes.InvalidArgument, "Unable to find all entities")
	}

	call := &models.IPCall{Pending: true}
	switch req.Type {
	case excomms.IPCallType_VIDEO:
		call.Type = models.IPCallTypeVideo
	case excomms.IPCallType_AUDIO:
		call.Type = models.IPCallTypeAudio
	default:
		return nil, grpcErrorf(codes.InvalidArgument, "Unknown call type %s", req.Type.String())
	}

	call.Participants = make([]*models.IPCallParticipant, 0, len(leres.Entities))
	var callerPar *models.IPCallParticipant
	for i, e := range leres.Entities {
		if e.AccountID == "" {
			return nil, grpcErrorf(codes.InvalidArgument, "Participant %s missing account ID", e.ID)
		}
		p := &models.IPCallParticipant{
			EntityID:  e.ID,
			AccountID: e.AccountID,
		}
		p.Identity, err = generateIPCallIdentity()
		if err != nil {
			return nil, grpcErrorf(codes.Internal, "Failed to generate identity: %s", err)
		}
		if e.ID == req.CallerEntityID {
			p.Role = models.IPCallParticipantRoleCaller
			p.State = models.IPCallStateAccepted
			callerPar = p
		} else {
			p.Role = models.IPCallParticipantRoleRecipient
			p.State = models.IPCallStatePending
		}
		call.Participants[i] = p
	}

	if err := e.dal.CreateIPCall(ctx, call); err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}

	rcall, err := e.transformIPCallToResponse(call, callerPar)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}

	return &excomms.InitiateIPCallResponse{Call: rcall}, nil
}

func (e *excommsService) PendingIPCalls(ctx context.Context, req *excomms.PendingIPCallsRequest) (*excomms.PendingIPCallsResponse, error) {
	if req.AccountID == "" {
		return nil, grpcErrorf(codes.InvalidArgument, "AccountID required")
	}
	calls, err := e.dal.PendingIPCallsForAccount(ctx, req.AccountID)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	res := &excomms.PendingIPCallsResponse{
		Calls: make([]*excomms.IPCall, 0, len(calls)),
	}
	for _, c := range calls {
		var par *models.IPCallParticipant
		for _, p := range c.Participants {
			if p.AccountID == req.AccountID {
				par = p
				break
			}
		}
		if par == nil {
			// Sanity check, this is an internal consistency error since the pending calls should only include calls with the account as a participant
			return nil, grpcErrorf(codes.Internal, "Participant not found for account %s even though call %s was returned", req.AccountID, c.ID)
		}
		call, err := e.transformIPCallToResponse(c, par)
		if err != nil {
			return nil, grpcErrorf(codes.Internal, err.Error())
		}
		res.Calls = append(res.Calls, call)
	}
	return res, nil
}

func (e *excommsService) UpdateIPCall(ctx context.Context, req *excomms.UpdateIPCallRequest) (*excomms.UpdateIPCallResponse, error) {
	if req.IPCallID == "" {
		return nil, grpcErrorf(codes.InvalidArgument, "IPCallID is required")
	}
	if req.AccountID == "" {
		return nil, grpcErrorf(codes.InvalidArgument, "AccountID is required")
	}
	callID, err := models.ParseIPCallID(req.IPCallID)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, "IPCallID is invalid")
	}
	newState, err := transformIPCallStateToModel(req.State)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, err.Error())
	}
	if newState == models.IPCallStatePending {
		return nil, grpcErrorf(codes.InvalidArgument, "Cannot transition to the PENDING State")
	}
	var call *models.IPCall
	var par *models.IPCallParticipant
	err = e.dal.Transact(func(dl dal.DAL) error {
		call, err = e.dal.IPCall(ctx, callID, dal.ForUpdate)
		if errors.Cause(err) == dal.ErrIPCallNotFound {
			return grpcErrorf(codes.NotFound, "IPCall %s not found", callID)
		} else if err != nil {
			return grpcErrorf(codes.Internal, err.Error())
		}
		for _, p := range call.Participants {
			if p.AccountID == req.AccountID {
				par = p
				break
			}
		}
		if par == nil {
			return grpcErrorf(codes.PermissionDenied, "Account %s not a participant in %s", req.AccountID, callID)
		}
		if newState == par.State {
			// Nothing to do
			return nil
		}
		// Validate that the new state is a valid transition from the current state
		if _, ok := validIPCallParicipantStateTransitions[ipCallStateTransition{from: par.State, to: newState}]; !ok {
			return grpcErrorf(codes.InvalidArgument, "Cannot transition from state %s to %s", par.State, newState)
		}
		// Update the participant so we don't have to refetch when returning the response
		par.State = newState
		if err := e.dal.UpdateIPCallParticipant(ctx, callID, req.AccountID, newState); err != nil {
			return errors.Trace(err)
		}
		if call.Pending {
			// If any participant reached an end-state then the call is no longer pending
			switch newState {
			case models.IPCallStateCompleted, models.IPCallStateDeclined, models.IPCallStateFailed:
				// Update the call so we don't have to refetch when returning the response
				call.Pending = false
				if err := e.dal.UpdateIPCall(ctx, callID, false); err != nil {
					return errors.Trace(err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	rcall, err := e.transformIPCallToResponse(call, par)
	if err != nil {
		return nil, grpcErrorf(codes.Internal, err.Error())
	}
	return &excomms.UpdateIPCallResponse{Call: rcall}, nil
}

func (e *excommsService) transformIPCallToResponse(call *models.IPCall, par *models.IPCallParticipant) (*excomms.IPCall, error) {
	var token string
	var err error
	if par != nil {
		token, err = generateIPCallToken(par.Identity, e.twilioVideoConfigSID).ToJWT(e.twilioApplicationSID, e.twilioSigningKeySID, e.twilioSigningKey)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	c := &excomms.IPCall{
		ID:           call.ID.String(),
		Pending:      call.Pending,
		Token:        token,
		Participants: make([]*excomms.IPCallParticipant, 0, len(call.Participants)),
	}
	switch call.Type {
	case models.IPCallTypeVideo:
		c.Type = excomms.IPCallType_VIDEO
	case models.IPCallTypeAudio:
		c.Type = excomms.IPCallType_AUDIO
	default:
		return nil, errors.Trace(fmt.Errorf("unknown call type %s for call %s", call.Type, call.ID))
	}
	for _, p := range call.Participants {
		cp := &excomms.IPCallParticipant{
			AccountID: p.AccountID,
			EntityID:  p.EntityID,
			Identity:  p.Identity,
		}
		switch p.Role {
		case models.IPCallParticipantRoleCaller:
			cp.Role = excomms.IPCallParticipantRole_CALLER
		case models.IPCallParticipantRoleRecipient:
			cp.Role = excomms.IPCallParticipantRole_RECIPIENT
		default:
			return nil, errors.Trace(fmt.Errorf("unknown role %s for ipcall %s participant account %s", p.Role, call.ID, p.AccountID))
		}
		switch p.State {
		case models.IPCallStateAccepted:
			cp.State = excomms.IPCallState_ACCEPTED
		case models.IPCallStateDeclined:
			cp.State = excomms.IPCallState_DECLINED
		case models.IPCallStateCompleted:
			cp.State = excomms.IPCallState_COMPLETED
		case models.IPCallStateConnected:
			cp.State = excomms.IPCallState_CONNECTED
		case models.IPCallStateFailed:
			cp.State = excomms.IPCallState_FAILED
		case models.IPCallStatePending:
			cp.State = excomms.IPCallState_PENDING
		default:
			return nil, errors.Trace(fmt.Errorf("unknown state %s for ipcall %s participant account %s", p.State, call.ID, p.AccountID))
		}
		c.Participants = append(c.Participants, cp)
	}
	return c, nil
}

func transformIPCallStateToModel(state excomms.IPCallState) (models.IPCallState, error) {
	switch state {
	case excomms.IPCallState_ACCEPTED:
		return models.IPCallStateAccepted, nil
	case excomms.IPCallState_DECLINED:
		return models.IPCallStateDeclined, nil
	case excomms.IPCallState_COMPLETED:
		return models.IPCallStateCompleted, nil
	case excomms.IPCallState_CONNECTED:
		return models.IPCallStateConnected, nil
	case excomms.IPCallState_FAILED:
		return models.IPCallStateFailed, nil
	case excomms.IPCallState_PENDING:
		return models.IPCallStatePending, nil
	}
	return "", errors.Trace(fmt.Errorf("unknown ipcall state %s", state))
}

func generateIPCallToken(identity, configProfileSID string) *twilio.AccessToken {
	return &twilio.AccessToken{
		Identity: identity,
		Grants: []twilio.Grant{twilio.ConversationsGrant{
			ConfigurationProfileSID: configProfileSID,
		}},
		TTL: ipCallTokenTTL,
	}
}

func generateIPCallIdentity() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", errors.Trace(err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}