package xotel

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sptGabriel/gb-libs/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/metadata"

	traceSdk "go.opentelemetry.io/otel/sdk/trace"
)

func SetupTracing(
	ctx context.Context,
	r *resource.Resource,
	cfg config.AppConfig,
) (trace.Tracer, func(context.Context) error, error) {
	var appTracer trace.Tracer
	var shutdown func(context.Context) error

	if !cfg.Enabled {
		provider := noop.NewTracerProvider()
		appTracer = provider.Tracer(cfg.Name)
		otel.SetTracerProvider(provider)
		shutdown = func(ctx context.Context) error { return nil }

		return appTracer, shutdown, nil
	}

	if cfg.TraceEndpoint == "" {
		return nil, nil, fmt.Errorf("opentelemetry traces endpoint not specified")
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.TraceEndpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithURLPath("/v1/traces"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize opentelemetry traces exporter: %w", err)
	}

	provider := traceSdk.NewTracerProvider(
		traceSdk.WithResource(r),
		traceSdk.WithBatcher(exporter),
	)
	shutdown = provider.Shutdown
	otel.SetTracerProvider(provider)
	appTracer = provider.Tracer(cfg.Name)
	return appTracer, shutdown, nil
}

func InjectOtelContext(ctx context.Context, carrier map[string]string) {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	p.Inject(ctx, propagation.MapCarrier(carrier))
}

func ExtractOtelContext(ctx context.Context, carrier map[string]string) context.Context {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return p.Extract(ctx, propagation.MapCarrier(carrier))
}

func ExtractOtelContextFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return p.Extract(ctx, metadataTextMap{md})
}

// metadataTextMap implements the TextMapCarrier interface for gRPC metadata.
type metadataTextMap struct {
	md metadata.MD
}

func (mtm metadataTextMap) Get(key string) string {
	values := mtm.md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// Set is not used for extraction, so it's a no-op.
func (mtm metadataTextMap) Set(key, value string) {
	// no-op for extraction
}

func (mtm metadataTextMap) Keys() []string {
	keys := make([]string, 0, len(mtm.md))
	for k := range mtm.md {
		keys = append(keys, k)
	}
	return keys
}

// AmqpHeadersCarrier implements the TextMapCarrier interface for AMQP metadata.
type AmqpHeadersCarrier map[string]string

func (a AmqpHeadersCarrier) Get(key string) string {
	v, ok := a[key]
	if !ok {
		return ""
	}

	return v
}

func (a AmqpHeadersCarrier) Set(key, value string) {
	a[key] = value
}

func (a AmqpHeadersCarrier) Keys() []string {
	i := 0
	r := make([]string, len(a))

	for k := range a {
		r[i] = k
		i++
	}

	return r
}

func (a AmqpHeadersCarrier) Serialize() amqp091.Table {
	t := amqp091.Table{}
	for k, v := range a {
		t[k] = v
	}

	return t
}

func Deserialize(t amqp091.Table) AmqpHeadersCarrier {
	c := AmqpHeadersCarrier{}
	for k, v := range t {
		s, ok := v.(string)
		if !ok {
			continue
		}

		c.Set(k, s)
	}

	return c
}

// Injects the tracing from the context into the AMQP headers carrier
func InjectAMQPHeaders(ctx context.Context, h AmqpHeadersCarrier) {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	p.Inject(ctx, h)
}

// Extracts the tracing from the AMQP headers and puts it into the context
func ExtractAMQPHeaders(ctx context.Context, headers amqp091.Table) context.Context {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return p.Extract(ctx, Deserialize(headers))
}
