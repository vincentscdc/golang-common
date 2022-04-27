package otelinit

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

// WithGRPCTraceExporter allows you to send your traces to the collector target
// collectorTarget is the address of the collector, e.g. "127.0.0.1:4317"
func WithGRPCTraceExporter(ctx context.Context, collectorTarget string) func(*provider) error {
	return func(pvd *provider) error {
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(collectorTarget),
			otlptracegrpc.WithInsecure(),
		)

		traceExporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to create grpc traceExporter: %w", err)
		}

		pvd.traceExporter = traceExporter

		return nil
	}
}
