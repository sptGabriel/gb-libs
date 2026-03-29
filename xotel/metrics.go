package xotel

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sptGabriel/gb-libs/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func SetupMetrics(
	ctx context.Context,
	r *resource.Resource,
	cfg config.AppConfig,
) (func(context.Context) error, error) {
	var shutdown func(context.Context) error

	if !cfg.Enabled {
		otel.SetMeterProvider(noop.NewMeterProvider())
		shutdown = func(ctx context.Context) error { return nil }

		return shutdown, nil
	}

	if cfg.MetricsEndPoint == "" {
		return nil, fmt.Errorf("opentelemetry metrics endpoint not specified")
	}

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.MetricsEndPoint),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithURLPath("/v1/metrics"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics exporter: %w", err)
	}

	provider := metricSdk.NewMeterProvider(
		metricSdk.WithResource(r),
		metricSdk.WithReader(
			metricSdk.NewPeriodicReader(exporter, metricSdk.WithInterval(time.Second)),
		),
	)

	shutdown = provider.Shutdown
	otel.SetMeterProvider(provider)

	err = setupMemoryHeapMetric(provider, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup memory heap metric: %w", err)
	}

	return shutdown, nil
}

func setupMemoryHeapMetric(provider *metricSdk.MeterProvider, cfg config.AppConfig) error {
	meter := provider.Meter(cfg.Name)

	_, err := meter.Float64ObservableGauge(
		"system.memory.heap",
		metric.WithDescription("Memory usage of the allocated heap objects."),
		metric.WithUnit("MiBy"),
		metric.WithFloat64Callback(func(ctx context.Context, o metric.Float64Observer) error {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			o.Observe(float64(memStats.HeapAlloc) / 1048576)
			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}
