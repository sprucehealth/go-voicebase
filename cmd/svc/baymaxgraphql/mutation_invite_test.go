package main

import (
	"encoding/json"
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/gqlctx"
	"github.com/sprucehealth/backend/cmd/svc/baymaxgraphql/internal/models"
	"github.com/sprucehealth/backend/device"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/invite"
	"github.com/sprucehealth/backend/test"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

func TestAssociateInviteMutation(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	var acc *models.Account
	ctx = gqlctx.WithAccount(ctx, acc)
	sh := &device.SpruceHeaders{DeviceID: "deviceID"}
	ctx = gqlctx.WithSpruceHeaders(ctx, sh)

	g.inviteC.Expect(mock.NewExpectation(g.inviteC.LookupInvite, &invite.LookupInviteRequest{
		Token: "token",
	}).WithReturns(&invite.LookupInviteResponse{
		Values: []*invite.AttributionValue{{Key: "foo", Value: "bar"}},
	}, nil))

	g.inviteC.Expect(mock.NewExpectation(g.inviteC.SetAttributionData, &invite.SetAttributionDataRequest{
		DeviceID: "deviceID",
		Values:   []*invite.AttributionValue{{Key: "foo", Value: "bar"}},
	}).WithReturns(&invite.SetAttributionDataResponse{}, nil))

	res := g.query(ctx, `
		mutation _ {
			associateInvite(input: {
				clientMutationId: "a1b2c3",
				token: "token",
			}) {
				clientMutationId
				success
				values {
					key
					value
				}
			}
		}`, nil)
	b, err := json.MarshalIndent(res, "", "\t")
	test.OK(t, err)
	test.Equals(t, `{
	"data": {
		"associateInvite": {
			"clientMutationId": "a1b2c3",
			"success": true,
			"values": [
				{
					"key": "foo",
					"value": "bar"
				}
			]
		}
	}
}`, string(b))
}

func TestAssociateInviteMutation_NotFound(t *testing.T) {
	g := newGQL(t)
	defer g.finish()

	ctx := context.Background()
	var acc *models.Account
	ctx = gqlctx.WithAccount(ctx, acc)
	sh := &device.SpruceHeaders{DeviceID: "deviceID"}
	ctx = gqlctx.WithSpruceHeaders(ctx, sh)

	g.inviteC.Expect(mock.NewExpectation(g.inviteC.LookupInvite, &invite.LookupInviteRequest{
		Token: "token",
	}).WithReturns(&invite.LookupInviteResponse{}, grpcErrorf(codes.NotFound, "not found")))

	res := g.query(ctx, `
		mutation _ {
			associateInvite(input: {
				clientMutationId: "a1b2c3",
				token: "token",
			}) {
				clientMutationId
				success
				errorCode
			}
		}`, nil)
	b, err := json.MarshalIndent(res, "", "\t")
	test.OK(t, err)
	test.Equals(t, `{
	"data": {
		"associateInvite": {
			"clientMutationId": "a1b2c3",
			"errorCode": "INVALID_INVITE",
			"success": false
		}
	}
}`, string(b))
}