package server

import (
	"testing"

	"github.com/sprucehealth/backend/cmd/svc/invite/internal/dal"
	"github.com/sprucehealth/backend/cmd/svc/invite/internal/models"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"golang.org/x/net/context"
)

// Build time check for matching against the interface
var _ dal.DAL = &mockDAL{}

type mockDAL struct {
	*mock.Expector
}

func newMockDAL(t *testing.T) *mockDAL {
	return &mockDAL{&mock.Expector{T: t}}
}

func (dl *mockDAL) AttributionData(ctx context.Context, deviceID string) (map[string]string, error) {
	r := dl.Expector.Record(deviceID)
	return r[0].(map[string]string), mock.SafeError(r[1])
}

func (dl *mockDAL) SetAttributionData(ctx context.Context, deviceID string, values map[string]string) error {
	r := dl.Expector.Record(deviceID, values)
	return mock.SafeError(r[0])
}

func (dl *mockDAL) InsertInvite(ctx context.Context, invite *models.Invite) error {
	r := dl.Expector.Record(invite)
	return mock.SafeError(r[0])
}

func (dl *mockDAL) InviteForToken(ctx context.Context, token string) (*models.Invite, error) {
	r := dl.Expector.Record(token)
	return r[0].(*models.Invite), mock.SafeError(r[1])
}
