package xotel

import (
	"fmt"

	"github.com/sptGabriel/defaults/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func SetupResource(cfg config.AppConfig) (*resource.Resource, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.Name),
			semconv.ServiceVersion(cfg.Version),
			attribute.String("service.environment", cfg.Environment),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize opentelemetry resource. err: %v", err)
	}

	return r, nil
}
