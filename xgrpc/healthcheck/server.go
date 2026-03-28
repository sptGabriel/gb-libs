package healthcheck

// import (
// 	"context"

// 	"github.com/waycarbon/go-ecosystem-defaults/v2/config"
// 	libgrpc

// 	health
// )

// type healthCheckServer struct {
// 	health.UnimplementedHealthCheckServiceServer
// 	config config.AppConfig
// 	// SemVer number, typically found in the service pkg package.
// 	version string
// }

// func NewHealthCheckServer(config config.AppConfig, version string) health.HealthCheckServiceServer {
// 	return &healthCheckServer{config: config, version: version}
// }

// func RegisterGRPCService(server *libgrpc.Server, config config.AppConfig, version string) {
// 	healthServer := NewHealthCheckServer(config, version)
// 	health.RegisterHealthCheckServiceServer(server.GetGRPCServer(), healthServer)
// }

// func (s *healthCheckServer) CheckHealth(
// 	ctx context.Context,
// 	in *health.CheckHealthRequest,
// ) (*health.CheckHealthResponse, error) {
// 	return &health.CheckHealthResponse{
// 		Service: s.config.AppName(),
// 		Version: s.version,
// 	}, nil
// }
