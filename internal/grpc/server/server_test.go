package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Misha-Mayskiy/HNC-proto/gen/go/user"
)

// MockService mocks the SettingsService interface
type MockService struct {
	mock.Mock
}

func (m *MockService) GetSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserSettingsResponse), args.Error(1)
}

func (m *MockService) UpdateSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateUserSettingsResponse), args.Error(1)
}

func (m *MockService) AnalyzeReview(ctx context.Context, req *pb.AnalyzeReviewRequest) (*pb.AnalyzeReviewResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AnalyzeReviewResponse), args.Error(1)
}

// MockClient mocks the CRUDClient interface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateUserProfileResponse), args.Error(1)
}

func TestGetUserSettings_Success(t *testing.T) {
	mockSvc := new(MockService)
	mockClient := new(MockClient)

	cachedResp := &pb.GetUserSettingsResponse{
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	mockSvc.On("GetSettings", mock.Anything, mock.MatchedBy(func(req *pb.GetUserSettingsRequest) bool {
		return req.UserId == "user123"
	})).Return(cachedResp, nil)

	srv := New(mockSvc, mockClient)

	req := &pb.GetUserSettingsRequest{UserId: "user123"}
	resp, err := srv.GetUserSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, cachedResp, resp)
	mockSvc.AssertCalled(t, "GetSettings", mock.Anything, req)
}

func TestGetUserSettings_Error(t *testing.T) {
	mockSvc := new(MockService)
	mockClient := new(MockClient)

	mockSvc.On("GetSettings", mock.Anything, mock.MatchedBy(func(req *pb.GetUserSettingsRequest) bool {
		return req.UserId == "user456"
	})).Return(nil, errors.New("cache error"))

	srv := New(mockSvc, mockClient)

	req := &pb.GetUserSettingsRequest{UserId: "user456"}
	resp, err := srv.GetUserSettings(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateUserSettings_Success(t *testing.T) {
	mockSvc := new(MockService)
	mockClient := new(MockClient)

	updateResp := &pb.UpdateUserSettingsResponse{
		Theme:       "light",
		PickedModel: "claude-3",
		Font:        "serif",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	mockSvc.On("UpdateSettings", mock.Anything, mock.MatchedBy(func(req *pb.UpdateUserSettingsRequest) bool {
		return req.UserId == "user789"
	})).Return(updateResp, nil)

	srv := New(mockSvc, mockClient)

	req := &pb.UpdateUserSettingsRequest{
		UserId:      "user789",
		Theme:       "light",
		PickedModel: "claude-3",
		Font:        "serif",
	}
	resp, err := srv.UpdateUserSettings(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, updateResp, resp)
	mockSvc.AssertCalled(t, "UpdateSettings", mock.Anything, req)
}

func TestAnalyzeReview_Success(t *testing.T) {
	mockSvc := new(MockService)
	mockClient := new(MockClient)

	analyzeResp := &pb.AnalyzeReviewResponse{
		ReviewId: "uuid-123",
		Status:   "QUEUED",
	}

	mockSvc.On("AnalyzeReview", mock.Anything, mock.MatchedBy(func(req *pb.AnalyzeReviewRequest) bool {
		return req.UserId == "u1" && req.Text == "hello"
	})).Return(analyzeResp, nil)

	srv := New(mockSvc, mockClient)

	req := &pb.AnalyzeReviewRequest{UserId: "u1", Text: "hello"}
	resp, err := srv.AnalyzeReview(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, analyzeResp, resp)
	mockSvc.AssertCalled(t, "AnalyzeReview", mock.Anything, req)
}

func TestCreateUserProfile_ProxiesToClient(t *testing.T) {
	mockSvc := new(MockService)
	mockClient := new(MockClient)

	profile := &pb.UserProfile{
		UserId:      "user999",
		Username:    "testuser",
		CompanyName: "Acme Inc",
		PhoneNumber: "+1234567890",
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
		UpdatedAt:   timestamppb.New(time.Now()),
	}

	createResp := &pb.CreateUserProfileResponse{
		Profile: profile,
	}

	mockClient.On("CreateUserProfile", mock.Anything, mock.MatchedBy(func(req *pb.CreateUserProfileRequest) bool {
		return req.UserId == "user999"
	})).Return(createResp, nil)

	srv := New(mockSvc, mockClient)

	req := &pb.CreateUserProfileRequest{
		UserId:      "user999",
		Username:    "testuser",
		CompanyName: "Acme Inc",
		PhoneNumber: "+1234567890",
		Theme:       "dark",
		PickedModel: "gpt-4",
		Font:        "monospace",
	}

	resp, err := srv.CreateUserProfile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, createResp, resp)
	mockClient.AssertCalled(t, "CreateUserProfile", mock.Anything, req)
	mockSvc.AssertNotCalled(t, "GetSettings")
}
