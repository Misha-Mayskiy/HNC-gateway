package server

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	customerclient "github.com/shvdev1/HackNeChange/api-gateway/internal/clients/customer"
	pb "github.com/shvdev1/HackNeChange/api-gateway/internal/gen"
	"github.com/shvdev1/HackNeChange/api-gateway/internal/service"
)

// Server implements gRPC server for UserProfileService
type Server struct {
	pb.UnimplementedUserProfileServiceServer
	service *service.Service
	client  *customerclient.Client
}

func New(svc *service.Service, client *customerclient.Client) *Server {
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
