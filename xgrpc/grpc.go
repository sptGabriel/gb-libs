package xgrpc

import (
	"context"
	"log/slog"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type UnaryServerInterceptor = grpc.UnaryServerInterceptor

type Server struct {
	grpcServer *grpc.Server
	logger     *slog.Logger
}

func NewServer(unaryInterceptors []UnaryServerInterceptor, logger *slog.Logger, opts ...grpc.ServerOption) *Server {
	interceptOptions := grpc.ChainUnaryInterceptor(unaryInterceptors...)
	options := []grpc.ServerOption{
		interceptOptions,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}
	options = append(options, opts...)

	return &Server{
		grpcServer: grpc.NewServer(options...),
		logger:     logger,
	}
}

func (s *Server) EnableReflection() *Server {
	reflection.Register(s.grpcServer)
	return s
}

func (s *Server) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}

func (s *Server) GetServiceRegistrar() grpc.ServiceRegistrar {
	return s.grpcServer
}

func (s *Server) RegisterService(registerFunc func(grpc.ServiceRegistrar)) {
	registerFunc(s.grpcServer)
}

func (s *Server) Start(listener net.Listener, addr string) error {
	s.logger.Info("starting gRPC server...", slog.String("address", addr))
	if err := s.grpcServer.Serve(listener); err != nil {
		s.logger.Error("failed to start server", slog.Any("err", err))
		return err
	}
	return nil
}

func (s *Server) MustStart(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("failed to start server", slog.Any("err", err))
		panic(err)
	}
	err = s.Start(listener, addr)
	if err != nil {
		panic(err)
	}
}

func (s *Server) MustStartWithContext(ctx context.Context, addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("failed to start server", slog.Any("err", err))
		panic(err)
	}

	go func() {
		err := s.Start(listener, addr)
		if err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
	s.logger.Info("gRPC server stop requested")
	s.grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")
}
