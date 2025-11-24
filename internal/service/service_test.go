package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/shvdev1/HackNeChange/api-gateway/internal/gen"
)

// MockStorage mocks the redisstorage.Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Get(ctx context.Context, userID string) (*pb.GetUserSettingsResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserSettingsResponse), args.Error(1)
}

func (m *MockStorage) Set(ctx context.Context, userID string, data *pb.GetUserSettingsResponse) error {
	args := m.Called(ctx, userID, data)
	return args.Error(0)
}

func (m *MockStorage) Invalidate(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockCustomerClient mocks the CustomerServiceClient interface
type MockCustomerClient struct {
	mock.Mock
}

func (m *MockCustomerClient) GetSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserSettingsResponse), args.Error(1)
}

func (m *MockCustomerClient) UpdateSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateUserSettingsResponse), args.Error(1)
}

func (m *MockCustomerClient) CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateUserProfileResponse), args.Error(1)
}

// TestGetSettings_CacheHit tests cache-aside hit scenario
func TestGetSettings_CacheHit(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	cachedResp := &pb.GetUserSettingsResponse{
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	mockStorage.On("Get", mock.Anything, "user123").Return(cachedResp, nil)

	svc := New(mockStorage, mockClient)
	req := &pb.GetUserSettingsRequest{UserId: "user123"}

	resp, err := svc.GetSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, cachedResp, resp)
	mockStorage.AssertExpectations(t)
	mockClient.AssertNotCalled(t, "GetSettings")
}

// TestGetSettings_CacheMiss tests cache-aside miss: fetch from client, store in background
func TestGetSettings_CacheMiss(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	freshResp := &pb.GetUserSettingsResponse{
		Theme:       "light",
		PickedModel: "claude-3",
		Font:        "serif",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	mockStorage.On("Get", mock.Anything, "user456").Return(nil, nil)
	mockStorage.On("Set", mock.Anything, "user456", freshResp).Return(nil)
	mockClient.On("GetSettings", mock.Anything, mock.MatchedBy(func(req *pb.GetUserSettingsRequest) bool {
		return req.UserId == "user456"
	})).Return(freshResp, nil)

	svc := New(mockStorage, mockClient)
	req := &pb.GetUserSettingsRequest{UserId: "user456"}

	resp, err := svc.GetSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, freshResp, resp)
	mockStorage.AssertCalled(t, "Get", mock.Anything, "user456")
	mockClient.AssertCalled(t, "GetSettings", mock.Anything, req)

	// Give background goroutine time to complete
	time.Sleep(100 * time.Millisecond)
	mockStorage.AssertCalled(t, "Set", mock.Anything, "user456", freshResp)
}

// TestGetSettings_ClientError tests error from downstream service
func TestGetSettings_ClientError(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	mockStorage.On("Get", mock.Anything, "user789").Return(nil, nil)
	mockClient.On("GetSettings", mock.Anything, mock.MatchedBy(func(req *pb.GetUserSettingsRequest) bool {
		return req.UserId == "user789"
	})).Return(nil, errors.New("customer service unavailable"))

	svc := New(mockStorage, mockClient)
	req := &pb.GetUserSettingsRequest{UserId: "user789"}

	resp, err := svc.GetSettings(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockStorage.AssertNotCalled(t, "Set")
}

// TestUpdateSettings_Success tests update and cache invalidation
func TestUpdateSettings_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	updateResp := &pb.UpdateUserSettingsResponse{
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	mockClient.On("UpdateSettings", mock.Anything, mock.MatchedBy(func(req *pb.UpdateUserSettingsRequest) bool {
		return req.UserId == "user999"
	})).Return(updateResp, nil)
	mockStorage.On("Invalidate", mock.Anything, "user999").Return(nil)

	svc := New(mockStorage, mockClient)
	req := &pb.UpdateUserSettingsRequest{
		UserId:      "user999",
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
	}

	resp, err := svc.UpdateSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, updateResp, resp)
	mockClient.AssertCalled(t, "UpdateSettings", mock.Anything, req)
	mockStorage.AssertCalled(t, "Invalidate", mock.Anything, "user999")
}

// TestUpdateSettings_ClientError tests error from downstream
func TestUpdateSettings_ClientError(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	mockClient.On("UpdateSettings", mock.Anything, mock.MatchedBy(func(req *pb.UpdateUserSettingsRequest) bool {
		return req.UserId == "userXXX"
	})).Return(nil, errors.New("database error"))

	svc := New(mockStorage, mockClient)
	req := &pb.UpdateUserSettingsRequest{
		UserId:      "userXXX",
		Theme:       "light",
		PickedModel: "claude-2",
		Font:        "sans-serif",
	}

	resp, err := svc.UpdateSettings(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockStorage.AssertNotCalled(t, "Invalidate")
}

// TestGetSettings_NilRequest handles nil or empty user ID
func TestGetSettings_NilRequest(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	svc := New(mockStorage, mockClient)

	// Test nil request
	resp, err := svc.GetSettings(context.Background(), nil)
	assert.Nil(t, err)
	assert.Nil(t, resp)

	// Test empty user ID
	resp, err = svc.GetSettings(context.Background(), &pb.GetUserSettingsRequest{UserId: ""})
	assert.Nil(t, err)
	assert.Nil(t, resp)

	mockStorage.AssertNotCalled(t, "Get")
	mockClient.AssertNotCalled(t, "GetSettings")
}

// TestUpdateSettings_NilRequest handles nil or empty user ID
func TestUpdateSettings_NilRequest(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)

	svc := New(mockStorage, mockClient)

	// Test nil request
	resp, err := svc.UpdateSettings(context.Background(), nil)
	assert.Nil(t, err)
	assert.Nil(t, resp)

	// Test empty user ID
	resp, err = svc.UpdateSettings(context.Background(), &pb.UpdateUserSettingsRequest{UserId: ""})
	assert.Nil(t, err)
	assert.Nil(t, resp)

	mockClient.AssertNotCalled(t, "UpdateSettings")
	mockStorage.AssertNotCalled(t, "Invalidate")
}
