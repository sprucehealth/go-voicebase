package mock

import (
	"testing"

	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/media"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Compile time check to make sure the mock conforms to the interface
var _ media.MediaClient = &Client{}

// Client is a mock for the directory service client.
type Client struct {
	*mock.Expector
}

// New returns an initialized Client.
func New(t *testing.T) *Client {
	return &Client{&mock.Expector{T: t}}
}

// MediaInfos implements media.MediaClient
func (c *Client) MediaInfos(ctx context.Context, in *media.MediaInfosRequest, opts ...grpc.CallOption) (*media.MediaInfosResponse, error) {
	rets := c.Expector.Record(in)
	if len(rets) == 0 {
		return nil, nil
	}
	return rets[0].(*media.MediaInfosResponse), mock.SafeError(rets[1])
}

// ClaimMedia implements media.MediaClient
func (c *Client) ClaimMedia(ctx context.Context, in *media.ClaimMediaRequest, opts ...grpc.CallOption) (*media.ClaimMediaResponse, error) {
	rets := c.Expector.Record(in)
	if len(rets) == 0 {
		return nil, nil
	}
	return rets[0].(*media.ClaimMediaResponse), mock.SafeError(rets[1])
}

// UpdateMedia implements media.MediaClient
func (c *Client) UpdateMedia(ctx context.Context, in *media.UpdateMediaRequest, opts ...grpc.CallOption) (*media.UpdateMediaResponse, error) {
	rets := c.Expector.Record(in)
	if len(rets) == 0 {
		return nil, nil
	}
	return rets[0].(*media.UpdateMediaResponse), mock.SafeError(rets[1])
}
