package tracing

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
	tp     *tracesdk.TracerProvider
)

// InitTracing initializes Jaeger tracing for a service
func InitTracing(serviceName string) (func(), error) {
	jaegerURL := os.Getenv("JAEGER_ENDPOINT")
	if jaegerURL == "" {
		jaegerURL = "http://jaeger:14268/api/traces"
	}

	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource with service name
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
		resource.WithFromEnv(),
		resource.WithProcess(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp = tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get tracer for this service
	tracer = otel.Tracer(serviceName)

	log.Printf("Tracing initialized for service: %s", serviceName)

	// Return cleanup function
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}, nil
}

// GetTracer returns the tracer for this service
func GetTracer() trace.Tracer {
	if tracer == nil {
		// Fallback to no-op tracer if not initialized
		return trace.NewNoopTracerProvider().Tracer("noop")
	}
	return tracer
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name)
}

// StartSpanWithParent starts a new span as a child of the parent span
func StartSpanWithParent(ctx context.Context, name string) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name)
}

// GetPropagator returns the global text map propagator
func GetPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}
