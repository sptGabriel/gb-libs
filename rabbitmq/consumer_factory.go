package rabbitmq

import (
	"context"
	"log/slog"
	"time"

	"github.com/sptGabriel/gb-libs/config"
	"github.com/sptGabriel/gb-libs/xlog"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type MessageHandler func(ctx context.Context, message []byte) error

type QueueConsumerFactory struct {
	DB     *gorm.DB
	Logger xlog.Logger
	Config *config.AppConfig
	Tracer trace.Tracer
}

func NewQueueConsumerFactory(
	db *gorm.DB,
	logger xlog.Logger,
	config *config.AppConfig,
	tracer trace.Tracer,
) *QueueConsumerFactory {
	return &QueueConsumerFactory{
		DB:     db,
		Logger: logger.With(slog.String("component", "queue-consumer-factory")),
		Config: config,
		Tracer: tracer,
	}
}

func (f *QueueConsumerFactory) GetClientConfig() ClientConfig {
	return ClientConfig{
		ConnectionUrl:                f.Config.GetExtra("RABBITMQ_URL"),
		ReconnectDelay:               5 * time.Second,
		ChannelReinitializationDelay: 2 * time.Second,
		PublishTimeout:               2 * time.Second,
		WaitForConnectionTimeout:     5 * time.Second,
	}
}

func (f *QueueConsumerFactory) GetQueueConfig(queueName, routingKey string) RabbitMQQueueConfig {
	return RabbitMQQueueConfig{
		Exchange:     f.Config.GetExtra("INTEGRATION_EVENTS_EXCHANGE"),
		Queue:        f.Config.GetExtra(queueName),
		RoutingKeys:  []string{f.Config.GetExtra(routingKey)},
		ExchangeType: TopicExchange,
	}
}

func (f *QueueConsumerFactory) SetupConsumer(
	queueConfigName string,
	routingKeyName string,
	worker MessageHandler,
) (*RabbitMQConsumer, error) {
	clientConfig := f.GetClientConfig()
	queueConfig := f.GetQueueConfig(queueConfigName, routingKeyName)

	consumer, err := NewRabbitMQConsumer(
		clientConfig,
		queueConfig,
		f.DB,
		f.Logger,
		f.Config,
		f.Tracer,
		worker,
	)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}
