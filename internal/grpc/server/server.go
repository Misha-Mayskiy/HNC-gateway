package server

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/shvdev1/HackNeChange/api-gateway/internal/gen"
)

// SettingsService defines the interface for settings operations (cache-aside)
type SettingsService interface {
	GetSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error)
	UpdateSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error)
	AnalyzeReview(ctx context.Context, req *pb.AnalyzeReviewRequest) (*pb.AnalyzeReviewResponse, error)
}

// CRUDClient defines the interface for CRUD operations (proxy to downstream)
type CRUDClient interface {
	CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error)
}

// Server implements gRPC server for UserProfileService
type Server struct {
	pb.UnimplementedUserProfileServiceServer
	service SettingsService
	client  CRUDClient
}

func New(svc SettingsService, client CRUDClient) *Server {
	return &Server{service: svc, client: client}
}

// Register registers server on grpcServer
func (s *Server) Register(grpcServer *grpc.Server) {
	pb.RegisterUserProfileServiceServer(grpcServer, s)
}

// CreateUserProfile proxies to downstream customer service
func (s *Server) CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error) {
	return s.client.CreateUserProfile(ctx, req)
}

// GetUserSettings uses service cache-aside logic
func (s *Server) GetUserSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error) {
	return s.service.GetSettings(ctx, req)
}

// UpdateUserSettings uses service logic to update and then invalidate cache
func (s *Server) UpdateUserSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error) {
	return s.service.UpdateSettings(ctx, req)
}

func (s *Server) AnalyzeReview(ctx context.Context, req *pb.AnalyzeReviewRequest) (*pb.AnalyzeReviewResponse, error) {
	return s.service.AnalyzeReview(ctx, req)
}

// Run starts the grpc server
func Run(listenAddr string, srv *Server) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	srv.Register(grpcServer)
	log.Printf("gRPC server listening on %s", listenAddr)
	return grpcServer.Serve(l)
}
