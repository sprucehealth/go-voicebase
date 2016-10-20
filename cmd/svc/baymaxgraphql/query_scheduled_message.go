package main

import (
	"context"
	"sort"

	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/models"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/raccess"
	baymaxgraphqlsettings "github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/settings"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/backend/svc/settings"
	"github.com/sprucehealth/backend/svc/threading"
	"github.com/sprucehealth/graphql"
)

var scheduledMessageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ScheduledMessage",
	Fields: graphql.Fields{
		"id":                    &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"threadItem":            &graphql.Field{Type: graphql.NewNonNull(threadItemType)},
		"scheduledForTimestamp": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
	},
})

func getScheduledMessages(ctx context.Context, ram raccess.ResourceAccessor, threadID, organizationID string) ([]*models.ScheduledMessage, error) {
	resp, err := ram.ScheduledMessages(ctx, &threading.ScheduledMessagesRequest{
		LookupKey: &threading.ScheduledMessagesRequest_ThreadID{
			ThreadID: threadID,
		},
		// At the graphql layer just show pending
		Status: []threading.ScheduledMessageStatus{threading.SCHEDULED_MESSAGE_STATUS_PENDING},
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	scheduledMessages := transformScheduledMessagesToResponse(ctx, resp.ScheduledMessages, organizationID)
	sort.Sort(scheduledMessageByScheduledFor(scheduledMessages))
	return scheduledMessages, nil
}

func transformScheduledMessagesToResponse(ctx context.Context, ms []*threading.ScheduledMessage, organizationID string) []*models.ScheduledMessage {
	rms := make([]*models.ScheduledMessage, len(ms))
	for i, m := range ms {
		rms[i] = transformScheduledMessageToResponse(ctx, m, organizationID)
	}
	return rms
}

func transformScheduledMessageToResponse(ctx context.Context, m *threading.ScheduledMessage, organizationID string) *models.ScheduledMessage {
	return &models.ScheduledMessage{
		ID:           m.ID,
		ScheduledFor: m.ScheduledFor,
		ThreadItem: &models.ThreadItem{
			ID:             m.ID,
			Internal:       m.Internal,
			Timestamp:      m.Modified,
			ActorEntityID:  m.ActorEntityID,
			OrganizationID: organizationID,
			Data:           m.GetMessage(),
		},
	}
}

func isScheduledMessagesEnabled() func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		svc := serviceFromParams(p)
		ctx := p.Context
		acc := gqlctx.Account(ctx)
		if acc.Type != auth.AccountType_PROVIDER {
			return false, nil
		}
		var orgID string
		switch s := p.Source.(type) {
		case *models.Organization:
			if s == nil {
				return false, nil
			}
			orgID = s.ID
		case *models.Thread:
			if s == nil {
				return false, nil
			}
			orgID = s.OrganizationID
		default:
			return nil, errors.Errorf("Unhandled source type %T for isScheduledMessagesEnabled, returning false", s)
		}
		booleanValue, err := settings.GetBooleanValue(ctx, svc.settings, &settings.GetValuesRequest{
			NodeID: orgID,
			Keys: []*settings.ConfigKey{
				{
					Key: baymaxgraphqlsettings.ConfigKeyScheduledMessages,
				},
			},
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
		return booleanValue.Value, nil
	}
}

type scheduledMessageByScheduledFor []*models.ScheduledMessage

func (s scheduledMessageByScheduledFor) Len() int      { return len(s) }
func (s scheduledMessageByScheduledFor) Swap(a, b int) { s[a], s[b] = s[b], s[a] }
func (s scheduledMessageByScheduledFor) Less(a, b int) bool {
	return s[a].ScheduledFor > s[b].ScheduledFor
}
