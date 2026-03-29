package xgrpc

import (
	"fmt"

	"github.com/sptGabriel/gb-libs/config"
	"github.com/sptGabriel/gb-libs/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceConfig struct {
	Name    string // Service  name for logging
	EnvKey  string // Enviroment variable for the service address
	Address string // Optional direct address (takes precedence over EnvKey)
}

type ClientFactory interface {
	Close() error
	Connect(ServiceConfig) (*grpc.ClientConn, error)
}

type clientFactory struct {
	config      *config.AppConfig
	logger      xlog.Logger
	connections map[string]*grpc.ClientConn
}

func (c *clientFactory) GetConnection(serviceConfig ServiceConfig) (*grpc.ClientConn, error) {
	if conn, ok := c.connections[serviceConfig.Name]; ok {
		return conn, nil
	}

	var serviceAddress string
	if serviceConfig.Address != "" {
		serviceAddress = serviceConfig.Address
	}

	if serviceConfig.EnvKey != "" && serviceAddress == "" {
		c.config.LoadDynamic(serviceConfig.EnvKey)
		serviceAddress = c.config.GetExtra(serviceConfig.EnvKey)

		if serviceAddress == "" {
			return nil, fmt.Errorf("%s not set", serviceConfig.EnvKey)
		}
	}

	if serviceAddress == "" {
		return nil, fmt.Errorf("no address or envrioment key provided for service %s", serviceConfig.Name)
	}

	conn, err := grpc.NewClient(serviceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s service: %w", serviceConfig.Name, err)
	}

	c.connections[serviceConfig.Name] = conn

	return conn, nil
}

func (c *clientFactory) Close() error {
	var lastErr error
	for serviceName, conn := range c.connections {
		if err := conn.Close(); err != nil {
			c.logger.Error("Closing connection", "service", serviceName, "error", err)
			lastErr = err

			continue
		}

		c.logger.Info("closed connection", "service", serviceName)
	}

	return lastErr
}

func NewClientFactory(config *config.AppConfig, logger xlog.Logger) *clientFactory {
	return &clientFactory{
		config:      config,
		logger:      logger,
		connections: make(map[string]*grpc.ClientConn),
	}
}
