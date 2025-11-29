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

// MockProducer mocks the EventProducer interface (Kafka)
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// TestGetSettings_CacheHit tests cache-aside hit scenario
func TestGetSettings_CacheHit(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)
	mockProducer := new(MockProducer)

	cachedResp := &pb.GetUserSettingsResponse{
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	mockStorage.On("Get", mock.Anything, "user123").Return(cachedResp, nil)

	// Передаем mockProducer третьим аргументом
	svc := New(mockStorage, mockClient, mockProducer)
	req := &pb.GetUserSettingsRequest{UserId: "user123"}

	resp, err := svc.GetSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, cachedResp, resp)
	mockStorage.AssertExpectations(t)
	mockClient.AssertNotCalled(t, "GetSettings")
}

// TestGetSettings_CacheMiss tests cache-aside miss
func TestGetSettings_CacheMiss(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)
	mockProducer := new(MockProducer)

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

	svc := New(mockStorage, mockClient, mockProducer)
	req := &pb.GetUserSettingsRequest{UserId: "user456"}

	resp, err := svc.GetSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, freshResp, resp)
	mockStorage.AssertCalled(t, "Get", mock.Anything, "user456")
	mockClient.AssertCalled(t, "GetSettings", mock.Anything, req)

	time.Sleep(100 * time.Millisecond) // wait for goroutine
	mockStorage.AssertCalled(t, "Set", mock.Anything, "user456", freshResp)
}

// TestGetSettings_ClientError tests error from downstream service
func TestGetSettings_ClientError(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)
	mockProducer := new(MockProducer)

	mockStorage.On("Get", mock.Anything, "user789").Return(nil, nil)
	mockClient.On("GetSettings", mock.Anything, mock.MatchedBy(func(req *pb.GetUserSettingsRequest) bool {
		return req.UserId == "user789"
	})).Return(nil, errors.New("customer service unavailable"))

	svc := New(mockStorage, mockClient, mockProducer)
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
	mockProducer := new(MockProducer)

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

	svc := New(mockStorage, mockClient, mockProducer)
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

// TestAnalyzeReview_Success (NEW TEST)
func TestAnalyzeReview_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)
	mockProducer := new(MockProducer)

	req := &pb.AnalyzeReviewRequest{
		UserId: "user123",
		Text:   "Great app",
		Source: "web",
	}

	// Expect SendMessage to be called with correct key and payload type
	mockProducer.On("SendMessage", "user123", mock.MatchedBy(func(val interface{}) bool {
		// We can cast to ReviewPayload to check fields if we exported it or defined it in test
		// For now just checking it's not nil is enough for the mock match
		return val != nil
	})).Return(nil)

	svc := New(mockStorage, mockClient, mockProducer)

	resp, err := svc.AnalyzeReview(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "QUEUED", resp.Status)
	assert.NotEmpty(t, resp.ReviewId)

	mockProducer.AssertExpectations(t)
}

// TestAnalyzeReview_ProducerError (NEW TEST)
func TestAnalyzeReview_ProducerError(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClient := new(MockCustomerClient)
	mockProducer := new(MockProducer)

	mockProducer.On("SendMessage", mock.Anything, mock.Anything).Return(errors.New("kafka error"))

	svc := New(mockStorage, mockClient, mockProducer)
	req := &pb.AnalyzeReviewRequest{UserId: "u1", Text: "text"}

	resp, err := svc.AnalyzeReview(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}
