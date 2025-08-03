package grpc

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/vagonaizer/authenitfication-service/api/proto/generated"
	"github.com/vagonaizer/authenitfication-service/internal/transport/grpc/handlers"
	"github.com/vagonaizer/authenitfication-service/internal/transport/grpc/interceptors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type Server struct {
	server          *grpc.Server
	authHandler     *handlers.AuthGRPCHandler
	userHandler     *handlers.UserGRPCHandler
	authInterceptor *interceptors.AuthInterceptor
	logInterceptor  *interceptors.LoggingInterceptor
	logger          *logger.Logger
}

func NewServer(
	authHandler *handlers.AuthGRPCHandler,
	userHandler *handlers.UserGRPCHandler,
	authInterceptor *interceptors.AuthInterceptor,
	logInterceptor *interceptors.LoggingInterceptor,
	logger *logger.Logger,
) *Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logInterceptor.Unary(),
			authInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(
			logInterceptor.Stream(),
			authInterceptor.Stream(),
		),
	)

	generated.RegisterAuthServiceServer(server, authHandler)
	generated.RegisterUserServiceServer(server, userHandler)

	reflection.Register(server)

	return &Server{
		server:          server,
		authHandler:     authHandler,
		userHandler:     userHandler,
		authInterceptor: authInterceptor,
		logInterceptor:  logInterceptor,
		logger:          logger,
	}
}

func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.logger.Infof("gRPC server starting on %s", address)
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	s.logger.Info("shutting down gRPC server")
	s.server.GracefulStop()
}
