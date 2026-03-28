package rabbitmq

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sptGabriel/defaults/xlog"
	"github.com/sptGabriel/defaults/xotel"
)

var (
	ErrNotConnected                 = errors.New("not connected to a server")
	ErrClientConnectionAlreadyClose = errors.New("connection already closed")
)

type ExchangeType string

const (
	TopicExchange   = ExchangeType("topic")
	FanoutExchange  = ExchangeType("fanout")
	DirectExchange  = ExchangeType("direct")
	HeadersExchange = ExchangeType("headers")
)

type ClientConfig struct {
	ConnectionUrl                string
	ReconnectDelay               time.Duration
	PublishTimeout               time.Duration
	WaitForConnectionTimeout     time.Duration
	ChannelReinitializationDelay time.Duration
}

type RabbitMQQueueConfig struct {
	Queue           string
	Exchange        string
	RoutingKeys     []string
	ExchangeType    ExchangeType
	ConsumerTimeout time.Duration
}

type Client struct {
	m       *sync.RWMutex
	isReady bool

	done            chan struct{}
	waitForChan     chan struct{}
	ch              *amqp.Channel
	conn            *amqp.Connection
	config          ClientConfig
	notifyChanClose chan *amqp.Error
	notifyConnClose chan *amqp.Error
	logger          xlog.Logger
	tracer          trace.Tracer
}

func (c *Client) Connection() *amqp.Connection {
	return c.conn
}

func (c *Client) Channel() *amqp.Channel {
	return c.ch
}

func (c *Client) Config() ClientConfig {
	return c.config
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.config.ConnectionUrl)
	if err != nil {
		c.logger.Error("failed to dial connection", "error", err)
		return err
	}

	c.changeConnection(conn)
	c.logger.Debug("RabbitMQ successful connection")

	return nil
}

func (c *Client) changeConnection(conn *amqp.Connection) {
	c.conn = conn
	c.notifyConnClose = make(chan *amqp.Error, 1)
	c.conn.NotifyClose(c.notifyConnClose)
}

func (c *Client) handleReinitialization() bool {
	for {
		c.m.Lock()
		c.isReady = false
		c.m.Unlock()

		err := c.initialize(c.conn)
		if err != nil {
			c.logger.Error("RabbitMQ failed to initialize channel")
			select {
			case <-c.done:
				return true
			case err := <-c.notifyConnClose:
				c.logger.Debug(
					"RabbitMQ connection closed while initializing channel",
					slog.Any("reason", err),
				)
				return false
			case <-time.After(c.config.ChannelReinitializationDelay):
				continue
			}
		}

		select {
		case <-c.done:
			return true
		case err := <-c.notifyConnClose:
			c.logger.Debug(
				"RabbitMQ connection closed, reconnecting...",
				slog.Any("reason", err),
			)
			return false
		case err := <-c.notifyChanClose:
			c.logger.Debug(
				"RabbitMQ channel closed, will start new initialization attempt",
				slog.Any("reason", err),
			)
		}
	}
}

func (c *Client) initialize(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		c.logger.Error("RabbitMQ failed to open channel", "error", err)
		return err
	}

	c.changeChannel(ch)

	c.m.Lock()
	c.isReady = true
	c.m.Unlock()

	c.logger.Debug("RabbitMQ channel initialized")
	close(c.waitForChan)
	c.waitForChan = make(chan struct{})
	return nil
}

func (c *Client) changeChannel(ch *amqp.Channel) {
	c.ch = ch
	c.notifyChanClose = make(chan *amqp.Error, 1)
	c.ch.NotifyClose(c.notifyChanClose)
}

func (c *Client) waitForConnection() error {
	c.m.RLock()
	if !c.isReady {
		c.m.RUnlock()

		c.logger.Debug("RabbitMQ connection/channel not ready, waiting to be ready...")
		select {
		case <-time.After(c.config.WaitForConnectionTimeout):
			c.logger.Error(
				"RabbitMQ timed out while waiting for connection",
				slog.Duration("timeout_duration", c.config.WaitForConnectionTimeout),
			)
			return ErrNotConnected
		case <-c.waitForChan:
			c.logger.Debug(
				"RabbitMQ connection ready",
				"from",
				"wait for channel",
			)
		}
	} else {
		c.m.RUnlock()
		c.logger.Debug("RabbitMQ connection ready")
	}

	return nil
}

// Declares an exchange and queue, then binds them.
func (c *Client) NewQueue(config RabbitMQQueueConfig) error {
	err := c.waitForConnection()
	if err != nil {
		return err
	}

	c.logger.Debug("RabbitMQ connection is ready for queue creation")

	if err := c.DeclareExchange(
		config.ExchangeType,
		config.Exchange,
	); err != nil {
		return err
	}

	if config.ConsumerTimeout == 0 {
		config.ConsumerTimeout = 1 * time.Minute
	}

	if _, err := c.ch.QueueDeclare(
		config.Queue,
		true,
		true,
		false,
		false,
		amqp.Table{
			amqp.ConsumerTimeoutArg: config.ConsumerTimeout.Milliseconds(),
		},
	); err != nil {
		return err
	}

	c.logger.Debug(
		"RabbitMQ queue declared successfully",
		"queue", config.Queue,
	)

	for _, routingKey := range config.RoutingKeys {
		if err := c.ch.QueueBind(
			config.Queue,
			routingKey,
			config.Exchange,
			false,
			nil,
		); err != nil {
			return err
		}

		c.logger.Debug(
			"RabbitMQ queue bound to exchange successfully",
			"queue", config.Queue,
			"exchange", config.Exchange,
			"routing-key", routingKey,
		)
	}

	return nil
}

func (c *Client) safeConsume(
	ctx context.Context,
	config RabbitMQQueueConfig,
	consumerName string,
) (<-chan amqp.Delivery, error) {
	c.m.Lock()
	if !c.isReady {
		c.m.Unlock()
		return nil, ErrNotConnected
	}
	c.m.Unlock()

	return c.ch.ConsumeWithContext(
		ctx,
		config.Queue,
		consumerName,
		false,
		false,
		false,
		false,
		nil,
	)
}

// Blocks until ctx is done, or RabbitMQ channel/connection breaks
func (c *Client) StartConsuming(
	ctx context.Context,
	config RabbitMQQueueConfig,
	consumerName string,
	workerFun func(context.Context, []byte) error,
) error {
	msgs, err := c.safeConsume(ctx, config, consumerName)
	if err != nil {
		c.logger.Error("RabbitMQ failed to start consuming", "error", err)
		return err
	}

	// must be different from c.notifyChanClose because the rabbitMQ
	// library sends only one notification and c.notifyChanClose
	// already has a receiver in c.handleReconnect()
	chClosedNotifier := make(chan *amqp.Error, 1)
	c.ch.NotifyClose(chClosedNotifier)

	backoffRate := 2
	backoffCeiling := 5 * time.Minute

	go func(ctx context.Context) {
		c.logger.Info("RabbitMQ worker running and awaiting for messages...")

		backoff := 1 * time.Second

		for {
			select {
			case msg := <-msgs:
				ctx := xotel.ExtractAMQPHeaders(ctx, msg.Headers)
				ctx, span := c.tracer.Start(
					ctx,
					"rabbitmq.consumer",
					trace.WithSpanKind(trace.SpanKindConsumer),
				)

				payload := msg.Body
				c.logger.Debug(
					"received message",
					slog.Int("size-in-bytes", len(payload)),
				)

				err := workerFun(ctx, payload)
				if err != nil {
					c.logger.Error("error processing message", "error", err)
					span.RecordError(err)
					span.SetStatus(codes.Error, "error processing message")

					err = msg.Nack(false, false)
					if err != nil {
						c.logger.Error("error not acknowledging message", "error", err)
						span.RecordError(err)
						span.SetStatus(codes.Error, "error not acknowledging message")
					}

					span.End()
					continue
				}

				if span.IsRecording() {
					span.AddEvent("message processed")
				}

				c.logger.Debug("done processing message")

				err = msg.Ack(false)
				if err != nil {
					c.logger.Error("error acknowledging message", "error", err)
					span.RecordError(err)
					span.SetStatus(codes.Error, "error acknowledging message")
				}

				c.logger.Debug("done acknowledging message")
				span.End()

			case amqErr := <-chClosedNotifier:
				c.logger.Debug("awaiting...", "backoff", backoff)
				time.Sleep(backoff)

				backoff *= time.Duration(backoffRate)
				if backoff > backoffCeiling {
					backoff = backoffCeiling
				}

				c.logger.Error(
					"RabbitMQ channel closed due error",
					"error", amqErr,
				)

				msgs, err = c.safeConsume(ctx, config, consumerName)
				if err != nil {
					c.logger.Error(
						"RabbitMQ failed to start consuming, retrying...",
						"error", err,
					)
					continue
				}

				chClosedNotifier = make(chan *amqp.Error, 1)
				c.ch.NotifyClose(chClosedNotifier)

			case <-ctx.Done():
				c.logger.Info(
					"RabbitMQ worker stopped",
					"consumer", consumerName,
				)
				return
			}
		}
	}(ctx)

	<-ctx.Done()
	c.logger.Info("RabbitMQ got exit signal")

	err = c.Close()
	if err != nil {
		c.logger.Error("RabbitMQ failed to close consumer", "error", err)
	}

	c.logger.Info("RabbitMQ done processing messages from queue")
	return nil
}

func (c *Client) DeclareExchange(
	exchangeType ExchangeType,
	exchange string,
) error {
	if err := c.ch.ExchangeDeclare(
		exchange,
		string(exchangeType),
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		c.logger.Error(
			"failed to declare RabbitMQ exchange",
			"exchange", exchange,
			"type", string(exchangeType),
			"error", err,
		)
		return err
	}

	c.logger.Debug(
		"RabbitMQ exchange declared successfully",
		"exchange", exchange,
		"type", string(exchangeType),
	)

	return nil
}

func (c *Client) Close() error {
	c.logger.Debug("RabbitMQ close triggered")

	c.m.Lock()
	defer c.m.Unlock()

	if !c.isReady {
		return ErrClientConnectionAlreadyClose
	}

	close(c.done)

	if err := c.ch.Close(); err != nil {
		c.logger.Error("failed to close RabbitMQ channel")
		return err
	}

	c.logger.Info("closed channel")

	if err := c.conn.Close(); err != nil {
		c.logger.Error("failed to close RabbitMQ connection")
		return err
	}

	c.logger.Info("closed connection")

	c.isReady = false
	return nil
}

// Waits for a connection error on notifyConnClose, and
// then continuously attempt to reconnect.
func (c *Client) handleReconnect() {
	for {
		c.m.Lock()
		c.isReady = false
		c.m.Unlock()

		c.logger.Debug("RabbitMQ reconnection attempt started")

		err := c.connect()
		if err != nil {
			c.logger.Error("RabbitMQ reconnection attempt failed")

			select {
			case <-c.done:
				return
			case <-time.After(c.config.ReconnectDelay):
			}

			continue
		}

		if done := c.handleReinitialization(); done {
			c.logger.Debug("Stopped reconnect process")
			break
		}
	}
}

func NewRabbitMQClient(
	config ClientConfig,
	logger xlog.Logger,
	tracer trace.Tracer,
) (*Client, error) {
	logger = logger.With("external_dependency", "rabbitmq")

	c := &Client{
		m:           &sync.RWMutex{},
		done:        make(chan struct{}),
		config:      config,
		logger:      logger,
		tracer:      tracer,
		isReady:     true,
		waitForChan: make(chan struct{}),
	}

	go c.handleReconnect()

	return c, nil
}
