package otelinit

import (
	"context"
	"fmt"
	"io"
	"log"
)

// Initialize a otel provider with a collector
func Example_initProviderCollector() {
	ctx := context.Background()

	shutdown, err := InitProvider(
		ctx,
		"simple-gohttp",
		WithGRPCTraceExporter(
			ctx,
			fmt.Sprintf("%s:%d", "127.0.0.1", 4317),
		),
	)
	if err != nil {
		log.Println("failed to initialize opentelemetry")

		return
	}

	defer func() {
		if err := shutdown(); err != nil {
			log.Println("failed to shutdown")
		}
	}()
}

// Initialize a otel provider and discard traces
// useful for dev
func Example_initProviderDiscardTraces() {
	ctx := context.Background()

	shutdown, err := InitProvider(
		ctx,
		"simple-gohttp",
		WithWriterTraceExporter(io.Discard),
	)
	if err != nil {
		log.Println("failed to initialize opentelemetry")

		return
	}

	defer func() {
		if err := shutdown(); err != nil {
			log.Println("failed to shutdown")
		}
	}()
}
