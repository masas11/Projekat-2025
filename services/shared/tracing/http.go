package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPClient wraps an http.Client with tracing
func HTTPClient(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	return &http.Client{
		Transport: otelhttp.NewTransport(
			client.Transport,
			otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)),
		),
		Timeout: client.Timeout,
	}
}

// HTTPHandler wraps an http.Handler with tracing
func HTTPHandler(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

// HTTPMiddleware creates tracing middleware for HTTP handlers
func HTTPMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http.request")
}
