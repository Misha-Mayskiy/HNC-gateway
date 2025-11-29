package customerclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	customer "github.com/Misha-Mayskiy/HNC-proto/gen/go/user"
)

// Client is a wrapper around the generated gRPC client
type Client struct {
	client customer.UserProfileServiceClient
	conn   *grpc.ClientConn
}

// New connects to customer service and returns the wrapped client
func New(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	c := customer.NewUserProfileServiceClient(conn)
	return &Client{client: c, conn: conn}, nil
}

// GetSettings calls downstream customer service GetUserSettings
func (c *Client) GetSettings(ctx context.Context, req *customer.GetUserSettingsRequest) (*customer.GetUserSettingsResponse, error) {
	return c.client.GetUserSettings(ctx, req)
}

// UpdateSettings calls downstream customer service UpdateUserSettings
func (c *Client) UpdateSettings(ctx context.Context, req *customer.UpdateUserSettingsRequest) (*customer.UpdateUserSettingsResponse, error) {
	return c.client.UpdateUserSettings(ctx, req)
}

// Forwards CreateUserProfile
func (c *Client) CreateUserProfile(ctx context.Context, req *customer.CreateUserProfileRequest) (*customer.CreateUserProfileResponse, error) {
	return c.client.CreateUserProfile(ctx, req)
}

// Close is a noop for now, but provided for symmetry if conn handling is changed
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
