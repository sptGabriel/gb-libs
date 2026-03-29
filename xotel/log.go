package xotel

import (
	"context"
	"fmt"

	"github.com/sptGabriel/gb-libs/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func SetupLogging(
	ctx context.Context,
	r *resource.Resource,
	cfg config.AppConfig,
) (*log.LoggerProvider, *otelslog.Handler, func(context.Context) error, error) {
	var provider *log.LoggerProvider
	var shutdown func(context.Context) error

	if !cfg.Enabled {
		provider = log.NewLoggerProvider()
		shutdown = func(ctx context.Context) error { return nil }

		return provider, buildHandler(cfg.Name, provider), shutdown, nil
	}

	if cfg.LogEndpoint == "" {
		return nil, nil, nil, fmt.Errorf("opentelemetry log endpoint not specified")
	}

	path := cfg.LogPath
	if path == "" {
		path = "/v1/logs"
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.LogEndpoint),
		otlploghttp.WithURLPath(path),
	}
	if !cfg.OTLPSecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize otlp log exporter: %w", err)
	}

	provider = log.NewLoggerProvider(
		log.WithResource(r),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)

	shutdown = provider.Shutdown

	return provider, buildHandler(cfg.Name, provider), shutdown, nil
}

func buildHandler(name string, provider *log.LoggerProvider) *otelslog.Handler {
	return otelslog.NewHandler(name,
		otelslog.WithLoggerProvider(provider),
		otelslog.WithSource(true),
	)
}
