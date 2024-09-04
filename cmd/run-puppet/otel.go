package main

import (
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func newTraceExporter() (trace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func newTraceProvider(traceExporter trace.SpanExporter) *trace.TracerProvider {
	// traceRes, err := resource.New(ctx,
	// 	resource.WithAttributes(semconv.ServiceNamespaceKey.String("api2")),
	// )
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second)),
		// trace.WithResource(),
	)
	return traceProvider
}
