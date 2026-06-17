package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracing configures OTLP HTTP trace export when OTEL_ENABLED=true.
func InitTracing(cfg *config.Config) (func(), error) {
	if cfg == nil || !cfg.Observability.OTELEnabled {
		return nil, nil
	}

	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Observability.OTELExporterEndpoint),
		otlptracehttp.WithInsecure(), // local collectors typically use HTTP without TLS
	)
	if err != nil {
		return nil, fmt.Errorf("otel exporter: %w", err)
	}

	serviceName := cfg.Observability.ServiceName
	if serviceName == "" {
		serviceName = "luxe-api"
	}

	// Do not merge resource.Default() — it uses a different semconv schema URL than v1.24.0.
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(cfg.Observability.ServiceVersion),
		semconv.DeploymentEnvironment(cfg.Observability.SentryEnvironment),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}, nil
}
