package rabbitmq

import (
	"context"

	"github.com/sptGabriel/gb-libs/config"
	"github.com/sptGabriel/gb-libs/xlog"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type RabbitMQConsumer struct {
	client      *Client
	db          *gorm.DB
	logger      xlog.Logger
	config      *config.AppConfig
	tracer      trace.Tracer
	queueConfig RabbitMQQueueConfig
	workerFun   func(context.Context, []byte) error
}

func NewRabbitMQConsumer(
	clientConfig ClientConfig,
	queueConfig RabbitMQQueueConfig,
	db *gorm.DB,
	logger xlog.Logger,
	config *config.AppConfig,
	tracer trace.Tracer,
	workerFun func(context.Context, []byte) error,
) (*RabbitMQConsumer, error) {
	rmqClient, err := NewRabbitMQClient(clientConfig, logger, tracer)
	if err != nil {
		return nil, err
	}

	consumer := &RabbitMQConsumer{
		client:      rmqClient,
		db:          db,
		logger:      logger,
		config:      config,
		tracer:      tracer,
		queueConfig: queueConfig,
		workerFun:   workerFun,
	}

	err = rmqClient.NewQueue(queueConfig)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ queue.")
		return nil, err
	}

	return consumer, nil
}

func (c *RabbitMQConsumer) Client() *Client {
	return c.client
}

func (c RabbitMQConsumer) Consume(ctx context.Context) {
	err := c.client.StartConsuming(
		ctx,
		c.queueConfig,
		c.config.Name,
		c.workerFun,
	)
	if err != nil {
		c.logger.Error("failed to consume messages", "error", err)
	}
}
