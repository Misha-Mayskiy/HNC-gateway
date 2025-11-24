package service

import (
	"context"
	"log"

	pb "github.com/shvdev1/HackNeChange/api-gateway/internal/gen"
	redisstorage "github.com/shvdev1/HackNeChange/api-gateway/internal/storage/redis"
)

// CustomerServiceClient defines the interface for calling downstream customer service
type CustomerServiceClient interface {
	GetSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error)
	UpdateSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error)
	CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error)
}

// Service provides business logic for the API gateway
type Service struct {
	store  redisstorage.Storage
	client CustomerServiceClient
}

// New creates a new service
func New(store redisstorage.Storage, client CustomerServiceClient) *Service {
	return &Service{store: store, client: client}
}

// GetSettings implements cache-aside: check cache, otherwise fetch from customer and store in background
func (s *Service) GetSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error) {
	if req == nil || req.UserId == "" {
		return nil, nil
	}
	// Try cache
	cached, err := s.store.Get(ctx, req.UserId)
	if err != nil {
		log.Printf("redis get error: %v", err)
	}
	if cached != nil {
		return cached, nil
	}

	// Not in cache - call customer service
	resp, err := s.client.GetSettings(ctx, req)
	if err != nil {
		return nil, err
	}
	// Save to redis in background
	go func(r *pb.GetUserSettingsResponse, userID string) {
		if err := s.store.Set(context.Background(), userID, r); err != nil {
			log.Printf("failed to set cache for user %s: %v", userID, err)
		}
	}(resp, req.UserId)

	return resp, nil
}

// UpdateSettings - call downstream and invalidate cache
func (s *Service) UpdateSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error) {
	if req == nil || req.UserId == "" {
		return nil, nil
	}
	resp, err := s.client.UpdateSettings(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.store.Invalidate(ctx, req.UserId); err != nil {
		log.Printf("failed to invalidate cache for user %s: %v", req.UserId, err)
	}
	return resp, nil
}
