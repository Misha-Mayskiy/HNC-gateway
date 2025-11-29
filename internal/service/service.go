package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	pb "github.com/shvdev1/HackNeChange/api-gateway/internal/gen"
	redisstorage "github.com/shvdev1/HackNeChange/api-gateway/internal/storage/redis"
)

// EventProducer интерфейс, чтобы не зависеть от kafka напрямую (для тестов удобно)
type EventProducer interface {
	SendMessage(key string, value interface{}) error
}

// ReviewPayload - то, что улетит в Кафку (должно совпадать с тем, что ждет process-service)
type ReviewPayload struct {
	ReviewID  string    `json:"review_id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// CustomerServiceClient defines the interface for calling downstream customer service
type CustomerServiceClient interface {
	GetSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error)
	UpdateSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error)
	CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error)
}

// Service provides business logic for the API gateway
type Service struct {
	store    redisstorage.Storage
	client   CustomerServiceClient
	producer EventProducer
}

// New creates a new service
func New(store redisstorage.Storage, client CustomerServiceClient, producer EventProducer) *Service {
	return &Service{
		store:    store,
		client:   client,
		producer: producer,
	}
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

// AnalyzeReview отправляет отзыв в Kafka для асинхронного анализа
func (s *Service) AnalyzeReview(ctx context.Context, req *pb.AnalyzeReviewRequest) (*pb.AnalyzeReviewResponse, error) {
	// 1. Генерируем UUID для отзыва
	reviewID := uuid.New().String()

	// 2. Собираем пейлоад
	payload := ReviewPayload{
		ReviewID:  reviewID,
		UserID:    req.UserId,
		Text:      req.Text,
		Source:    req.Source,
		CreatedAt: time.Now(),
	}

	// 3. Отправляем в Kafka (асинхронно для клиента, синхронно для кода)
	if err := s.producer.SendMessage(req.UserId, payload); err != nil {
		log.Printf("Failed to send review to kafka: %v", err)
		return nil, err
	}

	// 4. Сразу возвращаем ответ "В очереди"
	return &pb.AnalyzeReviewResponse{
		ReviewId: reviewID,
		Status:   "QUEUED",
	}, nil
}
